package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/anthonynsimon/bild/transform"
)

func ResizeImage(mimeType string, r io.Reader, width, height int) (*DataURL, error) {
	var img image.Image
	var err error

	switch mimeType {
	case `image/jpeg`:
		img, err = jpeg.Decode(r)
	case `image/png`:
		img, err = png.Decode(r)
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
