package main

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"hyperkit/core/airplayserver"
	"os"
)

func main() {
	// [------ Preset Config ------]
	presetHandler, err := NewPresetHandler(config.DefaultSolid)
	if err != nil {
		log.Panic(err)
	}

	aurora := service.NewOutlet()
	auroraName := characteristic.NewName()
	auroraName.Value = "Preset: Aurora"
	aurora.AddCharacteristic(auroraName.Characteristic)
	presetHandler.InitPreset(2, aurora)

	flow := service.NewOutlet()
	flowName := characteristic.NewName()
	flowName.Value = "Preset: Flow"
	flow.AddCharacteristic(flowName.Characteristic)
	presetHandler.InitPreset(1, flow)

	opps := service.NewOutlet()
	oppsName := characteristic.NewName()
	oppsName.Value = "Preset: Opps"
	opps.AddCharacteristic(oppsName.Characteristic)
	presetHandler.InitPreset(3, opps)

	rainbowBPM := service.NewOutlet()
	rainbowBPMName := characteristic.NewName()
	rainbowBPMName.Value = "Preset: Rainbow"
	rainbowBPM.AddCharacteristic(rainbowBPMName.Characteristic)
	presetHandler.InitPreset(4, rainbowBPM)

	candyCane := service.NewOutlet()
	candyCaneName := characteristic.NewName()
	candyCaneName.Value = "Preset: Candy"
	candyCane.AddCharacteristic(candyCaneName.Characteristic)
	presetHandler.InitPreset(5, candyCane)

	ledfx := service.NewOutlet()
	ledfxName := characteristic.NewName()
	ledfxName.Value = "Preset: Music"
	ledfx.AddCharacteristic(ledfxName.Characteristic)
	musicBridge, err := airplayserver.NewAirplayLedFXBridge("hypercube-audio", "/home/pi/ledfx/audio/stream", config.BtDeviceName)
	if err != nil {
		log.Panicf("Error creating new AirPlay/LEDfx BT bridge: %v\n", err)
	}
	presetHandler.AddLedFXBridge(musicBridge, ledfx)

	// [------ End Preset Config ------]

	// [------ Service Config ------]
	hyperCube.AddService(aurora.Service)
	hyperCube.AddService(flow.Service)
	hyperCube.AddService(opps.Service)
	hyperCube.AddService(rainbowBPM.Service)
	hyperCube.AddService(candyCane.Service)
	hyperCube.AddService(ledfx.Service)
	// [------ End Service Config ------]

	bridge.UpdateIDs()
	hyperCube.UpdateIDs()
	miscHandler.speedSelector.UpdateIDs()
	miscHandler.brightnessSelector.UpdateIDs()

	conf := hc.Config{Pin: "69694200"}
	t, err := hc.NewIPTransport(conf, bridge.Accessory, hyperCube.Accessory, miscHandler.speedSelector.Accessory, miscHandler.brightnessSelector.Accessory)
	if err != nil {
		log.Panic(err)
	}

	hc.OnTermination(func() {
		<-t.Stop()
		os.Exit(1)
	})
	t.Start()
}
