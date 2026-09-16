package twitter

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/gen2brain/avif"
	"github.com/movsb/taoblog/protocols/go/proto"
)

type testPostFiles map[int]fstest.MapFS

type testPostFS struct {
	fstest.MapFS
}

func TestAVIFToJPEG(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	source.Set(0, 0, color.NRGBA{R: 255, A: 255})
	var encoded bytes.Buffer
	if err := avif.Encode(&encoded, source); err != nil {
		t.Fatal(err)
	}
	converted, err := avifToJPEG(encoded.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := jpeg.Decode(bytes.NewReader(converted)); err != nil {
		t.Fatalf("converted data is not JPEG: %v", err)
	}
}

func (f testPostFS) ListFiles() ([]*proto.FileSpec, error) {
	files := make([]*proto.FileSpec, 0, len(f.MapFS))
	for name := range f.MapFS {
		files = append(files, &proto.FileSpec{Path: name})
	}
	return files, nil
}

func (f testPostFiles) ForPost(id int) fs.FS { return testPostFS{MapFS: f[id]} }

func TestPreview(t *testing.T) {
	publisher := &Publisher{config: func() Config {
		return Config{
			Enabled: true, ConsumerKey: "key", ConsumerSecret: "secret", AccessToken: "token", AccessTokenSecret: "token-secret",
		}
	}, files: testPostFiles{7: {
		"a.jpg":   &fstest.MapFile{Data: []byte("a")},
		"b.png":   &fstest.MapFile{Data: []byte("b")},
		"note.md": &fstest.MapFile{Data: []byte("ignored")},
	}}}
	preview, err := publisher.Preview(&proto.Post{Id: 7, Content: "<p>hello</p>"})
	if err != nil {
		t.Fatal(err)
	}
	if preview.Text != "hello" || len(preview.MediaUrls) != 2 {
		t.Fatalf("unexpected preview: %+v", preview)
	}
}

func TestHTMLText(t *testing.T) {
	got := htmlText(`<p>你好 <strong>world</strong></p><script>bad()</script><p>next</p>`)
	if got != "你好 world next" {
		t.Fatalf("got %q", got)
	}
}

func TestTruncatePostText(t *testing.T) {
	for _, tc := range []struct {
		text  string
		limit int
		want  string
	}{
		{"hello", 5, "hello"},
		{"abcdef", 5, "abc…"},
		{"你好世界", 6, "你好…"},
	} {
		if got := truncatePostText(tc.text, tc.limit); got != tc.want {
			t.Fatalf("truncatePostText(%q, %d) = %q, want %q", tc.text, tc.limit, got, tc.want)
		}
	}
}
