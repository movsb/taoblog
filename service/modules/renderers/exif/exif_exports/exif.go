package exif_exports

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

func Extract(r io.ReadCloser) (*Metadata, error) {
	cmd := exec.CommandContext(context.TODO(), `exiftool`, `-G`, `-s`, `-json`, `-`)
	cmd.Stdin = r
	output, err := cmd.Output()
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
