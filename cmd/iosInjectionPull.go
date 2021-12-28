package cmd

import (
	"os"
	"path"

	"github.com/ishani/gowebdav"
	"github.com/spf13/cobra"
)

const (
	cInjectionSamples = "injection_sample"
	cInjectionDB      = "injection_db"
)

func fileRemoteToLocal(remoteFile, localFile string, client *gowebdav.Client) error {

	bytes, err := client.Read(remoteFile)
	if err != nil {
		return err
	}

	// write bytes array to localFile
	file, err := os.OpenFile(
		localFile,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

//
var iosInjectionPullCmd = &cobra.Command{
	Use:   "ios-injection-pull",
	Short: "Fetch the new preset storage database from a 1.3.x Endlesss install",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		deviceIDs, err := LoadIOSDeviceIdsFromConfig()
		if err != nil {
			sqFatal(err)
		}

		// connect to the iOS device
		c := gowebdav.NewClient(deviceIDs.webDAV, "", "")
		err = c.Connect()
		if err != nil {
			sqFatal(err)
		}

		remotePresetDir := path.Join(iOSPathDataContainer, deviceIDs.dataGUID, iOSPathEndlesssDataRoot, iOSPresetDatabaseDir, iOSPresetDatabaseName)

		if _, err = os.Stat(cInjectionDB); os.IsNotExist(err) {
			os.MkdirAll(cInjectionDB, 0777)
		}

		localPresetDir := path.Join(cInjectionDB, iOSPresetDatabaseName)
		if _, err = os.Stat(localPresetDir); os.IsNotExist(err) {
			os.MkdirAll(localPresetDir, 0777)
		}

		err = fileRemoteToLocal(path.Join(remotePresetDir, iOSPresetDatabase_1), path.Join(localPresetDir, iOSPresetDatabase_1), c)
		if err != nil {
			sqFatal(err)
		}
		err = fileRemoteToLocal(path.Join(remotePresetDir, iOSPresetDatabase_2), path.Join(localPresetDir, iOSPresetDatabase_2), c)
		if err != nil {
			sqFatal(err)
		}
		err = fileRemoteToLocal(path.Join(remotePresetDir, iOSPresetDatabase_3), path.Join(localPresetDir, iOSPresetDatabase_3), c)
		if err != nil {
			sqFatal(err)
		}

		sqComplete()
	},
}

func init() {
	rootCmd.AddCommand(iosInjectionPullCmd)
}
