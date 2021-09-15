package cmd

import (
	b64 "encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/user"
	"regexp"
	"strings"
	"syscall"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/term"

	"github.com/cespare/xxhash"
	"github.com/gtank/cryptopasta"
)

var sqlog = log.New(os.Stdout, "", 0)

func sqFatal(v ...interface{}) {
	sqlog.Fatal("Error : ", styleFailure.Render(fmt.Sprint(v...)))
}
func sqComplete() {
	sqlog.Print(styleSuccess.Render("[ Complete ]"))
}

const (
	iOSPathDataContainer       = "/var/mobile/Containers/Data/Application/"
	iOSPathAppContainer        = "/var/containers/Bundle/Application/"
	iOSPathEndlesssInstruments = "Library/Application Support/Endlesss/Presets/Instruments/"
	iOSPathEndlesssApp         = "Endlesss.app/Assets/"
	iOSPathEndlesssImagePack   = "Endlesss.app/Assets/Images/Packs/"
)
const (
	osxPathEndlesssInstruments = "~/Library/Containers/fm.endlesss.app/Data/Library/Application Support/Endlesss/Presets/Instruments/"
	osxPathEndlesssImagePack   = "/Applications/Endlesss.app/Contents/Resources/Assets/Images/Packs/"
)

func readPwdFromTerminal(secretName string) (string, error) {

	fmt.Print("Enter ", secretName, " : ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	password := string(bytePassword)
	return strings.TrimSpace(password), nil
}

func stringToCryptoKeyBytes(secretString string, keyLength int) ([]byte, error) {

	// hash the cypher input to use as a random seed
	rngSeed := xxhash.Sum64String(secretString)
	rand.Seed(int64(rngSeed))

	key := make([]byte, keyLength)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// scrobble a string into an encrypted lump, expressed as base64
func encryptStringToBase64(inputString, secretString string) (string, error) {

	randomKey, err := stringToCryptoKeyBytes(secretString, 32)
	if err != nil {
		return "", err
	}
	var randomKey32 [32]byte
	copy(randomKey32[:], randomKey)

	cryptedBytes, err := cryptopasta.Encrypt([]byte(inputString), &randomKey32)
	if err != nil {
		return "", err
	}

	return b64.StdEncoding.EncodeToString(cryptedBytes), nil
}

// unscrobble the result of encryptStringToBase64
func decryptStringFromBase64(inputString, secretString string) (string, error) {

	randomKey, err := stringToCryptoKeyBytes(secretString, 32)
	if err != nil {
		return "", err
	}
	var randomKey32 [32]byte
	copy(randomKey32[:], randomKey)

	cryptedBytes, err := b64.StdEncoding.DecodeString(inputString)
	if err != nil {
		return "", err
	}

	decryptedBytes, err := cryptopasta.Decrypt(cryptedBytes, &randomKey32)
	if err != nil {
		return "", err
	}

	return string(decryptedBytes), nil
}

func indexOf(word string, data []string) int {
	for k, v := range data {
		if word == v {
			return k
		}
	}
	return -1
}

func indexOfCaseInvariant(word string, data []string) int {
	for k, v := range data {
		if strings.EqualFold(word, v) {
			return k
		}
	}
	return -1
}

func localFileCopy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func recursiveGatherFilesFromDir(path string) []string {

	var fileList []string

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		filePath := path + "/" + file.Name()

		if !file.IsDir() {
			fileList = append(fileList, filePath)
		} else {
			fileList = append(fileList, recursiveGatherFilesFromDir(filePath)...)
		}
	}
	return fileList
}

func isRunningAsRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("[isRunningAsRoot] Unable to get current user: %s", err)
	}
	return currentUser.Username == "root"
}

func fastFileHash(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher, _ := blake2b.New256(nil)
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	hash := hasher.Sum(nil)
	encodedHex := hex.EncodeToString(hash[:])

	return encodedHex, nil
}

var validateInstrument = regexp.MustCompile(`^[a-zA-Z0-9 ]+$`).MatchString

func isValidInstrumentName(name string) (bool, string) {

	if len(name) > 14 {
		return false, "Instrument name is too long, keep it below 14 characters"
	}

	if !validateInstrument(name) {
		return false, "Instrument name should be alphanumeric only"
	}

	return true, ""
}
