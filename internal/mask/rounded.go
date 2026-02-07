package mask

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

func GenerateRoundedMask(path string, width, height, radius int,fill color.NRGBA) error {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			alpha := fill.A

			// Top-left corner
			if x < radius && y < radius {
				if !insideCircle(x, y, radius, radius, radius) {
					alpha = 0
				}
			}

			// top-right
			if x >= width-radius && y < radius {
				if !insideCircle(x, y, width-radius-1, radius, radius) {
					alpha = 0
				}
			}

			// bottom-left
			if x < radius && y >= height-radius {
				if !insideCircle(x, y, radius, height-radius-1, radius) {
					alpha = 0
				}
			}

			// bottom-right
			if x >= width-radius && y >= height-radius {
				if !insideCircle(x, y, width-radius-1, height-radius-1, radius) {
					alpha = 0
				}
			}

			img.SetNRGBA(x, y, color.NRGBA{R: fill.R, G: fill.G, B: fill.B, A: alpha})
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

func insideCircle(x, y, cx, cy, r int) bool {
	dx := float64(x - cx)
	dy := float64(y - cy)
	return math.Hypot(dx, dy) <= float64(r)
}