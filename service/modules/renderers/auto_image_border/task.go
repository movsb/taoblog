package auto_image_border

import (
	"context"
	"io"
	"log"
	"mime"
	"path"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/storage"
)

func NewTask(store utils.PluginStorage, fs *storage.SQLite, invalidate func(id int)) *Task {
	t := &Task{
		store:      store,
		fs:         fs,
		invalidate: invalidate,
	}
	return t
}

type Task struct {
	store      utils.PluginStorage
	fs         *storage.SQLite
	invalidate func(id int)
}

type Calculate func(file *models.File, input io.Reader, r, g, b byte, ratio float32) (float32, error)

func (t *Task) Run(ctx context.Context, calc Calculate) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	t.run(calc)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.run(calc)
		}
	}
}

func (t *Task) run(calc Calculate) {
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
		if file.Meta.BorderContrastRatio != 0 {
			continue
		}

		if !strings.HasPrefix(mime.TypeByExtension(path.Ext(file.Path)), `image/`) {
			continue
		}

		if err := calcFile(t, file, calc); err != nil {
			log.Println(err, file)
			return
		}
	}

	if len(updated) > 0 {
		log.Println(`计算可访问性结束`)
	}

	t.store.SetInteger(`last`, now.Unix())
}

func calcFile(t *Task, file *models.File, calc Calculate) error {
	log.Println(`计算可访问性：`, file.PostID, file.Path)

	// debug
	// if file.Path == `IMG_0091.avif` {
	// 	file.Path = file.Path
	// }

	fp, err := t.fs.Open(file.PostID, file.Path)
	if err != nil {
		return err
	}
	defer fp.Close()

	// value := BorderContrastRatio(fp, 255, 255, 255, 1)
	value, err := calc(file, fp, 255, 255, 255, 1)
	if err != nil {
		return err
	}

	// 零为特殊值，计算过的始终不为零。
	if value < 0.001 {
		value = 0.001
	}
	file.Meta.BorderContrastRatio = value

	if err := t.fs.UpdateFileMeta(file.ID, &file.Meta); err != nil {
		return err
	}

	log.Println(`可访问性：`, file.PostID, file.Path, value)

	t.invalidate(file.PostID)

	return nil
}
