package main

import (
	"fmt"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"github.com/gorilla/websocket"
	"sync"
)

type MiscHandler struct {
	curSpeed      uint8
	curBrightness uint8

	speedSelector      *accessory.Outlet
	speedServices      map[uint8]*service.Outlet
	brightnessSelector *accessory.Outlet
	brightnessServices map[uint8]*service.Outlet
	mu                 *sync.Mutex
}

func NewMiscHandler() (m *MiscHandler) {
	m = &MiscHandler{
		speedSelector: accessory.NewOutlet(accessory.Info{
			Name:             "HyperCube Speed",
			Manufacturer:     "Carter Peel",
			Model:            "HyperCube v1.0.0",
			FirmwareRevision: "HyperKit v1.0.0",
			ID:               3,
		}),
		brightnessSelector: accessory.NewOutlet(accessory.Info{
			Name:             "HyperCube Brightness",
			Manufacturer:     "Carter Peel",
			Model:            "HyperCube v1.0.0",
			FirmwareRevision: "HyperKit v1.0.0",
			ID:               4,
		}),
		speedServices:      make(map[uint8]*service.Outlet),
		brightnessServices: make(map[uint8]*service.Outlet),
		mu:                 new(sync.Mutex),
	}

	speedName := characteristic.NewName()
	speedName.Value = "25"
	m.speedSelector.Outlet.AddCharacteristic(speedName.Characteristic)
	m.speedServices[25] = m.speedSelector.Outlet
	m.speedSelector.Outlet.On.OnValueRemoteUpdate(func(b bool) {
		for _, svc2 := range m.speedServices {
			if svc2 != m.speedSelector.Outlet {
				svc2.On.SetValue(false)
			}
		}
		m.SetSpeed(25)
	})

	brightnessName := characteristic.NewName()
	brightnessName.Value = "25"
	m.brightnessSelector.Outlet.AddCharacteristic(brightnessName.Characteristic)
	m.brightnessServices[25] = m.brightnessSelector.Outlet
	m.brightnessSelector.Outlet.On.OnValueRemoteUpdate(func(b bool) {
		for _, svc2 := range m.brightnessServices {
			if svc2 != m.brightnessSelector.Outlet {
				svc2.On.SetValue(false)
			}
		}
		m.SetBrightness(25)
	})

	speeds := []uint8{50, 75, 100, 125, 150, 175, 200, 225, 255}
	for _, speed := range speeds {
		speed := speed
		newOutlet := service.NewOutlet()
		name := characteristic.NewName()
		name.Value = fmt.Sprintf("%d", speed)
		newOutlet.AddCharacteristic(name.Characteristic)

		newOutlet.AddLinkedService(m.speedSelector.Outlet.Service)
		m.speedSelector.Outlet.AddLinkedService(newOutlet.Service)

		for _, v := range m.speedServices {
			newOutlet.AddLinkedService(v.Service)
		}
		newOutlet.On.OnValueRemoteUpdate(func(b bool) {
			for _, svc2 := range m.speedServices {
				if svc2 != newOutlet {
					svc2.On.SetValue(false)
				}
			}
			m.SetSpeed(speed)
		})
		m.speedServices[speed] = newOutlet
		m.speedSelector.AddService(newOutlet.Service)
	}

	for _, brightness := range speeds {
		brightness := brightness
		newOutlet := service.NewOutlet()
		name := characteristic.NewName()
		name.Value = fmt.Sprintf("%d", brightness)
		newOutlet.AddCharacteristic(name.Characteristic)

		newOutlet.AddLinkedService(m.brightnessSelector.Outlet.Service)
		m.brightnessSelector.Outlet.AddLinkedService(newOutlet.Service)

		for _, v := range m.brightnessServices {
			newOutlet.AddLinkedService(v.Service)
		}
		newOutlet.On.OnValueRemoteUpdate(func(b bool) {
			for _, svc2 := range m.brightnessServices {
				if svc2 != newOutlet {
					svc2.On.SetValue(false)
				}
			}
			m.SetBrightness(brightness)
		})
		m.brightnessServices[brightness] = newOutlet
		m.brightnessSelector.AddService(newOutlet.Service)
	}

	return m
}

func (sph *MiscHandler) SetSpeed(speed uint8) {
	sph.mu.Lock()
	defer sph.mu.Unlock()
	sph.curSpeed = speed
	if err := socket.WriteMessage(websocket.TextMessage, buildSpeedJson(speed)); err != nil {
		log.Printf("Error setting speed to %d (%.2f%%): %v\n", speed, (float64(speed)/255)*100, err)
		return
	}
	log.Printf("Set speed to %d (%.2f%%)\n", speed, (float64(speed)/255)*100)
}

func (sph *MiscHandler) SetBrightness(brightness uint8) {
	sph.mu.Lock()
	defer sph.mu.Unlock()
	sph.curBrightness = brightness
	if err := socket.WriteMessage(websocket.TextMessage, buildBrightnessJson(brightness)); err != nil {
		log.Printf("Error setting brightness to %d (%.2f%%): %v\n", brightness, (float64(brightness)/255)*100, err)
		return
	}
	log.Printf("Set brightness to %d (%.2f%%)\n", brightness, (float64(brightness)/255)*100)
}
