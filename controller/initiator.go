package controller

import (
	"goblocksync/data/configuration"
	"os/exec"
	"path/filepath"
	"os"
	"log"
	"encoding/gob"
	"io"
	"goblocksync/data/messages"
	"goblocksync/utils"
)

//Interface for Master and Slave both

type Initiator interface {
	GetConfig() configuration.Configuration
	Start() error
}

//Master

type master struct {
	Config        configuration.Configuration
}

func NewMaster(conf configuration.Configuration) master{
	return master{conf}
}

func (m master) GetConfig() configuration.Configuration {
	return m.Config
}

func (m master) Start() (err error) {

	// execute slave locally
	// connect slave with encoder/decoder
	inDecoder, outEncoder, err := execSlave()
	//defer cmd.Process.Kill()

	if err != nil {
		panic(err)
	}
	//send hello+version/receive hello+version
	err = outEncoder.Encode(messages.NewHelloInfo())
	if err != nil {
		panic(err)
	}

	var remoteHelloInfo messages.HelloInfo
	err  =inDecoder.Decode(&remoteHelloInfo)
	if err != nil {
		panic(err)
	}

	//choose protocol version
	inter := utils.Intersection(configuration.SupportedProtocols,remoteHelloInfo.SupportedProtocols)
	bestProtocol := utils.Max(inter)

	log.Println("Best selected protocol: ",bestProtocol)
	//send complemented configuration to slave
	//execute source or destination controller (for selected protocol version)
	//source := NewSource(m.Config)
	//err = source.Start()

	return err
}


//Slave

type slave struct {
	Config        configuration.Configuration
}

func NewSlave() slave{
	return slave{}
}

func (m slave) GetConfig() configuration.Configuration {
	return m.Config
}

func (m slave) Start() error {

	//send hello+version/receive hello+version
	//choose protocol version
	//receive complemented configuration from master
	//execute source or destination controller (for selected protocol version)

	//source := NewSource(m.Config)
	//err := source.Start()

	return nil
}

// Executes slave locally, inDecoder sends objects to the process, outEncoder reads objects from the process
func execSlave() (inDecoder *gob.Decoder, outEncoder *gob.Encoder, err error){
	execName := os.Args[0]

	cmd := exec.Command(execName, "-S")

	out, _ := cmd.StdinPipe()
	in, _ := cmd.StdoutPipe()
	outEncoder = gob.NewEncoder(out)
	inDecoder = gob.NewDecoder(in)

	err = cmd.Run()
	if err != nil {
		return nil,nil, err
	}
	return
}

func chooseProtocol(remoteSupported []int){

}