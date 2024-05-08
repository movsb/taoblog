package blog

import (
	"embed"
)

// NOTE：/* 才能加载 . _ 开头的文件，见 embed 的注释。
//
//go:embed statics templates/*
var Root embed.FS
