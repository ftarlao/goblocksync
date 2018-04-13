package configuration

import (
	"errors"
	"os"
)

type FileDetails struct {
	FileName string
	//file size [Bytes]
	size int64
	// true when this file represents a device (Linux/Unix/iOS)
	isDevice bool
}

func (f FileDetails) Update() (bool, error) {
	fileInfo, err := os.Stat("/path/to/file");
	if err != nil {
		return false, err
	}
	// get the size
	f.size = fileInfo.Size()
	f.isDevice = true //TODO to operate with block devices.. for real
}

//TODO Should few details about Master file names be masked .. and useless fields emptied? Less infos to the slave peer
type Configuration struct {
	IsMaster        bool
	IsSource        bool
	SourceFile      FileDetails
	DestinationFile FileDetails
	StartLoc        int64
	BlockSize       int64
}

func (c Configuration) Validate() (bool, error) {
	var err error
	correct := len(c.SourceFile.FileName) > 0 && len(c.DestinationFile.FileName) > 0
	if !correct {
		err = errors.New("please provide source and destination file names")
		return correct, err
	}
	correct = c.BlockSize > 0
	if !correct {
		err = errors.New("block size [byte] should be greater than zero")
		return correct, err
	}
	return correct, err
}

func (c Configuration) Complement() Configuration {
	conf := c
	conf.IsMaster = !c.IsMaster
	conf.IsSource = !c.IsSource
	return conf
}

// Hardcoded constants
// Number of hashes inside one HashGroupMessage
const HashGroupMessageSize = 4

//Bytes of supported read ahead, the file size the buffered hashes should cover (64M)
const HashChannelMaxBytes = 64 * 1024 * 1024

// Hash size [bytes]
const HashSize = 4

// Size of the HashGroupMessage channel buffer (max number elements)
const HashGroupChannelSize = HashChannelMaxBytes / (HashGroupMessageSize * HashSize)

var SupportedProtocols = []int {1}
