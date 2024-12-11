package main

import (
	"crypto/rand"
	"encoding/base64"
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

func deployCollection(c *Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		collectionName := args[0]
		fmt.Printf("Deploying collection: %s\n", collectionName)
		c.deploy(collectionName)
		os.Exit(0)
	}
}

func generateKeys() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		keys := make([]byte, 64) // Create a slice to hold 32 bytes
		_, err := rand.Read(keys)
		if err != nil {
			return
		}
		hashKey := base64.StdEncoding.EncodeToString(keys[0:32])
		blockKey := base64.StdEncoding.EncodeToString(keys[32:64])
		fmt.Println("COOKIE_HASH_KEY=" + string(hashKey))
		fmt.Println("COOKIE_BLOCK_KEY=" + string(blockKey))
	}
}

func ExecuteCLICommands(c *Context) {
	var deployCmd = &cobra.Command{
		Use:   "deploy [collectionName]",
		Short: "Deploy a specific collection",
		Long:  `Deploy the given collection name to the server.`,
		Args:  cobra.ExactArgs(1),
		Run:   deployCollection(c),
	}
	rootCmd.AddCommand(deployCmd)

	var keysCmd = &cobra.Command{
		Use:   "keys",
		Short: "Generate a set of keys",
		Long:  `Generate a set of keys`,
		Args:  cobra.MaximumNArgs(0),
		Run:   generateKeys(),
	}
	rootCmd.AddCommand(keysCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
