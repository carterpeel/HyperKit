package main

import (
	"fmt"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"github.com/gorilla/websocket"
	"hyperkit/core/airplayserver"
	"sync"
	"sync/atomic"
)

type PresetHandler struct {
	lastActive int

	Presets map[int]*service.Outlet
	mu      *sync.Mutex

	ledFxBridge  *airplayserver.AirplayServer
	ledFxSwitch  *service.Outlet
	musicEnabled uint32
}

func NewPresetHandler(defaultSolid bool) (p *PresetHandler, err error) {
	p = &PresetHandler{
		Presets:    make(map[int]*service.Outlet, 0),
		lastActive: -1, // Should be -1 by default
		mu:         &sync.Mutex{},
	}

	if defaultSolid {
		solid := service.NewOutlet()
		solidName := characteristic.NewName()
		solidName.Value = "Preset: Solid"
		solid.AddCharacteristic(solidName.Characteristic)
		hyperCube.AddService(solid.Service)
		p.InitPreset(69, solid)
		if defaultSolid {
			solid.On.SetValue(true)
		}
	}

	if err := socket.WriteMessage(websocket.TextMessage, []byte(`{"bri":255, "ps":69}`)); err != nil {
		return nil, fmt.Errorf("error booting WLED: %v", err)
	}

	hyperCube.Outlet.On.SetValue(true)

	hyperCube.Outlet.On.OnValueRemoteUpdate(func(b bool) {
		p.togglePower()
		if !b {
			for id, preset := range p.Presets {
				if preset.On.Value.(bool) == true {
					p.lastActive = id
				}
				go preset.On.SetValue(false)
			}
		} else if b {
			if p.lastActive > -1 {
				go p.enablePresetByID(p.lastActive)
			}
		}
	})
	return p, nil
}

func (p *PresetHandler) InitPreset(wledPresetID int, preset *service.Outlet) {
	preset.On.OnValueRemoteUpdate(func(b bool) {
		if !b {
			go p.enablePresetByID(-1)
			log.Println("Switching to solid color WLED preset...")
			return
		}
		if !hyperCube.Outlet.On.Value.(bool) {
			go p.togglePower()
			log.Println("Turning on HyperCube...")
			hyperCube.Outlet.On.SetValue(true)
		}
		p.enablePresetByID(wledPresetID)
		for _, linkedPreset := range p.Presets {
			if linkedPreset != preset {
				linkedPreset.On.SetValue(false)
			}
		}
	})
	p.Presets[wledPresetID] = preset
	for _, v := range p.Presets {
		if v != preset {
			preset.AddLinkedService(v.Service)
		}
		v.AddLinkedService(preset.Service)
	}
}

func (p *PresetHandler) togglePower() {
	p.mu.Lock()
	defer p.mu.Unlock()
	log.Println("Toggling HyperCube power...")
	if err := socket.WriteMessage(websocket.TextMessage, []byte(TogglePowerJson)); err != nil {
		log.Printf("Error turning off HyperCube: %v\n", err)
	}
}

func (p *PresetHandler) enablePresetByID(id int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	preset, ok := p.Presets[id]
	if !ok && id > -1 {
		log.Printf("Preset %d does not exist!\n", id)
		return
	}
	log.Printf("Switching to WLED preset with id: %d\n", id)
	if id > -1 {
		preset.On.SetValue(true)
	}
	if err := socket.WriteMessage(websocket.TextMessage, buildPresetJson(id)); err != nil {
		log.Printf("Error turning off LEDs: %v\n", err)
	}
}

func (p *PresetHandler) AddLedFXBridge(br *airplayserver.AirplayServer, toggleSwitch *service.Outlet) {
	p.mu.Lock()
	defer p.mu.Unlock()

	toggleSwitch.On.SetValue(false)

	toggleSwitch.On.OnValueRemoteUpdate(func(b bool) {
		if !b {
			defer atomic.StoreUint32(&p.musicEnabled, 0)
			toggleSwitch.On.SetValue(false)
			log.Printf("Stopping AirPlayBT bridge...")
			if err := br.Stop(); err != nil {
				log.Printf("Error stopping AirPlay BT bridge: %v\n", err)
			}
		} else if b {
			atomic.StoreUint32(&p.musicEnabled, 1)
			toggleSwitch.On.SetValue(true)
			log.Printf("Starting AirPlayBT bridge...")
			if err := br.Start(); err != nil {
				log.Printf("Error starting bridge: %v\n", err)
			}
		}
	})
}

func (p *PresetHandler) musicIsActive() bool {
	return atomic.LoadUint32(&p.musicEnabled) > 0
}
