package main

import (
	"flag"
	"fmt"
	"os"
)

// var version int = 0

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "goblocksync -s sourcefile -d destinationfile\n\n")
		flag.PrintDefaults()
	}
	sourceFile := flag.String("s", "", "Source file path")
	destinationFile := flag.String("d", "", "Destination file path")
	flag.Parse()

	fmt.Println("Goblocksync running...\n\nThe destination file will be synched with the source file")
	fmt.Println("DESTINATION FILE WILL BE OVERWRITTEN\n")
	fmt.Println("Source file:\t\t", *sourceFile)
	fmt.Println("Destination file:\t", *destinationFile)

}
