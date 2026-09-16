package twitter

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"io/fs"
	"mime"
	"net/url"
	"path"
	"regexp"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/gen2brain/avif"
	"github.com/movsb/taoblog/protocols/go/proto"
)

type postFiles interface {
	ForPost(int) fs.FS
}

type Publisher struct {
	config func() Config
	files  postFiles
}

func NewPublisher(config func() Config, files postFiles) *Publisher {
	return &Publisher{config: config, files: files}
}

type mediaCandidate struct {
	name, contentType string
}

func (p *Publisher) Preview(post *proto.Post) (*proto.TwitterPostPreview, error) {
	if p.config == nil || !p.config().Valid() {
		return nil, fmt.Errorf("Twitter 同步尚未启用或凭据不完整")
	}
	media, err := p.mediaCandidates(int(post.Id))
	if err != nil {
		return nil, err
	}
	preview := &proto.TwitterPostPreview{Text: truncatePostText(htmlText(post.Content), 280)}
	for _, item := range media {
		preview.MediaUrls = append(preview.MediaUrls,
			fmt.Sprintf("/v3/posts/%d/files/%s", post.Id, url.PathEscape(item.name)))
	}
	return preview, nil
}

func (p *Publisher) Publish(ctx context.Context, post *proto.Post) (string, error) {
	config := p.config()
	if !config.Valid() {
		return "", fmt.Errorf("Twitter 同步尚未启用或凭据不完整")
	}
	client := NewClient(config)
	media, err := p.mediaCandidates(int(post.Id))
	if err != nil {
		return "", err
	}
	mediaIDs, err := p.uploadMedia(ctx, client, int(post.Id), media)
	if err != nil {
		return "", err
	}
	return client.CreatePost(ctx, truncatePostText(htmlText(post.Content), 280), mediaIDs)
}

func (p *Publisher) mediaCandidates(postID int) ([]mediaCandidate, error) {
	postFS := p.files.ForPost(postID)
	lister, ok := postFS.(interface {
		ListFiles() ([]*proto.FileSpec, error)
	})
	if !ok {
		return nil, fmt.Errorf("文章文件系统不支持 ListFiles")
	}
	files, err := lister.ListFiles()
	if err != nil {
		return nil, err
	}
	var images []mediaCandidate
	var motion *mediaCandidate
	for _, file := range files {
		if file.ParentPath != "" || strings.HasPrefix(file.Path, ".") || strings.HasPrefix(file.Path, "_") {
			continue
		}
		contentType := file.Type
		if contentType == "" {
			contentType = mime.TypeByExtension(path.Ext(file.Path))
		}
		candidate := mediaCandidate{name: file.Path, contentType: contentType}
		switch {
		case contentType == "image/gif", strings.HasPrefix(contentType, "video/"):
			if motion == nil {
				motion = &candidate
			}
		case contentType == "image/jpeg" || contentType == "image/png" || contentType == "image/webp" || contentType == `image/avif`:
			images = append(images, candidate)
		}
	}
	if motion != nil {
		return []mediaCandidate{*motion}, nil
	}
	if len(images) > 4 {
		images = images[:4]
	}
	return images, nil
}

func (p *Publisher) uploadMedia(ctx context.Context, client *Client, postID int, media []mediaCandidate) ([]string, error) {
	postFS := p.files.ForPost(postID)
	ids := make([]string, 0, len(media))
	for _, item := range media {
		data, err := readFile(postFS, item.name)
		if err != nil {
			return nil, err
		}
		if item.contentType == "image/avif" {
			data, err = avifToJPEG(data)
			if err != nil {
				return nil, fmt.Errorf("转换 AVIF 图片 %q：%w", item.name, err)
			}
			item.contentType = "image/jpeg"
		}
		var id string
		if item.contentType == "image/gif" || strings.HasPrefix(item.contentType, "video/") {
			category := "tweet_video"
			if item.contentType == "image/gif" {
				category = "tweet_gif"
			}
			id, err = client.UploadChunked(ctx, item.contentType, category, data)
		} else {
			id, err = client.UploadImage(ctx, item.contentType, data)
		}
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func avifToJPEG(data []byte) ([]byte, error) {
	img, err := avif.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	// JPEG 不支持透明通道。使用白色背景，避免透明区域变黑。
	bounds := img.Bounds()
	flattened := image.NewRGBA(bounds)
	draw.Draw(flattened, bounds, &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(flattened, bounds, img, bounds.Min, draw.Over)

	var output bytes.Buffer
	if err := jpeg.Encode(&output, flattened, &jpeg.Options{Quality: 90}); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func readFile(fileSystem fs.FS, name string) ([]byte, error) {
	file, err := fileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(file)
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	return data, closeErr
}

var whitespace = regexp.MustCompile(`\s+`)

func htmlText(raw string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(raw))
	if err != nil {
		return strings.TrimSpace(raw)
	}
	doc.Find("script,style").Remove()
	doc.Find("br,p,div,li,blockquote,pre,h1,h2,h3,h4,h5,h6").Each(func(_ int, selection *goquery.Selection) {
		selection.AppendHtml(" ")
	})
	return strings.TrimSpace(whitespace.ReplaceAllString(doc.Text(), " "))
}

func runeWeight(r rune) int {
	if r <= unicode.MaxASCII {
		return 1
	}
	return 2
}

func truncatePostText(text string, limit int) string {
	text = strings.TrimSpace(text)
	weight := 0
	for _, r := range text {
		weight += runeWeight(r)
	}
	if weight <= limit {
		return text
	}
	const ellipsis = "…"
	var out strings.Builder
	weight = 0
	for _, r := range text {
		w := runeWeight(r)
		if weight+w+2 > limit {
			break
		}
		out.WriteRune(r)
		weight += w
	}
	return strings.TrimSpace(out.String()) + ellipsis
}
