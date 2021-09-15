package cmd

import (
	"path"

	"github.com/ishani/gowebdav"
	"github.com/spf13/cobra"
)

// NOTE removing built-in presets from /Instruments is useless unless you also remove the original
// 		.pack files, otherwise they will be unpacked again fresh when Endlesss boots
//		-- this function is mostly for tidying up / debugging on-device --
//
var iosObliterateCmd = &cobra.Command{
	Use:   "ios-obliterate",
	Short: "Empty the instrument cache on-device",
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

		// go fetch a list of all the presets in the instruments directory on the device
		remoteInstrumentDir := path.Join(iOSPathDataContainer, deviceIDs.dataGUID, iOSPathEndlesssInstruments)

		files, err := c.ReadDir(remoteInstrumentDir)
		if err != nil {
			sqFatal(err)
		}
		for _, file := range files {
			IOSRemoveInstrumentFromDevice(deviceIDs, file.Name())
		}

		sqComplete()
	},
}

func init() {
	rootCmd.AddCommand(iosObliterateCmd)
}
