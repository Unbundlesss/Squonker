package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/Unbundlesss/squonker/data"
	"github.com/spf13/cobra"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const (
	cMetadataRoot = "metadata"
	cOutputRoot   = "output"
)

var designToCompile = ""
var packToCompile = ""
var compileVerbose = false
var autoDeployIOS = false
var autoDeployOSX = false

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile an audio design ready for deployment",
	Run: func(cmd *cobra.Command, args []string) {

		sc := stateCache{}
		sc.wave = NewWaveStateCache()

		runCompiler(sc)
	},
}

type stateCache struct {
	wave *waveCache
}

func init() {
	rootCmd.AddCommand(compileCmd)

	compileCmd.Flags().StringVarP(&designToCompile, "design", "d", "", "Root directory of soundpack design")
	compileCmd.MarkFlagRequired("design")

	compileCmd.Flags().StringVarP(&packToCompile, "pack", "p", "", "Only compile the given pack by name")

	compileCmd.Flags().BoolVarP(&compileVerbose, "verbose", "v", false, "Print out more logging")

	compileCmd.Flags().BoolVar(&autoDeployIOS, "ios", false, "auto-deploy compiled instruments to iOS")
	compileCmd.Flags().BoolVar(&autoDeployOSX, "osx", false, "auto-deploy compiled instruments to OSX Studio")
}

func loadNoteTemplate(templateName string) (data.PresetTemplate, string, error) {
	// check if the template exists in the designToCompile folder first
	templatePath := path.Join(designToCompile, "metadata", "templates", "notes_template."+templateName+".json")
	if _, err := os.Stat(templatePath); err == nil {
		return data.LoadPresetTemplate(templatePath)
	}

	return data.LoadPresetTemplate(path.Join(cMetadataRoot, "templates", "notes_template."+templateName+".json"))
}

func loadDrumTemplate(templateName string) (data.PresetTemplate, string, error) {
	// check if the template exists in the designToCompile folder first
	templatePath := path.Join(designToCompile, "metadata", "templates", "drum_template."+templateName+".json")
	if _, err := os.Stat(templatePath); err == nil {
		return data.LoadPresetTemplate(templatePath)
	}

	return data.LoadPresetTemplate(path.Join(cMetadataRoot, "templates", "drum_template."+templateName+".json"))
}

// load presets.txt file, one name per line, add into a map
// we can check new names against the Endlesss standard presets to try and avoid duplicates
func loadPresetList() (map[string]string, error) {
	file, err := ioutil.ReadFile(path.Join(cMetadataRoot, "presets.txt"))
	if err != nil {
		return nil, err
	}
	presets := make(map[string]string)
	for _, line := range strings.Split(string(file), "\n") {
		if line != "" {
			lowerLine := strings.TrimSpace(strings.ToLower(line))
			presets[lowerLine] = line
		}
	}
	return presets, nil
}

// copy packImage to output if it exists
func copyPackImageIfPresent(packName, designName string) error {
	packImageJpeg := packName + "_2x.jpg"
	packImage := path.Join(designName, "images", packImageJpeg)
	if _, err := os.Stat(packImage); err == nil {
		_, err = localFileCopy(packImage, path.Join(cOutputRoot, packImageJpeg))
		if err != nil {
			return err
		}
	}
	return nil
}

func runCompiler(sc stateCache) {

	inputData, err := data.LoadInputData(designToCompile)
	if err != nil {
		sqFatal(err)
	}

	existingPresetNames, err := loadPresetList()
	if err != nil {
		sqFatal(err)
	}

	sqlog.Print("Building pack by ", inputData.Author)

	if _, err = os.Stat(cOutputRoot); os.IsNotExist(err) {
		os.MkdirAll(cOutputRoot, 0777)
	}

	warningsGenerated := 0
	warningsGenerated += drumkitCompiler(sc, inputData, existingPresetNames)
	warningsGenerated += notesCompiler(sc, inputData, existingPresetNames)

	if warningsGenerated > 0 {
		sqlog.Print(styleFailure.Render("Please check the log, there were compilation issues"))
	} else {
		sqComplete()
	}
}

func handlePackCompletion(packName string) {
	if autoDeployIOS {
		IOSDeployPack(packName, false)
	}
	if autoDeployOSX {
		OSXDeployPack(packName, false)
	}
	runtime.GC()
}

func notesCompiler(sc stateCache, inputData data.DataInput, existingPresetNames map[string]string) int {
	warningsGenerated := 0
	for _, notedef := range inputData.Notes {

		if len(packToCompile) > 0 && notedef.Name != packToCompile {
			sqlog.Print("Ignoring : ", notedef.Name)
			continue
		}

		sqlog.Print("============================================================")
		sqlog.Print("Note Pack : ", notedef.Name)

		err := copyPackImageIfPresent(notedef.Name, designToCompile)
		if err != nil {
			sqFatal(err)
		}

		inputDir := path.Join(designToCompile, "inputs", notedef.Directory)
		sqlog.Print("                  files : ", inputDir)

		// error if the path does not exist
		if _, err = os.Stat(inputDir); os.IsNotExist(err) {
			sqFatal(err)
		}

		presetIndex := 0

		// iterate all directories in the directory
		presetDirs, err := ioutil.ReadDir(inputDir)
		if err != nil {
			sqFatal(err)
		}
		for _, presetDir := range presetDirs {

			// ignore dot-prefix files and dirs, like .DS_Store et al
			if strings.HasPrefix(presetDir.Name(), ".") {
				continue
			}

			// this isn't a directory, skip it
			if !presetDir.IsDir() {
				sqlog.Print("               skipping : ", styleFailure.Render(presetDir.Name()))
				warningsGenerated++
				continue
			}
			sqlog.Print(styleNotice.Render(" - - - - - - - - - - - -  "), presetDir.Name())

			if ok, whyBad := isValidInstrumentName(presetDir.Name()); !ok {
				sqFatal(fmt.Errorf("%s : %s", styleFailure.Render(presetDir.Name()), whyBad))
			}

			presetName := presetDir.Name()
			presetNameLower := strings.ToLower(presetName)

			// check that this preset name doesn't collide
			if _, ok := existingPresetNames[presetNameLower]; ok {
				sqFatal("Halting; preset has conflicting / duplicate name : ", presetName)
			}
			// add to the list to keep track of duplicates
			existingPresetNames[presetNameLower] = presetName

			// ensure the output directory exists
			oggOutputPath := path.Join(cOutputRoot, notedef.Name, presetName, "Samples")
			if _, err = os.Stat(oggOutputPath); os.IsNotExist(err) {
				os.MkdirAll(oggOutputPath, 0777)
			}

			// get filename without extension
			oggFilename := presetName + ".ogg"
			oggOutput := path.Join(oggOutputPath, oggFilename)

			// grab the WAV(s) inside, ignoring everything else
			wavRootPath := path.Join(inputDir, presetDir.Name())
			wavFile := []string{}
			err = filepath.WalkDir(wavRootPath, func(s string, d fs.DirEntry, e error) error {
				if e != nil {
					return e
				}
				if strings.ToLower(filepath.Ext(d.Name())) == ".wav" {
					wavFile = append(wavFile, d.Name())
				}
				return nil
			})
			if err != nil {
				sqFatal(err)
			}
			// we only want one WAV please thank you
			if len(wavFile) != 1 {
				sqlog.Print("               skipping : ", styleFailure.Render(presetDir.Name()))
				sqlog.Print("                        : not enough / too many WAV files (", len(wavFile), ")")
				warningsGenerated++
				continue
			}

			filePathRelative := path.Join(wavRootPath, wavFile[0])

			wavState, wavStateFromCache, err := sc.wave.GetWavFileState(filePathRelative)
			if err != nil {
				sqFatal("GetWavFileState: ", err)
			}
			sqlog.Print("                        : ", wavFile[0], " : ", wavState.SampleRate, "Hz, ", wavState.ChannelCount, "ch, ", wavState.BitDepth, "b ", "[ cache:", wavStateFromCache, " H:", wavState.WaveHash[:8], "... ]")

			if wavState.LoopPointsValid {
				sqlog.Print("             loop start : ", wavState.LoopPointStart)
				sqlog.Print("          loop \"length\" : ", wavState.LoopPointLength)
				sqlog.Print("          total samples : ", wavState.SampleCount)
			}

			needsConversionToStereo := wavState.ChannelCount == 1

			// check mods to see if we override the template to use
			presetTemplate := notedef.Template
			for _, modPotential := range notedef.Mods {
				if strings.EqualFold(modPotential.Preset, presetName) {
					if modPotential.Template != "" {
						presetTemplate = modPotential.Template
						break
					}
				}
			}

			// load the chosen template
			inputTemplate, templateFilePath, err := loadNoteTemplate(presetTemplate)
			if err != nil {
				sqFatal(err)
			}
			sqlog.Print("               template : ", inputTemplate.Name, " (", templateFilePath, ")")

			volumeBoost := notedef.VolumeBoost

			// apply any other mod items to the template
			for _, modPotential := range notedef.Mods {

				// 'preset' can be a regex string, so we can match everything if required
				match, err := regexp.MatchString(modPotential.Preset, presetName)
				if err != nil {
					sqFatal(err)
				}

				if match {
					// go apply macro mods to the template
					if len(modPotential.Overrides) > 0 {
						for _, mod := range modPotential.Overrides {
							sqlog.Print("              macro mod : ", mod.Macro, " : ", mod.Override)

							// find the macro in the template
							macroIndex := indexOfCaseInvariant(mod.Macro, inputTemplate.MacroNames)
							if macroIndex < 0 {
								sqlog.Print(styleFailure.Render("macro not found in template: " + mod.Macro))
								warningsGenerated++
								continue
							}

							// patch the default
							inputTemplate.MacroDefaults[macroIndex] = mod.Override
						}
					}
				}
			}

			ffmpegArgMap := ffmpeg.KwArgs{
				"c:a": "libvorbis",
				"q":   "10",
			}
			if !wavState.LoopPointsValid {
				ffmpegArgMap["ar"] = "48000"
			}
			if needsConversionToStereo {
				ffmpegArgMap["ac"] = "2"
				volumeBoost += 3
			}
			ffmpegArgMap["filter:a"] = fmt.Sprintf("volume=%fdB", volumeBoost)
			if volumeBoost != 0 {
				sqlog.Print("                    -=> : ", ffmpegArgMap["filter:a"])
			}

			sqlog.Print("                    -=> : ", oggOutput)

			// launch ffmpeg task to do conversion
			ffmpegStream := ffmpeg.Input(filePathRelative).
				Output(oggOutput, ffmpegArgMap).
				OverWriteOutput()
			if compileVerbose {
				ffmpegStream.ErrorToStdOut() // optional barf out ffmpeg's progress to stdout
			}
			err = ffmpegStream.Run()
			if err != nil {
				sqFatal("FFMPEG: ", err)
			}

			inputTemplate.Author = inputData.Author
			inputTemplate.UserID = inputData.Author
			inputTemplate.Name = presetName
			inputTemplate.Type = presetName
			inputTemplate.Description = notedef.Note
			inputTemplate.PackName = notedef.Name
			inputTemplate.PackSortOrder = presetIndex

			inputTemplate.LoopStart = wavState.LoopPointStart
			inputTemplate.LoopLength = wavState.LoopPointLength
			inputTemplate.LoopingMode = wavState.LoopPointsValid

			// see if this note is present in the Order array, and if so, use its position there as the pack sort order instead
			customPresetIndex := indexOfCaseInvariant(presetName, notedef.Order)
			if customPresetIndex != -1 {
				inputTemplate.PackSortOrder = customPresetIndex
				sqlog.Print("         pack order set : ", inputTemplate.PackSortOrder)
			} else {
				// otherwise just increment the existing order
				presetIndex++
			}

			inputTemplate.Samples = []string{presetName}

			outputDescriptorFilename := path.Join(cOutputRoot, notedef.Name, presetName, presetName+".json")

			file, err := json.MarshalIndent(inputTemplate, "", " ")
			if err != nil {
				sqFatal(err)
			}
			err = ioutil.WriteFile(outputDescriptorFilename, file, 0644)
			if err != nil {
				sqFatal(err)
			}

			sqlog.Print("")
		}
		handlePackCompletion(notedef.Name)
	}
	return warningsGenerated
}

func drumkitCompiler(sc stateCache, inputData data.DataInput, existingPresetNames map[string]string) int {
	warningsGenerated := 0
	for _, drumdef := range inputData.Drumkits {

		if len(packToCompile) > 0 && drumdef.Name != packToCompile {
			sqlog.Print("Ignoring : ", drumdef.Name)
			continue
		}

		runtime.GC()

		sqlog.Print("============================================================")
		sqlog.Print("Drum Pack : ", drumdef.Name)

		err := copyPackImageIfPresent(drumdef.Name, designToCompile)
		if err != nil {
			sqFatal(err)
		}

		drumRootDir := path.Join(designToCompile, "inputs", drumdef.Directory)
		sqlog.Print("             files : ", drumRootDir)

		presetIndex := 0

		for _, drumPreset := range drumdef.Presets {
			presetRootDir := path.Join(drumRootDir, drumPreset.Directory)

			sqlog.Print(styleNotice.Render(" - - - - - - - - - - - -  "), drumPreset.Name)

			if ok, whyBad := isValidInstrumentName(drumPreset.Name); !ok {
				sqFatal(fmt.Errorf("%s : %s", styleFailure.Render(drumPreset.Name), whyBad))
			}

			// error if the path does not exist
			if _, err := os.Stat(presetRootDir); os.IsNotExist(err) {
				sqFatal(err)
			}

			// ensure the output directory exists
			oggOutputPath := path.Join(cOutputRoot, drumdef.Name, drumPreset.Name, "Samples")
			if _, err := os.Stat(oggOutputPath); os.IsNotExist(err) {
				os.MkdirAll(oggOutputPath, 0777)
			}

			// use the top-level template, or use a per-preset one if present
			presetTemplate := drumdef.Template
			if drumPreset.Template != "" {
				presetTemplate = drumPreset.Template
				sqlog.Print("           template set : ", presetTemplate)
			}

			// load the chosen template, fresh each time in case previous runs had changed anything
			inputTemplate, templateFilePath, err := loadDrumTemplate(presetTemplate)
			if err != nil {
				sqFatal(err)
			}
			sqlog.Print("               template : ", inputTemplate.Name, " (", templateFilePath, ")")

			sampleIndex := 0
			inputTemplate.Samples = []string{}
			inputTemplate.TriggerIcons = []string{}

			wavSamples, err := ioutil.ReadDir(presetRootDir)
			if err != nil {
				sqFatal(err)
			}
			for _, wavSample := range wavSamples {
				wavRootPath := path.Join(presetRootDir, wavSample.Name())

				wavState, wavStateFromCache, err := sc.wave.GetWavFileState(wavRootPath)
				if err != nil {
					sqFatal("GetWavFileState: ", err)
				}
				sqlog.Print("                        : ", wavSample.Name(), " : ", wavState.SampleRate, "Hz, ", wavState.ChannelCount, "ch, ", wavState.BitDepth, "b ", "[ cache:", wavStateFromCache, " H:", wavState.WaveHash[:8], "... ]")

				if wavState.LoopPointsValid {
					sqlog.Print("                warning : loop points ignored on drum samples")
				}
				needsConversionToStereo := wavState.ChannelCount == 1

				// figure a trigger icon out; iterate the IconMap and match against the wavSample name
				chosenTriggerIcon := "Zap"
				for _, icon := range inputData.Iconmap {
					// prefer starts-with-regex first, then anywhere-in-string
					for _, possibleRegex := range icon.Regex {
						if len(possibleRegex) > 0 && strings.HasPrefix(wavSample.Name(), possibleRegex) {
							chosenTriggerIcon = icon.Svg
						}
					}
					for _, possibleRegex := range icon.Regex {
						if len(possibleRegex) > 0 && strings.Contains(wavSample.Name(), possibleRegex) {
							chosenTriggerIcon = icon.Svg
						}
					}
				}
				inputTemplate.TriggerIcons = append(inputTemplate.TriggerIcons, chosenTriggerIcon)

				// drum samples are just named as 1.ogg, 2.ogg, etc to allow easier replacement on-device
				sampleIndex++
				oggFilename := fmt.Sprintf("%d.ogg", sampleIndex)
				inputTemplate.Samples = append(inputTemplate.Samples, fmt.Sprintf("%d", sampleIndex))

				oggOutput := path.Join(oggOutputPath, oggFilename)
				sqlog.Print("                    -=> : pad icon [", chosenTriggerIcon, "], ", oggOutput)

				volumeBoost := drumdef.VolumeBoost

				ffmpegArgMap := ffmpeg.KwArgs{
					"c:a": "libvorbis",
					"q":   "10",
					"ar":  "48k",
				}
				if needsConversionToStereo {
					ffmpegArgMap["ac"] = "2"
					volumeBoost += 3
				}
				ffmpegArgMap["filter:a"] = fmt.Sprintf("volume=%fdB", volumeBoost)
				if volumeBoost != 0 {
					sqlog.Print("                    -=> : ", ffmpegArgMap["filter:a"])
				}

				// launch ffmpeg task to do conversion
				ffmpegStream := ffmpeg.Input(wavRootPath).
					Output(oggOutput, ffmpegArgMap).
					OverWriteOutput()
				if compileVerbose {
					ffmpegStream.ErrorToStdOut() // optional barf out ffmpeg's progress to stdout
				}
				err = ffmpegStream.Run()
				if err != nil {
					sqFatal(err)
				}

				sqlog.Print("")
			}

			// go apply macro mods to the template
			if len(drumPreset.Overrides) > 0 {
				sqlog.Print("                        : applying template modifications")

				for _, mod := range drumPreset.Overrides {
					sqlog.Print("              macro mod : ", mod.Macro, " : ", mod.Override)

					// find the macro in the template
					macroIndex := indexOfCaseInvariant(mod.Macro, inputTemplate.MacroNames)
					if macroIndex < 0 {
						sqlog.Print(styleFailure.Render("macro not found in template: " + mod.Macro))
						warningsGenerated++
						continue
					}

					// patch the default
					inputTemplate.MacroDefaults[macroIndex] = mod.Override
				}
			}

			inputTemplate.Author = inputData.Author
			inputTemplate.UserID = inputData.Author
			inputTemplate.Name = drumPreset.Name
			inputTemplate.Type = drumPreset.Name
			inputTemplate.Description = drumPreset.Note
			inputTemplate.PackName = drumdef.Name
			inputTemplate.PackSortOrder = presetIndex

			// no looping in a drum-hole, captain
			inputTemplate.LoopLength = 1.0
			inputTemplate.LoopStart = 0
			inputTemplate.LoopingMode = false

			presetIndex++

			outputDescriptorFilename := path.Join(cOutputRoot, inputTemplate.PackName, inputTemplate.Name, inputTemplate.Name+".json")

			file, err := json.MarshalIndent(inputTemplate, "", " ")
			if err != nil {
				sqFatal(err)
			}
			err = ioutil.WriteFile(outputDescriptorFilename, file, 0644)
			if err != nil {
				sqFatal(err)
			}
		}

		handlePackCompletion(drumdef.Name)
	}
	return warningsGenerated
}
