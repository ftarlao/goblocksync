package messages

import (
	"encoding/gob"
	"errors"
	"goblocksync/data/configuration"
)

type Message interface {
	GetMessageID() byte
}

func EncodeMessage(encoder *gob.Encoder, m Message) error {
	encoder.Encode(m.GetMessageID())
	err := encoder.Encode(m)
	return err
}

func DecodeMessage(decoder *gob.Decoder) (m Message, err error) {
	var msgID byte
	err = decoder.Decode(&msgID)
	if err != nil {
		return
	}
	switch msgID {
	case HelloInfoMessageID:
		var msg HelloInfoMessage
		err = decoder.Decode(&msg)
		m = msg
	case HashGroupMessageID:
		var msg HashGroupMessage
		err = decoder.Decode(&msg)
		m = msg
	case configuration.ConfigurationMessageID:
		var msg configuration.Configuration
		err = decoder.Decode(&msg)
		m = msg
	default:
		err = errors.New("unknown message ID")
	}
	return
}