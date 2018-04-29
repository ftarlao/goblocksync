package test

import (
	"testing"
	"io/ioutil"
	"os"
	"math"
	"math/rand"
	"github.com/ftarlao/goblocksync/data/messages"
	"github.com/ftarlao/goblocksync/controller/routines"
	"time"
	"path/filepath"
	"bytes"
)

const TestTimeout = 5 * time.Second

func TestHasherImpl(t *testing.T) {
	t.Log("Started Hasher Test")
	fileSizeBytes := int64(8511)
	blockSizeBytes := int64(32)
	periodSizeBytes := int64(96)

	t.Log("File size [bytes]: ",fileSizeBytes,", block size [bytes]: ",blockSizeBytes)

	//Extimated number of hashes
	var expectedNumberHash int = int(math.Ceil(float64(fileSizeBytes) / float64(blockSizeBytes)))

	//Create temp file
	f, err := createTmpFile(fileSizeBytes,periodSizeBytes,0)
	if err != nil {
		t.Errorf("Cannot open tmp file for testing")
	}

	//Init hashing facility
	hasher := routines.NewHasherImpl(blockSizeBytes, f, 0)
	outMsg := hasher.GetOutMsgChannel()
	hasher.Start()

	//Read hashing messages from the Hasher
	numHash := 0
	hashStorage := make([][]byte, 0, expectedNumberHash)

	var msg messages.Message
	MainLoop:
	for msg == nil || msg.GetMessageID() != messages.EndMessageID {
		select {
		case msg = <- outMsg:
			switch msg.GetMessageID() {
			case messages.HashGroupMessageID:
				hMsg := msg.(*messages.HashGroupMessage)
				numHash+= len(hMsg.HashGroup)
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
	t.Log("Expected number of Hashes: ", expectedNumberHash,", returned: ", numHash)

	//Check number of Hash
	if numHash != expectedNumberHash {
		t.Error("wrong number of Hash")
		return
	}

	//Check periodicity
	periodSteps := int(periodSizeBytes / blockSizeBytes)
	t.Log("Expected the Hashes to repeat every ", periodSteps," blocks (the periodicity)")
	//note last hash is ignored (the block can be incomplete)
	for i:=0; i< len(hashStorage)-periodSteps-1; i++{
		if !bytes.Equal(hashStorage[i],hashStorage[i+periodSteps]){
			t.Error("Hash number ",i," is not equal to Hash number ",i+periodSteps)
			return
		}
	}
	t.Log("Periodicity OK")

}

func createTmpFile(size int64, periodBytes int64, seed int64) (f *os.File, err error){
	rand.Seed(seed)
	sequence := make([]byte, periodBytes)
	tempFileName, err := filepath.Abs("../tmp")
	if err!=nil {
		return nil,err
	}
	f, err = ioutil.TempFile(tempFileName,"data")
	if err!=nil {
		return nil,err
	}
	data := make([]byte,size)

	for  i := int64(0); i < size; i+=int64(len(sequence)) {
		copy(data[i:],sequence)
	}
	_,err = f.Write(data)
	return f,err
}