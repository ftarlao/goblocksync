package messages

import "goblocksync/data/configuration"

const EndMessageID byte = 4

type EndMessage struct {
}

func NewEndMessage() EndMessage {
	return EndMessage{}
}

func (EndMessage) GetMessageID() byte {
	return EndMessageID
}
