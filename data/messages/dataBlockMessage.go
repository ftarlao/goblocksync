package messages

const DataBlockMessageID byte = 5

type DataBlockMessage struct {
	StartLoc  int64
	Data []byte
	//Hash of data, normally is null
	Hash []byte
}

func NewDataBlockMessage(startLoc int64, dataBlock []byte) *DataBlockMessage {
	return &(DataBlockMessage{StartLoc:startLoc, Data:dataBlock})
}

func (*DataBlockMessage) GetMessageID() byte {
	return DataBlockMessageID
}