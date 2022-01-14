package bluetoothproxy

import (
	"fmt"
	"strings"

	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/bluez/profile/adapter"
	"github.com/muka/go-bluetooth/bluez/profile/device"
	log "github.com/sirupsen/logrus"
)

type BluetoothProxy struct {
	dev        *device.Device1
	adapter    *adapter.Adapter1
	deviceName string
}

func ProxyBluetoothDevice(deviceName string) (bt *BluetoothProxy, err error) {
	bt = &BluetoothProxy{
		dev:        new(device.Device1),
		deviceName: deviceName,
	}

	return bt, nil
}

func (bt *BluetoothProxy) Connect() (err error) {
	if bt.adapter, err = adapter.GetDefaultAdapter(); err != nil {
		return fmt.Errorf("error getting default adapter: %v", err)
	}
	_ = bt.adapter.SetPowered(true)
	devs, err := bt.adapter.GetDevices()
	if err != nil {
		return fmt.Errorf("error getting devices: %v", err)
	}
	for _, v := range devs {
		if strings.Contains(strings.ToLower(v.Properties.Name), strings.ToLower(bt.deviceName)) {
			bt.deviceName = v.Properties.Name
			connected, err := v.GetConnected()
			if err != nil {
				log.Warnf("Error getting device connection status for '%s': %v\n", v.Properties.Name, err)
				continue
			}
			if connected {
				log.Infof("Already connected to requested device. No need to reconnect.\n")
				return nil
			} else {
				if err := v.Connect(); err != nil {
					return fmt.Errorf("error connecting to device: %v", err)
				}
				return nil
			}
		}
	}

	log.Debug("Starting discovery...")
	discovery, cancel, err := api.Discover(bt.adapter, nil)
	if err != nil {
		return fmt.Errorf("error running discovery: %v", err)
	}
	defer cancel()

	if err := bt.adapter.FlushDevices(); err != nil {
		return fmt.Errorf("error flushing devices: %v", err)
	}

	for ev := range discovery {
		if ev.Type == adapter.DeviceRemoved {
			continue
		}

		if bt.dev, err = device.NewDevice1(ev.Path); err != nil {
			log.Errorf("%s: %s", ev.Path, err)
			continue
		}

		log.Infof("name=%s addr=%s rssi=%d", bt.dev.Properties.Name, bt.dev.Properties.Address, bt.dev.Properties.RSSI)
		if strings.Contains(strings.ToLower(bt.dev.Properties.Name), strings.ToLower(bt.deviceName)) {
			bt.deviceName = bt.dev.Properties.Name
			log.Warnf("Found requested device! (mac: %s)\n", bt.dev.Properties.Address)
			break
		} else {
			bt.dev = nil
		}
	}
	if bt.dev != nil {
		log.Warnf("Connecting now...\n")
		if err := bt.dev.Connect(); err != nil {
			return fmt.Errorf("error connecting to device: %v", err)
		}
		log.Warnf("BLUETOOTH DEVICE CONNECTED")
		return nil
	}
	return fmt.Errorf("could not find device")
}

func (bt *BluetoothProxy) Disconnect() {
	bt.dev = nil
	bt.adapter.Close()
}
