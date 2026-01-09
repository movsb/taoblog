package styling

import (
	"embed"
	"io/fs"
	"os"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/modules/version"
)

//go:embed root
var _embed embed.FS
var _local = os.DirFS(dir.SourceAbsoluteDir().Join())

func Root() fs.FS {
	fsys := utils.IIF(version.DevMode(), _local, fs.FS(_embed))
	return utils.Must1(fs.Sub(fsys, `root`))
}
