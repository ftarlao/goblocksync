package controller

import (
	"fmt"
	"goblocksync/data/configuration"
	"goblocksync/utils"
	"os"
)

type Source interface {
	GetConfig() configuration.Configuration
	Start() error
}

type sourceV1 struct {
	Config        configuration.Configuration
	startPosition int64
	sourceFile    *os.File
}

func (s sourceV1) GetConfig() configuration.Configuration {
	return s.Config
}

func (s sourceV1) Start() error {

	f, err := os.Open(s.Config.SourceFile.FileName)
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

func NewSource(config configuration.Configuration) sourceV1 {
	s := sourceV1{Config: config}
	s.startPosition = config.StartLoc
	return s
}
