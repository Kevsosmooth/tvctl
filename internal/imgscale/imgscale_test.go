package imgscale

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

func encode(t *testing.T, w, h int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestFitWidthShrinks(t *testing.T) {
	out, err := FitWidth(encode(t, 3840, 2160), 1280)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Width != 1280 || cfg.Height != 720 {
		t.Errorf("size = %dx%d", cfg.Width, cfg.Height)
	}
}

func TestFitWidthLeavesSmallImages(t *testing.T) {
	in := encode(t, 800, 450)
	out, err := FitWidth(in, 1280)
	if err != nil || !bytes.Equal(in, out) {
		t.Errorf("small image changed, err %v", err)
	}
	if out, _ := FitWidth(in, 0); !bytes.Equal(in, out) {
		t.Error("width 0 should keep the image")
	}
}

func TestFitWidthRejectsGarbage(t *testing.T) {
	if _, err := FitWidth([]byte("not a png"), 100); err == nil {
		t.Error("expected decode error")
	}
}
