package storage

import (
	"bytes"
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
	meta *taorm.DB
	data *DataStore
}

type DataStore struct {
	data *taorm.DB
}

func NewDataStore(data *sql.DB) *DataStore {
	return &DataStore{
		data: taorm.NewDB(data),
	}
}

func NewSQLite(meta *sql.DB, data *DataStore) *SQLite {
	return &SQLite{
		meta: taorm.NewDB(meta),
		data: data,
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
	if err := fs.meta.Select(fileFieldsWithoutData).Find(&files); err != nil {
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

			Digest: f.Digest,
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

const fileFieldsWithoutData = `id,created_at,updated_at,post_id,path,mode,mod_time,size,meta,digest`

func (fs *SQLiteForPost) Open(name string) (std_fs.File, error) {
	fullName := path.Clean(path.Join(fs.dir, name))
	var file models.File
	if err := fs.s.meta.Select(fileFieldsWithoutData).Where(`post_id=?`, fs.pid).Where(`path=?`, fullName).Find(&file); err != nil {
		if taorm.IsNotFoundError(err) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}

	return file.FsFile(&_Reader{
		// 这里用数据数据库，而不是元数据数据库。
		db:   fs.s.data,
		meta: &file,
	}), nil
}

// 获取文件的内容数据。
// 零大小数据不能走这里。
func (d *DataStore) GetFile(postID int, digest string, size int) (io.ReadSeeker, error) {
	var file models.FileData
	if err := d.data.Select(`data`).Where(`post_id=? AND digest=?`, postID, digest).Find(&file); err != nil {
		return nil, err
	}
	if len(file.Data) != size {
		return nil, fmt.Errorf(`文件内容数据已损坏。`)
	}
	return bytes.NewReader(file.Data), nil
}

func (d *DataStore) UpdateData(postID int, odlDigest, newDigest string, data []byte) error {
	r, err := d.data.From(models.FileData{}).Where(`post_id=? AND digest=?`, postID, odlDigest).UpdateMap(taorm.M{
		`digest`: newDigest,
		`data`:   data,
	})
	if err != nil {
		return fmt.Errorf(`更新文件失败：%w`, err)
	}
	n, err := r.RowsAffected()
	if err != nil {
		return fmt.Errorf(`更新文件失败：%w`, err)
	}
	if n != 1 {
		return fmt.Errorf(`更新文件失败：没有更新到任何行`)
	}
	return nil
}

func (d *DataStore) CreateData(postID int, digest string, data []byte) error {
	dataModel := models.FileData{
		PostID: postID,
		Digest: digest,
		Data:   data,
	}
	if err := d.data.Model(&dataModel).Create(); err != nil {
		return fmt.Errorf(`创建文件数据失败：%w`, err)
	}
	return nil
}

func (d *DataStore) DeleteData(postID int, digest string) error {
	return d.data.From(models.FileData{}).Where(`post_id=? AND digest=?`, postID, digest).Delete()
}

// 只有 PostID 和 Digest 字段。
func (d *DataStore) ListAllFiles() ([]*models.FileData, error) {
	var files []*models.FileData
	d.data.Select(`post_id,digest`).MustFind(&files)
	return files, nil
}

type _Reader struct {
	db   *DataStore
	meta *models.File
	data io.ReadSeeker
}

func (r *_Reader) prepare() error {
	if r.data == nil {
		if r.meta.Size == 0 {
			r.data = bytes.NewReader(nil)
		} else {
			f, err := r.db.GetFile(r.meta.PostID, r.meta.Digest, int(r.meta.Size))
			if err != nil {
				return fmt.Errorf(`获取文件数据失败：%w`, err)
			}
			r.data = f
		}
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
	if err := fs.s.meta.Select(fileFieldsWithoutData).Where(`post_id=?`, fs.pid).Find(&files); err != nil {
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

			Digest: f.Digest,
		})
	}
	return specs, nil
}

func (fs *SQLiteForPost) Delete(name string) error {
	fullName := path.Clean(path.Join(fs.dir, name))
	var file models.File
	if err := fs.s.meta.Select(`id,digest`).Where(`post_id=?`, fs.pid).Where(`path=?`, fullName).Find(&file); err != nil {
		if taorm.IsNotFoundError(err) {
			return os.ErrNotExist
		}
		return err
	}
	if err := fs.s.meta.Model(&file).Delete(); err != nil {
		return err
	}
	return fs.s.data.DeleteData(fs.pid, file.Digest)
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
	if err := fs.s.meta.Select(fileFieldsWithoutData).Where(`post_id=?`, fs.pid).
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

func (fs *SQLiteForPost) UpdateCaption(name string, caption *proto.FileSpec_Meta_Source) error {
	var file models.File
	fullName := path.Clean(path.Join(fs.dir, name))
	if err := fs.s.meta.Where(`post_id=? AND path=?`, fs.pid, fullName).Find(&file); err != nil {
		return err
	}
	file.Meta.Source = caption
	_, err := fs.s.meta.Model(&file).UpdateMap(taorm.M{
		`updated_at`: time.Now().Unix(),
		`meta`:       file.Meta,
	})
	if err != nil {
		return fmt.Errorf(`更新文件失败：%w`, err)
	}
	return nil
}

func (fs *SQLiteForPost) Write(spec *proto.FileSpec, r io.Reader) error {
	if fs.pid <= 0 {
		return fmt.Errorf(`没有指定文件编号`)
	}

	if !std_fs.ValidPath(spec.Path) || spec.Path == "." {
		return fmt.Errorf(`无效文件名：%q`, spec.Path)
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

	meta := models.FileMetaFromProto(spec.Meta)
	models.Encrypt(&meta.Encryption, data)

	var old models.File

	digest := models.Digest(data)

	// 更新文件。
	if err := fs.s.meta.Where(`post_id=? AND path=?`, fs.pid, fullName).Find(&old); err == nil {
		_, err := fs.s.meta.Model(&old).UpdateMap(taorm.M{
			`updated_at`: now.Unix(),
			`mode`:       spec.Mode,
			`mod_time`:   spec.Time,
			`size`:       spec.Size,
			`meta`:       meta,
			`digest`:     digest,
		})
		if err != nil {
			return fmt.Errorf(`更新文件失败：%w`, err)
		}
		// 零大小文件不存放数据。
		if spec.Size == 0 {
			return fs.s.data.DeleteData(fs.pid, old.Digest)
		}
		return fs.s.data.UpdateData(fs.pid, old.Digest, digest, data)
	}

	// 创建新文件。
	file := models.File{
		CreatedAt: now.Unix(),
		UpdatedAt: now.Unix(),
		PostID:    fs.pid,
		Path:      fullName,
		Mode:      spec.Mode,
		ModTime:   int64(spec.Time),
		Size:      spec.Size,
		Meta:      meta,
		Digest:    digest,
	}
	if err := fs.s.meta.Model(&file).Create(); err != nil {
		return fmt.Errorf(`创建文件失败：%w`, err)
	}

	// 零大小文件不存放数据。
	if spec.Size == 0 {
		return nil
	}

	return fs.s.data.CreateData(fs.pid, file.Digest, data)
}
