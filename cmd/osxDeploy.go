package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/Unbundlesss/squonker/data"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var osxPackToDeploy string
var osxDesignToDeploy string
var osxCleanBeforeDeploy bool

var osxDeployCmd = &cobra.Command{
	Use:   "osx-deploy",
	Short: "Install an single pack or all packs from a design to a local install of Studio",
	Long:  `Install an single pack or all packs from a design to a local install of Studio`,
	Run: func(cmd *cobra.Command, args []string) {

		deployPack := len(osxPackToDeploy) > 0
		deployDesign := len(osxDesignToDeploy) > 0
		if deployPack && deployDesign {
			cmd.Help()
			sqFatal("You must specify either a pack (-p) OR design (-d) to deploy, not both")
		}

		if deployPack {
			OSXDeployPack(osxPackToDeploy, osxCleanBeforeDeploy)
		} else if deployDesign {
			OSXDeployDesign(osxDesignToDeploy, osxCleanBeforeDeploy)
		} else {
			cmd.Help()
			os.Exit(0)
		}

		sqComplete()
	},
}

func OSXDeployPack(packName string, cleanPresets bool) {

	packRootPath := path.Join("output", packName)
	if _, err := os.Stat(packRootPath); os.IsNotExist(err) {
		sqFatal("Cannot find pack directory : ", packRootPath)
	}

	osxInstrumentDirectory, err := homedir.Expand(osxPathEndlesssInstruments)
	if err != nil {
		sqFatal(err)
	}

	weAreRoot := isRunningAsRoot()

	packImageJpeg := packName + "_2x.jpg"
	packImage := path.Join("output", packImageJpeg)
	// only copy if the pack image exists in /output
	if _, err := os.Stat(packImage); err == nil {
		// .. and if we have a hope of actually copying it - need to have admin rights to
		// start blitting files into app containers
		if weAreRoot {
			installedPackImage := path.Join(osxPathEndlesssImagePack, packImageJpeg)
			_, err = localFileCopy(packImage, installedPackImage)
			if err != nil {
				sqFatal(err)
			}
		} else {
			sqlog.Print("Pack images only copied if run as root")
		}
	}

	// only copy instruments around if not root to avoid spilling weird file permissions everywhere
	if !weAreRoot {

		// iterate all instruments to deploy in root path
		instrumentDirs, err := ioutil.ReadDir(packRootPath)
		if err != nil {
			sqFatal(err)
		}

		if cleanPresets {
			for _, instrumentDir := range instrumentDirs {
				instName := instrumentDir.Name()
				err = OSXRemoveInstrumentFromDevice(instName)
				if err != nil {
					sqlog.Print("        could not clean : ", instName)
				}
			}
		}

		for _, instDir := range instrumentDirs {

			instName := instDir.Name()

			sqlog.Print("              deploying : ", instName)

			installedInstrumentDir := path.Join(osxInstrumentDirectory, instName)
			os.MkdirAll(installedInstrumentDir, 0755)

			installedInstrumentSamplesDir := path.Join(installedInstrumentDir, "Samples")
			os.MkdirAll(installedInstrumentSamplesDir, 0755)

			packFiles := recursiveGatherFilesFromDir(path.Join(packRootPath, instName))
			for _, packFile := range packFiles {
				localFile := packFile
				installedFile := strings.Replace(packFile, packRootPath, osxInstrumentDirectory, 1)

				//sqlog.Print("              from : ", localFile)
				//sqlog.Print("              to : ", remoteFile)

				// read localFile as binary byte array
				localFileByteArray, err := ioutil.ReadFile(localFile)
				if err != nil {
					sqFatal(err)
				}

				err = os.WriteFile(installedFile, localFileByteArray, os.FileMode(0777))
				if err != nil {
					sqFatal(err)
				}
			}
		}
	} else {
		sqlog.Print("Running as root, not copying instruments")
	}
}

func OSXDeployDesign(designName string, cleanPresets bool) {

	sqlog.Print("       examining design : ", designName)
	inputData, err := data.LoadInputData(designName)
	if err != nil {
		sqFatal(err)
	}

	// iterate all notes and drumkits to deploy
	for _, notePack := range inputData.Notes {
		OSXDeployPack(notePack.Name, cleanPresets)
	}
	for _, drumPack := range inputData.Drumkits {
		OSXDeployPack(drumPack.Name, cleanPresets)
	}
}

func init() {
	// conditionally registered in osxInit.go

	osxDeployCmd.Flags().StringVarP(&osxPackToDeploy, "pack", "p", "", "Name of soundpack to deploy to Studio")
	osxDeployCmd.Flags().StringVarP(&osxDesignToDeploy, "design", "d", "", "Name of design to deploy to Studio")

	osxDeployCmd.Flags().BoolVar(&osxCleanBeforeDeploy, "clean", false, "Remove instruments from this pack before copying new data")
}
