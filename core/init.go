package main

import (
	"fmt"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	hclog "github.com/brutella/hc/log"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"
	"hyperkit/core/airplayserver"
	"io"
	"io/ioutil"
	flog "log"
	"os"
)

func init() {
	// Parse the config file
	configBytes, err := ioutil.ReadFile("/etc/hyperkit.conf")
	if err != nil {
		flog.Fatalf("Error reading config file: %v\n", err)
	}

	config = new(Config)
	if err := yaml.Unmarshal(configBytes, config); err != nil {
		flog.Fatalf("Error parsing config file: %v\n", err)
	}

	if len(config.WledIP) <= 0 {
		flog.Fatalf("Error: 'wled_ip' must not be omitted in /etc/hyperkit.conf\n")
	}

	if len(config.BtDeviceName) <= 0 {
		flog.Fatalf("Error: 'bluetooth_device' must not be omitted in /etc/hyperkit.conf\n")
	}

	if config.DefaultSpeed == 0 {
		config.DefaultSpeed = 127
	}
	if config.DefaultBrightness == 0 {
		config.DefaultBrightness = 255
	}

	if len(config.LogFile) <= 0 {
		config.LogFile = "/var/log/hyperkit.log"
	}

	// Initialize the websocket global
	if socket, _, err = websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s/ws", config.WledIP), nil); err != nil {
		flog.Fatalf("Error dialing websocket: %v\n", err)
	}

	if config.Debug {
		hclog.Debug.Enable()
		log = hclog.Debug.Logger
		go SocketReader(socket)
		airplayserver.StartProfiler()
	} else {
		log = flog.New(os.Stdout, "[HYPERKIT-INFO]: ", flog.LstdFlags)
	}

	logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error opening logfile: %v\n", err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	bridge = accessory.NewBridge(accessory.Info{
		Name:             "HyperBridge",
		Manufacturer:     "Carter Peel",
		Model:            "HyperBridge v1.0.0",
		FirmwareRevision: "HyperKit v1.0.0",
		ID:               1,
	})

	hyperCube = accessory.NewOutlet(accessory.Info{
		Name:             "HyperCube",
		Manufacturer:     "Carter Peel",
		Model:            "HyperCube v1.0.0",
		FirmwareRevision: "HyperKit v1.0.0",
		ID:               2,
	})

	miscHandler = NewMiscHandler()
	miscHandler.SetSpeed(config.DefaultSpeed)
	miscHandler.SetBrightness(config.DefaultBrightness)

	hyperCubeName := characteristic.NewName()
	hyperCubeName.Value = "Power"
	hyperCube.Outlet.Service.AddCharacteristic(hyperCubeName.Characteristic)
}

type Config struct {
	WledIP            string `yaml:"wled_ip,omitempty"`
	DefaultSolid      bool   `yaml:"use_default_solid,omitempty"`
	DefaultSpeed      uint8  `yaml:"default_speed,omitempty"`
	DefaultBrightness uint8  `yaml:"default_brightness,omitempty"`
	Debug             bool   `yaml:"debug_logging,omitempty"`
	LogFile           string `yaml:"logfile,omitempty"`
	BtDeviceName      string `yaml:"bluetooth_device,omitempty"`
}

func SocketReader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// print out that message for clarity
		fmt.Println(string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}

	}
}
