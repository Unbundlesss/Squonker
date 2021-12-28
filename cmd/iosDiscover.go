package cmd

import (
	"log"
	"path"

	"github.com/ishani/gowebdav"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var webdavAddr string

var iosDiscoverCmd = &cobra.Command{
	Use:   "ios-discover",
	Short: "Map an iOS device to find Endlesss",
	Long:  `Given a jailbroken iOS device IP, go analyse it for Endlessian presence, stash results in config for use elsewhere`,

	Run: func(cmd *cobra.Command, args []string) {
		doIOSDiscover()
	},
}

func init() {
	rootCmd.AddCommand(iosDiscoverCmd)

	iosDiscoverCmd.PersistentFlags().StringVar(&webdavAddr, "webdav", "", "WebDAV address of iOS device")
	iosDiscoverCmd.MarkPersistentFlagRequired("webdav")
	viper.BindPFlag(cfgWebDAV, iosDiscoverCmd.PersistentFlags().Lookup("webdav"))
}

func doIOSDiscover() {

	c := gowebdav.NewClient(webdavAddr, "", "")
	err := c.Connect()
	if err != nil {
		sqFatal(err)
	}

	sqlog.Printf("Searching iOS device for Endlesss install GUIDs, please wait ...")

	// go look for a matching path pattern to discern which GUID represents the endlesss app
	files, err := c.ReadDir(iOSPathDataContainer)
	if err != nil {
		sqFatal(err)
	}
	for _, file := range files {
		testPathForEndlesss := path.Join(iOSPathDataContainer, file.Name(), iOSPathEndlesssDataRoot)
		fi, err := c.Stat(testPathForEndlesss)
		if (err == nil) &&
			(fi.IsDir()) {

			sqlog.Printf(" > Data GUID found : %s", file.Name())

			viper.Set(cfgDataGUID, file.Name())
			break
		}
	}

	// same again but looking for the app container
	files, err = c.ReadDir(iOSPathAppContainer)
	if err != nil {
		sqFatal(err)
	}
	for _, file := range files {
		testPathForEndlesss := path.Join(iOSPathAppContainer, file.Name(), iOSPathEndlesssApp)
		fi, err := c.Stat(testPathForEndlesss)
		if (err == nil) &&
			(fi.IsDir()) {

			sqlog.Printf(" >  App GUID found : %s", file.Name())

			viper.Set(cfgAppGUID, file.Name())
			break
		}
	}

	// save IP
	viper.Set(cfgWebDAV, webdavAddr)

	// write viper config
	err = viper.WriteConfig()
	if err != nil {
		log.Fatal(err)
	}

	sqlog.Printf("iOS device mapped, ready for use")
	sqComplete()
}
