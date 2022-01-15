package core

import (
	"fmt"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"github.com/gorilla/websocket"
	"hyperkit/core/airplayserver"
	"hyperkit/core/apiconn"
	"hyperkit/core/util"
)

type Core struct {
	presetHandler *PresetHandler
	presets       map[string]*Preset
	airplayServer *airplayserver.AirplayServer
	airplaySwitch *service.Outlet
	miscHandler   *MiscHandler
	homekitPin    [8]uint
	socket        *websocket.Conn
	bridge        *accessory.Bridge
	menuOutlet    *accessory.Outlet
	config        *Config
}

func NewCore(airplayName, audioNamedPipePath, bluetoothDevice string, homekitPin [8]uint) (c *Core, err error) {
	c = &Core{
		presets:    make(map[string]*Preset),
		homekitPin: homekitPin,
		bridge: accessory.NewBridge(accessory.Info{
			Name:             "HyperBridge",
			Manufacturer:     "Carter Peel",
			Model:            "HyperBridge v1.0.0",
			FirmwareRevision: "HyperKit v1.0.0",
			ID:               1,
		}),
		menuOutlet: accessory.NewOutlet(accessory.Info{
			Name:             "WLED-HyperKit",
			Manufacturer:     "Carter Peel",
			Model:            "HyperCube v1.0.0",
			FirmwareRevision: "HyperKit v1.0.0",
			ID:               2,
		}),
	}

	// Create a power button
	menuName := characteristic.NewName()
	menuName.Value = "Power"
	c.menuOutlet.Outlet.Service.AddCharacteristic(menuName.Characteristic)

	// Config file setup
	if c.config, err = InitConfig(); err != nil {
		return nil, fmt.Errorf("error initializing config: %v", err)
	}

	if c.socket, err = InitWebSocket(c.config.WledIP); err != nil {
		return nil, fmt.Errorf("error initializing websocket: %v", err)
	}

	// Preset Handler (HomeKit)
	if c.presetHandler, err = c.NewPresetHandler(); err != nil {
		return nil, fmt.Errorf("error creating preset handler: %v", err)
	}

	// AirPlay2 server (audio proxy)
	if c.airplayServer, err = airplayserver.NewAirplayLedFXBridge(airplayName, audioNamedPipePath, bluetoothDevice); err != nil {
		return nil, fmt.Errorf("error creating new AirPlay2 server: %v", err)
	}

	// Miscellaneous menu
	c.miscHandler = c.NewMiscHandler()
	// Set default speed and brightness
	c.miscHandler.SetSpeed(c.config.DefaultSpeed)
	c.miscHandler.SetBrightness(c.config.DefaultBrightness)

	// Create an airplay switch
	c.airplaySwitch = service.NewOutlet()
	airplaySwitchName := characteristic.NewName()
	airplaySwitchName.Value = "AirPlay2"
	c.airplaySwitch.AddCharacteristic(airplaySwitchName.Characteristic)

	c.presetHandler.AddLedFXBridge(c.airplayServer, c.airplaySwitch)

	c.menuOutlet.AddService(c.airplaySwitch.Service)
	c.menuOutlet.UpdateIDs()

	return c, nil
}

func (c *Core) LoadPresetsFromWled() (err error) {
	presets, err := apiconn.GetAllPresets(c.config.WledIP)
	if err != nil {
		return fmt.Errorf("error getting all presets: %v", err)
	}
	for _, preset := range presets {
		if preset.ID <= 0 {
			continue
		}
		if err := c.AddWledPreset(preset.Name, preset.ID); err != nil {
			return fmt.Errorf("error adding WLED preset: %v", err)
		}
	}
	return nil
}

func (c *Core) AddWledPreset(name string, id int) error {
	if id <= 0 {
		return fmt.Errorf("preset ID must be greater than 0")
	}

	preset := service.NewOutlet()
	presetName := characteristic.NewName()
	presetName.Value = name
	preset.AddCharacteristic(presetName.Characteristic)
	c.presetHandler.InitPreset(id, preset)

	c.menuOutlet.AddService(preset.Service)
	c.menuOutlet.UpdateIDs()

	return nil
}

func (c *Core) Start() (err error) {
	pin, err := util.PinArrayToString(c.homekitPin)
	if err != nil {
		return fmt.Errorf("error converting pin array to string: %v", err)
	}

	t, err := hc.NewIPTransport(hc.Config{Pin: pin}, c.bridge.Accessory, c.menuOutlet.Accessory, c.miscHandler.speedSelector.Accessory, c.miscHandler.brightnessSelector.Accessory)
	if err != nil {
		return fmt.Errorf("error creating new transport: %v", err)
	}

	hc.OnTermination(func() {
		<-t.Stop()
	})
	t.Start()
	return nil
}

/*func main() {
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
*/
