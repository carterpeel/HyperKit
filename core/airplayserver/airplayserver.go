package airplayserver

import (
	"fmt"
	"github.com/carterpeel/bobcaygeon/raop"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"sync"
)

type AirplayServer struct {
	plyr           *LocalPlayer
	svc            *raop.AirplayServer
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
		return nil, fmt.Errorf("error creating new player: %v", err)
	}
	log.Infof("Created local player with hook to named pipe '%s'\n", pipeFilePath)
	a.svc = raop.NewAirplayServer(8044, advertisementName, a.plyr)
	log.Infof("Created AirPlay server with advertisementName '%s'\n", advertisementName)

	return
}

func (a *AirplayServer) Start() error {
	if err := unix.Mkfifo("/home/pi/ledfx/audio/stream", 0600); err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("error creating FIFO file: %v", err)
		}
	}
	go func() {
		a.containerMutex.Lock()
		defer a.containerMutex.Unlock()
		log.Infof("Starting LEDfx container...\n")
		if out, err := exec.Command("docker", "start", "ledfx").CombinedOutput(); err != nil {
			log.Errorf("error restarting LEDfx Docker container: %v ([\"docker\", \"restart\", \"ledfx\"]: %s)", err, string(out))
		}
	}()

	go a.svc.Start(false, true)

	log.Printf("Connecting to BT device...\n")
	if err := a.plyr.btpx.Connect(); err != nil {
		return fmt.Errorf("error connecting to BT device: %v", err)
	}

	return nil
}

func (a *AirplayServer) Stop() error {
	go func() {
		a.containerMutex.Lock()
		defer a.containerMutex.Unlock()
		if out, err := exec.Command("docker", "stop", "ledfx").CombinedOutput(); err != nil {
			log.Errorf("error stopping LEDfx Docker container: %v ([\"docker\", \"stop\", \"ledfx\"]: %s", err, string(out))
		}
	}()
	log.Infof("Stopping AirPlay server...")
	a.svc.Stop()
	log.Infof("Stopped AirPlay server successfully.")

	return nil
}
