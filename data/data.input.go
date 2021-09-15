package data

import (
	"encoding/json"
	"io/ioutil"
	"path"
)

type DataInput struct {
	Iconmap []struct {
		Svg   string   `json:"svg"`
		Regex []string `json:"regex"`
	} `json:"iconmap"`
	Author string `json:"author"`
	Notes  []struct {
		Directory   string   `json:"directory"`
		Name        string   `json:"name"`
		Template    string   `json:"template"`
		Note        string   `json:"note"`
		Order       []string `json:"order,omitempty"`
		VolumeBoost float64  `json:"volume_boost"`
		Mods        []struct {
			Preset    string `json:"preset"`
			Template  string `json:"template"`
			Overrides []struct {
				Macro    string  `json:"macro"`
				Override float64 `json:"override"`
			} `json:"overrides,omitempty"`
		} `json:"mods,omitempty"`
	} `json:"notes"`
	Drumkits []struct {
		Directory   string  `json:"directory"`
		Name        string  `json:"name"`
		Template    string  `json:"template"`
		VolumeBoost float64 `json:"volume_boost"`
		Presets     []struct {
			Name      string `json:"name"`
			Directory string `json:"directory"`
			Note      string `json:"note"`
			Template  string `json:"template"`
			Overrides []struct {
				Macro    string  `json:"macro"`
				Override float64 `json:"override"`
			} `json:"overrides,omitempty"`
		} `json:"presets"`
	} `json:"drumkits"`
}

func LoadInputData(designFolder string) (DataInput, error) {

	data := DataInput{}

	file, err := ioutil.ReadFile(path.Join(designFolder, "design.json"))
	if err != nil {
		return data, err
	}

	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		return data, err
	}
	return data, nil
}
