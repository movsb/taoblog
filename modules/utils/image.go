package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/anthonynsimon/bild/transform"
	"github.com/gen2brain/avif"
)

// 调整图片的大小。
//
// 支持 JPEG/PNG/AVIF。
//
// 所有元信息会丢失。
func ResizeImage(mimeType string, r io.Reader, width, height int) (*DataURL, error) {
	var img image.Image
	var err error

	switch mimeType {
	case `image/jpeg`:
		img, err = jpeg.Decode(r)
	case `image/png`:
		img, err = png.Decode(r)
	case `image/avif`:
		img, err = avif.Decode(r)
	default:
		return nil, fmt.Errorf(`不能处理图片格式：%s`, mimeType)
	}

	if err != nil {
		return nil, err
	}

	img = transform.Resize(img, width, height, transform.Lanczos)

	buf := bytes.NewBuffer(nil)
	if err := png.Encode(buf, img); err != nil {
		return nil, err
	}

	return &DataURL{
		Type: `image/png`,
		Data: buf.Bytes(),
	}, nil
}
