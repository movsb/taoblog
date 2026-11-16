package open_graph

import (
	"image"
	"math"
)

const backgroundBlurRadius = 4

func filterBackground(img *image.RGBA) {
	// 沿用之前的对比度 -0.6，RGB 向中灰靠拢，alpha 保持原值。
	var contrast [256]uint8
	for i := range contrast {
		contrast[i] = uint8(((float64(i)/255-0.5)*0.4 + 0.5) * 255)
	}
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	for y := range height {
		row := img.Pix[y*img.Stride : y*img.Stride+width*4]
		for x := 0; x < len(row); x += 4 {
			row[x] = contrast[row[x]]
			row[x+1] = contrast[row[x+1]]
			row[x+2] = contrast[row[x+2]]
		}
	}

	scratch := image.NewRGBA(img.Bounds())
	blurBackgroundPass(scratch, img, true)
	blurBackgroundPass(img, scratch, false)
}

func blurBackgroundPass(dst, src *image.RGBA, horizontal bool) {
	// 半径 4 的可分离高斯核，与之前的背景模糊权重一致。
	var weights [2*backgroundBlurRadius + 1]float64
	var total float64
	for i := range weights {
		d := float64(i - backgroundBlurRadius)
		weights[i] = math.Exp(-d * d / (4 * backgroundBlurRadius))
		total += weights[i]
	}
	for i := range weights {
		weights[i] /= total
	}

	width, height := src.Bounds().Dx(), src.Bounds().Dy()
	length, step := height, src.Stride
	if horizontal {
		length, step = width, 4
	}
	for y := range height {
		for x := range width {
			position := y
			if horizontal {
				position = x
			}
			srcOffset := y*src.Stride + x*4
			var rgba [4]float64
			for i, weight := range weights {
				p := min(max(position+i-backgroundBlurRadius, 0), length-1)
				offset := srcOffset + (p-position)*step
				pixel := src.Pix[offset : offset+4]
				for c, value := range pixel {
					rgba[c] += float64(value) * weight
				}
			}
			dstOffset := y*dst.Stride + x*4
			for c, value := range rgba {
				dst.Pix[dstOffset+c] = uint8(value)
			}
		}
	}
}
