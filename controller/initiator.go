package controller

import (
	"errors"

	"github.com/ftarlao/goblocksync/controller/routines"
	"github.com/ftarlao/goblocksync/data/configuration"
	"github.com/ftarlao/goblocksync/data/messages"
	"github.com/ftarlao/goblocksync/utils"
	"io"
	"log"
	"os"
	"os/exec"
)

//Interface for Master and Slave both

type Initiator interface {
	GetConfig() configuration.Configuration
	Start() error
}

//Master

type master struct {
	Config configuration.Configuration
}

func NewMaster(conf configuration.Configuration) master {
	return master{conf}
}

func (m master) GetConfig() configuration.Configuration {
	return m.Config
}

func (m master) Start() (err error) {

	//TODO to understand golang logging and change/remove prints with 'professional' stuff
	// execute slave locally and connect slave with encoder/decoder
	in, out, err := execSlave()
	//defer cmd.Process.Kill()
	if err != nil {
		return err
	}

	// perform Handshake
	bestProtocol, err := handshake(in, out)
	if err != nil {
		return err
	}
	log.Println("Best selected protocol: ", *bestProtocol)

	//Prepare encoder/decoder based on configuration
	_, outEncoder := routines.EncoderInOut(in, out)
	//send complemented configuration to slave
	remoteConf := m.Config.Complement()
	err = outEncoder.Encode(remoteConf)
	if err != nil {
		return err
	}

	//execute source or destination controller (for selected protocol version)
	source, err := NewSource(m.Config, *bestProtocol, in, out)
	if err != nil {
		return err
	}
	err = source.Start()
	return err
}

//Slave

type slave struct {
	Config configuration.Configuration
}

func NewSlave() slave {
	return slave{}
}

func (m slave) GetConfig() configuration.Configuration {
	return m.Config
}

func (m slave) Start() error {

	//send hello+version/receive hello+version, choose protocol version
	protocol, err := handshake(os.Stdin, os.Stdout)
	if err != nil {
		return err
	}

	//receive complemented configuration from master
	//execute source or destination controller (for selected protocol version)

	if m.Config.IsSource {
		source, err := NewSource(m.Config, *protocol, os.Stdin, os.Stdout)
		if err != nil {
			return err
		}
		err = source.Start()
	} else {
		destination, err := NewDestination(m.Config, *protocol, os.Stdin, os.Stdout)
		if err != nil {
			return err
		}
		err = destination.Start()

	}

	return nil
}

// Executes slave locally, inDecoder sends objects to the process, outEncoder reads objects from the process
func execSlave() (in io.Reader, out io.Writer, err error) {
	execName := os.Args[0]

	cmd := exec.Command(execName, "-S")

	out, _ = cmd.StdinPipe()
	in, _ = cmd.StdoutPipe()

	err = cmd.Start()
	if err != nil {
		return nil, nil, err
	}
	return
}

func handshake(in io.Reader, out io.Writer) (bestProtocol *int, err error) {
	inDecoder, outEncoder := routines.EncoderInOut(in, out)

	// send hello+version
	err = messages.EncodeMessage(outEncoder, messages.NewHelloInfo())
	//err = outEncoder.Encode(messages.NewHelloInfo())
	if err != nil {
		return bestProtocol, err
	}
	// receive hello+version
	var remoteHelloInfo *messages.HelloInfoMessage
	m, err := messages.DecodeMessage(inDecoder)
	//err = inDecoder.Decode(&remoteHelloInfo)
	remoteHelloInfo = m.(*messages.HelloInfoMessage)
	if err != nil {
		return bestProtocol, err
	}

	// let's choose protocol version
	inter := utils.Intersection(configuration.SupportedProtocols, remoteHelloInfo.SupportedProtocols)
	if len(inter) == 0 {
		return bestProtocol, errors.New("master and slave protocols versions are no compatible")
	}
	bestProtocol = utils.Max(inter)
	return bestProtocol, err
}
