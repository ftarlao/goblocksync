package controller

import (
	"fmt"
	"goblocksync/data/configuration"
	"goblocksync/utils"
	"os"
)

type Destination interface {
	GetConfig() configuration.Configuration
	Start() error
}

type destinationV1 struct {
	Config        configuration.Configuration
	startPosition int64
}

func (s destinationV1) GetConfig() configuration.Configuration {
	return s.Config
}

func (s destinationV1) Start() error {

	f, err := os.Open(s.Config.DestinationFile.FileName)
	if err != nil {
		fmt.Println(err)
		return err
	}

	hasher := utils.NewHasherImpl(s.Config.BlockSize, f, s.Config.StartLoc)
	success, err := hasher.Start()
	if !success {
		fmt.Println(err)
	}
	return err
}

func NewDestination(config configuration.Configuration) destinationV1 {
	return destinationV1{Config: config}
}
