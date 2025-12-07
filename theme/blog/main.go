package blog

import (
	"embed"

	"github.com/movsb/taoblog/modules/utils/dir"
)

//go:embed statics templates/*
var Root embed.FS

//go:generate sass --style compressed --no-source-map styles/style.scss statics/style.css

var SourceRelativeDir = dir.SourceRelativeDir()
