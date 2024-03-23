package main

import (
	"fmt"
	"log"
	"os"

	intl "apcandsons.com/bisqit/internal"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bisq",
	Short: "This is a CLI tool for bisqit, a light weight application switcher",
	Long:  `A fast and flexisble application switcher build with love in Go. Complete documentation is available at https://github.com/apcandsons/bisqit`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes bisqit configuration",
	Long:  "This command initializes the bisqit configuration file (bisqit.yaml)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing bisqit")
		var cfg = intl.Config{}
		cfg.Init()
		cfg.WriteToFile("bisqit.yaml")
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage bisqit server",
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the bisqit server",
	Long:  `This command starts the bisqit server, making it ready to manage application version instances`,
	Run: func(cmd *cobra.Command, args []string) {
		var cfg = intl.Config{}
		err := cfg.ReadFromFile("bisqit.yaml")
		if err != nil {
			log.Fatalf("Error reading config file: %v", err)
		}

		var repo = intl.Repository{}
		var svr = intl.Server{}

		svr.Init(&cfg, &repo)
		if err := svr.Start(); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(serverStartCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
