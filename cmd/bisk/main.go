package main

import (
	"fmt"
	"log"
	"os"

	intl "apcandsons.com/bisket/internal"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bisk",
	Short: "This is a CLI tool for bisket, a light weight application switcher",
	Long:  `A fast and flexible application switcher build with love in Go. Complete documentation is available at https://github.com/apcandsons/bisket`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes bisket configuration",
	Long:  "This command initializes the bisket configuration file (bisket.yaml)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing bisket")
		var cfg = intl.Config{}
		cfg.Init()
		cfg.WriteToFile("bisket.yaml")
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage bisket server",
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the bisket server",
	Long:  `This command starts the bisket server, making it ready to manage application version instances`,
	Run: func(cmd *cobra.Command, args []string) {
		var cfg = intl.Config{}
		err := cfg.ReadFromFile("bisket.yaml")
		if err != nil {
			log.Fatalf("Error reading config file: %v", err)
		}

		var svr = intl.Server{Conf: cfg}
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
