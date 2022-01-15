package apiconn

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func GetAllPresets(wledIP string) (pd []PresetData, err error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/presets.json", wledIP))
	if err != nil {
		return nil, fmt.Errorf("error connecting to WLED API: %v", err)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading data from response: %v", err)
	}
	pdMap := make(map[string]*PresetData)

	if err := json.Unmarshal(bodyBytes, &pdMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %v", err)
	}

	pd = make([]PresetData, 0)

	for id, pst := range pdMap {
		if pst.ID, err = strconv.Atoi(id); err != nil {
			return nil, fmt.Errorf("error converting WLED string id '%s' to numerical value: %v", id, err)
		}
		pd = append(pd, *pst)
	}

	return pd, nil
}

type PresetData struct {
	Name        string `json:"n"`
	ShortName   string `json:"ql"`
	On          bool   `json:"on"`
	Brightness  int    `json:"bri"`
	Transition  int    `json:"transition"`
	MainSegment int    `json:"mainseg"`
	ID          int
}
