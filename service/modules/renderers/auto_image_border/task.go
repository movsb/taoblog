package auto_image_border

import (
	"context"
	"io"
	"log"
	"mime"
	"path"
	"runtime"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
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

func (t *Task) Run(ctx context.Context, s proto.Utils_RegisterAutoImageBorderHandlerServer) {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	t.run(s)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.run(s)
		case <-s.Context().Done():
			log.Println(`handler server context down`)
			return
		}
	}
}

func (t *Task) run(s proto.Utils_RegisterAutoImageBorderHandlerServer) {
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

		if err := calcFile(t, file, s); err != nil {
			log.Println(err, file)
			return
		}

		if mime.TypeByExtension(path.Ext(file.Path)) == `image/avif` {
			time.Sleep(time.Second * 10)
			log.Println(`运行垃圾回收`)
			runtime.GC()
		}
	}

	if len(updated) > 0 {
		log.Println(`计算可访问性结束`)
	}

	t.store.SetInteger(`last`, now.Unix())
}

func calcFile(t *Task, file *models.File, s proto.Utils_RegisterAutoImageBorderHandlerServer) error {
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
	value, err := remoteCalc(s, file, fp, 255, 255, 255, 1)
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

func remoteCalc(s proto.Utils_RegisterAutoImageBorderHandlerServer, file *models.File, fp io.Reader, r, g, b byte, ratio float32) (value float32, outErr error) {
	defer utils.CatchAsError(&outErr)

	log.Println(`发送数据`)

	utils.Must(s.Send(&proto.AutoImageBorderRequest{
		PostId: uint32(file.PostID),
		Path:   file.Path,

		Data:  utils.Must1(io.ReadAll(fp)),
		R:     uint32(r),
		G:     uint32(g),
		B:     uint32(b),
		Ratio: ratio,
	}))

	log.Println(`等待接收数据`)

	return utils.Must1(s.Recv()).GetValue(), nil
}
