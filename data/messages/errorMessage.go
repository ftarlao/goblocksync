package messages

const ErrorMessageID byte = 2

type ErrorMessage struct {
	Err error
}

func NewErrorMessage(err error) *ErrorMessage {
	return &ErrorMessage{err}
}

func (*ErrorMessage) GetMessageID() byte {
	return ErrorMessageID
}
