package styling

import (
	"embed"
	"io/fs"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
)

//go:embed index.md
var Index []byte

//go:embed root/*
var _embed embed.FS

var Root = utils.Must1(fs.Sub(_embed, `root`))

var Dir = dir.SourceAbsoluteDir()
