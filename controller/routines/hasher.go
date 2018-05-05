package routines

import (
	"bufio"
	"errors"
	"github.com/ftarlao/goblocksync/data/configuration"
	"github.com/ftarlao/goblocksync/data/messages"
	"github.com/ftarlao/goblocksync/utils"
	"io"
	"os"
	"sync"
	"time"
)

type Hasher interface {
	GetOutMsgChannel() chan messages.Message
	Start() error
	Stop() error
	GetCurrentPosition() int64
	IsRunning() bool
}

type hasherImpl struct {
	// Block size in bytes
	blockSize int64
	// File descriptor
	fileDesc *os.File
	// Current position of hashing operator in bytes; the next hash is for the [currentLoc,currentLoc+blockSize) portion
	currentLoc int64
	// Output chan for the obtained hashes
	outMsgChannel chan messages.Message
	// Internal data channel
	readDataChannel chan messages.Message
	// Current running status
	running bool
	// lockHasher on Start
	lockHasher  sync.Mutex
	// current hashing function
	hashingFunc func([]byte, int) []byte
	// channel for stop signals
	stopChannel chan bool
}

func (h *hasherImpl) GetOutMsgChannel() chan messages.Message {
	return h.outMsgChannel
}

func (h *hasherImpl) Start() error {
	h.lockHasher.Lock()
	if h.running {
		return errors.New("the 'hasher' is already running")
	}
	h.running = true
	h.lockHasher.Unlock()

	go dataReader(h)

	go hasherRoutine(h)

	return nil
}

func dataReader(hasher *hasherImpl) {
	defer func() {
		if r := recover(); r!=nil { //this is very unlikely to happen, defensive
			hasher.readDataChannel <- messages.NewErrorMessage(r.(error))
		}
		hasher.stopChannel <- true
	}()

	//seek to the start position
	_, err := hasher.fileDesc.Seek(hasher.currentLoc, 0)
	if err != nil {
		hasher.readDataChannel <- messages.NewErrorMessage(err)
		return
	}
	//..better to put a read buffer
	fBuffered := bufio.NewReaderSize(hasher.fileDesc, int(5*hasher.blockSize))

	var n1 = 0

	for err == nil && hasher.running {
		//fmt.Println("Block ", numHashes, "Start position [byte] ", h.currentLoc)
		dataBlock := make([]byte, hasher.blockSize)
		n1, err = io.ReadFull(fBuffered, dataBlock)
		//An error is sent to the hashing part..
		if err != nil && !utils.IsEOF(err) {
			hasher.readDataChannel <- messages.NewErrorMessage(err)
		}
		if n1 > 0 {
			//allocating a struct and array is inefficient but may be the right thing to parallelize the Hashing part later
			hasher.readDataChannel <- messages.NewDataBlockMessage(hasher.currentLoc, dataBlock[:n1])
			hasher.currentLoc += int64(n1)
		}
		if utils.IsEOF(err) {
			hasher.readDataChannel <- messages.NewEndMessage()
			return
		}
	}
}

func hasherRoutine(hasher *hasherImpl) {
	defer func() {
		if r := recover(); r!=nil {
			hasher.outMsgChannel <- messages.NewErrorMessage(r.(error))
		}
		hasher.running = false
		hasher.stopChannel <- true
	}()

	var err error
	var msg messages.Message
	currentMessage := messages.NewHashGroupMessage(hasher.currentLoc)
	for err == nil && hasher.running {
		// Create new HashGroupMessage
		msg = <-hasher.readDataChannel

		switch msg.GetMessageID() {
		case messages.DataBlockMessageID:
			msgDataBlock := msg.(*messages.DataBlockMessage)
			if currentMessage.IsFull() {
				hasher.outMsgChannel <- currentMessage
				// Create new HashGroupMessage
				currentMessage = messages.NewHashGroupMessage(msgDataBlock.StartLoc)
			}

			hash := hasher.hashingFunc(msgDataBlock.Data, configuration.HashSize)
			currentMessage.AddHash(hash)
		case messages.EndMessageID:
			if !currentMessage.IsEmpty() {
				currentMessage.TruncHashGroup()
				hasher.outMsgChannel <- currentMessage
			}
			hasher.outMsgChannel <- msg
			return
		case messages.ErrorMessageID:
			hasher.outMsgChannel <- msg
			return
		default:
			hasher.outMsgChannel <- messages.NewErrorMessage(errors.New("unexpected msg type provided from data reader goroutine to the hashing goroutine"))
			return
		}
	}
}

func (n *hasherImpl) Stop() error{
	n.lockHasher.Lock()
	n.running = false
	for i := 0; i < 2; i++ {
		select {
		case <-n.stopChannel:
		case <-time.After(stopTimeout):
			return errors.New("stop timeout")
		}
	}
	n.lockHasher.Unlock()
	return nil
}

func (h *hasherImpl) GetCurrentPosition() int64 {
	return h.currentLoc
}

func (h *hasherImpl) IsRunning() bool {
	return h.running
}

func NewHasherImpl(blockSize int64, fileDesc *os.File, startLoc int64, hashingFunc func([]byte, int) []byte) Hasher {
	instance := hasherImpl{blockSize: blockSize, fileDesc: fileDesc, currentLoc: startLoc, running: false}
	instance.outMsgChannel = make(chan messages.Message, configuration.HashGroupChannelSize)
	instance.readDataChannel = make(chan messages.Message, 5)
	instance.hashingFunc = hashingFunc
	return &instance
}

// very dumb 'size' bit hash, ...for tests only
func DummyHash(data []byte, size int) (hash []byte) {
	hash = make([]byte, size)
	for i, elem := range data {
		hash[i%size] = hash[i%size] ^ elem
	}
	return
}
