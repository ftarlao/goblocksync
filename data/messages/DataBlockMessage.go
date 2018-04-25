package messages

import (
	"errors"
	"github.com/ftarlao/goblocksync/data/configuration"
)

const DataBlockMessageID byte = 5

type DataBlockMessage struct {
	StartLoc  int64
	Data []byte
	//Hash of data, normally is null
	Hash []hash
}

func NewDataBlockMessage(startLoc int64, dataBlock []byte) *DataBlockMessage {
	return &(DataBlockMessage{StartLoc:startLoc, Data:dataBlock})
}

func (*DataBlockMessage) GetMessageID() byte {
	return DataBlockMessageID
}