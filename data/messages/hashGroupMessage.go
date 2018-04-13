package messages

import "goblocksync/data/configuration"

type HashGroupMessage struct {
	StartLoc  int64
	HashGroup [][]byte
}

func NewHashMessage(startLoc int64) HashGroupMessage {
	return HashGroupMessage{startLoc,make([][]byte,configuration.HashGroupMessageSize)}
}

func (m HashGroupMessage) TruncHashGroup(size int){
	m.HashGroup = m.HashGroup[:size]
}
