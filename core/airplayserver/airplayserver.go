package airplayserver

import (
	"fmt"
	"github.com/carterpeel/bobcaygeon/raop"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"hyperkit/core/ledfx"
	"os"
	"sync"
)

type AirplayServer struct {
	plyr           *LocalPlayer
	svc            *raop.AirplayServer
	ledfxctl       *ledfx.Controller
	containerMutex *sync.Mutex

	muted bool

	album   string
	artist  string
	title   string
	artwork []byte
}

func NewAirplayLedFXBridge(advertisementName, pipeFilePath, btDeviceName string) (a *AirplayServer, err error) {
	a = &AirplayServer{
		containerMutex: &sync.Mutex{},
		muted:          false,
		artwork:        make([]byte, 0),
	}
	log.Infof("Creating local player...\n")

	if a.plyr, err = NewBluetoothPlayer(pipeFilePath, btDeviceName); err != nil {
		return nil, fmt.Errorf("error creating new player: %w", err)
	}
	log.Infof("Created local player with hook to named pipe '%s'\n", pipeFilePath)
	a.svc = raop.NewAirplayServer(8044, advertisementName, a.plyr)
	log.Infof("Created AirPlay server with advertisementName '%s'\n", advertisementName)

	if a.ledfxctl, err = ledfx.NewController(); err != nil {
		return nil, fmt.Errorf("error creating new LedFX controller: %w", err)
	}

	return
}

func (a *AirplayServer) Start() error {
	if err := unix.Mkfifo("/home/pi/ledfx/audio/stream", 0600); err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("error creating FIFO file: %w", err)
		}
	}
	go func() {
		a.containerMutex.Lock()
		defer a.containerMutex.Unlock()
		log.Infof("Starting LEDfx container...\n")
		if err := a.ledfxctl.Resume(); err != nil {
			log.Errorf("Error resuming LedFX Docker container: %v\n", err)
		}
	}()

	go a.svc.Start(false, true)

	return nil
}

func (a *AirplayServer) Stop() error {
	go func() {
		a.containerMutex.Lock()
		defer a.containerMutex.Unlock()
		if err := a.ledfxctl.Pause(); err != nil {
			log.Errorf("Error pausing LEDfx Docker container: %v\n", err)
		}
	}()
	log.Infof("Stopping AirPlay server...")
	a.svc.Stop()
	log.Infof("Stopped AirPlay server successfully.")

	return nil
}
