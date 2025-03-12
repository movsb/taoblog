package sass

import (
	"log"
	"os"
	"os/exec"
	"path"
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

// TODO 未处理带空格的路径。
// 使用 sass --watch 不易 debounce，所以没使用。
// NOTE: 只在 DevMode 下执行。
func Watch(dir string, input, output string) {
	if !version.DevMode() {
		// log.Println(`非开发模式，不观察样式：`, dir)
		return
	}

	log.Println(`动态样式观察：`, dir, input)
	exitOnError := true

	bundle := func() {
		// TODO 不要忽略警告。
		cmd := exec.Command(`sass`, `-q`, `--no-source-map`, input, output)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			if exitOnError {
				log.Fatalln(err)
			} else {
				log.Println(err)
			}
		} else {
			log.Println(`样式更新：`, dir)
		}
	}

	exitOnError = true
	bundle()
	exitOnError = false

	go func() {
		debouncer := utils.NewDebouncer(time.Second, bundle)
		fsDir := utils.NewDirFSWithNotify(dir).(utils.FsWithChangeNotify)
		for event := range fsDir.Changed() {
			switch event.Op {
			case fsnotify.Create, fsnotify.Remove, fsnotify.Write:
				// 只关心 .scss 文件
				if path.Ext(event.Name) != `.scss` {
					break
				}
				debouncer.Enter()
			}
		}
	}()
}
