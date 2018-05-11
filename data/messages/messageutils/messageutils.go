package messageutils

import (
	"github.com/ftarlao/goblocksync/data/configuration"
	"github.com/ftarlao/goblocksync/data/messages"
	"math/rand"
)

//Message generation methods may be invoked in the same thread or sequentially in order to obtain an exact location-wise sequence
//..elsewhere there are no issues

var randomDumbHashGroupCount int64 = 0

func DumbHashGroupMessage(seed int64, blockSizeBytes int64) (*messages.HashGroupMessage, int64) {
	hashGroupMsg := messages.NewHashGroupMessage(randomDumbHashGroupCount * configuration.HashGroupMessageSize * blockSizeBytes)

	for i := 0; i < configuration.HashGroupMessageSize; i++ {
		hash := make([]byte, configuration.HashSize)
		for j := range hash {
			hash[j] = byte(i % 256)
		}
		hashGroupMsg.AddHash(hash)
	}

	randomDumbHashGroupCount++
	return hashGroupMsg, configuration.HashGroupMessageSize*configuration.HashSize + 8 + 2
}

var randomDumbBlockCount int64 = 0

func DumbDataBlockMessage(blockSizeBytes int64) (*messages.DataBlockMessage, int64) {
	data := make([]byte, blockSizeBytes)
	for j := range data {
		data[j] = byte(j % 256)
	}
	dataBlockMsg := messages.NewDataBlockMessage(randomDumbBlockCount, data)
	randomDumbBlockCount++
	return dataBlockMsg, blockSizeBytes + 8
}

var randomHashGroupCount int64 = 0

func RandomHashGroupMessage(rgen *rand.Rand, blockSizeBytes int64) *messages.HashGroupMessage {
	numHash := rgen.Intn(configuration.HashGroupMessageSize)
	hashGroupMsg := messages.NewHashGroupMessage(randomHashGroupCount * int64(numHash) * blockSizeBytes)

	for i := 0; i < numHash; i++ {
		hash := make([]byte, configuration.HashSize)
		rgen.Read(hash)
		hashGroupMsg.AddHash(hash)
	}

	randomHashGroupCount++
	return hashGroupMsg
}

var randomDataBlockCount int64 = 0

func RandomDataBlockMessage(rgen *rand.Rand, blockSizeBytes int64) *messages.DataBlockMessage {
	data := make([]byte, blockSizeBytes)
	rgen.Read(data)

	dataBlockMsg := messages.NewDataBlockMessage(randomDataBlockCount*blockSizeBytes, data)

	randomDataBlockCount++
	return dataBlockMsg
}
