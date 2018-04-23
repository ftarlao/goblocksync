package routines

import (
	"encoding/gob"
	"github.com/ftarlao/goblocksync/data/configuration"
	"github.com/ftarlao/goblocksync/data/messages"
	"io"
	"sync"
	"github.com/ftarlao/goblocksync/utils"
)

type networkManager struct {
	InChannel     io.Reader
	OutChannel    io.Writer
	inDecoder     *gob.Decoder
	outEncoder    *gob.Encoder
	inMsgChannel  chan messages.Message
	outMsgChannel chan messages.Message
	lockIn          sync.Mutex
	lockOut          sync.Mutex
	// Current running status
	runningIn  bool
	runningOut bool
}

func NewNetworkManager(in io.Reader, out io.Writer) (n networkManager) {
	inDecoder, outEncoder := EncoderInOut(in, out)
	n = networkManager{InChannel: in, OutChannel: out,
		inDecoder: inDecoder, outEncoder: outEncoder,
		inMsgChannel: make(chan messages.Message, configuration.NetworkChannelsSize),
		outMsgChannel: make(chan messages.Message, configuration.NetworkChannelsSize),
		runningIn:false, runningOut:false}
	return
}

func (n *networkManager) GetInMsgChannel() chan messages.Message {
	return n.inMsgChannel
}

func (n *networkManager) GetOutMsgChannel() chan messages.Message {
	return n.outMsgChannel
}

func (n *networkManager) Start() (err error) {
	//Synchronized method
	n.lockIn.Lock()
	n.lockOut.Lock()

	n.runningOut = true
	n.runningIn = true
	//Write messages routine
	go func() {
		defer n.lockOut.Unlock()
		defer func() {
			if r := recover(); r != nil {
				n.stopOn(r.(error))
				return
			}
		}()

		var errGo error
		for n.runningOut {
			select {
			case msg := <-n.outMsgChannel:
				errGo = messages.EncodeMessage(n.outEncoder,msg)
				utils.Check(errGo)
			}
		}
		errGo = messages.EncodeMessage(n.outEncoder,messages.NewEndMessage())
		utils.Check(errGo)
	}()

	//Read messages routine
	go func() {
		defer n.lockIn.Unlock()
		defer func() {
			if r := recover(); r != nil {
				n.stopOn(r.(error))
				return
			}
		}()

		for n.runningIn {
			m, errGo := messages.DecodeMessage(n.inDecoder)
			utils.Check(errGo)
			n.inMsgChannel <- m
		}

		n.inMsgChannel <- messages.NewEndMessage()
	}()

	//No flush nor sync exists for Reader/Writer
	return
}

func (n *networkManager) stopOn(err error){
	n.runningOut = false
	n.runningIn = false
	emsg := messages.NewErrorMessage(err)
	n.inMsgChannel <- emsg
}

func (n *networkManager) Stop() (err error) {
	n.runningIn = false
	n.runningOut = false

	//Wait stop
	defer func() {
		n.lockIn.Unlock()
		n.lockOut.Unlock()
	}()
	n.lockIn.Lock()
	n.lockOut.Lock()

	return nil
}


//Utils
func EncoderInOut(in io.Reader, out io.Writer) (inDecoder *gob.Decoder, outEncoder *gob.Encoder) {
	outEncoder = gob.NewEncoder(out)
	inDecoder = gob.NewDecoder(in)
	return
}
