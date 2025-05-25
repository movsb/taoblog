package sass

import (
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
)

func WatchAsync(dir string, input, output string) {
	go Watch(dir, input, output)
}

func WatchDefaultAsync(dir string) {
	go Watch(dir, `style.scss`, `style.css`)
}

// 使用 sass --watch 不易 debounce，所以没使用。
// NOTE: 只在 DevMode 下执行。
func Watch(dir string, input, output string) {
	if !version.DevMode() {
		log.Fatalln(`非开发模式，不能观察样式：`, dir)
	}

	// 去掉可能的绝对路径前缀，用于日志打印。
	dirStripped := dir
	if _, after, found := strings.Cut(dir, version.NameLowercase); found {
		dirStripped = after[1:]
	}

	log.Println(`动态样式观察：`, dirStripped, input)
	exitOnError := true

	bundle := func() {
		cmd := exec.Command(`sass`, `-q`, `--no-source-map`, input, output)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			if exitOnError {
				log.Fatalln(dirStripped, err)
			} else {
				log.Println(dirStripped, err)
			}
		} else {
			log.Println(`样式更新：`, dirStripped)
		}
	}

	exitOnError = true
	bundle()
	exitOnError = false

	go func() {
		debouncer := utils.NewDebouncer(time.Second, bundle)
		watchFS := utils.NewOSDirFS(dir).(utils.WatchFS)
		events, close := utils.Must2(watchFS.Watch())
		defer close()
		for event := range events {
			if event.Has(fsnotify.Create | fsnotify.Remove | fsnotify.Write) {
				// 只关心 .scss 文件
				if path.Ext(event.Name) != `.scss` {
					continue
				}
				debouncer.Enter()
			}
		}
	}()
}
