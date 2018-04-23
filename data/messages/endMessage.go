package messages

const EndMessageID byte = 4

type EndMessage struct {
}

func NewEndMessage() EndMessage {
	return EndMessage{}
}

func (EndMessage) GetMessageID() byte {
	return EndMessageID
}
