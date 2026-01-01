package open_graph

import (
	"bytes"
	"context"
	"crypto/md5"
	"embed"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/anthonynsimon/bild/adjust"
	"github.com/anthonynsimon/bild/blur"
	"github.com/anthonynsimon/bild/transform"
	"github.com/fogleman/gg"
	"github.com/gen2brain/avif"
	"github.com/golang/freetype/truetype"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/modules/version"
	"github.com/phuslu/lru"
	"golang.org/x/image/font"
)

const (
	padding          = 80   // 整体内边距
	fullWidth        = 1200 // Open Graph 总尺寸
	fullHeight       = 630
	avatarRadius     = 100
	lineHeight       = 1.3
	chineseFontPath  = `经典粗宋简/经典粗宋简.ttf`
	fontSizeSiteName = 80
	fontSizeTitle    = 55
	fontSizeExcerpt  = 40
)

var debug = false

// 根据指定的参数生成 OpenGraph 图像（PNG格式）。
//
// excerpt（摘要）会被自动裁剪以适合高度。
func GenerateImage(siteName string, title, excerpt string, avatar, background io.Reader) (_ []byte, outErr error) {
	// defer utils.CatchAsError(&outErr)

	// 初始化字体。
	if !loadAllFonts() {
		return nil, fmt.Errorf(`failed to load font`)
	}

	dc := gg.NewContext(fullWidth, fullHeight)

	// 背景图片打底。
	// 背景图片默认每篇文章总是不一样，所以本函数内不缓存，需要外部缓存。
	drawBackground(dc, background)

	y := padding
	w := 0
	h := 0

	// 绘制站点名。
	w = fullWidth - padding*2 - avatarRadius*2 - padding
	y += drawString(dc, siteName, padding, y, 0xFFFFFFFF, fontSiteName, w, 0)

	y += 25
	// 绘制标题
	w = fullWidth - padding*2 - avatarRadius*2 - padding
	y += drawString(dc, title, padding, y, 0xFFF0F0F0, fontTitle, w, 0)

	y += 35

	// 绘制摘要。
	w = fullWidth - padding*2
	h = fullHeight - y - padding
	y += drawString(dc, excerpt, padding, y, 0xFFF0F0F0, fontExcerpt, w, h)

	// 绘制头像
	// 奇怪，不知道为什么要放后面才正确，否则文字无法显示。
	drawAvatar(dc, avatar)

	// 导出
	output := bytes.NewBuffer(nil)
	utils.Must(png.Encode(output, dc.Image()))

	return output.Bytes(), nil
}

func drawString(dc *gg.Context, s string, x, y int, color uint32, face font.Face, maxWidth, maxHeight int) int {
	dc.SetRGB255(int(color>>0&0xFF), int(color>>8&0xFF), int(color>>16&0xFF))
	dc.SetFontFace(face)
	s, n := breakString(dc, s, maxWidth, maxHeight)
	dc.DrawStringWrapped(s, float64(x), float64(y), 0, 0, float64(maxWidth), lineHeight, gg.AlignLeft)

	height := n * int(dc.FontHeight()*lineHeight)

	if debug {
		dc.SetRGB(.5, 0, 0)
		dc.DrawRectangle(float64(x), float64(y), float64(maxWidth), float64(height))
		dc.Stroke()
	}

	return height
}

func drawBackground(dc *gg.Context, background io.Reader) {
	// 默认黑色背景
	dc.SetRGB(0, 0, 0)
	dc.Clear()

	if background != nil {
		img := utils.Must1(decodeImage(background))

		scaled := transform.Resize(img, fullWidth, fullHeight, transform.Lanczos)
		dimmed := adjust.Contrast(scaled, -.6)
		blurred := blur.Gaussian(dimmed, 4)

		dc.DrawImage(blurred, 0, 0)
	}

	// 调试：内边距
	if debug {
		dc.SetRGB(.5, 0, 0)
		dc.DrawRectangle(padding, padding, fullWidth-padding*2, fullHeight-padding*2)
		dc.Stroke()
	}
}

func drawAvatar(dc *gg.Context, avatar io.Reader) {
	if avatar != nil {
		img := utils.Must1(decodeImageWithCache(avatar))
		scaled := transform.Resize(img, avatarRadius*2, avatarRadius*2, transform.Lanczos)
		dc.Push()
		dc.DrawCircle(fullWidth-padding-avatarRadius, padding+avatarRadius, avatarRadius)
		dc.Clip()
		dc.DrawImageAnchored(scaled, fullWidth-padding-avatarRadius, padding+avatarRadius, 0.5, 0.5)
		dc.Pop()
	}
}

// gg 只能按空格折行，否则会超出绘制区域。
// 这里手动加换行符以达到效果。
// 写得乱七八糟。
func breakString(dc *gg.Context, s string, maxWidth, maxHeight int) (string, int) {
	lines := []string{}
	runes := []rune(s)
	line := []rune{}
	height := 0
	shouldFix := true

	fixHeight := func() bool {
		height += int(dc.FontHeight() * lineHeight)
		if maxHeight > 0 && float64(height) > float64(maxHeight) && len(lines) > 0 {
			lines = lines[:len(lines)-1]
			if len(lines) <= 0 {
				return true
			}
			last := []rune(lines[len(lines)-1])
			last = last[:max(0, len(last)-5)] // 实际上减3就够了
			last = append(last, '.', '.', '.')
			lines[len(lines)-1] = string(last)
			line = line[:0]
			return true
		}
		return false
	}

	for i := 0; i < len(runes); i++ {
		line = append(line, runes[i])
		w, _ := dc.MeasureString(string(line))
		if runes[i] == '\n' {
			lines = append(lines, string(line[:len(line)-1]))
			height += int(dc.FontHeight() * lineHeight)
			if maxHeight > 0 && float64(height) > float64(maxHeight) {
				break
			}
			line = line[:0]
			continue
		}
		// maxWidth 不要太小，否则会 panic
		if w > float64(maxWidth) {
			lines = append(lines, string(line[:len(line)-1]))
			line = line[:0]
			if fixHeight() {
				shouldFix = false
				break
			}
			i--
			continue
		}
	}

	if len(line) > 0 {
		lines = append(lines, string(line))
	}

	if shouldFix {
		fixHeight()
	}

	return strings.Join(lines, "\n"), len(lines)
}

type _CheckSum [md5.Size]byte

var imageCache = lru.NewTTLCache[_CheckSum, image.Image](8)

func decodeImageWithCache(r io.Reader) (image.Image, error) {
	dup := utils.MemDupReader(r)

	md5sum := md5.New()
	io.Copy(md5sum, dup())
	sum := md5sum.Sum(nil)
	sum2 := [md5.Size]byte{}
	copy(sum2[:], sum)

	img, err, _ := imageCache.GetOrLoad(context.Background(), sum2, func(ctx context.Context, cs _CheckSum) (image.Image, time.Duration, error) {
		img, err := decodeImage(dup())
		return img, time.Hour, err
	})

	return img, err
}

func decodeImage(r io.Reader) (image.Image, error) {
	dup := utils.MemDupReader(r)
	img, _, err1 := image.Decode(dup())
	if err1 != nil {
		var err2 error
		img, err2 = avif.Decode(dup())
		if err2 != nil {
			return nil, errors.Join(err1, err2)
		}
	}
	return img, nil
}

var (
	onceLoadFonts                        sync.Once
	fontSiteName, fontTitle, fontExcerpt font.Face
)

func loadAllFonts() bool {
	onceLoadFonts.Do(func() {
		fsys := utils.IIF[fs.FS](version.DevMode(), localChineseFontDir, embedChineseFontDir)
		fontBytes := utils.Must1(fs.ReadFile(fsys, chineseFontPath))
		font := utils.Must1(truetype.Parse(fontBytes))

		fontSiteName = truetype.NewFace(font, &truetype.Options{Size: fontSizeSiteName})
		fontTitle = truetype.NewFace(font, &truetype.Options{Size: fontSizeTitle})
		fontExcerpt = truetype.NewFace(font, &truetype.Options{Size: fontSizeExcerpt})
	})

	return fontSiteName != nil && fontTitle != nil && fontExcerpt != nil
}

//go:embed 经典粗宋简
var embedChineseFontDir embed.FS
var localChineseFontDir = os.DirFS(dir.SourceAbsoluteDir().Join())

//go:generate bash -c "! [ -d 经典粗宋简 ] && curl -L -o 经典粗宋简.tar https://github.com/movsb/taoblog/raw/refs/heads/assets/assets/%E7%BB%8F%E5%85%B8%E7%B2%97%E5%AE%8B%E7%AE%80.tar && tar xvf 经典粗宋简.tar"
