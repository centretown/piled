package lights

import "image/color"

type ColorItem struct {
	c color.Color
}

const colorMAX = 255

var (
	ColorTable = map[string]color.Color{
		"red":     color.RGBA{R: colorMAX},
		"green":   color.RGBA{G: colorMAX},
		"blue":    color.RGBA{B: colorMAX},
		"yellow":  color.RGBA{R: colorMAX, G: colorMAX},
		"cyan":    color.RGBA{B: colorMAX, G: colorMAX},
		"magenta": color.RGBA{R: colorMAX, B: colorMAX},
		"white":   color.RGBA{R: colorMAX, G: colorMAX, B: colorMAX},
		"black":   color.RGBA{},
	}
)
