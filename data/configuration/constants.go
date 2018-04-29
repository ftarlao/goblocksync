package configuration

// Hardcoded constants
const Version = 0
const MinorVersion = 1

// Number of hashes inside one HashGroupMessage
const HashGroupMessageSize = 50

//Bytes for buffered hashes (64M)
const HashMaxBytes = 64 * 1024 * 1024

// Hash size [bytes], this is currently used by the dumb hash function
const HashSize = 4

// Size of the HashGroupMessage channel buffer (max number elements in the channel)
const HashGroupChannelSize = HashMaxBytes / (HashGroupMessageSize * HashSize)

var SupportedProtocols = []int{1}

// Max number of messages in the message queue, this should be only a small buffer (we have TCP buffers, other queues..)
// The effective max size [bytes] depends on the message types, max block size.. it should range (approximately) between:
// BlockSize * NetworkChannelsSize > size_bytes > HashGroupMessageSize * HashSize * NetworkChannelsSize
const NetworkChannelsSize = 100
