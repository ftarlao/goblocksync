package messages

const ErrorMessageID byte = 2

type ErrorMessage struct {
	Err string
}

func NewErrorMessage(err error) *ErrorMessage {
	return &ErrorMessage{err.Error()}
}

func (*ErrorMessage) GetMessageID() byte {
	return ErrorMessageID
}
