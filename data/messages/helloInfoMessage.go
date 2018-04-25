package messages

import "github.com/ftarlao/goblocksync/data/configuration"

const HelloInfoMessageID byte = 1

type HelloInfoMessage struct {
	Hello              string
	SupportedProtocols []int
}

func NewHelloInfo() *HelloInfoMessage {
	return &HelloInfoMessage{"goblocksync", configuration.SupportedProtocols}
}

func (*HelloInfoMessage) GetMessageID() byte {
	return HelloInfoMessageID
}
