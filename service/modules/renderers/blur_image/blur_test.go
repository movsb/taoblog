package blur_image

import (
	"fmt"
	"image/png"
	"os"
	"testing"

	"github.com/buckket/go-blurhash"
)

func TestEncode(t *testing.T) {
	t.SkipNow()
	// Generate the BlurHash for a given image
	imageFile, _ := os.Open("1.png")
	loadedImage, err := png.Decode(imageFile)
	str, _ := blurhash.Encode(8, 8, loadedImage)
	if err != nil {
		// Handle errors
	}
	fmt.Printf("Hash: %s\n", str)

	// Generate an image for a given BlurHash
	// Width will be 300px and Height will be 500px
	// Punch specifies the contrasts and defaults to 1
	img, err := blurhash.Decode(str, 32, 32, 1)
	if err != nil {
		// Handle errors
	}
	f, _ := os.Create("test_blur.png")
	_ = png.Encode(f, img)
}
