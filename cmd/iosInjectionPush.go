package cmd

import (
	"github.com/ishani/gowebdav"
	"github.com/spf13/cobra"
)

//
var iosInjectionPushCmd = &cobra.Command{
	Use:   "ios-injection-push",
	Short: "",
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

		sqComplete()
	},
}

func init() {
	rootCmd.AddCommand(iosInjectionPullCmd)
}
