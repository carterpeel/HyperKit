package bluetoothproxy

import (
	"fmt"
	"hyperkit/core/errorTypes"
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

func (bt *BluetoothProxy) ConnectAudioOutput() (err error) {
	if bt.adapter, err = adapter.GetDefaultAdapter(); err != nil {
		return fmt.Errorf("error getting default adapter: %v", err)
	}

	log.Infof("Searching for Bluetooth devices containing '%s'...\n", strings.ToLower(bt.deviceName))
	switch err := bt.ConnectFromCurrentDeviceList(); err != nil {
	case err == errorTypes.BtDeviceDoesNotExist:
		go func() {
			if err := bt.StartScanner(); err != nil {
				log.Errorf("Error starting Bluetooth runInBackground scanner: %v", err)
			}
		}()
		return nil
	case err == errorTypes.BtDeviceDown:
		log.Infof("Indefinitely attempting to connect to device '%s'...\n", bt.dev.Properties.Name)
		go bt.TryConnect(-1)
		return nil
	}
	return fmt.Errorf("error finding device from existing list: %v", err)
}

func (bt *BluetoothProxy) TryConnect(retries int) {
	for i := 0; i != retries; i++ {
		if err := bt.dev.Connect(); err != nil {
			if !errorTypes.IsBtDevDown(err) {
				log.Warnf("Weird/unexpected error returned during connection attempt %d: %v\n", i, err)
			}
			continue
		}
		log.Infof("Successfully connected to device '%s'\n", bt.dev.Properties.Name)
		return
	}
}

func (bt *BluetoothProxy) ConnectFromCurrentDeviceList() (err error) {
	_ = bt.adapter.SetPowered(true)
	devs, err := bt.adapter.GetDevices()
	if err != nil {
		return fmt.Errorf("error getting devices: %v", err)
	}
	for _, v := range devs {
		if strings.Contains(strings.ToLower(v.Properties.Name), strings.ToLower(bt.deviceName)) {
			bt.dev = v
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
					if errorTypes.IsBtDevDown(err) {
						return errorTypes.BtDeviceDown
					} else {
						log.Warnf("error connecting to '%s': %v\n", v.Properties.Name, err)
						continue
					}
				}
				return nil
			}
		}
	}
	return errorTypes.BtDeviceDoesNotExist
}

func (bt *BluetoothProxy) StartScanner() (err error) {
	log.Infof("Scanning for Bluetooth devices matching the criteria '%s'...\n", bt.deviceName)
	discovery, cancel, err := api.Discover(bt.adapter, nil)
	if err != nil {
		return fmt.Errorf("error running discovery: %v", err)
	}
	go func() {
		defer cancel()
		for ev := range discovery {
			if ev.Type == adapter.DeviceRemoved {
				continue
			}

			if bt.dev, err = device.NewDevice1(ev.Path); err != nil {
				log.Errorf("%s: %s", ev.Path, err)
				continue
			}
			if strings.Contains(strings.ToLower(bt.dev.Properties.Name), strings.ToLower(bt.deviceName)) {
				bt.deviceName = bt.dev.Properties.Name
				log.Infof("Found requested device! (mac: %s)\n", bt.dev.Properties.Address)
				break
			} else {
				log.Infof("BT device with address '%s' discovered but did not match the connection criteria\n", bt.dev.Properties.Address)
				bt.dev = nil
			}
		}
		if bt.dev != nil {
			log.Infoln("Connecting now...")
			if err := bt.dev.Connect(); err != nil {
				log.Errorf("Error connecting to device '%s': %v\n", bt.dev.Properties.Name, err)
				return
			}
			log.Infoln("BLUETOOTH DEVICE CONNECTED")
			return
		}
	}()
	return nil
}

// Deprecated
/*func (bt *BluetoothProxy) Connect() (err error) {
	log.Infof("Searching for Bluetooth device '%s'...\n", bt.deviceName)
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

	log.Debugln("Starting discovery...")
	discovery, cancel, err := api.Discover(bt.adapter, nil)
	if err != nil {
		return fmt.Errorf("error running discovery: %v", err)
	}
	defer cancel()

	for ev := range discovery {
		if ev.Type == adapter.DeviceRemoved {
			continue
		}

		if bt.dev, err = device.NewDevice1(ev.Path); err != nil {
			log.Errorf("%s: %s", ev.Path, err)
			continue
		}

		log.Infof("BT device with name '%s' discovered but did not match the connection criteria\n", bt.dev.Properties.Name)
		if strings.Contains(strings.ToLower(bt.dev.Properties.Name), strings.ToLower(bt.deviceName)) {
			bt.deviceName = bt.dev.Properties.Name
			log.Warnf("Found requested device! (mac: %s)\n", bt.dev.Properties.Address)
			break
		} else {
			bt.dev = nil
		}
	}
	if bt.dev != nil {
		log.Infoln("Connecting now...")
		if err := bt.dev.Connect(); err != nil {
			return fmt.Errorf("error connecting to device: %v", err)
		}
		log.Infoln("BLUETOOTH DEVICE CONNECTED")
		return nil
	}
	return fmt.Errorf("could not find device")
}
*/

// Deprecated
/*func (bt *BluetoothProxy) Disconnect() {
	bt.dev = nil
	bt.adapter.Close()
}*/
