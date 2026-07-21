package main

import (
	"html/template"
	"image/color"
	"log"
	"os"
	"strings"
	"testing"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

func TestFormatColor(t *testing.T) {
	f := func(c color.Color, cmp string) {
		s := FormatColor(c)
		if s != cmp {
			t.Fail()
		}
		t.Log(s)
	}

	f(color.RGBA{R: 255, B: 255}, "#ff00ff")
	f(color.RGBA{R: 127}, "#7f0000")
	f(color.RGBA{B: 31}, "#00001f")
	f(color.RGBA{G: 253}, "#00fd00")
}

var test_opt = ws2811.Option{
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

func TestLoad(t *testing.T) {
	const pattern = "www/*.html"
	templ, err := template.New("").ParseGlob(pattern)
	if err != nil {
		log.Fatalln("Parse global templates", pattern, err)
	}

	folder := buildCustomColors()
	err = templ.Lookup("folder").Execute(os.Stdout, &folder)
	if err != nil {
		log.Fatalln("Parse global templates", pattern, err)
	}
}

func buildCustomColors() FolderData {
	grids := make([]GridData, 0)
	cards := make([]CardData, 0)
	// for key := range lights.ColorTable {
	key := "custom1"
	cards = append(cards, CardData{
		ID:     "color-" + key,
		Title:  strings.ToTitle(key[0:1]) + key[1:],
		Value:  "#ff0000",
		Traits: NewTraits(TraitLocal, TraitAction, TraitColor, TraitLabel),
	})
	// }
	grids = append(grids, GridData{ID: "led-colors", Cards: cards, Hide: false})
	return FolderData{
		ID:    "led-colors",
		Title: "Colors",
		Open:  true,
		Grids: grids,
	}
}
