package data

import (
	"errors"
)

type Configuration struct {
	SourceFileName      string
	DestinationFileName string
}

func (c Configuration) Validate() (bool, error) {
	correct := len(c.SourceFileName) > 0 && len(c.DestinationFileName) > 0
	var err error
	if !correct {
		err = errors.New("Please provide source and destination filenames")
	}
	return correct, err
}
