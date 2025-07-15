package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
)

type File struct {
	ID        int
	CreatedAt int64
	UpdatedAt int64
	PostID    int
	Path      string
	Mode      uint32
	ModTime   int64
	Size      uint32
	Meta      FileMeta
	Digest    string
	Data      []byte
}

// 元数据。
// 由 State.sys 返回文件拿到。
type FileMeta struct {
	// 如果是图片，则包含宽高。
	// 只能是空值。上传的时候浏览器计算。
	Width, Height int

	Encryption FileEncryptionMeta
}

type FileEncryptionMeta struct {
	Key    []byte // 加密密钥（32字节）
	Nonce  []byte // 初始化向量（12字节）
	Digest string // 加密后的摘要
	Size   int    // 加密后的大小
}

func Encrypt(m *FileEncryptionMeta, data []byte) {
	m.Key = make([]byte, 32)
	m.Nonce = make([]byte, 12)
	rand.Read(m.Key)
	rand.Read(m.Nonce)
	aes := utils.Must1(aes.NewCipher(m.Key[:]))
	aead := utils.Must1(cipher.NewGCM(aes))
	buf := make([]byte, 0, len(data)+aead.Overhead())
	encrypted := aead.Seal(buf, m.Nonce[:], data, nil)
	m.Size = len(encrypted)
	m.Digest = Digest(encrypted)
}

func Digest(data []byte) string {
	d := md5.New()
	d.Write(data)
	s := d.Sum(nil)
	return fmt.Sprintf(`%x`, s)
}

func (m FileMeta) ToProto() *proto.FileSpec_Meta {
	return &proto.FileSpec_Meta{
		Width:  int32(m.Width),
		Height: int32(m.Height),
	}
}

func FileMetaFromProto(m *proto.FileSpec_Meta) FileMeta {
	if m == nil {
		return FileMeta{}
	}
	return FileMeta{
		Width:  int(m.Width),
		Height: int(m.Height),
	}
}

func (m FileMeta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *FileMeta) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	}
	return fmt.Errorf(`cannot convert %T to files.meta`, value)
}

func (File) TableName() string {
	return `files`
}

func (f *File) FsFile(r io.ReadSeeker) fs.File {
	return &_FsFile{f: f, ReadSeeker: r}
}

func (f *File) InfoFile() fs.FileInfo {
	return &_InfoFile{f: f}
}

func (f *File) GetImageDimension() (int, int) {
	return f.Meta.Width, f.Meta.Height
}

type _FsFile struct {
	f *File
	io.ReadSeeker
}

func (f *_FsFile) Stat() (fs.FileInfo, error) { return &_InfoFile{f.f}, nil }
func (f *_FsFile) Close() error               { return nil }

type _InfoFile struct{ f *File }

func (f *_InfoFile) Name() string       { return path.Base(f.f.Path) }
func (f *_InfoFile) Size() int64        { return int64(f.f.Size) }
func (f *_InfoFile) Mode() fs.FileMode  { return fs.FileMode(f.f.Mode) }
func (f *_InfoFile) ModTime() time.Time { return time.Unix(f.f.ModTime, 0).Local() }
func (f *_InfoFile) IsDir() bool        { return f.Mode().IsDir() }
func (f *_InfoFile) Sys() any           { return f.f }
