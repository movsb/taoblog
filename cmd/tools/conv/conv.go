package conv

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/spf13/cobra"
)

func AddCommands(parent *cobra.Command) {
	convCmd := cobra.Command{
		Use:   `conv <files/dirs...>`,
		Short: `图片/视频格式转换。`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			utils.Must1(exec.LookPath(`exiftool`))
			utils.Must1(exec.LookPath(`avifenc`))
			utils.Must1(exec.LookPath(`sips`))
			utils.Must1(exec.LookPath(`ffmpeg`))

			wg := sync.WaitGroup{}

			handleFile := func(ctx context.Context, path string) {
				switch strings.ToLower(filepath.Ext(path)) {
				case `.heic`, `.jpeg`, `.jpg`, `png`:
					wg.Add(1)
					go func() {
						defer wg.Done()
						convertImage(ctx, path, `.`)
					}()
				case `.mov`:
					wg.Add(1)
					go func() {
						defer wg.Done()
						convertVideo(ctx, path, `.`)
					}()
				default:
					log.Println(`忽略文件：`, path)
				}
			}

			for _, arg := range args {
				info := utils.Must1(os.Stat(arg))
				if info.IsDir() {
					entries := utils.Must1(os.ReadDir(arg))
					reImage := regexp.MustCompile(`^(?i:IMG_E\d+\.HEIC)$`)
					reVideo := regexp.MustCompile(`^(?i:IMG_E\d+\.MOV)$`)
					for _, entry := range entries {
						if reImage.MatchString(entry.Name()) {
							continue
						}
						if reVideo.MatchString(entry.Name()) {
							continue
						}
						handleFile(cmd.Context(), filepath.Join(arg, entry.Name()))
					}
				} else {
					handleFile(cmd.Context(), arg)
				}
			}

			wg.Wait()
		},
	}
	parent.AddCommand(&convCmd)
}

func run(ctx context.Context, cmd ...string) {
	c := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		log.Println(err)
	}
}

func convertImage(ctx context.Context, path string, outDir string) {
	originPath := path

	if ext := filepath.Ext(path); strings.EqualFold(ext, `.heic`) {
		baseName := filepath.Base(path)
		jpgName := strings.TrimSuffix(baseName, ext) + `.jpg`
		tmpJpgPath := filepath.Join(os.TempDir(), jpgName)
		run(ctx, `sips`, `-s`, `format`, `jpeg`, path, `-o`, tmpJpgPath)
		defer os.Remove(tmpJpgPath)
		path = tmpJpgPath
	}

	outName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) + `.avif`
	fullOutName := filepath.Join(outDir, outName)
	run(ctx, `avifenc`, path, fullOutName)
	defer os.Remove(fullOutName + `_original`)
	run(ctx, `exiftool`, `-tagsFromFile`, originPath, fullOutName)
}

func convertVideo(ctx context.Context, path string, outDir string) {
	name := filepath.Base(path)
	name = strings.TrimSuffix(name, filepath.Ext(name)) + `.mp4`
	outName := filepath.Join(outDir, name)
	// -i 要放在 -preset 前面，为什么？
	run(ctx, `ffmpeg`, `-y`, `-i`, path, `-preset`, `veryslow`, outName)
}
