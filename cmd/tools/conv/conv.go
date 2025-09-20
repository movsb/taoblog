package conv

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/spf13/cobra"
)

type Result struct {
	Path string
	Err  error
}

func AddCommands(parent *cobra.Command) {
	convCmd := cobra.Command{
		Use:   `conv <dir>`,
		Short: `图片/视频格式转换。`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			utils.Must1(exec.LookPath(`exiftool`))
			utils.Must1(exec.LookPath(`avifenc`))
			utils.Must1(exec.LookPath(`sips`))
			utils.Must1(exec.LookPath(`ffmpeg`))

			s := NewServer(args[0])
			go s.Run(context.Background())

			select {}
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

func convertHEIC(ctx context.Context, path string) string {
	if ext := filepath.Ext(path); strings.EqualFold(ext, `.heic`) {
		dir := filepath.Dir(path)
		baseName := filepath.Base(path)
		jpgName := strings.TrimSuffix(baseName, ext) + `.jpg`
		jpgPath := filepath.Join(dir, jpgName)
		if _, err := os.Stat(jpgPath); err != nil {
			run(ctx, `sips`, `-s`, `format`, `jpeg`, path, `-o`, jpgPath)
		}
		return jpgPath
	}
	return path
}

func convertImage(ctx context.Context, dir, name string, q int) (string, error) {
	outName := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name)) + `.avif`
	fullOutName := filepath.Join(dir, outName)

	// avifenc ../IMG_0923.JPG 1.avif
	// Directly copied JPEG pixel data (no YUV conversion): ../IMG_0923.JPG
	// XMP extraction failed: invalid multiple standard XMP segments
	// Cannot read input file: ../IMG_0923.JPG
	err := run(ctx, `avifenc`, `-q`, fmt.Sprint(q), filepath.Join(dir, name), fullOutName)
	if err != nil {
		err = run(ctx, `ffmpeg`, `-y`, `-i`, filepath.Join(dir, name), fullOutName)
	}
	if err != nil {
		return outName, err
	}
	defer os.Remove(fullOutName + `_original`)
	return outName, run(ctx, `exiftool`, `-tagsFromFile`, filepath.Join(dir, name), fullOutName)
}

func convertVideo(ctx context.Context, dir, name string, crf int) (string, error) {
	outName := strings.TrimSuffix(name, filepath.Ext(name)) + `.mp4`
	fullOutPath := filepath.Join(dir, outName)
	return outName, run(ctx, `ffmpeg`, `-y`, `-i`, filepath.Join(dir, name), `-crf`, fmt.Sprint(crf), fullOutPath)
}
