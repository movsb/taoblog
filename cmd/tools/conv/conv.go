package conv

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/spf13/cobra"
)

type Result struct {
	Path string
	Err  error
}

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
			results := make(chan Result)

			handleFile := func(ctx context.Context, path string) {
				switch strings.ToLower(filepath.Ext(path)) {
				case `.heic`, `.jpeg`, `.jpg`, `.png`:
					wg.Add(1)
					go func() {
						defer wg.Done()
						if err := convertImage(ctx, path, `.`); err != nil {
							results <- Result{
								Path: path,
								Err:  err,
							}
						}
					}()
				case `.mov`:
					wg.Add(1)
					go func() {
						defer wg.Done()
						if err := convertVideo(ctx, path, `.`); err != nil {
							results <- Result{
								Path: path,
								Err:  err,
							}
						}
					}()
				default:
					log.Println(`忽略文件：`, path)
				}
			}

			for _, arg := range args {
				info := utils.Must1(os.Stat(arg))
				if info.IsDir() {
					entries := utils.Must1(os.ReadDir(arg))
					for _, entry := range entries {
						handleFile(cmd.Context(), filepath.Join(arg, entry.Name()))
					}
				} else {
					handleFile(cmd.Context(), arg)
				}
			}

			go func() {
				wg.Wait()
				close(results)
			}()

			var errors []Result
			for r := range results {
				errors = append(errors, r)
			}

			for _, e := range errors {
				log.Println(e.Path, e.Err)
			}
		},
	}
	parent.AddCommand(&convCmd)
}

func run(ctx context.Context, cmd ...string) error {
	c := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func convertImage(ctx context.Context, path string, outDir string) error {
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

	// avifenc ../IMG_0923.JPG 1.avif
	// Directly copied JPEG pixel data (no YUV conversion): ../IMG_0923.JPG
	// XMP extraction failed: invalid multiple standard XMP segments
	// Cannot read input file: ../IMG_0923.JPG
	err := run(ctx, `avifenc`, path, fullOutName)
	if err != nil {
		err = run(ctx, `ffmpeg`, `-y`, `-i`, path, fullOutName)
	}
	if err != nil {
		return err
	}
	defer os.Remove(fullOutName + `_original`)
	return run(ctx, `exiftool`, `-tagsFromFile`, originPath, fullOutName)
}

func convertVideo(ctx context.Context, path string, outDir string) error {
	name := filepath.Base(path)
	name = strings.TrimSuffix(name, filepath.Ext(name)) + `.mp4`
	outName := filepath.Join(outDir, name)
	// -i 要放在 -preset 前面，为什么？
	return run(ctx, `ffmpeg`, `-y`, `-i`, path, `-preset`, `veryslow`, outName)
}
