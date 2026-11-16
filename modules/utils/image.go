package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/gen2brain/avif"
	"golang.org/x/image/draw"
)

// DecodeImage 用原有的 AVIF 解码器处理 AVIF，其他格式走标准库。
func DecodeImage(r io.Reader) (image.Image, error) {
	br := bufio.NewReader(r)
	if isAVIF(br) {
		return avif.Decode(br)
	}
	img, _, err := image.Decode(br)
	return img, err
}

// DecodeImageConfig 与 DecodeImage 使用相同的格式分派。
func DecodeImageConfig(r io.Reader) (image.Config, error) {
	br := bufio.NewReader(r)
	if isAVIF(br) {
		return avif.DecodeConfig(br)
	}
	cfg, _, err := image.DecodeConfig(br)
	return cfg, err
}

func isAVIF(r *bufio.Reader) bool {
	b, err := r.Peek(12)
	if err != nil {
		return false
	}
	// 与原有 AVIF 解码器注册到 image 的签名一致。
	return string(b[4:]) == "ftypavif" || string(b[4:]) == "ftypavis"
}

// ResizeRGBA 将图片缩放到指定尺寸。
func ResizeRGBA(src image.Image, width, height int) *image.RGBA {
	if width <= 0 || height <= 0 || src.Bounds().Empty() {
		return image.NewRGBA(image.Rectangle{})
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)
	return dst
}

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

	img = ResizeRGBA(img, width, height)

	buf := bytes.NewBuffer(nil)
	if err := png.Encode(buf, img); err != nil {
		return nil, err
	}

	return &DataURL{
		Type: `image/png`,
		Data: buf.Bytes(),
	}, nil
}
