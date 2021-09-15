package cmd

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/EndlesssTrackClub/squonker/data"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var osxPackToDecompile string

var osxDecompileCmd = &cobra.Command{
	Use:   "osx-decompile",
	Short: "Take a given soundpack from Studio and turn it into a Sqonker design definintion",
	Run: func(cmd *cobra.Command, args []string) {
		OSXDecompilePack(osxPackToDecompile)
	},
}

func OSXDecompilePack(packToDecompile string) {

	// go find all the instruments that belong to the given pack
	instruments, err := OSXFindInstrumentsInPack(packToDecompile)
	if err != nil {
		sqFatal(err)
	}

	for _, instrument := range instruments {
		sqlog.Print(instrument)
	}
}

// returns list of full paths to instruments in the given pack; to their .json preset, to be exact
func OSXFindInstrumentsInPack(packToDecompile string) ([]string, error) {

	osxInstrumentDirectory, err := homedir.Expand(osxPathEndlesssInstruments)
	if err != nil {
		sqFatal(err)
	}

	instruments := []string{}
	err = filepath.WalkDir(osxInstrumentDirectory, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if strings.ToLower(filepath.Ext(d.Name())) == ".json" {
			template, _, err := data.LoadPresetTemplate(s)
			if err != nil {
				return err
			}
			if template.PackName == packToDecompile {
				instruments = append(instruments, s)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return instruments, nil
}

func init() {
	// conditionally registered in osxInit.go

	osxDecompileCmd.Flags().StringVarP(&osxPackToDecompile, "pack", "p", "", "Name of soundpack to decompile to a squad")
	osxDecompileCmd.MarkFlagRequired("pack")
}
