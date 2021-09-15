package cmd

import (
	"io/ioutil"
	"os"
	"path"
)

func OSXRecursiveWalk(path string) ([]string, error) {

	var fileList []string

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return []string{}, err
	}
	for _, file := range files {
		filePath := path + "/" + file.Name()
		fileList = append(fileList, filePath)

		if file.IsDir() {
			recusiveFiles, err := OSXRecursiveWalk(filePath)
			if err != nil {
				return []string{}, err
			}

			fileList = append(fileList, recusiveFiles...)
		}
	}
	return fileList, nil
}

func OSXRemoveInstrumentFromDevice(instrumentName string) error {

	sqlog.Print("               deleting : ", instrumentName)
	instrumentDir := path.Join(osxPathEndlesssInstruments, instrumentName)

	fileList, err := OSXRecursiveWalk(instrumentDir)
	if err != nil {
		return err
	}
	// prepend actual preset dir
	fileList = append([]string{instrumentDir}, fileList...)

	for i := len(fileList) - 1; i >= 0; i-- {

		err = os.Remove(fileList[i])
		if err != nil {
			return err
		}
	}
	return nil
}
