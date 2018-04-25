package main

import (
	"flag"
	"fmt"
	"github.com/ftarlao/goblocksync/controller"
	"github.com/ftarlao/goblocksync/data/configuration"
	"os"
	"time"
)

func main() {
	globalConfig, isMaster, err := parseArgs()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(3) //let's look for hardcoded error codes
	}
	if isMaster {
		fmt.Println("Goblocksync command executed")

		fmt.Println("The destination file will be synched with the source file")
		fmt.Println("DESTINATION FILE WILL BE OVERWRITTEN\n")
		fmt.Println("Source file:\t\t", globalConfig.SourceFile.FileName)
		fmt.Println("Destination file:\t", globalConfig.DestinationFile.FileName)

		//Start Master
		master := controller.NewMaster(*globalConfig)
		master.Start()

		time.Sleep(3000 * time.Millisecond)
	} else {

		slave := controller.NewSlave()
		slave.Start()

		time.Sleep(3000 * time.Millisecond)
	}

}

// returns configuration, isMaster boolean, and in case.. an error. Configuration is nil for slave
func parseArgs() (*configuration.Configuration, bool, error) {
	flag.Usage = func() {
		fmt.Print(os.Stderr, "goblocksync -s sourcefile -d destinationfile\n\n")
		flag.PrintDefaults()
	}

	sourceFileName := flag.String("s", "", "Source file path")
	destinationFileName := flag.String("d", "", "Destination file path")
	isSlave := flag.Bool("S", false, "Enables slave mode, the other arguments are ignored")
	flag.Parse()

	if *isSlave {
		return nil, false, nil
	}
	// When master we parse

	//TODO to assign the isSource matters for future remote connections
	// populate the configuration
	globalConfig := configuration.Configuration{!*isSlave, true, configuration.FileDetails{FileName: *sourceFileName},
		configuration.FileDetails{FileName: *destinationFileName}, 0, 4096}

	// validate the configuration
	_, err := globalConfig.Validate()
	return &globalConfig, !*isSlave, err
}
