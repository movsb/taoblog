// Package exiftool runs the embedded ExifTool without an external process.
package exiftool

import (
	"context"
	"io"

	embedded "github.com/lbe/go-exiftool-wasm"
	"github.com/movsb/taoblog/modules/utils"
)

func Command(stdin io.Reader, args ...string) ([]byte, error) {
	return CommandContext(context.Background(), stdin, args...)
}

func CommandContext(ctx context.Context, stdin io.Reader, args ...string) (output []byte, err error) {
	// TODO: 按需复用 NewServer（空闲后关闭），避免每次初始化 Perl。
	// 需先支持自定义 WASI 内存文件系统：Server 的 stdin 用于接收命令，
	// 图片需通过虚拟路径访问；写入还需支持 Seek、重命名和删除。
	// Keep interpreter failures on the error path, as with the former subprocess.
	defer utils.CatchAsError(&err)
	return embedded.CommandContext(ctx, stdin, args...)
}
