package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/movsb/taoblog/modules/exiftool"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
)

func TestFileUploadDownloadGPS(t *testing.T) {
	r := Serve(t.Context())
	p := utils.Must1(r.client.Blog.CreatePost(r.user1, &proto.Post{
		SourceType: "markdown", Source: "# GPS upload test",
	}))

	readTags := func(t *testing.T, name string) map[string]any {
		t.Helper()
		output := utils.Must1(exiftool.Command(nil, "-G", "-s", "-json", "-Artist", "-gps*", name))
		var records []map[string]any
		if err := json.Unmarshal(output, &records); err != nil || len(records) != 1 {
			t.Fatalf("invalid metadata: %s, error: %v", output, err)
		}
		return records[0]
	}

	for _, ext := range []string{".jpg", ".avif"} {
		t.Run(ext, func(t *testing.T) {
			if ext == ".jpg" {
				if _, err := exec.LookPath("avifenc"); err != nil {
					t.Skip("JPEG conversion requires avifenc")
				}
			}
			input := filepath.Join(t.TempDir(), "照片"+ext)
			if ext == ".jpg" {
				var b bytes.Buffer
				utils.Must(jpeg.Encode(&b, image.NewRGBA(image.Rect(0, 0, 2, 2)), nil))
				utils.Must(os.WriteFile(input, b.Bytes(), 0o600))
			} else {
				data := utils.Must1(os.ReadFile("../service/modules/renderers/exif/testdata/exif.avif"))
				utils.Must(os.WriteFile(input, data, 0o600))
			}
			utils.Must1(exiftool.Command(nil, "-overwrite_original", "-Artist=测试作者",
				"-GPSLatitude=31.2304", "-GPSLatitudeRef=N", "-GPSLongitude=121.4737", "-GPSLongitudeRef=E",
				"-XMP-exif:GPSLatitude=31.2304", "-XMP-exif:GPSLongitude=121.4737", input))
			before := readTags(t, input)
			for _, tag := range []string{"EXIF:GPSLatitude", "EXIF:GPSLongitude", "XMP:GPSLatitude", "XMP:GPSLongitude"} {
				if before[tag] == nil {
					t.Fatalf("fixture missing %s: %v", tag, before)
				}
			}
			data := utils.Must1(os.ReadFile(input))

			for _, drop := range []bool{false, true} {
				t.Run(fmt.Sprintf("drop_gps_tags=%t", drop), func(t *testing.T) {
					name := fmt.Sprintf("照片-%s-%t%s", strings.TrimPrefix(ext, "."), drop, ext)
					var body bytes.Buffer
					parts := multipart.NewWriter(&body)
					utils.Must(parts.WriteField("spec", string(utils.Must1(jsonPB.Marshal(&proto.FileSpec{
						Path: name, Size: uint32(len(data)),
					})))))
					utils.Must1(utils.Must1(parts.CreateFormFile("data", name)).Write(data))
					// Match files.mjs: JSON options follows the file part.
					utils.Must(parts.WriteField("options", fmt.Sprintf(`{"drop_gps_tags":%t}`, drop)))
					utils.Must(parts.Close())

					endpoint := r.server.JoinPath("/v3/posts", fmt.Sprint(p.Id), "files")
					req := utils.Must1(http.NewRequestWithContext(t.Context(), http.MethodPost, endpoint, &body))
					req.Header.Set("Content-Type", parts.FormDataContentType())
					r.addAuth(req, r.user1ID)
					rsp := utils.Must1(http.DefaultClient.Do(req))
					uploadBody := utils.Must1(io.ReadAll(rsp.Body))
					utils.Must(rsp.Body.Close())
					if rsp.StatusCode != http.StatusOK {
						t.Fatalf("upload: %s: %s", rsp.Status, uploadBody)
					}
					var uploaded struct {
						Spec *proto.FileSpec `json:"spec"`
					}
					utils.Must(json.Unmarshal(uploadBody, &uploaded))
					wantPath := strings.TrimSuffix(name, ext) + ".avif"
					if uploaded.Spec == nil || uploaded.Spec.Path != wantPath || uploaded.Spec.Type != "image/avif" {
						t.Fatalf("unexpected upload response: %s", uploadBody)
					}

					// Download through HTTP from the real file service and database.
					req = utils.Must1(http.NewRequestWithContext(t.Context(), http.MethodGet,
						r.server.JoinPath("/v3/posts", fmt.Sprint(p.Id), "files", uploaded.Spec.Path), nil))
					r.addAuth(req, r.user1ID)
					rsp = utils.Must1(http.DefaultClient.Do(req))
					downloaded := utils.Must1(io.ReadAll(rsp.Body))
					utils.Must(rsp.Body.Close())
					if rsp.StatusCode != http.StatusOK {
						t.Fatalf("download: %s: %s", rsp.Status, downloaded)
					}
					if len(downloaded) == 0 || uint32(len(downloaded)) != uploaded.Spec.Size {
						t.Fatalf("downloaded size = %d, uploaded size = %d", len(downloaded), uploaded.Spec.Size)
					}
					output := filepath.Join(t.TempDir(), uploaded.Spec.Path)
					utils.Must(os.WriteFile(output, downloaded, 0o600))
					after := readTags(t, output)
					if drop {
						for tag := range after {
							if strings.Contains(tag, ":GPS") {
								t.Errorf("GPS tag remains: %s=%v", tag, after[tag])
							}
						}
					} else {
						for tag, value := range before {
							if strings.Contains(tag, ":GPS") && !reflect.DeepEqual(after[tag], value) {
								t.Errorf("%s: got %v, want %v", tag, after[tag], value)
							}
						}
					}
					if after["EXIF:Artist"] != "测试作者" {
						t.Errorf("artist metadata changed: %v", after)
					}
				})
			}
		})
	}
}
