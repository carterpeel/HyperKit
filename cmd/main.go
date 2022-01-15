package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"hyperkit/core"
	"os"
)

var (
	AirPlayName   string
	AudioPipePath string
	BtDevice      string
)

func init() {
	flag.StringVar(&AirPlayName, "airPlayName", "HyperKit-Audio", "The advertisement name for the AirPlay2 server. (default: HyperKit-Audio)")
	flag.StringVar(&AudioPipePath, "audioPipePath", "/home/pi/ledfx/audio/stream", "The fully qualified path to your LedFX audio pipe file. (default: '/home/pi/ledfx/audio/stream')")
	flag.StringVar(&BtDevice, "bluetoothDevice", "", "The name of the BlueTooth audio device to proxy audio to. (required)")

	flag.Parse()

	if BtDevice == "" {
		log.Errorln("'-bluetoothDevice' flag is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

}

func main() {
	c, err := core.NewCore(AirPlayName, AudioPipePath, BtDevice, [8]uint{6, 9, 6, 9, 4, 2, 0, 0})
	if err != nil {
		log.Fatalf("Error creating new core: %v\n", err)
	}
	if err := c.LoadPresetsFromWled(); err != nil {
		log.Fatalf("Error loading presets from WLED: %v\n", err)
	}
	log.Fatalln(c.Start())
}
