package storage

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"io"
	std_fs "io/fs"
	"mime"
	"os"
	"path"
	"slices"
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

func (fs *SQLite) AllFiles() (map[int][]*proto.FileSpec, error) {
	var files []*models.File
	if err := fs.db.Select(fileFieldsWithoutData).Find(&files); err != nil {
		return nil, err
	}
	m := make(map[int][]*proto.FileSpec)
	for _, f := range files {
		m[f.PostID] = append(m[f.PostID], &proto.FileSpec{
			Path: f.Path,
			Mode: f.Mode,
			Size: f.Size,
			Time: uint32(f.ModTime),
			Type: mime.TypeByExtension(path.Ext(f.Path)),
			Meta: f.Meta.ToProto(),
		})
	}
	return m, nil
}

func (fs *SQLite) ForPost(id int) (std_fs.FS, error) {
	return &SQLiteForPost{s: fs, pid: id, dir: ``}, nil
}

var _ interface {
	std_fs.FS
	std_fs.StatFS
	std_fs.SubFS
	utils.DeleteFS
	utils.WriteFS
} = (*SQLiteForPost)(nil)

const fileFieldsWithoutData = `id,created_at,updated_at,post_id,path,mode,mod_time,size,meta`

func (fs *SQLiteForPost) Open(name string) (std_fs.File, error) {
	fullName := path.Clean(path.Join(fs.dir, name))
	var file models.File
	if err := fs.s.db.Select(fileFieldsWithoutData).Where(`post_id=?`, fs.pid).Where(`path=?`, fullName).Find(&file); err != nil {
		if taorm.IsNotFoundError(err) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}

	return file.FsFile(&_Reader{db: fs.s.db, pid: fs.pid, path: fullName, size: int(file.Size)}), nil
}

type _Reader struct {
	db   *taorm.DB
	pid  int
	path string

	// 不用再从表里面读一遍。
	// 以后可能独立出 data 表？
	size int

	data io.ReadSeeker
}

func (r *_Reader) prepare() error {
	if r.data == nil {
		var file models.File
		if err := r.db.Select(`data`).Where(`post_id=? AND path=?`, r.pid, r.path).Find(&file); err != nil {
			return err
		}
		if len(file.Data) != r.size {
			return fmt.Errorf(`文件内容数据已损坏。`)
		}
		r.data = bytes.NewReader(file.Data)
	}
	return nil
}

func (r *_Reader) Seek(offset int64, whence int) (int64, error) {
	if err := r.prepare(); err != nil {
		return 0, err
	}
	return r.data.Seek(offset, whence)
}
func (r *_Reader) Read(p []byte) (int, error) {
	if err := r.prepare(); err != nil {
		return 0, err
	}
	return r.data.Read(p)
}

func (fs *SQLiteForPost) ListFiles() ([]*proto.FileSpec, error) {
	if fs.dir != `` && fs.dir != `.` {
		return nil, errors.New(`不支持列举子目录文件。`)
	}
	var files []*models.File
	if err := fs.s.db.Select(fileFieldsWithoutData).Where(`post_id=?`, fs.pid).Find(&files); err != nil {
		return nil, err
	}
	// TODO 为了前端显示方便，这里临时按时间排序。
	slices.SortFunc(files, func(a, b *models.File) int {
		return -int(a.UpdatedAt - b.UpdatedAt)
	})
	specs := make([]*proto.FileSpec, 0, len(files))
	for _, f := range files {
		specs = append(specs, &proto.FileSpec{
			Path: f.Path,
			Mode: f.Mode,
			Size: f.Size,
			Time: uint32(f.ModTime),
			Type: mime.TypeByExtension(path.Ext(f.Path)),
			Meta: f.Meta.ToProto(),
		})
	}
	return specs, nil
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
			`meta`:       models.FileMetaFromProto(spec.Meta),
			`digest`:     digest(data),
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
			Meta:      models.FileMetaFromProto(spec.Meta),
			Digest:    digest(data),
			Data:      data,
		}
		return fs.s.db.Model(&file).Create()
	}
}

func digest(data []byte) string {
	d := md5.New()
	d.Write(data)
	s := d.Sum(nil)
	return fmt.Sprintf(`%x`, s)
}
