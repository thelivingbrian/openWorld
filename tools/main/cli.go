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

func runWithContext(c *Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		collectionName := args[0]
		fmt.Printf("Deploying collection: %s\n", collectionName)
		c.deploy(collectionName)
		os.Exit(0)
	}
}

func ExecuteCLICommands(c *Context) {
	var deployCmd = &cobra.Command{
		Use:   "deploy [collectionName]",
		Short: "Deploy a specific collection",
		Long:  `Deploy the given collection name to the server.`,
		Args:  cobra.ExactArgs(1),
		Run:   runWithContext(c),
	}
	rootCmd.AddCommand(deployCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
