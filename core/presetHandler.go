package core

import (
	"fmt"
	"github.com/brutella/hc/service"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
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
	core         *Core
}

func (c *Core) NewPresetHandler() (p *PresetHandler, err error) {
	p = &PresetHandler{
		Presets:    make(map[int]*service.Outlet, 0),
		lastActive: -1, // Should be -1 by default
		mu:         &sync.Mutex{},
		core:       c,
	}

	if err := c.socket.WriteMessage(websocket.TextMessage, []byte(`{"bri":255, "ps":69}`)); err != nil {
		return nil, fmt.Errorf("error booting WLED: %v", err)
	}

	c.menuOutlet.Outlet.On.SetValue(true)

	c.menuOutlet.Outlet.On.OnValueRemoteUpdate(func(b bool) {
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
		if !p.core.menuOutlet.Outlet.On.Value.(bool) {
			go p.togglePower()
			log.Println("Turning on HyperCube...")
			p.core.menuOutlet.Outlet.On.SetValue(true)
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
	if err := p.core.socket.WriteMessage(websocket.TextMessage, []byte(TogglePowerJson)); err != nil {
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
	if err := p.core.socket.WriteMessage(websocket.TextMessage, buildPresetJson(id)); err != nil {
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
