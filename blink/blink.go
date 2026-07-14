package main

import (
	"flag"
	"image/color"
	"led/lights"
	"log"
	"os"
	"os/signal"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

const (
	brightness = 15
	ledCounts  = 60
	sleepTime  = 0
	gpioPin    = 18
)

var opt = ws2811.Option{
	Frequency: ws2811.TargetFreq,
	DmaNum:    ws2811.DefaultDmaNum,
	Channels: []ws2811.ChannelOption{
		{
			GpioPin:    gpioPin,
			LedCount:   ledCounts,
			Brightness: brightness,
			StripeType: ws2811.WS2812Strip,
			Invert:     false,
			// Gamma:      gamma8,
		},
	},
}

func main() {
	brightness := uint(15)
	duration := uint(1000)
	phase := uint(40)
	pb := flag.Uint("b", brightness, "set the brightness")
	pd := flag.Uint("d", duration, "set the duration in milliseconds")
	ph := flag.Uint("p", phase, "set the phase time in milliseconds")
	flag.Parse()
	if *pb <= 100 {
		brightness = *pb
	}
	if *pd <= 100000 {
		duration = *pd
	}
	if *ph <= 1000 {
		phase = *ph
	}

	log.Printf("Brightness=%v", brightness)
	log.Printf("Duration=%v", duration)
	log.Printf("Phase=%v", phase)
	pi := lights.NewPiLED(&opt)
	if err := pi.Init(); err != nil {
		return
	}
	defer func() {
		pi.Clear(0)
		pi.Finish()
		log.Println("Finished, done.")
	}()

	go pi.ShowBlink(0,
		[]color.Color{
			color.RGBA{R: 255},
			color.RGBA{G: 255},
			color.RGBA{B: 255},
		},
		uint32(brightness), uint32(duration), uint32(phase))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs
	log.Println("Signal")
	pi.Stop()
}
