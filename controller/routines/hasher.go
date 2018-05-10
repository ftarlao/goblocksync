package routines

import (
	"bufio"
	"errors"
	"github.com/ftarlao/goblocksync/data/configuration"
	"github.com/ftarlao/goblocksync/data/messages"
	"github.com/ftarlao/goblocksync/utils"
	"io"
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

const (  // iota is reset to 0
	STOPPED = iota  // c0 == 0
	RUNNING = iota  // c1 == 1
	SHUTDOWN = iota  // c2 == 2
)

type hasherImpl struct {
	// Block size in bytes
	blockSize int64
	// File descriptor
	fileDesc io.ReadSeeker
	// Current position of hashing operator in bytes; the next hash is for the [currentLoc,currentLoc+blockSize) portion
	currentLoc int64
	// Output chan for the obtained hashes
	outMsgChannel chan messages.Message
	// Internal data channel
	readDataChannel chan messages.Message
	// Current running status
	running int
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
	if h.running != STOPPED {
		return errors.New("the 'hasher' is already running")
	}
	h.running = RUNNING
	h.lockHasher.Unlock()

	go dataReader(h)

	go hasherRoutine(h)

	return nil
}

func dataReader(n *hasherImpl) {
	defer func() {
		if r := recover(); r!=nil { //this is very unlikely to happen, defensive
			n.readDataChannel <- messages.NewErrorMessage(r.(error))
		}
	}()
	defer func(){
		n.stopChannel <- true
	}()

	//seek to the start position
	_, err := n.fileDesc.Seek(n.currentLoc, 0)
	if err != nil {
		n.readDataChannel <- messages.NewErrorMessage(err)
		return
	}
	//..better to put a read buffer
	fBuffered := bufio.NewReaderSize(n.fileDesc, int(5*n.blockSize))

	var n1 = 0

	for err == nil && n.running == RUNNING {
		//fmt.Println("Block ", numHashes, "Start position [byte] ", h.currentLoc)
		dataBlock := make([]byte, n.blockSize)
		n1, err = io.ReadFull(fBuffered, dataBlock)
		//An error is sent to the hashing part..
		if err != nil && !utils.IsEOF(err) {
			n.readDataChannel <- messages.NewErrorMessage(err)
		}
		if n1 > 0 {
			//allocating a struct and array is inefficient but may be the right thing to parallelize the Hashing part later
			n.readDataChannel <- messages.NewDataBlockMessage(n.currentLoc, dataBlock[:n1])
			n.currentLoc += int64(n1)
		}
		if utils.IsEOF(err) {
			n.readDataChannel <- messages.NewEndMessage()
			return
		}
	}
}

func hasherRoutine(n *hasherImpl) {
	defer func() {
		if r := recover(); r!=nil {
			n.outMsgChannel <- messages.NewErrorMessage(r.(error))
		}
	}()
	defer func(){
		n.running = SHUTDOWN
		n.stopChannel <- true
	}()

	var err error
	var msg messages.Message
	currentMessage := messages.NewHashGroupMessage(n.currentLoc)
	for err == nil && n.running == RUNNING {
		// Create new HashGroupMessage
		msg = <-n.readDataChannel

		switch msg.GetMessageID() {
		case messages.DataBlockMessageID:
			msgDataBlock := msg.(*messages.DataBlockMessage)
			if currentMessage.IsFull() {
				n.outMsgChannel <- currentMessage
				// Create new HashGroupMessage
				currentMessage = messages.NewHashGroupMessage(msgDataBlock.StartLoc)
			}

			hash := n.hashingFunc(msgDataBlock.Data, configuration.HashSize)
			currentMessage.AddHash(hash)
		case messages.EndMessageID:
			if !currentMessage.IsEmpty() {
				currentMessage.TruncHashGroup()
				n.outMsgChannel <- currentMessage
			}
			n.outMsgChannel <- msg
			return
		case messages.ErrorMessageID:
			n.outMsgChannel <- msg
			return
		default:
			n.outMsgChannel <- messages.NewErrorMessage(errors.New("unexpected msg type provided from data reader goroutine to the hashing goroutine"))
			return
		}
	}
}

func (n *hasherImpl) Stop() error{
	n.lockHasher.Lock()
	n.running = STOPPED
	for i := 0; i < 2; i++ {
		select {
		case <- n.stopChannel:
		case <-time.After(stopTimeout):	return errors.New("stop timeout")
		}
	}
	n.lockHasher.Unlock()
	return nil
}

func (n *hasherImpl) GetCurrentPosition() int64 {
	return n.currentLoc
}

func (n *hasherImpl) IsRunning() bool {
	return ! (n.running == STOPPED)
}

func NewHasherImpl(blockSize int64, fileDesc io.ReadSeeker, startLoc int64, hashingFunc func([]byte, int) []byte) Hasher {
	instance := hasherImpl{
		blockSize: blockSize,
		fileDesc: fileDesc,
		currentLoc: startLoc,
		running: STOPPED}

	instance.outMsgChannel = make(chan messages.Message, configuration.HashGroupChannelSize)
	instance.readDataChannel = make(chan messages.Message, configuration.DataMaxBytes / blockSize)
	instance.stopChannel = make(chan bool, 3)
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

// returns zero-filled hash array, no ops performed
func FakeHash(data []byte, size int) (hash []byte) {
	hash = make([]byte, size)
	return hash
}