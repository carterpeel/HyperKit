package airplayserver

import (
	"testing"
)

func TestNewAirplayServer(t *testing.T) {
	srv, err := NewAirplayLedFXBridge("airplayserver_test", "/home/pi/ledfx/audio/stream", "")
	if err != nil {
		t.Fatalf("Error creating new AirPlay LedFX bridge: %v\n", err)
	}
	t.Logf("Created bridge...")

	if err := srv.Start(); err != nil {
		t.Fatalf("Error starting LedFX bridge: %v\n", err)
	}
}
