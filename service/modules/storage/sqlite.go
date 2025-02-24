package storage

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	std_fs "io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
	"github.com/movsb/taorm"
)

type SQLite struct {
	db *taorm.DB
}

func NewSQLite(db *sql.DB) *SQLite {
	return &SQLite{
		db: taorm.NewDB(db),
	}
}

type SQLiteForPost struct {
	s   *SQLite
	pid int
	dir string
}

var _ interface {
	theme_fs.FS
} = (*SQLite)(nil)

func (fs *SQLite) Root() (std_fs.FS, error) {
	// panic(`未实现根目录`)
	return fs.ForPost(0)
}
func (fs *SQLite) ForPost(id int) (std_fs.FS, error) {
	return &SQLiteForPost{s: fs, pid: id, dir: ``}, nil
}

var _ interface {
	std_fs.FS
	std_fs.StatFS
	std_fs.ReadDirFS
	std_fs.SubFS
	utils.DeleteFS
	utils.WriteFS
} = (*SQLiteForPost)(nil)

const fileFieldsWithoutData = `id,created_at,updated_at,post_id,path,mode,mod_time,size`

func (fs *SQLiteForPost) Open(name string) (std_fs.File, error) {
	fullName := path.Clean(path.Join(fs.dir, name))
	var file models.File
	if err := fs.s.db.Select(fileFieldsWithoutData).Where(`post_id=?`, fs.pid).Where(`path=?`, fullName).Find(&file); err != nil {
		if taorm.IsNotFoundError(err) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}

	return file.FsFile(&_Reader{db: fs.s.db, pid: fs.pid, path: fullName}), nil
}

type _Reader struct {
	db   *taorm.DB
	pid  int
	path string

	data io.Reader
}

func (r *_Reader) Read(p []byte) (int, error) {
	if r.data == nil {
		var file models.File
		if err := r.db.Select(`data`).Where(`post_id=? AND path=?`, r.pid, r.path).Find(&file); err != nil {
			return 0, err
		}
		r.data = bytes.NewBuffer(file.Data)
	}
	return r.data.Read(p)
}

func (fs *SQLiteForPost) Delete(name string) error {
	fullName := path.Clean(path.Join(fs.dir, name))
	var file models.File
	if err := fs.s.db.Select(`id`).Where(`post_id=?`, fs.pid).Where(`path=?`, fullName).Find(&file); err != nil {
		if taorm.IsNotFoundError(err) {
			return os.ErrNotExist
		}
		return err
	}
	return fs.s.db.Model(&file).Delete()
}

// 用了 SQLite 的特殊函数 glob
func (fs *SQLiteForPost) ReadDir(name string) ([]std_fs.DirEntry, error) {
	dir := path.Clean(path.Join(fs.dir, name))
	pattern := path.Join(dir, `*`)
	var files []*models.File
	if err := fs.s.db.Select(fileFieldsWithoutData).Where(`post_id=?`, fs.pid).
		Where(`path glob ?`, pattern).Find(&files); err != nil {
		if taorm.IsNotFoundError(err) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	prefix := dir + `/`
	if prefix == `./` {
		prefix = ``
	}
	var entries []std_fs.DirEntry
	for _, file := range files {
		after, found := strings.CutPrefix(file.Path, prefix)
		if !found {
			continue
		}
		file.Path = after // 修改了原数据
		entries = append(entries, file.DirEntry())
	}
	return entries, nil
}

func (fs *SQLiteForPost) Stat(name string) (std_fs.FileInfo, error) {
	fullName := path.Clean(path.Join(fs.dir, name))
	if fullName == `.` {
		return (&models.File{
			Path:    `.`,
			Mode:    uint32(std_fs.ModeDir) | 0755,
			ModTime: time.Now().Unix(),
			Size:    0,
		}).InfoFile(), nil
	}

	var file models.File
	if err := fs.s.db.Select(fileFieldsWithoutData).Where(`post_id=?`, fs.pid).
		Where(`path=?`, fullName).Find(&file); err != nil {
		if taorm.IsNotFoundError(err) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	return file.InfoFile(), nil
}

func (fs *SQLiteForPost) Sub(dir string) (std_fs.FS, error) {
	return &SQLiteForPost{
		s:   fs.s,
		pid: fs.pid,
		dir: path.Join(fs.dir, dir),
	}, nil
}

func (fs *SQLiteForPost) Write(spec *proto.FileSpec, r io.Reader) error {
	if fs.pid <= 0 {
		return fmt.Errorf(`没有指定文件编号`)
	}
	fullName := path.Clean(path.Join(fs.dir, spec.Path))
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if len(data) != int(spec.Size) {
		return fmt.Errorf(`文件大小不相等：want:%d vs got:%d`, spec.Size, len(data))
	}

	now := time.Now()

	var old models.File
	if err := fs.s.db.Where(`post_id=? AND path=?`, fs.pid, fullName).Find(&old); err == nil {
		_, err := fs.s.db.Model(&old).UpdateMap(taorm.M{
			`updated_at`: now.Unix(),
			`mode`:       spec.Mode,
			`mod_time`:   spec.Time,
			`size`:       spec.Size,
			`data`:       data,
		})
		return err
	} else {
		file := models.File{
			CreatedAt: now.Unix(),
			UpdatedAt: now.Unix(),
			PostID:    fs.pid,
			Path:      fullName,
			Mode:      spec.Mode,
			ModTime:   int64(spec.Time),
			Size:      spec.Size,
			Data:      data,
		}
		return fs.s.db.Model(&file).Create()
	}
}
