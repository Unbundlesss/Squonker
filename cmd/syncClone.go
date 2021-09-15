package cmd

import (
	"encoding/hex"
	"fmt"
	"os"
	"regexp"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var syncAddName string
var syncAddGitAddr string
var syncAddGitAccessKey bool
var syncCopyGitAccessKeyFrom string

var IsValidSquadName = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

var syncAddCmd = &cobra.Command{
	Use:   "sync-clone",
	Short: "Clone a soundpack git repository",
	Run: func(cmd *cobra.Command, args []string) {

		secretPhrase := askForMasterSecret()

		// remove "squad" prefix from syncAddName if present
		if syncAddName != "" && len(syncAddName) > 6 {
			if syncAddName[:6] == "squad." {
				syncAddName = syncAddName[6:]
			}
		}
		if len(syncAddName) <= 1 {
			sqFatal("supplied name is not long enough; ", syncAddName)
		}
		if !IsValidSquadName(syncAddName) {
			sqFatal("supplied name must be letters and digits only; ", syncAddName)
		}

		// yaml has some restrictions on key name contents (eg. lowercase, no spaces, etc) so
		// smooch it into a hex string to store that instead
		syncKeyName := hex.EncodeToString([]byte(syncAddName))
		// sqlog.Print("Hashed sync name : ", syncKeyName)
		sqlog.Print("")

		cloneOptions := git.CloneOptions{
			URL:      syncAddGitAddr,
			Progress: os.Stdout,
		}

		if syncAddGitAccessKey {

			// if we already have a stashed access token, borrow it from a named sync and store it for this new one
			if syncCopyGitAccessKeyFrom != "" {

				gitKeys := viper.GetStringMap("git-keys")
				if gitKeys == nil {
					sqFatal("git-keys config not found")
				}

				// encode to hex string to retrieve from the existing list
				syncKeyNameHex := hex.EncodeToString([]byte(syncCopyGitAccessKeyFrom))

				if _, ok := gitKeys[syncKeyNameHex]; !ok {
					sqFatal("git-key not found: ", syncCopyGitAccessKeyFrom)
				}

				// decrypt it for use with the clone operation
				decryptedKey, err := decryptStringFromBase64(fmt.Sprintf("%v", gitKeys[syncKeyNameHex]), secretPhrase)
				if err != nil {
					sqFatal(err)
				}

				cloneOptions.Auth = &http.BasicAuth{
					Username: "null",
					Password: decryptedKey,
				}

				// .. and then just copy the already-encrypted key to the config
				gitKeys[syncKeyName] = gitKeys[syncKeyNameHex]
				viper.SetDefault("git-keys", gitKeys)
				viper.WriteConfig()

				sqlog.Println(styleSuccess.Render(fmt.Sprint("Access key copied from [", syncCopyGitAccessKeyFrom, "]")))

			} else {

				// ask user to provide access token
				sqlog.Print(styleNotice.Render("Please enter the access key to use when cloning private repositories"))
				gitAccessKey, err := readPwdFromTerminal("git-access-key")
				if err != nil {
					sqFatal(err)
				}
				sqlog.Print("")

				// stash it in clone options for immediate use
				cloneOptions.Auth = &http.BasicAuth{
					Username: "null",
					Password: gitAccessKey,
				}

				// crypt against the given master password and save result in config
				cryptedPhrase, err := encryptStringToBase64(gitAccessKey, secretPhrase)
				if err != nil {
					sqFatal(err)
				}

				// cache the encrypted key in the config file
				gitKeys := viper.GetStringMap("git-keys")
				gitKeys[syncKeyName] = cryptedPhrase
				viper.SetDefault("git-keys", gitKeys)
				viper.WriteConfig()
				sqlog.Println(styleSuccess.Render("Access key encrypted and stored in Squonker config"))
			}
		}

		// if no git repo given, this command can just set the key above and then exit
		if len(syncAddGitAddr) > 0 {
			sqlog.Print("Cloning repository ...")

			_, err := git.PlainClone("squad."+syncAddName, false, &cloneOptions)
			if err != nil {
				sqlog.Print(styleFailure.Render("Unable to clone repository!"))
				sqFatal(err)
			}

			gitSyncs := viper.GetStringMap("git-syncs")
			gitSyncs[syncKeyName] = syncAddGitAddr

			viper.SetDefault("git-syncs", gitSyncs)
			viper.WriteConfig()

			sqlog.Print(styleSuccess.Render("[Complete]"))

		} else {
			sqlog.Print("No repository specified, skipping clone")
		}
	},
}

func init() {
	rootCmd.AddCommand(syncAddCmd)

	syncAddCmd.Flags().StringVarP(&syncAddName, "name", "n", "", "local name for the folder")
	syncAddCmd.MarkFlagRequired("name")

	syncAddCmd.Flags().StringVarP(&syncAddGitAddr, "git", "g", "", "git repository address to add for syncing")

	syncAddCmd.Flags().BoolVar(&syncAddGitAccessKey, "key", false, "optional prompt for git personal access key for private repos")

	syncAddCmd.Flags().StringVarP(&syncCopyGitAccessKeyFrom, "samekey", "s", "", "use the same key as the given existing sync name")
}
