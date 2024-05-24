package blog

import (
	"embed"

	"github.com/movsb/taoblog/modules/utils/dir"
)

// NOTE：/* 才能加载 . _ 开头的文件，见 embed 的注释。
//
//go:embed statics templates/*
var Root embed.FS

func SourceRelativeDir() dir.Dir {
	return dir.SourceRelativeDir()
}
