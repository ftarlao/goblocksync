package routines

import (
	"bufio"
	"fmt"
	"github.com/ftarlao/goblocksync/data/configuration"
	"github.com/ftarlao/goblocksync/data/messages"
	"io"
	"os"
)

type Hasher interface {
	GetOutMsgChannel() chan messages.Message
	Start() error
	Stop()
	GetCurrentLoc() int64
	//SetCurrentLoc(int64)
	isRunning() bool
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
	// Current running status
	running bool
	// Communicate stop to goroutine
	statusChannel chan messages.Message
}

func (h *hasherImpl) GetOutMsgChannel() chan messages.Message {
	return h.outMsgChannel
}

func (h *hasherImpl) Start() error {
	h.running = false

	_, err := h.fileDesc.Seek(h.currentLoc, 0)
	fBuffered := bufio.NewReader(h.fileDesc)

	if err != nil {
		fmt.Println(err)
		return err
	}
	// Dovrei leggere per hash solo quando non ci sono altri dati da leggere o scrivere? Come schedulare tra le due attività?
	// Saltare a pendolo tra una e l'altra?
	go func() {
		var errCycle error = nil
		var n1, numHashes int //numHashes is the size of the current HashGroupMessage
		var hashGroupMessage messages.HashGroupMessage
		for errCycle == nil && h.running {
			// Create new HashGroupMessage
			if numHashes >= configuration.HashGroupMessageSize {
				fmt.Println("msg:", hashGroupMessage)
				h.outMsgChannel <- hashGroupMessage
				h.outMsgChannel = nil //defensive, should not be used again
				numHashes = 0
			}

			if numHashes == 0 {
				hashGroupMessage = messages.NewHashMessage(h.currentLoc)
			}

			select {
			default:
				//fmt.Println("Block ", numHashes, "Start position [byte] ", h.currentLoc)
				dataBlock := make([]byte, h.blockSize)
				n1, errCycle = io.ReadFull(fBuffered, dataBlock)
				if n1 > 0 {
					hash := dummyHash(dataBlock, configuration.HashSize)
					hashGroupMessage.HashGroup[numHashes] = hash
					numHashes++
					h.currentLoc += int64(n1)
				}

			}
		}

		//Let's send the last one if not-empty
		if numHashes > 0 {
			hashGroupMessage.TruncHashGroup(numHashes)
			fmt.Println("msg:", hashGroupMessage)
			h.outMsgChannel <- hashGroupMessage
		}
		// send EOF
		h.outMsgChannel <- messages.NewHashMessageEOF()
	}()
	return nil
}

func (h *hasherImpl) Stop() {
	h.running = false
	return
}

func (h *hasherImpl) GetCurrentLoc() int64 {
	return h.currentLoc
}

func (h *hasherImpl) isRunning() bool {
	return h.running
}

func NewHasherImpl(blockSize int64, fileDesc *os.File, startLoc int64) Hasher {
	instance := &hasherImpl{blockSize: blockSize, fileDesc: fileDesc, currentLoc: startLoc}
	instance.outMsgChannel = make(chan messages.Message, configuration.HashGroupChannelSize)
	instance.running = false
	return instance
}

// very dumb 'size' bit hash, for tests only
func dummyHash(data []byte, size int) (hash []byte) {
	hash = make([]byte, size)
	for i, elem := range data {
		hash[i%size] = hash[i%size] ^ elem
	}
	return
}
