package controller

import (
	"errors"
	"fmt"
	"github.com/ftarlao/goblocksync/controller/routines"
	"github.com/ftarlao/goblocksync/data/configuration"
	"github.com/ftarlao/goblocksync/data/messages"
	"io"
	"os"
)

//DESTINATION

type Destination interface {
	GetConfig() configuration.Configuration
	Start() error
}

type destinationV1 struct {
	Config        configuration.Configuration
	startPosition int64
	in            io.Reader
	out           io.Writer
}

func (s destinationV1) GetConfig() configuration.Configuration {
	return s.Config
}

func (d destinationV1) Start() error {

	f, err := os.Open(d.Config.DestinationFile.FileName)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//Prepare data Pipes
	_, outEncoder := routines.EncoderInOut(d.in, d.out)

	// Start hasher
	hasher := routines.NewHasherImpl(d.Config.BlockSize, f, d.Config.StartLoc, routines.DummyHash)
	err = hasher.Start()
	if err != nil {
		fmt.Println(err)
	}

	//Network flow Handler, single point (and thread) to send and receive messages
	//Should be split in two goroutines or we should add bigger buffers for Pipe Writer/Reader
	hashChan := hasher.GetOutMsgChannel()
	count := 0
	//Separate in different goroutines.. gosh
_:
	for {
		//send
		select {
		case hashMsg := <-hashChan:
			messages.EncodeMessage(outEncoder, hashMsg)
			//utils.DoNothing(hashMsg)
			count++
		default:

		}

		//receive
		select {
		default:

		}

		if d.Config.IsMaster {
			fmt.Println("Blocks: ", count)
		}
	}

	return err
}

func NewDestination(config configuration.Configuration, protocolVersion int, in io.Reader, out io.Writer) (d Destination, err error) {
	switch protocolVersion {
	case 1:
		d = destinationV1{Config: config, in: in, out: out}
	default:
		return nil, errors.New("protocol version not supported (mismatch between declared versions and available versions)")
	}
	return
}

//SOURCE

type Source interface {
	GetConfig() configuration.Configuration
	Start() error
}

type sourceV1 struct {
	Config     configuration.Configuration
	sourceFile *os.File
	in         io.Reader
	out        io.Writer
}

func (s sourceV1) GetConfig() configuration.Configuration {
	return s.Config
}

func (s sourceV1) Start() error {

	f, err := os.Open(s.Config.SourceFile.FileName)
	if err != nil {
		return err
	}

	//Prepare data Pipes
	inDecoder, _ := routines.EncoderInOut(s.in, s.out)

	// Start hasher
	hasher := routines.NewHasherImpl(s.Config.BlockSize, f, s.Config.StartLoc, routines.DummyHash)
	err = hasher.Start()
	if err != nil {
		fmt.Println(err)
	}

	//Network flow Handler, single point (and thread) to send and receive messages
	//Should be split in two goroutines or we should add bigger buffers for Pipe Writer/Reader
	//hashChan := hasher.GetOutMsgChannel()

	count := 0
_:
	for {
		//send
		m, err := messages.DecodeMessage(inDecoder)
		if err != nil {
			return err
		}
		hashMsg := m.(*messages.HashGroupMessage)
		hashMsg.StartLoc = 1
		count++
		if s.Config.IsMaster {
			fmt.Println("Blocks: ", count)
		}
	}

	return err
}

func NewSource(config configuration.Configuration, protocolVersion int, in io.Reader, out io.Writer) (s Source, err error) {
	switch protocolVersion {
	case 1:
		s = sourceV1{Config: config, in: in, out: out}
	default:
		return nil, errors.New("protocol version not supported (mismatch between declared versions and available versions)")
	}
	return s, err
}
