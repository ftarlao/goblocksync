package messages

import (
	"github.com/ftarlao/goblocksync/data/configuration"
)

const HashGroupMessageID byte = 0

type HashGroupMessage struct {
	StartLoc  int64
	NumHash   int16
	HashGroup [][]byte
}

func NewHashGroupMessage(startLoc int64) *HashGroupMessage {
	return &HashGroupMessage{startLoc, 0, make([][]byte, configuration.HashGroupMessageSize)}
}

func (m *HashGroupMessage) TruncHashGroup() {
	m.HashGroup = m.HashGroup[:m.NumHash]
}

func (*HashGroupMessage) GetMessageID() byte {
	return HashGroupMessageID
}

// Add an hash at the end of the group, this method does not check if there is free capacity, I encourage checking the
// isFull return value
func (m *HashGroupMessage) AddHash(hash []byte) (isFull bool) {
	m.HashGroup[m.NumHash] = hash
	m.NumHash++
	return m.IsFull()
}

func (m *HashGroupMessage) IsFull() (isFull bool) {
	isFull = configuration.HashGroupMessageSize <= m.NumHash //<= i.e., always be defensive
	return
}

func (m *HashGroupMessage) IsEmpty() (isEmpty bool) {
	return m.NumHash == 0
}
