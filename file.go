package thing2

import (
	"encoding/json"
	"io/ioutil"
)

var (
	filename = GetEnv("DEVICES_FILE", "devices.json")
)

func fileWriteDevices() error {
	data, err := json.MarshalIndent(devices, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func fileReadDevices() error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	devicesMu.Lock()
	defer devicesMu.Unlock()
	err = json.Unmarshal(data, &devices)
	if err != nil {
		return err
	}
	return nil
}
