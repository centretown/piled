package main

import (
	"bytes"
	"fmt"
	"html/template"
	"image/color"
	"led/lights"
	"led/socket"
	"log"
	"net/http"
	"strconv"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

type LightChannel struct {
	ID      string
	Title   string
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

type Traits map[string]any

const (
	TraitLocal  = "local"
	TraitGlobal = "global"
	TraitAction = "action"
	TraitColor  = "color"
	TraitNumber = "number"
	TraitLabel  = "label"
)

func NewTraits(members ...string) (traits Traits) {
	traits = make(Traits)
	for _, member := range members {
		traits[member] = nil
	}
	return
}

type CardData struct {
	ID     string
	Title  string
	Value  string
	Traits Traits
	Bounds Bounds
	Path   string
}

type StatusData struct {
	Color      string
	Brightness string
}

func (card *CardData) IsGlobal() (ok bool) { return card.HasTrait(TraitGlobal) }
func (card *CardData) IsLocal() (ok bool)  { return card.HasTrait(TraitLocal) }
func (card *CardData) IsAction() (ok bool) { return card.HasTrait(TraitAction) }
func (card *CardData) IsColor() (ok bool)  { return card.HasTrait(TraitColor) }
func (card *CardData) IsNumber() (ok bool) { return card.HasTrait(TraitNumber) }
func (card *CardData) IsLabel() (ok bool)  { return card.HasTrait(TraitLabel) }

func (card *CardData) HasTrait(member string) (ok bool) {
	_, ok = card.Traits[member]
	return
}

var currentStatus = StatusData{Color: "none", Brightness: "none"}

func CurrentStatus() *StatusData {
	return &currentStatus
}

type ValueUpdate struct {
	ID    string
	Value string
}

func SetCurrentStatus(sock *socket.Server, tmpl *template.Template, color, brightness string) {
	currentStatus.Color = color
	currentStatus.Brightness = brightness
	var (
		colorUpdate      = &ValueUpdate{ID: "status-color", Value: color}
		brightnessUpdate = &ValueUpdate{ID: "status-brightness", Value: brightness}
	)
	buf := bytes.NewBufferString("")
	tmpl.Lookup("value_update").Execute(buf, colorUpdate)
	tmpl.Lookup("value_update").Execute(buf, brightnessUpdate)
	sock.Broadcast(buf.String())
}

var (
	statusCards = []CardData{{
		ID:    "status-color",
		Title: "Color",
		Value: currentStatus.Color,
	}, {
		ID:    "status-brightness",
		Title: "Brightness",
		Value: currentStatus.Brightness,
	}}
	statusFolder = &FolderData{
		ID:    "status-header",
		Title: "Status",
		Open:  true,
		Grids: []GridData{
			{ID: "status-header", Cards: statusCards},
		},
	}
)

func loadDynamic(opt *ws2811.Option, templ *template.Template) func(w http.ResponseWriter, r *http.Request) {
	folders := make([]FolderData, 0)
	folders = append(folders, buildColors())
	folders = append(folders, buildChannels(opt)...)
	folders = append(folders, buildOptions(opt))
	return func(w http.ResponseWriter, r *http.Request) {
		var data struct {
			Header  []FolderData
			Folders []FolderData
			Status  *StatusData
		}

		data.Header = []FolderData{*statusFolder, buildPreferences()}
		data.Folders = folders
		data.Status = CurrentStatus()
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
		optionGrids = []GridData{{ID: "light-options-grid", Cards: optionCards, Hide: true}}
		optionData  = FolderData{
			ID:    "light-options-grid",
			Title: "Driver",
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
			ID:    "channel" + chanIndex,
			Title: "Channel",
			Value: chanIndex,
		})
		cards = append(cards, CardData{
			ID:    "led-count" + chanIndex,
			Title: "Led Count",
			Value: strconv.Itoa(ch.LedCount),
		})
		cards = append(cards, CardData{
			ID:    "parm-rows" + chanIndex,
			Title: "Rows",
			Value: "1",
		})
		cards = append(cards, CardData{
			ID:    "parm-columns" + chanIndex,
			Title: "Columns",
			Value: strconv.Itoa(ch.LedCount),
		})
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

		grids = append(grids, GridData{ID: "led-channel" + chanIndex, Cards: cards, Hide: true})
		folders = append(folders, FolderData{
			ID:    "led-channel" + chanIndex,
			Title: "Settings",
			Open:  false,
			Grids: grids,
		})
	}
	return folders
}

func FormatColor(c color.Color) string {
	return fmt.Sprintf("#%06x", lights.FromColor(c))
}

func buildColors() FolderData {
	grids := make([]GridData, 0)
	cards := make([]CardData, 0)
	i := 0
	for _, v := range lights.ColorTable {
		index := strconv.Itoa(i)
		cards = append(cards, CardData{
			ID:     "color-" + index,
			Title:  "Color" + index,
			Value:  FormatColor(v),
			Traits: NewTraits(TraitLocal, TraitAction, TraitColor, TraitLabel),
		})
		i++
	}
	grids = append(grids, GridData{ID: "led-colors", Cards: cards, Hide: true})
	return FolderData{
		ID:    "led-colors",
		Title: "Colors",
		Open:  false,
		Grids: grids,
	}
}

func buildPreferences() FolderData {
	grids := make([]GridData, 0)
	for channel := range opt.Channels {
		index := strconv.Itoa(channel)
		cards := make([]CardData, 0)
		cards = append(cards, CardData{
			ID:    "color-channel",
			Title: "Channel",
			Value: strconv.Itoa(channel),
		})
		cards = append(cards, CardData{
			ID:     "brightness" + strconv.Itoa(channel),
			Title:  "Brightness",
			Value:  strconv.Itoa(100),
			Traits: NewTraits(TraitLocal, TraitNumber),
			Bounds: Bounds{
				Min:  1,
				Max:  100,
				Step: 1,
			},
		})
		grids = append(grids, GridData{ID: "preferences" + index, Cards: cards})
	}
	folder := FolderData{
		ID:    "preferences0",
		Title: "Preferences",
		Open:  false,
		Grids: grids,
	}
	return folder
}
