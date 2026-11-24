package exif_exports

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/movsb/taoblog/modules/exiftool"
)

func Extract(r io.ReadCloser) (*Metadata, error) {
	// TODO: 支持 WASI 内存文件挂载后，以虚拟路径交给常驻实例读取，
	// 在保留流式输入接口、不增加磁盘临时文件的前提下复用 Perl 实例。
	output, err := exiftool.Command(r, `-G`, `-s`, `-json`, `-`)
	if err != nil {
		return nil, err
	}
	var md []*Metadata
	if err := json.Unmarshal(output, &md); err != nil {
		return nil, err
	}
	if len(md) <= 0 {
		return nil, fmt.Errorf(`没有提取到元数据`)
	}
	return md[0], nil
}
