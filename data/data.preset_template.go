package data

import (
	"encoding/json"
	"io/ioutil"
)

type PresetTemplate struct {
	Interaction string `json:"interaction"`
	InputMode   int    `json:"inputMode"`
	Params      struct {
		Tuning struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"tuning"`
		Reverb struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"reverb"`
		ModSustain struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"modSustain"`
		Pitch struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"pitch"`
		Levels struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"levels"`
		AmpAttack struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"ampAttack"`
		AmpSustain struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"ampSustain"`
		AmpDecay struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"ampDecay"`
		ModDecay struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"modDecay"`
		FilterCutoff struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"filterCutoff"`
		FilterEnvAmount struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"filterEnvAmount"`
		ModAttack struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"modAttack"`
		FilterResonance struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"filterResonance"`
		ModRelease struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"modRelease"`
		AmpRelease struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"ampRelease"`
		PitchEnv struct {
			StartValue   float64   `json:"startValue"`
			DefaultValue float64   `json:"defaultValue"`
			Multipliers  []float64 `json:"multipliers"`
		} `json:"pitchEnv"`
	} `json:"params"`
	Samples           []string  `json:"samples"`
	Tune              float64   `json:"tune"`
	LoopStart         float64   `json:"loopStart"`
	FilterPitchAmount float64   `json:"filterPitchAmount"`
	MacroMappings     []int     `json:"macroMappings"`
	MacroNames        []string  `json:"macroNames"`
	LoopLength        float64   `json:"loopLength"`
	UserID            string    `json:"user_id"`
	PackName          string    `json:"packName"`
	TriggerIcons      []string  `json:"triggerIcons"`
	PackType          string    `json:"packType"`
	EngineType        string    `json:"engineType"`
	FilterType        int       `json:"filterType"`
	ProductIds        []string  `json:"productIds"`
	MacroDefaults     []float64 `json:"macroDefaults"`
	Created           int64     `json:"created"`
	Icon              string    `json:"icon"`
	Tags              []string  `json:"tags"`
	AppVersion        int       `json:"app_version"`
	Author            string    `json:"author"`
	LoopingMode       bool      `json:"loopingMode"`
	ID                string    `json:"_id"`
	Description       string    `json:"description"`
	Engine            string    `json:"engine"`
	PackSortOrder     int       `json:"packSortOrder"`
	Colour            string    `json:"colour"`
	Name              string    `json:"name"`
	Type              string    `json:"type"`
}

func LoadPresetTemplate(templatePath string) (PresetTemplate, string, error) {

	file, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return PresetTemplate{}, templatePath, err
	}
	data := PresetTemplate{}

	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		return data, templatePath, err
	}
	return data, templatePath, nil
}
