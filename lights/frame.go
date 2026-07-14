package lights

import "image/color"

type Frame []uint32

func NewFrame(size int) Frame {
	return make(Frame, size)
}

func (fr Frame) Clear() Frame {
	for i := range fr {
		fr[i] = 0
	}
	return fr
}

func (fr Frame) Copy(cp Frame) Frame {
	copy(fr, cp)
	// for i := range cp {
	// 	fr[i] = cp[i]
	// }
	return fr
}

func (fr Frame) Fill(uc uint32) Frame {
	for i := range fr {
		fr[i] = uc
	}
	return fr
}

func (fr Frame) FillRGB(c color.RGBA) Frame {
	return fr.Fill(FromRGB(c))
}

func (fr Frame) FillColor(c color.Color) Frame {
	return fr.Fill(FromColor(c))
}

func (fr Frame) Merge(fg Frame, alpha uint8) Frame {
	if len(fr) != len(fg) {
		LogLengthMisMatch("Merge", len(fr), len(fg))
		return fr
	}
	result := NewFrame(len(fr))
	front, back := AlphaMasks(alpha)
	for i := range fr {
		result[i] = (fr[i] & back) | (fg[i] & front)
	}
	return result
}
