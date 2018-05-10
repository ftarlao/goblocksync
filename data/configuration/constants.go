package configuration

import "github.com/ftarlao/goblocksync/utils"

// Hardcoded constants
const MajorVersion = 0
const Version = 1
const PatchVersion = 0

// Number of hashes inside one HashGroupMessage
const HashGroupMessageSize = 200

//Bytes for buffered hashes (64M)
const HashMaxBytes = 64 * utils.MB
//Bytes for buffered queued data (64M)
const DataMaxBytes = 64 * utils.MB

// Hash size [bytes], this is currently used by the dumb hash function
const HashSize = 32

// Size of the HashGroupMessage channel buffer (max number elements in the channel)
const HashGroupChannelSize = HashMaxBytes / (HashGroupMessageSize * HashSize)

var SupportedProtocols = []int{1}

// Max number of messages in the message queue, this should be only a small buffer (we have TCP buffers, other queues..)
// The effective max size [bytes] depends on the message types, max block size.. it should range (approximately) between:
// BlockSize * NetworkChannelsSize > size_bytes > HashGroupMessageSize * HashSize * NetworkChannelsSize
const NetworkMaxBytes = 64 * utils.MB
