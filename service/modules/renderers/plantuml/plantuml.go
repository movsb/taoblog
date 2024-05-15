package plantuml

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
)

func fetch(ctx context.Context, server, format, compressed string) ([]byte, error) {
	// TODO 看看能不能不 embed metadata。
	u, err := url.Parse(server)
	if err != nil {
		return nil, err
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
