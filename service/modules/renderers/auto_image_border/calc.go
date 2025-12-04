package auto_image_border

import (
	"image"
	"image/color"
	"io"
	"iter"
	"log"
	"math"

	_ "github.com/gen2brain/avif"
	_ "golang.org/x/image/webp"
)

// ratio: 相对亮度比值为多少被视为高对比度。
// https://www.w3.org/TR/WCAG21/#contrast-minimum
// 返回值：表示多少个边缘点符合高对比度（范围：[0,1]）。
func BorderContrastRatio(f io.Reader, r, g, b byte, ratio float64) float32 {
	img, _, err := image.Decode(f)
	if err != nil {
		log.Println(err)
		return 0
	}

	l1 := relativeLuminance(r, g, b)

	points := 0
	good := 0

	for color := range nextPoint(img) {
		points++

		l2 := relativeLuminance(unpremultiply(color.RGBA()))
		rr := contrastRatio(l1, l2)
		// log.Println(`rr:`, rr)
		if rr > ratio {
			good++
		}
	}

	return float32(good) / float32(points)
}

func nextPoint(img image.Image) iter.Seq[color.Color] {
	bounds := img.Bounds()

	// 因为 bounds 不要求在原点，这里把它们移动到原点。
	max := bounds.Max.Sub(bounds.Min)

	return func(yield func(color.Color) bool) {
		for y := 0; y < max.Y; y++ {
			if y == 0 || y == max.Y-1 {
				for x := 0; x < max.X; x++ {
					// log.Println(x, y)
					if !yield(img.At(x, y)) {
						return
					}
				}
			} else {
				// log.Println(0, y)
				if !yield(img.At(0, y)) {
					return
				}
				// log.Println(max.X-1, y)
				if !yield(img.At(max.X-1, y)) {
					return
				}
			}
		}
	}
}

func unpremultiply(r16, g16, b16, a16 uint32) (r, g, b byte) {
	if a16 == 0 {
		return 0, 0, 0
	}

	// a8 = uint8(a16 / 257)

	r = uint8(r16 * 255 / a16)
	g = uint8(g16 * 255 / a16)
	b = uint8(b16 * 255 / a16)
	return
}

func contrastRatio(l1, l2 float64) float64 {
	if l2 > l1 {
		l1, l2 = l2, l1
	}

	return (l1 + 0.05) / (l2 + 0.05)
}

// https://www.w3.org/TR/WCAG21/#dfn-relative-luminance
func relativeLuminance(r, g, b byte) float64 {
	norm := func(c byte) float64 {
		return float64(c) / 255
	}
	rc := func(c float64) float64 {
		if c <= 0.04045 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*rc(norm(r)) + 0.7152*rc(norm(g)) + 0.0722*rc(norm(b))
}
