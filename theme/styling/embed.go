package styling

import (
	"embed"
	"io/fs"

	"github.com/movsb/taoblog/modules/utils"
)

//go:embed index.md
var Index []byte

//go:embed root/*
var _Root embed.FS

var Root = utils.Must1(fs.Sub(_Root, `root`))
