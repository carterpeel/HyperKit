package main

import (
	"fmt"
)

func buildPresetJson(id int) []byte {
	return []byte(fmt.Sprintf(`{"ps":%d}`, id))
}
func buildSpeedJson(speed uint8) []byte {
	return []byte(fmt.Sprintf(`{"seg":[{"id":0,"sx":%d}]}`, speed))
}
func buildBrightnessJson(brightness uint8) []byte {
	return []byte(fmt.Sprintf("{\"on\":true,\"bri\":%d}", brightness))
}

type StateSegment struct {
	StateSegmentTop State `json:"state"`
}

type State struct {
	Bri int       `json:"bri,omitempty"`
	Seg []Segment `json:"seg,omitempty"`
}
type Segment struct {
	Sx  int `json:"sx,omitempty"`
	Bri int `json:"bri,omitempty"`
}
