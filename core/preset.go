package core

import "github.com/brutella/hc/service"

type Preset struct {
	service      *service.Service
	name         string
	wledPresetId int
}
