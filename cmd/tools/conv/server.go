package conv

import (
	"context"
	"embed"
	_ "embed"
	"encoding/json"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/modules/version"
)

type Server struct {
	dir string
}

type File struct {
	Name string
	Type string
	Size string
}

func init() {
	mime.AddExtensionType(`.HEIC`, `image/heic`)
}

func NewServer(dir string) *Server {
	return &Server{
		dir: dir,
	}
}

var (
	//go:embed index.html script.js style.css
	_embed embed.FS
	_local = utils.NewOSDirFS(dir.SourceAbsoluteDir().Join())
)

func (s *Server) Run(ctx context.Context) {
	mux := http.NewServeMux()

	mux.HandleFunc(`GET /`, func(w http.ResponseWriter, r *http.Request) {
		fs := utils.IIF(version.DevMode(), _local, fs.FS(_embed))
		http.ServeFileFS(w, r, fs, path.Clean(r.URL.Path))
	})

	mux.HandleFunc(`GET /api/files`, func(w http.ResponseWriter, r *http.Request) {
		entries := utils.Must1(fs.ReadDir(os.DirFS(s.dir), `.`))
		files := []File{}
		for _, entry := range entries {
			if !entry.Type().IsRegular() {
				continue
			}
			info := utils.Must1(entry.Info())
			typ := mime.TypeByExtension(path.Ext(entry.Name()))
			if strings.HasPrefix(typ, `image/`) {
				if typ == `image/avif` {
					continue
				}

				if ext := filepath.Ext(entry.Name()); strings.ToLower(ext) == `.heic` {
					jpg1 := strings.TrimSuffix(entry.Name(), ext) + `.JPG`
					jpg2 := strings.TrimSuffix(entry.Name(), ext) + `.jpg`
					_, err1 := os.Stat(filepath.Join(s.dir, jpg1))
					_, err2 := os.Stat(filepath.Join(s.dir, jpg2))
					if err1 == nil || err2 == nil {
						continue
					}
					convertHEIC(context.Background(), filepath.Join(s.dir, entry.Name()))
					continue
				}

				files = append(files, File{
					Name: entry.Name(),
					Type: typ,
					Size: utils.ByteCountIEC(info.Size()),
				})
			} else if strings.HasPrefix(typ, `video/`) {
				if typ == `video/mp4` || typ == `video/webm` {
					continue
				}
				files = append(files, File{
					Name: entry.Name(),
					Type: typ,
					Size: utils.ByteCountIEC(info.Size()),
				})
			}
		}
		json.NewEncoder(w).Encode(files)
	})

	mux.HandleFunc(`POST /api/image`, func(w http.ResponseWriter, r *http.Request) {
		var imageArgs struct {
			Name string
			Q    int
		}

		if err := json.NewDecoder(r.Body).Decode(&imageArgs); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		outName, err := convertImage(r.Context(), s.dir, imageArgs.Name, imageArgs.Q)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		info := utils.Must1(os.Stat(filepath.Join(s.dir, outName)))
		json.NewEncoder(w).Encode(map[string]any{
			`Size`: utils.ByteCountIEC(info.Size()),
			`Path`: outName,
		})
	})

	mux.HandleFunc(`POST /api/video`, func(w http.ResponseWriter, r *http.Request) {
		var videoArgs struct {
			Name string
			CRF  int
		}

		if err := json.NewDecoder(r.Body).Decode(&videoArgs); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		outName, err := convertVideo(r.Context(), s.dir, videoArgs.Name, videoArgs.CRF)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		info := utils.Must1(os.Stat(filepath.Join(s.dir, outName)))
		json.NewEncoder(w).Encode(map[string]any{
			`Size`: utils.ByteCountIEC(info.Size()),
			`Path`: outName,
		})
	})

	mux.HandleFunc(`GET /in/`, func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(s.dir, strings.TrimPrefix(r.URL.Path, `/in/`))
		http.ServeFile(w, r, path)
	})

	mux.HandleFunc(`GET /out/`, func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(s.dir, strings.TrimPrefix(r.URL.Path, `/out/`))
		http.ServeFile(w, r, path)
	})

	lis := utils.Must1(net.Listen(`tcp4`, `:3367`))
	defer lis.Close()

	log.Printf(`HTTP: http://%s`, lis.Addr().String())

	if err := http.Serve(lis, mux); err != nil {
		log.Fatalln(err)
	}
}
