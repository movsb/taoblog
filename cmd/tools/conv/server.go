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
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

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
	Time time.Time
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
		files := []File{}
		const maxFiles = 100

		skips := map[string]struct{}{}

		filepath.WalkDir(s.dir, func(path string, d fs.DirEntry, err error) error {
			if len(files) > maxFiles {
				log.Println(`too many files`)
				return filepath.SkipAll
			}

			if _, ok := skips[path]; ok {
				return nil
			}

			if err != nil {
				log.Println(err)
				return nil
			}

			if !d.Type().IsRegular() {
				return nil
			}

			info := utils.Must1(d.Info())
			typ := mime.TypeByExtension(filepath.Ext(d.Name()))
			if strings.HasPrefix(typ, `image/`) {
				if typ == `image/avif` {
					return nil
				}

				if ext := filepath.Ext(d.Name()); strings.ToLower(ext) == `.heic` {
					jpg1 := strings.TrimSuffix(d.Name(), ext) + `.JPG`
					jpg2 := strings.TrimSuffix(d.Name(), ext) + `.jpg`
					_, err1 := os.Stat(filepath.Join(s.dir, jpg1))
					_, err2 := os.Stat(filepath.Join(s.dir, jpg2))
					if err1 == nil || err2 == nil {
						return nil
					}
					jpgPath := convertHEIC(context.Background(), path)
					info := utils.Must1(os.Stat(jpgPath))
					files = append(files, File{
						Name: jpgPath,
						Type: typ,
						Size: utils.ByteCountIEC(info.Size()),
						Time: info.ModTime(),
					})
					skips[jpgPath] = struct{}{}
					return nil
				}

				files = append(files, File{
					Name: path,
					Type: typ,
					Size: utils.ByteCountIEC(info.Size()),
					Time: info.ModTime(),
				})
			} else if strings.HasPrefix(typ, `video/`) {
				if typ == `video/mp4` || typ == `video/webm` {
					return nil
				}
				files = append(files, File{
					Name: path,
					Type: typ,
					Size: utils.ByteCountIEC(info.Size()),
					Time: info.ModTime(),
				})
			}

			return nil
		})

		slices.SortFunc(files, func(a, b File) int {
			return int(a.Time.Unix()) - int(b.Time.Unix())
		})

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

		output, err := convertImage(r.Context(), filepath.Join(s.dir, imageArgs.Name), s.dir, imageArgs.Q)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		info := utils.Must1(os.Stat(output))
		json.NewEncoder(w).Encode(map[string]any{
			`Size`: utils.ByteCountIEC(info.Size()),
			`Path`: output,
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

		output, err := convertVideo(r.Context(), filepath.Join(s.dir, videoArgs.Name), s.dir, videoArgs.CRF)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		info := utils.Must1(os.Stat(output))
		json.NewEncoder(w).Encode(map[string]any{
			`Size`: utils.ByteCountIEC(info.Size()),
			`Path`: output,
		})
	})

	mux.HandleFunc(`GET /in/`, func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(s.dir, strings.TrimPrefix(r.URL.Path, `/in/`))
		http.ServeFile(w, r, path)
	})

	mux.HandleFunc(`GET /out/`, func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Clean(strings.TrimPrefix(r.URL.Path, `/out/`))
		http.ServeFile(w, r, path)
	})

	lis := utils.Must1(net.Listen(`tcp4`, `:3367`))
	defer lis.Close()

	addr := `http://` + lis.Addr().String()
	log.Printf(`HTTP: %s`, addr)
	exec.Command(`open`, addr).Run()

	if err := http.Serve(lis, mux); err != nil {
		log.Fatalln(err)
	}
}
