package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	instrument "github.com/Unbundlesss/squonker/data"
	"github.com/spf13/cobra"
)

var instrumentSearchPath string

// copy 'Instruments' folder from
//
// STUDIO : ~/Library/Containers/fm.endlesss.app/Data/Library/Application Support/Endlesss/Presets/
//    IOS : /var/mobile/Containers/Data/Application/<GUID>/Library/Application Support/Endlesss/Presets/
//
// .. to somewhere local if you want to run an analysis pass on all the preset data
var analysisCmd = &cobra.Command{
	Use:   "analysis",
	Short: "Run data analysis across instrument presets",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runAnalysis(instrumentSearchPath)
	},
}

func init() {
	rootCmd.AddCommand(analysisCmd)

	analysisCmd.Flags().StringVarP(&instrumentSearchPath, "inst", "i", "", "Root directory of Endlesss instrument cache")
	analysisCmd.MarkFlagRequired("inst")
}

// map of macro names to CSV output files
type csvOutput struct {
	csvWriter     *csv.Writer
	csvFileHandle *os.File
}

func runAnalysis(instrumentPath string) {

	instDirs, err := ioutil.ReadDir(instrumentPath)
	if err != nil {
		sqFatal(err)
	}

	// "<instrument name>_<icon>" -> active CSV file handle
	var macroMap = map[string]csvOutput{}

	for _, instDir := range instDirs {
		if !instDir.IsDir() {
			sqlog.Print("             skipping : ", instDir.Name())
			continue
		}

		instrumentJsonPath := path.Join(instrumentPath, instDir.Name(), instDir.Name()+".json")
		instInteraction, instIcon, err := instrument.GetInstrumentType(instrumentJsonPath)

		if err != nil {
			sqlog.Print("             skipping : ", instDir.Name(), ": ", err)
			continue
		}
		sqlog.Print("            examining : ", instDir.Name(), " | ", instInteraction, " : ", instIcon)

		err = runPresetAnalysis(macroMap, instDir.Name(), instrumentJsonPath, instIcon)
		if err != nil {
			sqlog.Print("               failed : ", err)
			continue
		}
	}

	// close and flush out all CSV files
	for _, output := range macroMap {
		output.csvWriter.Flush()
		output.csvFileHandle.Close()
	}
}

func runPresetAnalysis(macroMap map[string]csvOutput, instrumentName, instrumentJsonPath, instrumentIcon string) error {

	file, err := ioutil.ReadFile(instrumentJsonPath)
	if err != nil {
		return err
	}
	data := instrument.PresetTemplate{}

	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		return err
	}

	// iterate macro names
	for _, macroName := range data.MacroNames {

		if len(macroName) == 0 {
			continue
		}

		// differentiate in the CSV storage by the icon (e.g. notes / bass / drumkit)
		macroMapName := instrumentIcon + "_" + macroName

		// get index into the various multipliers[] arrays to look at
		macroIndex := indexOf(macroName, data.MacroNames)

		// see if we already know about this macro name; if we don't create a new CSV to fill in for it
		if _, ok := macroMap[macroMapName]; !ok {

			// make the root path for what kind of instrument we're writing out
			csvPath := path.Join("analysis", instrumentIcon)
			if _, err = os.Stat(csvPath); os.IsNotExist(err) {
				os.MkdirAll(csvPath, 0777)
			}

			csvFile, err := os.Create(path.Join(csvPath, macroName+".csv"))
			if err != nil {
				return err
			}

			// create a new CSV writer instance and prep the data
			csvFileWriter := csv.NewWriter(csvFile)
			macroMap[macroMapName] = csvOutput{csvWriter: csvFileWriter, csvFileHandle: csvFile}

			// write title row
			macroMap[macroMapName].csvWriter.Write([]string{
				"Instrument",
				"Defaults",
				"Levels",
				"Tuning",
				"Pitch",
				"PitchEnv",
				"Reverb",
				"ModAttack",
				"ModDecay",
				"ModSustain",
				"ModRelease",
				"AmpAttack",
				"AmpSustain",
				"AmpDecay",
				"AmpRelease",
				"FilterType",
				"FilterCutoff",
				"FilterEnvAmount",
				"FilterResonance",
			})
		}

		csvData := []string{data.Name}

		// all the "if len(data.Params.ModDecay.Multipliers) == 8 {" stuff to cope with old drum presets that are mysteriously
		// missing certain parameter blocks

		csvData = append(csvData, fmt.Sprintf("%v", data.MacroDefaults[macroIndex]))
		csvData = append(csvData, fmt.Sprintf("%v", data.Params.Levels.Multipliers[macroIndex]))
		csvData = append(csvData, fmt.Sprintf("%v", data.Params.Tuning.Multipliers[macroIndex]))
		csvData = append(csvData, fmt.Sprintf("%v", data.Params.Pitch.Multipliers[macroIndex]))
		csvData = append(csvData, fmt.Sprintf("%v", data.Params.PitchEnv.Multipliers[macroIndex]))
		csvData = append(csvData, fmt.Sprintf("%v", data.Params.Reverb.Multipliers[macroIndex]))
		csvData = append(csvData, fmt.Sprintf("%v", data.Params.ModAttack.Multipliers[macroIndex]))
		if len(data.Params.ModDecay.Multipliers) == 8 {
			csvData = append(csvData, fmt.Sprintf("%v", data.Params.ModDecay.Multipliers[macroIndex]))
		} else {
			csvData = append(csvData, "0000")
		}
		if len(data.Params.ModSustain.Multipliers) == 8 {
			csvData = append(csvData, fmt.Sprintf("%v", data.Params.ModSustain.Multipliers[macroIndex]))
		} else {
			csvData = append(csvData, "0000")
		}
		csvData = append(csvData, fmt.Sprintf("%v", data.Params.ModRelease.Multipliers[macroIndex]))
		csvData = append(csvData, fmt.Sprintf("%v", data.Params.AmpAttack.Multipliers[macroIndex]))
		if len(data.Params.AmpSustain.Multipliers) == 8 {
			csvData = append(csvData, fmt.Sprintf("%v", data.Params.AmpSustain.Multipliers[macroIndex]))
		} else {
			csvData = append(csvData, "0000")
		}
		if len(data.Params.AmpDecay.Multipliers) == 8 {
			csvData = append(csvData, fmt.Sprintf("%v", data.Params.AmpDecay.Multipliers[macroIndex]))
		} else {
			csvData = append(csvData, "0000")
		}
		csvData = append(csvData, fmt.Sprintf("%v", data.Params.AmpRelease.Multipliers[macroIndex]))
		csvData = append(csvData, fmt.Sprintf("%v", data.FilterType))
		csvData = append(csvData, fmt.Sprintf("%v", data.Params.FilterCutoff.Multipliers[macroIndex]))
		csvData = append(csvData, fmt.Sprintf("%v", data.Params.FilterEnvAmount.Multipliers[macroIndex]))
		csvData = append(csvData, fmt.Sprintf("%v", data.Params.FilterResonance.Multipliers[macroIndex]))

		macroMap[macroMapName].csvWriter.Write(csvData)
	}
	return nil
}
