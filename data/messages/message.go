package messages

import (
	"encoding/gob"
	"errors"
	"github.com/ftarlao/goblocksync/data/configuration"
)

type Message interface {
	GetMessageID() byte
}

func EncodeMessage(encoder *gob.Encoder, m Message) error {
	id := m.GetMessageID()
	encoder.Encode(id)
	err := encoder.Encode(m)
	return err
}

func DecodeMessage(decoder *gob.Decoder) (m Message, err error) {
	var msgID byte
	err = decoder.Decode(&msgID)
	if err != nil {
		return nil, err
	}
	switch msgID {
	case HashGroupMessageID:
		var msg HashGroupMessage
		err = decoder.Decode(&msg)
		m = &msg
	case DataBlockMessageID:
		var msg DataBlockMessage
		err = decoder.Decode(&msg)
		m = &msg
	case configuration.ConfigurationMessageID:
		var msg configuration.Configuration
		err = decoder.Decode(&msg)
		m = &msg
	case ErrorMessageID:
		var msg ErrorMessage
		err = decoder.Decode(&msg)
		m = &msg
	case EndMessageID:
		var msg EndMessage
		err = decoder.Decode(&msg)
		m = &msg
	case HelloInfoMessageID:
		var msg HelloInfoMessage
		err = decoder.Decode(&msg)
		m = &msg
	default:
		err = errors.New("unknown message ID")
	}
	return
}
