package assets

import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/movsb/taoblog/modules/exiftool"
)

func metadataJPEG(t *testing.T) string {
	t.Helper()
	var b bytes.Buffer
	if err := jpeg.Encode(&b, image.NewRGBA(image.Rect(0, 0, 2, 2)), nil); err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "照片.jpg")
	if err := os.WriteFile(name, b.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := exiftool.Command(nil, "-overwrite_original", "-Artist=测试作者",
		"-GPSLatitude=31.2304", "-GPSLatitudeRef=N", "-GPSLongitude=121.4737", "-GPSLongitudeRef=E",
		"-XMP-exif:GPSLatitude=31.2304", "-XMP-exif:GPSLongitude=121.4737", name)
	if err != nil {
		t.Fatal(err)
	}
	return name
}

func readTags(t *testing.T, name string) map[string]any {
	t.Helper()
	b, err := exiftool.Command(nil, "-G", "-s", "-json", "-Artist", "-gps*", name)
	if err != nil {
		t.Fatal(err)
	}
	var records []map[string]any
	if err := json.Unmarshal(b, &records); err != nil || len(records) != 1 {
		t.Fatalf("invalid metadata JSON: %s, error: %v", b, err)
	}
	return records[0]
}

func TestCopyTagsToAVIFWithoutExternalExifTool(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	src := metadataJPEG(t)
	b, err := os.ReadFile("../../../service/modules/renderers/exif/testdata/exif.avif")
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(t.TempDir(), "输出.avif")
	if err := os.WriteFile(dst, b, 0o600); err != nil {
		t.Fatal(err)
	}
	CopyTags(src, dst)
	source, copied := readTags(t, src), readTags(t, dst)
	for _, tag := range []string{"EXIF:Artist", "EXIF:GPSLatitude", "EXIF:GPSLongitude"} {
		if source[tag] == nil || !reflect.DeepEqual(copied[tag], source[tag]) {
			t.Errorf("%s: got %v, want %v", tag, copied[tag], source[tag])
		}
	}
	if _, err := os.Stat(dst + "_original"); !os.IsNotExist(err) {
		t.Fatalf("unexpected backup file: %v", err)
	}
}

func TestDropGPSTagsPreservesOtherMetadataAndPixels(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	name := metadataJPEG(t)
	before := readTags(t, name)
	if before["EXIF:GPSLatitude"] == nil || before["XMP:GPSLatitude"] == nil {
		t.Fatalf("fixture must contain both EXIF and XMP GPS tags: %v", before)
	}
	pixels := func() image.Image {
		t.Helper()
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		img, err := jpeg.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		return img
	}
	originalPixels := pixels()
	if output, err := DropGPSTags(name); err != nil {
		t.Fatalf("DropGPSTags: %v: %s", err, output)
	}
	after := readTags(t, name)
	for tag := range after {
		if strings.Contains(tag, ":GPS") {
			t.Errorf("GPS tag remains: %s", tag)
		}
	}
	if after["EXIF:Artist"] != "测试作者" {
		t.Errorf("artist lost: %v", after)
	}
	if !reflect.DeepEqual(originalPixels, pixels()) {
		t.Error("image pixels changed")
	}
}
