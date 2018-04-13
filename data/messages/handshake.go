package messages

import "goblocksync/data/configuration"

type HelloInfo struct {
	Hello string
	SupportedProtocols []int
}

func NewHelloInfo() HelloInfo {
	return HelloInfo{"goblocksync",configuration.SupportedProtocols}
}