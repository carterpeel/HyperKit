package apiconn

import "testing"

func TestGetAllPresets(t *testing.T) {
	psts, err := GetAllPresets("hypercube.local")
	if err != nil {
		t.Fatalf("Error getting all presets: %v\n", err)
	}
	for _, v := range psts {
		t.Logf("Found preset: name=%v; id=%d\n", v.Name, v.ID)
	}
}
