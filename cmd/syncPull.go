package cmd

import (
	"encoding/hex"
	"fmt"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var syncPullCmd = &cobra.Command{
	Use:   "sync-pull",
	Short: "update all registered syncs",
	Run: func(cmd *cobra.Command, args []string) {

		secretPhrase := askForMasterSecret()

		gitSyncs := viper.GetStringMap("git-syncs")
		gitKeys := viper.GetStringMap("git-keys")
		for gitDir := range gitSyncs {
			syncKeyNameByte, err := hex.DecodeString(gitDir)
			if err != nil {
				sqFatal(err)
			}

			var gitAccessKeyDecrypted = ""
			if accessKey, ok := gitKeys[gitDir]; ok {
				gitAccessKeyDecrypted, err = decryptStringFromBase64(fmt.Sprintf("%v", accessKey), secretPhrase)
				if err != nil {
					sqFatal(err)
				}
			}

			err = SyncPull("squad."+string(syncKeyNameByte), gitAccessKeyDecrypted)
			if err != nil {
				sqFatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(syncPullCmd)
}

func SyncPull(gitDir, gitKey string) error {

	// check if the directory even exists
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		sqlog.Print("cannot find directory to sync : ", gitDir)
		return err
	}

	// we instantiate a new repository targeting the given path (the .git folder)
	r, err := git.PlainOpen(gitDir)
	if err != nil {
		return err
	}

	// get the working directory for the repository
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	pullOptions := git.PullOptions{
		RemoteName: "origin",
	}
	if len(gitKey) > 0 {
		pullOptions.Auth = &http.BasicAuth{
			Username: "null",
			Password: gitKey,
		}
	}

	// pull the latest changes from the origin remote and merge into the current branch
	sqlog.Print(gitDir, " : git pull origin")
	err = w.Pull(&pullOptions)
	// don't treat already up-to-date as an actual error
	if err != nil && err == git.NoErrAlreadyUpToDate {
		sqlog.Print(styleSuccess.Render(fmt.Sprint(gitDir, " : already up-to-date ")))
		return nil
	}
	if err != nil {
		return err
	}

	// print the latest commit that was just pulled
	ref, err := r.Head()
	if err != nil {
		return err
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return err
	}

	sqlog.Print(styleSuccess.Render(fmt.Sprint(gitDir, " : sync complete, latest commit ", commit.Hash)))
	return nil
}
