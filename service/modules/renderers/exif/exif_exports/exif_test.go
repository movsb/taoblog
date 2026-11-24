package exif_exports

import (
	"os"
	"testing"
)

func TestExtractAVIFWithoutExternalExifTool(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	f, err := os.Open("../testdata/exif.avif")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	m, err := Extract(f)
	if err != nil {
		t.Fatal(err)
	}
	if m.Make != "DJI" || m.Model != "FC3582" || m.MimeType != "image/avif" || m.ImageSize != "2x2" {
		t.Fatalf("unexpected image metadata: %+v", m)
	}
	if m.FNumber != 1.7 || m.FocalLength != "6.7 mm" || m.ExposureTime != "1/3000" || m.ISO != 130 {
		t.Fatalf("unexpected exposure metadata: %+v", m)
	}
	if m.CreateDate != "2024:06:28 14:28:48" || m.GPSPosition == "" || m.GPSAltitude == "" {
		t.Fatalf("missing date or GPS metadata: %+v", m)
	}
}

func TestSwapSizesRotationFormats(t *testing.T) {
	for _, rotation := range []IntOrString{"90", "270", "Rotate 90", "Rotate 270", "Rotate 90 CW", "Rotate 270 CW"} {
		if m := (Metadata{Rotation: rotation}); !m.SwapSizes() {
			t.Errorf("rotation %q should swap dimensions", rotation)
		}
	}
	for _, rotation := range []IntOrString{"", "0", "180", "Rotate 180", "Rotate 180 CW"} {
		if m := (Metadata{Rotation: rotation}); m.SwapSizes() {
			t.Errorf("rotation %q should not swap dimensions", rotation)
		}
	}
}
