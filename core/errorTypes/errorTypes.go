package errorTypes

import "errors"

var (
	BtDeviceDoesNotExist = errors.New("bluetooth device does not exist")
	BtDeviceDown         = errors.New("bluetooth device is down")
)
