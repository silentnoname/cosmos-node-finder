/*
Copyright © 2024 silent silentvalidator@gmail.com
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cosmos-node-finder",
	Short: "A cli tool to find cosmos chains public rpc nodes and live peers",
	Long: `Cosmos-node-finder is a cli tool to find cosmos chains public rpc nodes and live peers. 
You can get detailed info of these public rpc nodes(latest height,if it is an archive node, ...)`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
