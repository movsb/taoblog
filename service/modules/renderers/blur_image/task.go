package blur_image

import (
	"context"
	"encoding/base64"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"log"
	"mime"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/storage"
	"go.n16f.net/thumbhash"
)

func NewTask(ctx context.Context, store utils.PluginStorage, fs *storage.SQLite, invalidate func(pid int)) *Task {
	t := &Task{
		store:      store,
		fs:         fs,
		invalidate: invalidate,
	}
	go t.Run(ctx)
	return t
}

type Task struct {
	store      utils.PluginStorage
	fs         *storage.SQLite
	invalidate func(pid int)
}

func (t *Task) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.run()
		}
	}
}

func (t *Task) run() {
	now := time.Now()

	last, err := t.store.GetIntegerDefault(`last`, 0)
	if err != nil {
		log.Println(`退出：`, err)
		return
	}

	updated, err := t.fs.GetUpdatedFiles(time.Unix(last, 0))
	if err != nil {
		log.Println(err)
		return
	}
	for _, file := range updated {
		if file.Meta.ThumbHash != `` {
			continue
		}
		log.Println(`计算Hash：`, file.PostID, file.Path)
		fp, err := t.fs.Open(file.PostID, file.Path)
		if err != nil {
			log.Println(err)
			return
		}
		hash, err := calcHash(file, fp)
		if err != nil {
			log.Println(err)
			continue
		}
		if hash == `` {
			continue
		}
		file.Meta.ThumbHash = hash
		if err := t.fs.UpdateFileMeta(file.ID, &file.Meta); err != nil {
			log.Println(err)
			return
		}
		t.invalidate(file.PostID)
	}

	t.store.SetInteger(`last`, now.Unix())
}

func calcHash(info *models.File, fp fs.File) (string, error) {
	defer fp.Close()

	// if info.PostID != 1864 {
	// 	return ``, nil
	// }

	ext := path.Ext(info.Path)
	typ := mime.TypeByExtension(ext)
	if !strings.HasPrefix(typ, `image/`) {
		return ``, nil
	}

	var (
		err  error
		data image.Image
	)

	switch strings.ToLower(ext) {
	case `.jpg`, `.jpeg`:
		data, err = jpeg.Decode(fp)
	case `.png`:
		data, err = png.Decode(fp)
	case `.avif`:
		tmp, err2 := os.CreateTemp(``, ``)
		if err2 != nil {
			return ``, err2
		}
		defer os.Remove(tmp.Name())
		if _, err := io.Copy(tmp, fp); err != nil {
			return ``, err
		}
		tmp.Close()
		tmp2, err2 := os.CreateTemp(``, `*.png`)
		if err2 != nil {
			return ``, err2
		}
		defer os.Remove(tmp2.Name())
		cmd := exec.Command(`avifdec`, tmp.Name(), tmp2.Name())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return ``, err
		}
		fp, err2 := os.Open(tmp2.Name())
		if err2 != nil {
			return ``, err2
		}
		defer fp.Close()
		data, err = png.Decode(fp)
	default:
		log.Println(`不知道如何处理：`, info)
		return ``, nil
	}

	if err != nil {
		log.Println(err, info)
		return ``, err
	}

	hash := thumbhash.EncodeImage(data)
	return base64.StdEncoding.EncodeToString(hash), nil
}
