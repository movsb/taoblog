package conv

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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

func copyModTime(from, to string) {
	info1 := utils.Must1(os.Stat(from))
	modTime := info1.ModTime()
	os.Chtimes(to, time.Time{}, modTime)
}

func convertHEIC(ctx context.Context, path string) string {
	if ext := filepath.Ext(path); strings.EqualFold(ext, `.heic`) {
		dir := filepath.Dir(path)
		baseName := filepath.Base(path)
		jpgName := strings.TrimSuffix(baseName, ext) + `.jpg`
		jpgPath := filepath.Join(dir, jpgName)
		if _, err := os.Stat(jpgPath); err != nil {
			utils.Must(run(ctx, `sips`, `-s`, `format`, `jpeg`, path, `-o`, jpgPath))
			copyModTime(path, jpgPath)
		}
		return jpgPath
	}
	return path
}

func convertImage(ctx context.Context, input string, outputDir string, q int) (string, error) {
	output := filepath.Join(outputDir, strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))+`.avif`)

	// avifenc ../IMG_0923.JPG 1.avif
	// Directly copied JPEG pixel data (no YUV conversion): ../IMG_0923.JPG
	// XMP extraction failed: invalid multiple standard XMP segments
	// Cannot read input file: ../IMG_0923.JPG
	err := run(ctx, `avifenc`, `-q`, fmt.Sprint(q), input, output)
	if err != nil {
		err = run(ctx, `ffmpeg`, `-y`, `-i`, input, output)
	}
	if err != nil {
		return ``, err
	}

	defer os.Remove(output + `_original`)
	if err := run(ctx, `exiftool`, `-tagsFromFile`, input, output); err != nil {
		return ``, err
	}

	copyModTime(input, output)

	return output, nil
}

func convertVideo(ctx context.Context, input string, outputDir string, crf int) (string, error) {
	output := filepath.Join(outputDir, strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))+`.mp4`)
	if err := run(ctx, `ffmpeg`, `-y`, `-i`, input, `-crf`, fmt.Sprint(crf), output); err != nil {
		return ``, err
	}
	copyModTime(input, output)
	return output, nil
}
