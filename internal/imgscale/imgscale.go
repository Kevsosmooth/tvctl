// Package imgscale shrinks PNG screenshots so people and AI agents can view them quickly.
package imgscale

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"golang.org/x/image/draw"
)

// FitWidth returns the PNG scaled down to maxWidth, keeping its aspect ratio. Images
// already narrow enough, or maxWidth 0, come back unchanged.
func FitWidth(data []byte, maxWidth int) ([]byte, error) {
	if maxWidth <= 0 {
		return data, nil
	}
	src, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode screenshot: %w", err)
	}
	b := src.Bounds()
	if b.Dx() <= maxWidth {
		return data, nil
	}
	height := b.Dy() * maxWidth / b.Dx()
	dst := image.NewRGBA(image.Rect(0, 0, maxWidth, height))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, b, draw.Src, nil)
	var out bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := enc.Encode(&out, dst); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
