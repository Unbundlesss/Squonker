package data

import (
	"encoding/json"
	"io/ioutil"
)

type DataRawTemplate struct {
	Interaction string `json:"interaction"`
	Icon        string `json:"icon"`
}

// GetInstrumentInteraction returns the basic type of an instrument JSON; eg. ("notes", "bass") or ("drumpads", "drumtype")
func GetInstrumentType(filename string) (string, string, error) {

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", "", err
	}

	data := &DataRawTemplate{}
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		return "", "", err
	}
	return data.Interaction, data.Icon, nil
}
