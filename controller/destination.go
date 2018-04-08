package controller

import "os"

type Destination struct {
	currentPosition int64
	destinationFile     *os.File
	DestinationFileName string
}
