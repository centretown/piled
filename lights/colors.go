package lights

import "image/color"

type ColorItem struct {
	c color.Color
}

var (
	ColorTable = map[string]color.Color{
		"red":     color.RGBA{R: 255},
		"green":   color.RGBA{G: 255},
		"blue":    color.RGBA{B: 255},
		"yellow":  color.RGBA{R: 127, G: 127},
		"cyan":    color.RGBA{B: 127, G: 127},
		"magenta": color.RGBA{R: 127, B: 127},
		"white":   color.RGBA{R: 85, G: 85, B: 85},
		"black":   color.RGBA{},
	}
)
