package main

import (
	"html/template"
	"led/lights"
	"led/socket"
	"log"
	"net/http"
	"os"
	"os/signal"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

const (
	brightness = 128
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
	piled := lights.NewPiLED(&opt)
	if err := piled.Init(); err != nil {
		return
	}

	defer func() {
		piled.Clear(0)
		piled.Finish()
		log.Println("Finished, done.")
	}()

	hostUrl := getOutboundIP() + ":5000"
	mux := &http.ServeMux{}
	server := &http.Server{
		Addr:    hostUrl,
		Handler: mux,
	}

	const pattern = "www/*.html"
	templ, err := template.New("").ParseGlob(pattern)
	if err != nil {
		log.Fatalln("Parse global templates", pattern, err)
	}

	sockServer := socket.NewServer(templ)
	mux.HandleFunc("/events", sockServer.Events)
	sockServer.Run()

	fs := http.FileServer(http.Dir("www/"))
	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache")
		http.StripPrefix("/static/", fs).ServeHTTP(w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache")
		templ.ExecuteTemplate(w, "index.html", nil)
	})

	mux.HandleFunc("/load", loadDynamic(&opt, templ))

	setupHandlers(mux, piled, sockServer, templ)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Printf("Host at: %v '%v'", hostUrl, err)
		}
	}()
	log.Printf("Listening at %v", hostUrl)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs
	log.Println("Signal")
	piled.Stop()
}
