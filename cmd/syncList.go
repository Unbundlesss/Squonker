package cmd

import (
	"encoding/hex"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var syncListCmd = &cobra.Command{
	Use:   "sync-list",
	Short: "Display known syncs",
	Run: func(cmd *cobra.Command, args []string) {

		gitSyncs := viper.GetStringMap("git-syncs")
		gitKeys := viper.GetStringMap("git-keys")
		for gitDir := range gitSyncs {
			syncKeyNameByte, err := hex.DecodeString(gitDir)
			if err != nil {
				sqFatal(err)
			}

			keyState := ""
			if _, hasKey := gitKeys[gitDir]; hasKey {
				keyState = styleNotice.Render(" (+ access key)")
			}

			sqlog.Printf("%-16s : %-50s %s", string(syncKeyNameByte), gitSyncs[gitDir], keyState)
		}
		sqComplete()
	},
}

func init() {
	rootCmd.AddCommand(syncListCmd)
}
