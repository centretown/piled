package main

import (
	"fmt"
	"html/template"
	"image/color"
	"led/lights"
	"led/socket"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

func queryForm(r *http.Request) url.Values {
	err := r.ParseForm()
	if err != nil {
		lights.LogError("queryForm", err)
		return url.Values{}
	}
	values := r.Form
	return values
}

var (
	duration = 1000
	pulse    = 40
)

func setupHandlers(mux *http.ServeMux, piled *lights.PiLED, sock *socket.Server, tmpl *template.Template) {
	handleBasicColors(mux, piled, sock, tmpl)
	mux.HandleFunc("/blink", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			values := queryForm(r)
			channel := queryChannel(values, piled.ChannelCount())
			brightness := queryBrightness(values)
			piled.StartRun()
			piled.ShowBlink(channel,
				[]color.Color{
					color.RGBA{R: 255},
					color.RGBA{G: 255},
					color.RGBA{B: 255},
				},
				uint32(brightness), uint32(duration), uint32(pulse))
			piled.StopRun()
			log.Println("stop blink")
		}()
		log.Println("blink")
	})
	mux.HandleFunc("/pic", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			values := queryForm(r)
			channel := queryChannel(values, piled.ChannelCount())
			brightness := queryBrightness(values)
			piled.StartRun()
			piled.ShowFile(channel, "waves.jpg", uint32(brightness))
			piled.StopRun()
			log.Println("stop waves")
		}()
		log.Println("waves")
	})
}

func handleBasicColors(mux *http.ServeMux, piled *lights.PiLED, sock *socket.Server, tmpl *template.Template) {
	mux.HandleFunc("/color", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			values := queryForm(r)
			colorHex := queryColor(values)
			colorVal := lights.ToColor(colorHex)
			channel := queryChannel(values, piled.ChannelCount())
			brightness := queryBrightness(values)
			piled.StartRun()
			piled.ShowBytes(channel, []uint32{lights.FromColorBrightness(colorVal, uint32(brightness))})
			piled.StopRun()

			w.Write([]byte("OK"))
			SetCurrentStatus(sock, tmpl, "color", strconv.Itoa(int(brightness)))
		}()
	})

	mux.HandleFunc("/rgb", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			values := queryForm(r)
			uc := queryColors(values)
			channel := queryChannel(values, piled.ChannelCount())
			piled.StartRun()
			piled.ShowBytes(channel, []uint32{uc})
			piled.StopRun()
		}()
	})

	mux.HandleFunc("/clear", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			values := queryForm(r)
			channel := queryChannel(values, piled.ChannelCount())
			piled.StartRun()
			piled.Clear(channel)
			piled.StopRun()
		}()
	})
}

// thanks to Mr. Wong
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func queryRange(values url.Values, name string, min int, max int, fill int) int {
	fname := "queryRange"
	value, ok := values[name]
	if !ok {
		lights.LogError(fname, fmt.Errorf("%s not found", name))
		return fill
	}
	i, err := strconv.Atoi(value[0])
	if err != nil {
		lights.LogError(fname, err)
		return fill
	}

	if i > max || i < min {
		lights.LogError(fname,
			fmt.Errorf("%s out of bounds(%v-%v): %v", name, min, max, i))
		return fill
	}
	return i
}

func queryBrightness(values url.Values) uint8 {
	return uint8(queryRange(values, "brightness", 1, 100, 100))
}

func queryChannel(values url.Values, channelCount int) int {
	return queryRange(values, "channel", 0, channelCount-1, 0)
}

func queryColor(values url.Values) uint32 {
	fname := "queryColor"
	value, ok := values["value"]
	if !ok {
		lights.LogError(fname, fmt.Errorf("value not found"))
		return 0
	}
	log.Println(value[0])
	i, err := strconv.Atoi(value[0])
	if err != nil {
		lights.LogError(fname, err)
		return 0
	}
	return uint32(i)
}

func queryColors(values url.Values) uint32 {
	brightness := uint32(queryRange(values, "brightness", 1, 255, 255))
	red := uint32(queryRange(values, "r", 0, 255, 0))
	green := uint32(queryRange(values, "g", 0, 255, 0))
	blue := uint32(queryRange(values, "b", 0, 255, 0))
	log.Println(red, green, blue, brightness)
	return lights.FromColorBrightness(
		color.RGBA{R: uint8(red), G: uint8(green), B: uint8(blue)}, brightness)
}
