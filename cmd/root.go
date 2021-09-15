package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

const (
	cfgAppGUID  = "appGUID"
	cfgDataGUID = "dataGUID"
	cfgWebDAV   = "webdav"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "squonker",
	Short: "Let's squonk again",
	Long: `
Experimental Endlesss Soundpack Anarchy Engine
ishani 2021  //  use at your own terrible risk`,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.squonker.yaml)")
}

// initConfig reads in config file and ENV variables if set
func initConfig() {

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// search config in home directory with name ".squonker" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".squonker")
	}
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		sqFatal(err)
	}
	//	else {
	//		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	//	}
	viper.SafeWriteConfig()
}
