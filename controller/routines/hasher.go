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
)

type Hasher interface {
	GetOutMsgChannel() chan messages.Message
	Start() error
	Stop()
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
	runningHash bool
	runningRead bool

	// lockRead on Start
	lockRead sync.Mutex
	lockHash sync.Mutex

	hashingFunc func([]byte, int) []byte
}

func (h *hasherImpl) GetOutMsgChannel() chan messages.Message {
	return h.outMsgChannel
}

func (h *hasherImpl) Start() error {
	h.lockRead.Lock()
	h.lockHash.Lock()

	h.runningHash = true
	h.runningRead = true

	go dataReader(h)

	go hasherRoutine(h)

	return nil
}

func dataReader(hasher *hasherImpl) {
	defer hasher.lockRead.Unlock()

	//seek to the start position
	_, err := hasher.fileDesc.Seek(hasher.currentLoc, 0)
	if err != nil {
		hasher.readDataChannel <- messages.NewErrorMessage(err)
		return
	}
	//..better to put a read buffer
	fBuffered := bufio.NewReaderSize(hasher.fileDesc, int(5*hasher.blockSize))

	var n1 = 0

	for err == nil && hasher.runningRead {
		//fmt.Println("Block ", numHashes, "Start position [byte] ", h.currentLoc)
		dataBlock := make([]byte, hasher.blockSize)
		n1, err = io.ReadFull(fBuffered, dataBlock)
		//An error is sent to the hashing part..
		if err != nil && !utils.IsEOF(err) {
			hasher.readDataChannel <- messages.NewErrorMessage(err)
			return //Premature termination
		}
		if n1 > 0 {
			//allocating a struct and array is inefficient but may be the right thing to parallelize the Hashing part later
			hasher.readDataChannel <- messages.NewDataBlockMessage(hasher.currentLoc, dataBlock[:n1])
			hasher.currentLoc += int64(n1)
		}
		if utils.IsEOF(err) {
			hasher.readDataChannel <- messages.NewEndMessage()
		}
	}
}

func hasherRoutine(hasher *hasherImpl) {
	defer hasher.lockHash.Unlock()

	var err error
	var msg messages.Message
	currentMessage := messages.NewHashGroupMessage(hasher.currentLoc)
	for err == nil && hasher.runningHash {
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
			hasher.outMsgChannel <- messages.NewErrorMessage(errors.New("unexpected msg type provided from data reader to the hashing goroutine"))
			return
		}
	}
}

func (h *hasherImpl) Stop() {
	h.runningRead = false
	h.runningHash = false

	// waits for the goroutines to stop
	h.lockRead.Lock()
	h.lockHash.Lock()

	defer func() {
		h.lockRead.Unlock()
		h.lockHash.Unlock()
	}()
	return
}

func (h *hasherImpl) GetCurrentPosition() int64 {
	return h.currentLoc
}

func (h *hasherImpl) IsRunning() bool {
	return h.runningHash || h.runningRead
}

func NewHasherImpl(blockSize int64, fileDesc *os.File, startLoc int64, hashingFunc func([]byte, int) []byte) Hasher {
	instance := hasherImpl{blockSize: blockSize, fileDesc: fileDesc, currentLoc: startLoc, runningHash: false, runningRead: false}
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
