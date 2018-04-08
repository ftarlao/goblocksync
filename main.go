package main

import (
	"flag"
	"fmt"
	"os"
	"goblocksync/data"
	"goblocksync/controller"
)

// var version int = 0

func main() {
	globalConfig, correct, err := parseArgs()

	if !correct {
		fmt.Println("Error: ", err)
		os.Exit(3) //let's look for hardcoded error codes
	}

	fmt.Println("Goblocksync running...\n\nThe destination file will be synched with the source file")
	fmt.Println("DESTINATION FILE WILL BE OVERWRITTEN\n")
	fmt.Println("Source file:\t\t", globalConfig.SourceFileName)
	fmt.Println("Destination file:\t", globalConfig.DestinationFileName)

	// Access files
	source := controller.Source{}
	source.Config = globalConfig
	source.Start()
}

func parseArgs() (data.Configuration, bool, error) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "goblocksync -s sourcefile -d destinationfile\n\n")
		flag.PrintDefaults()
	}
	sourceFile := flag.String("s", "", "Source file path")
	destinationFile := flag.String("d", "", "Destination file path")
	flag.Parse()

	// populate the configuration
	globalConfig := data.Configuration{*sourceFile, *destinationFile}

	// validate the configuration
	correct, err := globalConfig.Validate()

	return globalConfig, correct, err
}
