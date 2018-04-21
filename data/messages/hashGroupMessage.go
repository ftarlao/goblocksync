package messages

import "goblocksync/data/configuration"

const HashGroupMessageID byte = 0

type HashGroupMessage struct {
	StartLoc  int64
	HashGroup [][]byte
}

func NewHashMessage(startLoc int64) HashGroupMessage {
	return HashGroupMessage{startLoc, make([][]byte, configuration.HashGroupMessageSize)}
}

func NewHashMessageEOF() HashGroupMessage {
	return HashGroupMessage{-1, nil}
}

func (m *HashGroupMessage) TruncHashGroup(size int) {
	m.HashGroup = m.HashGroup[:size]
}

func (m *HashGroupMessage) isEof() bool {
	return m.StartLoc < 0
}

func (HashGroupMessage) GetMessageID() byte {
	return HashGroupMessageID
}
