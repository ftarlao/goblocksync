package configuration

import (
	"errors"
	"github.com/ftarlao/goblocksync/utils"
	"os"
)

//TODO Should few details about Master file names be masked .. and useless fields emptied? Less infos to the slave peer
//TODO ?? build constructor?? More uniform with other messages?
const ConfigurationMessageID byte = 3

type Configuration struct {
	// IsMaster or IsSlave
	IsMaster bool
	// IsSource or IsDestination
	IsSource        bool
	SourceFile      FileDetails
	DestinationFile FileDetails
	// Starting file location [bytes]
	StartLoc int64
	// BlockSize [bytes]
	BlockSize int64
}

//TODO integrate validation
func (c *Configuration) Validate() (bool, error) {
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

//Creates the configuration that should be provided to the remote peer
func (c *Configuration) Complement() Configuration {
	conf := *c //copy value
	conf.IsMaster = !c.IsMaster
	conf.IsSource = !c.IsSource
	return conf
}

func (*Configuration) GetMessageID() byte {
	return ConfigurationMessageID
}

func (c *Configuration) EstimateNetworkChannelSize() int {
	maxMessageApproxSize := utils.IntMax(c.BlockSize, HashGroupMessageSize*HashSize)
	networkChannelSize := (NetworkMaxBytes / maxMessageApproxSize) / 2
	return int(networkChannelSize)
}

type FileDetails struct {
	FileName string
	//file size [Bytes]
	size int64
	// true when file represents a device (Linux/Unix/iOS)
	isDevice bool
}

func (f FileDetails) Update() (bool, error) {
	fileInfo, err := os.Stat("/path/to/file")
	if err != nil {
		return false, err
	}
	// get the size
	f.size = fileInfo.Size()
	f.isDevice = false
	//TODO to operate with block devices.. for real
	return true, err
}
