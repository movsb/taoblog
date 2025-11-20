package assets

import (
	"context"
	"fmt"
	"log"
	"mime"
	"os"
	"os/exec"
	pathpkg "path"
	"path/filepath"
	"strings"

	"github.com/movsb/taoblog/modules/utils"
)

func shouldConvertImage(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}

func isImageFile(path string) bool {
	typ := mime.TypeByExtension(pathpkg.Ext(path))
	return strings.HasPrefix(typ, "image/")
}

// 输入：
//
//	path: 原始文件路径，用于计算新的文件名，不用作为输入文件的路径。
//	input: 输入文件内容，文件路径。
//
// 输出：
//
//	修改后缀后的文件路径。
//	格式转换后的文件路径。
//
// 由使用者负责删除临时文件。
func ConvertToAVIF(ctx context.Context, path string, input string, keepTags bool) (_ string, _ string, outErr error) {
	defer utils.CatchAsError(&outErr)

	if !shouldConvertImage(path) {
		return "", "", fmt.Errorf("不支持的文件类型：%s", path)
	}

	ext := filepath.Ext(path)

	// 格式转换后的新的文件名。
	newPath := strings.TrimSuffix(path, ext) + `.avif`

	tmpOutputFile := utils.Must1(os.CreateTemp("", ""))
	tmpOutputFile.Close()

	log.Println(`转换成 AVIF 格式：`, input, "->", tmpOutputFile.Name())

	var args []string
	if keepTags {
		args = []string{input, tmpOutputFile.Name()}
	} else {
		args = []string{`--ignore-exif`, `--ignore-xmp`, `--ignore-icc`, input, tmpOutputFile.Name()}
	}

	cmd := exec.CommandContext(ctx, `avifenc`, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	utils.Must(cmd.Run())

	return newPath, tmpOutputFile.Name(), nil
}

// copyTags 从源文件复制元数据到目标文件。
func CopyTags(src, dst string) (string, error) {
	cmd := exec.Command("exiftool", "-overwrite_original", "-tagsFromFile", src, dst)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// 有些图像可能把 GPS 信息嵌在 XMP、ICC_Profile、MakerNotes 里（尤其是手机拍摄图）。
// 直接 -gps:all= 无法删除这些信息。
//
//   - [How do you delete GPS information in XMP metadata?](https://exiftool.org/forum/index.php?topic=4686.0)
//   - [Deleting Exif GPS Tags](https://exiftool.org/forum/index.php?topic=7868.0)
func DropGPSTags(file string) (string, error) {
	cmd := exec.Command("exiftool", "-overwrite_original", "-gps*=", file)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// func DropMakerNotes(file string) error {
// 	log.Println(`remove MakerNotes from`, file)
// 	cmd := exec.Command("exiftool", "-overwrite_original", "-MakerNotes=", file)
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	return cmd.Run()
// }
