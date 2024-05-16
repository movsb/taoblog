package plantuml

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

func fetch(ctx context.Context, server, format, compressed string, darkMode bool) ([]byte, error) {
	u, err := url.Parse(server)
	if err != nil {
		return nil, err
	}
	if darkMode {
		format = `d` + format
	}
	u.Path = path.Join(u.Path, format, compressed)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	// 就算是错，也有错误的 body as svg
	// if rsp.StatusCode != 200 {
	// 	return nil, err
	// }
	return io.ReadAll(io.LimitReader(rsp.Body, 1<<20))
}

var brotli = base64.NewEncoding(`0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_`)

// https://plantuml.com/en/text-encoding
func compress(source []byte) (string, error) {
	buf := bytes.NewBuffer(nil)
	zw, err := flate.NewWriter(buf, flate.BestCompression)
	if err != nil {
		return "", errors.New(`error create deflating writer`)
	}
	_, err = zw.Write(source)
	if err != nil {
		return "", fmt.Errorf(`error writing compression data: %w`, err)
	}
	if err := zw.Close(); err != nil {
		return "", fmt.Errorf(`error closing compressor: %w`, err)
	}
	return brotli.EncodeToString(buf.Bytes()), nil
}

// 移除：<?xml...?>
// 移除 <!--SRC=...-->
// https://regex101.com/r/ouWYn8/1
var stripRe = regexp.MustCompile(`<\?(?U:.+)\?>|<!--SRC=(?U:.+)-->`)

func strip(svg []byte) []byte {
	return stripRe.ReplaceAllLiteral(svg, nil)
}

func style(svg []byte, darkMode bool) []byte {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(svg))
	if err != nil {
		log.Println(`解析 svg 出错：`, err)
		return svg
	}

	// 添加一些特定的类，用于识别
	if svg := doc.Find(`svg`); true {
		// 透明：为了使图片预览功能使用页面相同的背景色。
		// NOTE：也可以包在 div 中，然后用 :nth-child(1/2) 来选择。
		svg.AddClass(`plantuml`, `transparent`)

		if darkMode {
			svg.AddClass(`dark`)
			// 默认隐藏，防止在 RSS 不加载样式时显示两张图片。
			svg.SetAttr(`style`, svg.AttrOr(`style`, ``)+`display:none;`)
		}
	}

	buf := bytes.NewBuffer(nil)
	goquery.Render(buf, doc.Find(`body`).Children().First())
	return buf.Bytes()
}
