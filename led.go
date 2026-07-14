package main

import (
	"image/color"
	"led/lights"
	"log"
	"os"
	"os/signal"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

const (
	brightness = 100
	ledCounts  = 60
	sleepTime  = 0
)

func main() {
	opt := ws2811.DefaultOptions
	opt.Channels[0].Brightness = brightness
	opt.Channels[0].LedCount = ledCounts
	piled := lights.NewPiLED(&opt)
	if err := piled.Init(); err != nil {
		return
	}

	defer func() {
		piled.Clear(0)
		piled.Finish()
		log.Println("Finished, done.")
	}()

	sigs := make(chan os.Signal, 1)
	done := make(chan int)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		log.Println("Signal")
		piled.Stop()
		done <- 1
	}()

	piled.LogStatus()
	log.Println("ShowRGBA")
	piled.ShowRGB(0, []color.RGBA{
		{B: 71},
		{B: 71},
		{B: 71},
		{B: 71},
		{B: 71},
		// {B: 255},
	})

	// piled.ShowJPEG(0, "waves.jpg")
	<-done
}
