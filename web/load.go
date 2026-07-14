package main

import (
	"fmt"
	"html/template"
	"led/lights"
	"log"
	"net/http"
	"strconv"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

type LightChannel struct {
	ID      string
	Options *ws2811.ChannelOption
}

type FolderData struct {
	ID    string
	Title string
	Open  bool
	Grids []GridData
}

type GridData struct {
	ID    string
	Hide  bool
	Cards []CardData
}

type Bounds struct {
	Min  uint
	Max  uint
	Step uint
}

type CardData struct {
	ID     string
	Title  string
	Value  string
	Edit   bool
	Bounds Bounds
	Path   string
}

func loadDynamic(opt *ws2811.Option, templ *template.Template) func(w http.ResponseWriter, r *http.Request) {
	folders := make([]FolderData, 0)
	for channel := range opt.Channels {
		folders = append(folders, buildParameters(channel, opt))
	}
	folders = append(folders, buildChannels(opt)...)
	folders = append(folders, buildOptions(opt))
	return func(w http.ResponseWriter, r *http.Request) {
		var data struct {
			Folders []FolderData
		}
		data.Folders = folders
		if len(folders) < 1 {
			log.Fatalln("No folders")
		}

		w.Header().Add("Cache-Control", "no-cache")
		err := templ.Lookup("dynamic").Execute(w, &data)
		if err != nil {
			log.Println("error is here")
			log.Fatalln(err)
		}
	}
}

func buildOptions(opt *ws2811.Option) FolderData {
	var (
		optionCards = []CardData{
			CardData{
				ID:    "frequency",
				Title: "Frequency",
				Value: strconv.Itoa(opt.Frequency),
			},
			CardData{
				ID:    "dma",
				Title: "DMA Channel",
				Value: strconv.Itoa(opt.DmaNum),
			},
			CardData{
				ID:    "render-wait",
				Title: "Render Wait Time",
				Value: strconv.Itoa(opt.RenderWaitTime),
			},
		}
		optionGrids = []GridData{{ID: "light-options-grid", Cards: optionCards}}
		optionData  = FolderData{
			ID:    "light-options-grid",
			Title: "Driver Settings",
			Open:  false,
			Grids: optionGrids,
		}
	)
	return optionData
}

func buildChannels(opt *ws2811.Option) []FolderData {
	folders := make([]FolderData, 0)
	grids := make([]GridData, 0)
	for chanNum, ch := range opt.Channels {
		chanIndex := strconv.Itoa(chanNum)
		cards := make([]CardData, 0)
		cards = append(cards, CardData{
			ID:    "gpio" + chanIndex,
			Title: "GPIO Pin",
			Value: strconv.Itoa(ch.GpioPin),
		})
		cards = append(cards, CardData{
			ID:    "invert" + chanIndex,
			Title: "Invert",
			Value: strconv.FormatBool(ch.Invert),
		})
		cards = append(cards, CardData{
			ID:    "led-count" + chanIndex,
			Title: "Led Count",
			Value: strconv.Itoa(ch.LedCount),
		})
		cards = append(cards, CardData{
			ID:    "led-type" + chanIndex,
			Title: "Led Type",
			Value: strconv.Itoa(ch.StripeType),
		})
		cards = append(cards, CardData{
			ID:    "led-brightness" + chanIndex,
			Title: "Led Brightness",
			Value: strconv.Itoa(ch.Brightness),
		})
		cards = append(cards, CardData{
			ID:    "led-wshift" + chanIndex,
			Title: "White Shift",
			Value: strconv.Itoa(ch.WShift),
		})
		cards = append(cards, CardData{
			ID:    "led-rshift" + chanIndex,
			Title: "Red Shift",
			Value: strconv.Itoa(ch.RShift),
		})
		cards = append(cards, CardData{
			ID:    "led-gshift" + chanIndex,
			Title: "Green Shift",
			Value: strconv.Itoa(ch.GShift),
		})
		cards = append(cards, CardData{
			ID:    "led-bshift" + chanIndex,
			Title: "Blue Shift",
			Value: strconv.Itoa(ch.BShift),
		})
		cards = append(cards, CardData{
			ID:    "led-gamma" + chanIndex,
			Title: "Gamma Table",
			Value: strconv.FormatBool(len(ch.Gamma) > 0),
		})

		grids = append(grids, GridData{ID: "led-channel" + chanIndex, Cards: cards})
		folders = append(folders, FolderData{
			ID:    "led-channel" + chanIndex,
			Title: "Channel " + chanIndex + " Settings",
			Open:  false,
			Grids: grids,
		})
	}
	return folders
}

func buildParameters(channel int, opt *ws2811.Option) FolderData {
	var (
		index  = strconv.Itoa(channel)
		folder = FolderData{
			ID:    "parm-channel" + index,
			Title: "Channel " + index + " Parameters",
			Open:  false,
		}
		channelLength = len(opt.Channels)
	)

	if channel >= channelLength {
		lights.LogError("buildParameters",
			fmt.Errorf("channel (%v) exceeds maximum (%v)",
				channel, channelLength-1))
		return folder
	}

	var (
		ledCount = opt.Channels[channel].LedCount

		grids = make([]GridData, 0)
		cards = []CardData{
			CardData{
				ID:    "parm-channel" + index,
				Title: "Channel",
				Value: index,
			},
			CardData{
				ID:    "parm-brightness" + index,
				Title: "Brightness",
				Value: strconv.Itoa(100),
				Edit:  true,
				Bounds: Bounds{
					Min:  1,
					Max:  100,
					Step: 1,
				},
			},
			CardData{
				ID:    "parm-rows" + index,
				Title: "Rows",
				Value: strconv.Itoa(1),
				Edit:  true,
				Bounds: Bounds{
					Min:  1,
					Max:  1,
					Step: 1,
				},
			},
			CardData{
				ID:    "parm-columns" + index,
				Title: "Columns",
				Value: strconv.Itoa(ledCount),
				Edit:  true,
				Bounds: Bounds{
					Min:  1,
					Max:  uint(ledCount),
					Step: 1,
				},
			},
		}
	)
	grids = append(grids, GridData{ID: "parm-channel" + index, Cards: cards})
	folder.Grids = grids
	return folder
}
