package media_size

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	_ "image/gif" // shut up
	_ "image/jpeg"
	_ "image/png"
	"io"
	"strings"
)

type Metadata struct {
	Width, Height int
}

// 常见文件格式（GIF/JPG/JPEG/PNG）支持，用标准库解析。
func normal(r io.Reader) (*Metadata, error) {
	config, _, err := image.DecodeConfig(r)
	if err != nil {
		return nil, err
	}
	return &Metadata{
		Width:  config.Width,
		Height: config.Height,
	}, nil
}

// https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute/viewBox
// https://www.geeksforgeeks.org/svg-viewbox-attribute/#
// https://www.digitalocean.com/community/tutorials/svg-svg-viewbox
func svg(r io.Reader) (*Metadata, error) {
	var decoded struct {
		// 有些有，没有的话……
		Width  string `xml:"width,attr"`
		Height string `xml:"height,attr"`
		// ……没有的话就用这个。
		ViewBox string `xml:"viewBox,attr"`
	}

	if err := xml.NewDecoder(r).Decode(&decoded); err != nil {
		return nil, err
	}

	var w, h int
	fmt.Sscanf(decoded.Width, `%d`, &w)
	fmt.Sscanf(decoded.Height, `%d`, &h)

	if w > 0 && h > 0 {
		return &Metadata{
			Width:  w,
			Height: h,
		}, nil
	}

	// 空格或者逗号分隔的
	var x, y, width, height float32
	fmt.Sscanf(
		// 不是很精确，但是该够用了。
		strings.ReplaceAll(decoded.ViewBox, ",", ""),
		"%f %f %f %f",
		&x, &y, &width, &height,
	)

	_, _ = x, y

	if width > 0 && height > 0 {
		return &Metadata{
			Width:  int(width),
			Height: int(height),
		}, nil
	}

	return nil, errors.New(`bad svg to get size`)
}

// 暴力出奇迹。
// https://blog.twofei.com/1068/
func avif(r io.Reader) (*Metadata, error) {
	sniff, err := io.ReadAll(io.LimitReader(r, 1<<10))
	if err != nil {
		return nil, err
	}
	index := bytes.Index(sniff, []byte{'i', 's', 'p', 'e'})
	if index == -1 || index < 4 || index-4+20 > len(sniff) {
		return nil, errors.New(`not avif`)
	}

	sniff = sniff[index-4:]
	index = 0

	var (
		bo      = binary.BigEndian
		size    = bo.Uint32(sniff[index+0:])
		version = bo.Uint32(sniff[index+8:])
		width   = bo.Uint32(sniff[index+12:])
		height  = bo.Uint32(sniff[index+16:])
	)

	if size == 0x14 && version == 0x00 &&
		(width > 0 && width < 65536) && // just in case
		(height > 0 && height < 65536) {
		return &Metadata{
			Width:  int(width),
			Height: int(height),
		}, nil
	}

	return nil, errors.New(`not avif`)
}
