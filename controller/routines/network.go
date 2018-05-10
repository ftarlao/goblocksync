package routines

import (
	"encoding/gob"
	"errors"
	"github.com/ftarlao/goblocksync/data/configuration"
	"github.com/ftarlao/goblocksync/data/messages"
	"github.com/ftarlao/goblocksync/utils"
	"io"
	"sync"
	"time"
)

type NetworkManager struct {
	//Input stream
	InStream io.Reader
	//Output stream
	OutStream io.Writer
	//Gob input decoder
	inDecoder *gob.Decoder
	//Gob output encoder
	outEncoder *gob.Encoder
	//Input channel for decoded messages
	inMsgChannel chan messages.Message
	//Output channel for encoded messages
	outMsgChannel chan messages.Message
	//locks to ensure complete stop and avoid double start
	lockNetManager sync.Mutex
	// Current running status
	running       bool
	startDisabled bool
	stopChannel   chan bool
}

const channelWaitTime = time.Second
const stopTimeout = 4 * time.Second

func NewNetworkManager(blocksize int64, in io.Reader, out io.Writer) (n NetworkManager) {

	maxMessageApproxSize := utils.IntMax(blocksize, configuration.HashGroupMessageSize*configuration.HashSize)
	NetworkChannelSize := (configuration.NetworkMaxBytes / maxMessageApproxSize)/2

	inDecoder, outEncoder := EncoderInOut(in, out)
	n = NetworkManager{
		InStream:      in,
		OutStream:     out,
		inDecoder:     inDecoder,
		outEncoder:    outEncoder,
		inMsgChannel:  make(chan messages.Message, NetworkChannelSize),
		outMsgChannel: make(chan messages.Message, NetworkChannelSize),
		running:       false,
		startDisabled: false,
		stopChannel:   make(chan bool, 3)}
	return
}

func (n *NetworkManager) GetInMsgChannel() chan messages.Message {
	return n.inMsgChannel
}

func (n *NetworkManager) GetOutMsgChannel() chan messages.Message {
	return n.outMsgChannel
}

func (n *NetworkManager) Start() (err error) {
	//Synchronized method
	n.lockNetManager.Lock()
	if n.running {
		return errors.New("already running")
	}
	if n.startDisabled {
		return errors.New("stopped, cannot run twice")
	}
	n.running = true
	n.startDisabled = true
	n.lockNetManager.Unlock()
	//Write messages routine
	go func() {
		defer func() {
			r := recover()
			n.stopOn(r)
			n.stopChannel <- true
		}()

		var errGo error
		for n.running {
			select {
			case msg := <-n.outMsgChannel:
				errGo = messages.EncodeMessage(n.outEncoder, msg)
				utils.Check(errGo)
			case <-time.After(channelWaitTime):
			}
		}

		//errGo = messages.EncodeMessage(n.outEncoder, messages.NewEndMessage())
		//utils.Check(errGo)
	}()

	//Read messages routine
	go func() {
		defer func() {
			r := recover()
			n.stopOn(r)
			n.stopChannel <- true
		}()

		for n.running {
			m, errGo := messages.DecodeMessage(n.inDecoder)
			utils.Check(errGo)
			select {
			case n.inMsgChannel <- m:
			case <-time.After(channelWaitTime):
			}
		}
	}()
	//No flush nor sync exists for Reader/Writer
	return
}

func (n *NetworkManager) stopOn(err interface{}) {

	//Only the first stopOn acts properly and notifies errors
	n.lockNetManager.Lock()
	if n.running {
		n.running = false
		if err != nil {
			eMsg := messages.NewErrorMessage(err.(error))
			n.inMsgChannel <- eMsg
		}

		//Force close of all the channels, readers and writers
		close(n.inMsgChannel)
		close(n.outMsgChannel)
		//Perform close of in/out only when Closer
		cReader, cSuccess := n.InStream.(io.ReadCloser)
		if cSuccess {
			cReader.Close()
		}
		cWriter, cSuccess := n.OutStream.(io.WriteCloser)
		if cSuccess {
			cWriter.Close()
		}
	}
	//...the stop notification is always performed
	n.lockNetManager.Unlock()
}

func (n *NetworkManager) Stop() (err error) {
	n.stopOn(nil)
	//Wait two stop signals
	for i := 0; i < 2; i++ {
		select {
		case <-n.stopChannel:
		case <-time.After(stopTimeout):
			return errors.New("stop timeout")
		}
	}
	return nil
}

func (n *NetworkManager) IsRunning() bool {
	return n.running
}

// Provides input and output gob encoder-decoder from a given Reader-Writer pair
func EncoderInOut(in io.Reader, out io.Writer) (inDecoder *gob.Decoder, outEncoder *gob.Encoder) {
	outEncoder = gob.NewEncoder(out)
	inDecoder = gob.NewDecoder(in)
	return
}
