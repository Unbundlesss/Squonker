// +build darwin

package cmd

func init() {
	rootCmd.AddCommand(osxDeployCmd)
	rootCmd.AddCommand(osxDecompileCmd)
}
