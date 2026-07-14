package lights

import (
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"time"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

type wsEngine interface {
	Init() error
	Render() error
	Wait() error
	Fini()
	Leds(channel int) []uint32
}

type PiLED struct {
	engine  wsEngine
	opt     *ws2811.Option
	stop    bool
	running bool
}

func NewPiLED(opt *ws2811.Option) *PiLED {
	return &PiLED{
		opt: opt,
	}
}

func (pi *PiLED) Init() error {
	engine, err := ws2811.MakeWS2811(pi.opt)
	if err != nil {
		return LogError("Init", err)
	}
	pi.engine = engine
	return LogError("Init", pi.engine.Init())
}

func (pi *PiLED) Finish() { pi.engine.Fini() }

func (pi *PiLED) ChannelCount() int { return len(pi.opt.Channels) }

func (pi *PiLED) Stop() { pi.stop = true }
func (pi *PiLED) StopRun() {
	pi.running = false
}
func (pi *PiLED) StartRun() {
	pi.stop = true
	for pi.running {
		time.Sleep(time.Millisecond * 10)
	}
	pi.running = true
	pi.stop = false
}

func (pi *PiLED) Clear(channel int) error {
	size := len(pi.engine.Leds(channel))
	for i := 0; i < size; i++ {
		pi.engine.Leds(channel)[i] = 0
	}
	return pi.engine.Render()
}

func (pi *PiLED) ShowBytes(channel int, buffer []uint32) error {
	next := 0
	last := len(buffer)
	if last == 0 {
		return LogNoBytes("ShowBytes")
	}

	size := len(pi.engine.Leds(channel))
	for i := 0; !pi.stop && i < size; i++ {
		pi.engine.Leds(channel)[i] = buffer[next]
		next++
		if next >= last {
			next = 0
		}
	}
	return LogError("ShowBytes", pi.engine.Render())
}

func (pi *PiLED) ShowRGB(channel int, buffer []color.RGBA) error {
	return pi.showRGBA(channel, buffer, FromRGB)
}

func (pi *PiLED) showRGBA(channel int, buffer []color.RGBA,
	from func(color.RGBA) uint32) error {

	next := 0
	last := len(buffer)
	if last == 0 {
		return LogNoBytes("ShowRGBA")
	}

	size := len(pi.engine.Leds(channel))
	for i := 0; !pi.stop && i < size; i++ {
		pi.engine.Leds(0)[i] = from(buffer[next])

		if next++; next >= last {
			next = 0
		}
	}
	err := pi.engine.Render()
	return LogError("ShowRGBA", err)
}

func (pi *PiLED) LogStatus() {
	log.Println("Frequency: ", pi.opt.Frequency)
	log.Println("DMA      : ", pi.opt.DmaNum)
	for _, ch := range pi.opt.Channels {
		log.Println("Leds        : ", ch.LedCount)
		log.Println("GPIO Pin    : ", ch.GpioPin)
		log.Println("Brightness  : ", ch.Brightness)
		log.Println("Invert      : ", ch.Invert)
		log.Println("Gamma       : ", len(ch.Gamma))
	}
}

var defColor color.Color = color.RGBA{R: 15, G: 15, B: 15, A: 255}

var alphaMasks = []uint8{0, 31, 63, 95, 127, 159, 191, 223, 255}
var alphaMaskSize = time.Duration(len(alphaMasks))

func (pi *PiLED) ShowBlink(channel int, colors []color.Color, brightness uint32, on_ms, phase uint32) {
	if len(colors) < 1 {
		LogNoColors("ShowBlink")
		return
	}
	frames := make([]Frame, len(colors))
	frameSize := int(pi.opt.Channels[channel].LedCount)
	for i, color := range colors {
		frame := NewFrame(frameSize)
		frame.Fill(FromColorBrightness(color, brightness))
		frames[i] = frame
	}

	frameCount := len(frames)
	prevFrame := NewFrame(frameSize).Clear()
	var drawFrame Frame

	for i := 0; !pi.stop; i++ {
		if i >= frameCount {
			i = 0
		}
		frame := frames[i]
		if !pi.stop {
			for alpha := range uint8(255) {
				drawFrame = prevFrame.Merge(frame, alpha)
				pi.ShowBytes(channel, drawFrame)
				time.Sleep(time.Millisecond * time.Duration(phase))
				prevFrame = drawFrame
				if pi.stop {
					break
				}
			}
			if !pi.stop {
				time.Sleep(time.Millisecond * time.Duration(on_ms))
			}
		}

		// if !pi.stop {
		// 	pi.Clear(channel)
		// 	time.Sleep(time.Millisecond * time.Duration(off_ms) / alphaMaskSize)
		// }
	}
}

func (pi *PiLED) ShowFile(channel int, fileName string, brightness uint32) {
	reader, err := os.Open(fileName)
	if err != nil {
		LogError("ShowFile", err)
		return
	}

	defer reader.Close()
	img, format, err := image.Decode(reader)
	if err != nil {
		LogError("ShowFile", err)
		return
	}

	var (
		rect = img.Bounds()
		pal  = make([]uint32, rect.Max.Y)
	)

	log.Printf("%v X %v format: %v", rect.Max.X, rect.Max.Y, format)

	for y := range rect.Max.Y {
		for x := range rect.Max.X {
			if pi.stop {
				return
			}
			// r, g, b, _ = img.At(x, y).RGBA()
			i := rect.Max.X - x - 1
			pal[i] = FromBackGround(img.At(x, y), brightness, 0x03_03_03)
		}

		if !pi.stop {
			pi.ShowBytes(0, pal)
			time.Sleep(time.Millisecond * 500)
		}
	}
}
