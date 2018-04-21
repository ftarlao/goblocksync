package routines

import (
	"encoding/gob"

	"io"
	"github.com/ftarlao/goblocksync/data/messages"
	"github.com/ftarlao/goblocksync/data/configuration"
)

type networkManager struct {
	InChannel     io.Reader
	OutChannel    io.Writer
	inDecoder     *gob.Decoder
	outEncoder    *gob.Encoder
	inMsgChannel  chan messages.Message
	outMsgChannel chan messages.Message
	statusChannel chan messages.Message
	// Current running status
	running bool
}

func NewNetworkManager(in io.Reader, out io.Writer) (n networkManager) {
	inDecoder, outEncoder := EncoderInOut(in, out)
	n = networkManager{in, out,
		inDecoder, outEncoder,
		make(chan messages.Message, configuration.NetworkChannelsSize),
		make(chan messages.Message, configuration.NetworkChannelsSize),
		make(chan messages.Message, 4), false}
	return
}

func (n *networkManager) GetInMsgChannel() chan messages.Message {
	return n.inMsgChannel
}

func (n *networkManager) GetOutMsgChannel() chan messages.Message {
	return n.outMsgChannel
}

func (n *networkManager) Start() (err error) {
	n.running = true

	//Write messages routine
	go func() {
		var err error
		for n.running {
			select {
			case msg := <-n.outMsgChannel:
				err = n.outEncoder.Encode(msg)
				if err != nil {
					n.running = false
					n.statusChannel <- messages.NewErrorMessage(err)
					return
				}
			}
		}
		n.statusChannel <- messages.NewEndMessage()
	}()

	//Read messages routine
	go func() {
		for n.running {
			m, err := messages.DecodeMessage(n.inDecoder)
			if err != nil {
				n.running = false
				n.statusChannel <- messages.NewErrorMessage(err)
				return
			}
			n.inMsgChannel <- m
		}
		n.statusChannel <- messages.NewEndMessage()
	}()

	//No flush nor sync exists for Reader/Writer

	return
}

func (n *networkManager) Stop() (err error) {
	n.running = false
	//We wait for exactly two non-error messages
	for i := 1; i < 2; i++ {
		select {
		case msg := <-n.statusChannel:
			if msg.GetMessageID() == messages.ErrorMessageID {
				err = msg.(messages.ErrorMessage).Err
			}
		}
	}
	return nil
}

//Utils

func EncoderInOut(in io.Reader, out io.Writer) (inDecoder *gob.Decoder, outEncoder *gob.Encoder) {
	outEncoder = gob.NewEncoder(out)
	inDecoder = gob.NewDecoder(in)
	return
}
