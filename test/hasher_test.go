package test

import (
	"bytes"
	"github.com/ftarlao/goblocksync/controller/routines"
	"github.com/ftarlao/goblocksync/data/messages"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const TestTimeout = 5 * time.Second

func TestHasherImpl(t *testing.T) {
	fileSizeBytes := int64(8511)
	blockSizeBytes := int64(32)
	periodSizeBytes := int64(96)
	paramTestHasherImpl(t,fileSizeBytes,blockSizeBytes,periodSizeBytes)
	fileSizeBytes = int64(21200)
	blockSizeBytes = int64(128)
	periodSizeBytes = int64(1024)
	paramTestHasherImpl(t,fileSizeBytes,blockSizeBytes,periodSizeBytes)
}

func paramTestHasherImpl(t *testing.T, fileSizeBytes,blockSizeBytes,periodSizeBytes int64) {
	t.Log("***Hasher Test***")
	t.Log("File size [bytes]: ", fileSizeBytes, ", block size [bytes]: ", blockSizeBytes)

	//Extimated number of hashes
	var expectedNumberHash int = int(math.Ceil(float64(fileSizeBytes) / float64(blockSizeBytes)))

	//Create temp file
	f, err := createTmpFile(fileSizeBytes, periodSizeBytes, 0)
	if err != nil {
		t.Errorf("Cannot open tmp file for testing")
	}
	defer f.Close()

	//Init hashing facility
	hasher := routines.NewHasherImpl(blockSizeBytes, f, 0, routines.DummyHash)
	outMsg := hasher.GetOutMsgChannel()
	hasher.Start()

	//Read hashing messages from the Hasher
	numHash := 0
	hashStorage := make([][]byte, 0, expectedNumberHash)

	var msg messages.Message
MainLoop:
	for msg == nil || msg.GetMessageID() != messages.EndMessageID {
		select {
		case msg = <-outMsg:
			switch msg.GetMessageID() {
			case messages.HashGroupMessageID:
				hMsg := msg.(*messages.HashGroupMessage)
				numHash += len(hMsg.HashGroup)
				hashStorage = append(hashStorage, hMsg.HashGroup...)
			case messages.ErrorMessageID:
				t.Error("error returned from hasher: ", msg.(*messages.ErrorMessage).Err)
				break MainLoop
			}
		case <-time.After(TestTimeout):
			t.Error("Timeout for Hasher, no EndMessage or no messages in queue")
			break MainLoop
		}

	}
	t.Log("Expected number of Hashes: ", expectedNumberHash, ", returned: ", numHash)

	//Check number of Hash
	if numHash != expectedNumberHash {
		t.Error("wrong number of Hash")
		return
	}

	//Check periodicity
	periodSteps := int(periodSizeBytes / blockSizeBytes)
	t.Log("Expected the Hashes to repeat every ", periodSteps, " blocks (the periodicity)")
	//note last hash is ignored (the block can be incomplete)
	for i := 0; i < len(hashStorage)-periodSteps-1; i++ {
		if i < periodSteps+1 && i < len(hashStorage) {
			t.Log("Hash ", i, ", value: ", hashStorage[i])
		}
		if !bytes.Equal(hashStorage[i], hashStorage[i+periodSteps]) {
			t.Error("Hash number ", i, " is not equal to Hash number ", i+periodSteps)
			return
		}
	}
	t.Log("Periodicity OK")

}

// Creates a tmp file in the tmp project folder, the file is randomly generated but can have a periodicity i.e. being made
// by a sequence that repeats.
// periodBytes = 0 means periodicity disabled
// seed is the random seed
func createTmpFile(size int64, periodBytes int64, seed int64) (f *os.File, err error) {

	if periodBytes == 0 {
		//Disable periodicity
		periodBytes = size
	}

	//Init the base random sequence
	source := rand.NewSource(seed)
	rgen := rand.New(source)
	sequence := make([]byte, periodBytes)
	rgen.Read(sequence)

	tempFileName, err := filepath.Abs("../tmp")
	if err != nil {
		return nil, err
	}
	f, err = ioutil.TempFile(tempFileName, "data")
	if err != nil {
		return nil, err
	}
	data := make([]byte, size)

	for i := int64(0); i < size; i += int64(len(sequence)) {
		copy(data[i:], sequence)
	}
	_, err = f.Write(data)
	return f, err
}
