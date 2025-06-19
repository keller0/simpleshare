package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	address    string
	port       string
	tmpFileDir string
)

func init() {

	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "7777", "listen port")
	rootCmd.PersistentFlags().StringVarP(&address, "address", "a", "127.0.0.1", "listen address")
	rootCmd.PersistentFlags().StringVarP(&tmpFileDir, "folder", "f", "tmpFile", "tmp file folder")

}

var rootCmd = &cobra.Command{
	Use:   "simpleshare",
	Short: "Share texts and files in local network",
	Long: `
  A Simple http service for share texts and files in local network
built in Go.
  source: https://github.com/keller0/simpleshare`,
	CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	Version:           "0.0.3",
	Example:           "./simpleshare  -a 127.0.0.1 -p 7777 -f tmpFile",
	Run: func(cmd *cobra.Command, args []string) {
		tDir := filepath.Join(".", tmpFileDir)
		err := os.MkdirAll(tDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
		fmt.Println("tmp file folder:", tDir)
		err = runServer()
		if err != nil {
			fmt.Println(err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
