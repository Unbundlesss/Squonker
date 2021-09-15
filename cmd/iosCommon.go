package cmd

import (
	"fmt"
	"path"

	"github.com/ishani/gowebdav"
	"github.com/spf13/viper"
)

// a structure with iOS device ids
type IOSDeviceIds struct {
	appGUID  string
	dataGUID string
	webDAV   string
}

// look up device ids from viper configuration and return them, or error if not found
func LoadIOSDeviceIdsFromConfig() (IOSDeviceIds, error) {

	appGUID := viper.GetString(cfgAppGUID)
	if appGUID == "" {
		return IOSDeviceIds{}, fmt.Errorf("cannot find iOS app GUID, please run ios-discover first")
	}

	dataGUID := viper.GetString(cfgDataGUID)
	if dataGUID == "" {
		return IOSDeviceIds{}, fmt.Errorf("cannot find iOS data GUID, please run ios-discover first")
	}

	webDAV := viper.GetString(cfgWebDAV)
	if webDAV == "" {
		return IOSDeviceIds{}, fmt.Errorf("cannot find iOS webDAV address, please run ios-discover first")
	}

	return IOSDeviceIds{appGUID, dataGUID, webDAV}, nil
}

func IOSRecursiveWalk(client *gowebdav.Client, path string) ([]string, error) {

	var fileList []string

	files, err := client.ReadDir(path)
	if err != nil {
		return []string{}, err
	}
	for _, file := range files {
		filePath := path + "/" + file.Name()
		fileList = append(fileList, filePath)

		if file.IsDir() {
			recusiveFiles, err := IOSRecursiveWalk(client, filePath)
			if err != nil {
				return []string{}, err
			}

			fileList = append(fileList, recusiveFiles...)
		}
	}
	return fileList, nil
}

func IOSRemoveInstrumentFromDevice(deviceIDs IOSDeviceIds, instrumentName string) error {

	sqlog.Print("               deleting : ", instrumentName)
	instrumentDir := path.Join(iOSPathDataContainer, deviceIDs.dataGUID, iOSPathEndlesssInstruments, instrumentName)

	c := gowebdav.NewClient(deviceIDs.webDAV, "", "")
	err := c.Connect()
	if err != nil {
		return err
	}

	fileList, err := IOSRecursiveWalk(c, instrumentDir)
	if err != nil {
		return err
	}
	// prepend actual instrument dir
	fileList = append([]string{instrumentDir}, fileList...)

	for i := len(fileList) - 1; i >= 0; i-- {

		err = c.Remove(fileList[i])
		if err != nil {
			return err
		}
	}
	return nil
}
