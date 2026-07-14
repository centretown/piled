package lights

import "image/color"

const MaxColor = uint8(255)
const MaxBright = uint32(100)

func ToRGBA(c uint32) color.RGBA {
	return color.RGBA{
		R: uint8((c & 0xff0000) >> 16),
		G: uint8((c & 0xff00) >> 8),
		B: uint8(c & 0xff),
		A: 0xff}
}

func FromRGB(c color.RGBA) uint32 {
	return uint32(c.R)<<16 |
		uint32(c.G)<<8 |
		uint32(c.B)
}

func FromRGBrightness(r, g, b uint32, brightness uint32) uint32 {
	r, g, b = r*brightness/MaxBright, g*brightness/MaxBright, b*brightness/MaxBright
	return r<<16 | g<<8 | b
}

func FromColor(c color.Color) uint32 {
	r, g, b, _ := c.RGBA()
	return uint32(uint8(r>>8))<<16 | uint32(uint8(g>>8))<<8 | uint32(uint8(b>>8))
}

func FromColorBrightness(c color.Color, brightness uint32) uint32 {
	r, g, b, _ := c.RGBA()
	r, g, b = r>>8, g>>8, b>>8
	r, g, b = (r*brightness)/MaxBright, (g*brightness)/MaxBright, (b*brightness)/MaxBright
	return r<<16 | g<<8 | b
}

func FromBackGround(c color.Color, brightness uint32, background uint32) uint32 {
	return FromColorBrightness(c, brightness) | background
}

func FromColorThreshold(r, g, b uint32, threshHold uint32) uint32 {
	r, g, b = r>>8, g>>8, (b >> 8)
	return (r<<16 | g<<8 | b) | threshHold
}

func AlphaMasks(alpha uint8) (front, back uint32) {
	front = uint32(alpha)<<16 |
		uint32(alpha)<<8 |
		uint32(alpha)
	remainder := 255 - alpha
	back = uint32(remainder)<<16 |
		uint32(remainder)<<8 |
		uint32(remainder)
	return
}
