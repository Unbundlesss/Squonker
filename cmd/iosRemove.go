package cmd

import (
	"github.com/spf13/cobra"
)

var iosPresetsToRemove []string

var iosRemoveCmd = &cobra.Command{
	Use:   "ios-remove",
	Short: "Cleanly remove a named instrument from an iOS device",
	Long:  `Pass instrument names and all the files inside their directory on the configured iOS device will be removed`,
	Run: func(cmd *cobra.Command, args []string) {

		// get configured iOS device or report the need to run ios-discover
		deviceIDs, err := LoadIOSDeviceIdsFromConfig()
		if err != nil {
			sqFatal(err)
		}

		// for each -i <Instrument>, go run a removal task
		for _, inst := range iosPresetsToRemove {
			IOSRemoveInstrumentFromDevice(deviceIDs, inst)
		}

		sqComplete()
	},
}

func init() {
	rootCmd.AddCommand(iosRemoveCmd)

	iosRemoveCmd.Flags().StringArrayVarP(&iosPresetsToRemove, "inst", "i", []string{}, "Instrument presets to remove from iOS device")
	iosRemoveCmd.MarkFlagRequired("presets")
}
