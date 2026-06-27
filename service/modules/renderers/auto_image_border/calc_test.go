package auto_image_border

import (
	"image/color"
	"log"
	"os"
	"path"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
)

func TestMinimal(t *testing.T) {
	for i, tc := range []struct {
		file       string
		background color.NRGBA
		good       bool
	}{
		{
			file:       `black.png`,
			background: color.NRGBA{255, 255, 255, 0},
			good:       true,
		},
		{
			file:       `black.png`,
			background: color.NRGBA{0, 0, 0, 0},
			good:       true,
		},
		{
			file:       `white.png`,
			background: color.NRGBA{255, 255, 255, 0},
			good:       false,
		},
		{
			file:       `white.png`,
			background: color.NRGBA{0, 0, 0, 0},
			good:       true,
		},
		{
			file:       `white.webp`,
			background: color.NRGBA{0, 0, 0, 0},
			good:       true,
		},
		{
			file:       `black.avif`,
			background: color.NRGBA{255, 255, 255, 0},
			good:       true,
		},
	} {
		f := utils.Must1(os.Open(path.Join(`testdata`, tc.file)))
		defer f.Close()

		value := BorderContrastRatio(f, tc.background.R, tc.background.G, tc.background.B, 1)
		if value > 0.75 != tc.good {
			t.Fatal(`not good:`, i+1, tc.file)
		}
	}
}

func TestFile(t *testing.T) {
	t.SkipNow()
	f := utils.Must1(os.Open(path.Join(`testdata`, `IMG_1303.avif`)))
	defer f.Close()
	value := BorderContrastRatio(f, 255, 255, 255, 1)
	log.Println(value)
}

/*
按 W3C/WCAG 公式，对比度是：

```text
(L1 + 0.05) / (L2 + 0.05)
```

其中 `L1` 是较亮颜色的相对亮度，`L2` 是较暗颜色的相对亮度。W3C 也说明计算值不应四舍五入后再判断阈值。([w3.org](https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html))

计算结果：

```text
#00D0F0  相对亮度: 0.5140302099
#CCC     相对亮度: 0.6038273389

contrast = 1.1592062399 : 1
```

所以两者的 WCAG contrast 约为 **1.16:1**。

这明显低于 WCAG AA 普通文本要求的 `4.5:1`，也低于大文本要求的 `3:1`。([w3.org](https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html))
*/
func TestRatio(t *testing.T) {
	l1 := relativeLuminance(0x00, 0xD0, 0xF0)
	l2 := relativeLuminance(0xCC, 0xCC, 0xCC)
	c := contrastRatio(l1, l2)
	t.Log(c)
}
