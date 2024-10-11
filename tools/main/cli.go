package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("No command line action entered.")
	},
	SilenceUsage: true,
}

var deployCmd = &cobra.Command{
	Use:   "deploy [collectionName]",
	Short: "Deploy a specific collection",
	Long:  `Deploy the given collection name to the server.`,
	Args:  cobra.ExactArgs(1), // Ensure exactly 1 argument is passed
	Run: func(cmd *cobra.Command, args []string) {
		collectionName := args[0]
		fmt.Printf("Deploying collection: %s\n", collectionName)
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

func ExecuteCLICommands() {
	// Check which command is being run
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
