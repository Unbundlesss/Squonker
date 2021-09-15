package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ishani/gowebdav"
	"github.com/spf13/cobra"
)

var iosPackToDeploy string
var iosCleanPresetsBeforeDeploy bool

func fileCopyRemote(localFile, remoteFile string, client *gowebdav.Client) error {
	// read localFile as binary byte array
	localFileByteArray, err := ioutil.ReadFile(localFile)
	if err != nil {
		return err
	}

	retries := 0
	for ; retries < 5; retries++ {
		err = client.Write(remoteFile, localFileByteArray, os.FileMode(0777))
		if err != nil {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	if retries == 5 {
		return fmt.Errorf("failed to copy")
	}
	return nil
}

var iosDeployCmd = &cobra.Command{
	Use:   "ios-deploy",
	Short: "Install an entire pack to a device",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		IOSDeployPack(iosPackToDeploy, iosCleanPresetsBeforeDeploy)
		sqComplete()
	},
}

func IOSDeployPack(packName string, cleanPresets bool) {

	deviceIDs, err := LoadIOSDeviceIdsFromConfig()
	if err != nil {
		sqFatal(err)
	}
	if packName == "" {
		sqlog.Println("You must specify a pack to deploy")
		return
	}

	sqlog.Println("[ Ensure your iOS device isn't sleeping ]")

	c := gowebdav.NewClient(deviceIDs.webDAV, "", "")
	err = c.Connect()
	if err != nil {
		sqFatal(err)
	}

	packRootPath := path.Join("output", packName)
	if _, err := os.Stat(packRootPath); os.IsNotExist(err) {
		sqFatal("Cannot find pack directory : ", packRootPath)
	}

	remoteInstallDir := path.Join(iOSPathDataContainer, deviceIDs.dataGUID, iOSPathEndlesssInstruments)

	// nice app container you have there shame if we installed new pack images into it
	packImageJpeg := packName + "_2x.jpg"
	packImage := path.Join("output", packImageJpeg)
	// only copy if the pack image exists in /output
	if _, err := os.Stat(packImage); err == nil {
		installedPackImage := path.Join(iOSPathAppContainer, deviceIDs.appGUID, iOSPathEndlesssImagePack, packImageJpeg)
		err = fileCopyRemote(packImage, installedPackImage, c)
		if err != nil {
			sqlog.Print("                unable to copy pack image")
		}
	}

	// iterate all instruments to deploy in root path
	instrumentDirs, err := ioutil.ReadDir(packRootPath)
	if err != nil {
		sqFatal(err)
	}

	if cleanPresets {
		for _, instrumentDir := range instrumentDirs {
			instName := instrumentDir.Name()
			err = IOSRemoveInstrumentFromDevice(deviceIDs, instName)
			if err != nil {
				sqlog.Print("        could not clean : ", instName)
			}
		}

		sqlog.Print("                pausing ... ")
		time.Sleep(3 * time.Second)
	}

	for _, instDir := range instrumentDirs {

		instrumentName := instDir.Name()

		sqlog.Print("              deploying : ", instrumentName)

		remoteInstrumentDir := path.Join(iOSPathDataContainer, deviceIDs.dataGUID, iOSPathEndlesssInstruments, instrumentName)
		c.MkdirAll(remoteInstrumentDir, 0755)

		remoteInstrumentSamplesDir := path.Join(remoteInstrumentDir, "Samples")
		c.MkdirAll(remoteInstrumentSamplesDir, 0755)

		packFiles := recursiveGatherFilesFromDir(path.Join(packRootPath, instrumentName))
		for _, packFile := range packFiles {

			localFile := packFile
			remoteFile := strings.Replace(packFile, packRootPath, remoteInstallDir, 1)
			//sqlog.Print("                from : ", localFile)
			//sqlog.Print("                  to : ", remoteFile)

			//sqlog.Print("                  xx : ", path.Dir(remoteFile))
			//c.MkdirAll(path.Dir(remoteFile), 0755)

			err = fileCopyRemote(localFile, remoteFile, c)
			if err != nil {
				sqlog.Print("        could not copy : ", remoteFile)
			}

		}
	}

	sqlog.Println("[ ... don't forget to reboot Endlesss! ]")
}

func init() {
	rootCmd.AddCommand(iosDeployCmd)

	iosDeployCmd.Flags().StringVarP(&iosPackToDeploy, "pack", "p", "", "Name of soundpack to deploy to iOS device")
	iosDeployCmd.MarkFlagRequired("pack")

	iosDeployCmd.Flags().BoolVar(&iosCleanPresetsBeforeDeploy, "clean", false, "Remove preset before copying new data")
}
