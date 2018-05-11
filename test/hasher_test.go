package test

import (
	"bytes"
	"github.com/ftarlao/goblocksync/controller/routines"
	"github.com/ftarlao/goblocksync/data/messages"
	"github.com/ftarlao/goblocksync/utils"
	"math"
	"math/rand"
	"testing"
	"time"
	"io"
)

const TestTimeout = 5 * time.Second

func TestUnitHasherImpl(t *testing.T) {
	fileSizeBytes := int64(8511)
	blockSizeBytes := int64(32)
	periodSizeBytes := int64(96)

	testHasherImpl(true, t, fileSizeBytes, blockSizeBytes, periodSizeBytes)

	//Corner case.. data is exactly int(N) blocks
	fileSizeBytes = int64(32 * utils.KB)
	blockSizeBytes = int64(128)
	periodSizeBytes = int64(utils.KB)

	testHasherImpl(true, t, fileSizeBytes, blockSizeBytes, periodSizeBytes)
}

//Creates file on disk or ram, calculates the hash(es), and checks, errors, hash(es) # and periodicity
func testHasherImpl(userRam bool, t *testing.T, fileSizeBytes, blockSizeBytes, periodSizeBytes int64) {
	t.Log("***Hasher Test***")
	t.Log("File size [bytes]: ", fileSizeBytes, ", block size [bytes]: ", blockSizeBytes)

	//Estimated number of hashes
	var expectedNumberHash = int(math.Ceil(float64(fileSizeBytes) / float64(blockSizeBytes)))

	var f io.ReadSeeker
	var err error

	if !userRam {
		//Create temp file
		fCloser, err := utils.CreateTmpFile(fileSizeBytes, periodSizeBytes, 0)
		if err != nil {
			t.Error("Cannot open tmp file for testing")
		}
		f = fCloser
		defer fCloser.Close()
	} else {
		//Simulates a file, in RAM
		f = utils.CreateTmpRamReader(fileSizeBytes,periodSizeBytes, 0)
	}
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
			case messages.EndMessageID:
				t.Log("EndMessage received from Hasher")
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

	//Checks periodicity
	periodSteps := int(periodSizeBytes / blockSizeBytes)
	t.Log("Expected the Hashes to repeat every ", periodSteps, " blocks (the periodicity)")
	//note last hash is ignored (the block can be incomplete)
	windowViewSize := len(hashStorage) - periodSteps - 1
	t.Log("First ", windowViewSize, " collected hash values:")
	for i := 0; i < windowViewSize; i++ {
		if i < periodSteps+1 && i < len(hashStorage) {
			t.Log("Hash ", i, ", value: ", hashStorage[i])
		}
		if !bytes.Equal(hashStorage[i], hashStorage[i+periodSteps]) {
			t.Error("Hash number ", i, " is not equal to Hash number ", i+periodSteps)
			return
		}
	}
	t.Log("Periodicity OK")

	err = hasher.Stop()
	if err != nil {
		t.Error(err)
	}
}


func TestBenchHasherImpl(t *testing.T) {
	t.Log("Test Read speed with no-ops hash algorithm")
	//Init the base random sequence
	source := rand.NewSource(11111)
	rgen := rand.New(source)
	var size int64 = utils.GB

	data := make([]byte, size)
	rgen.Read(data)
	fakeFile := bytes.NewReader(data)

	//Init hashing facility
	hasher := routines.NewHasherImpl(16*utils.MB, fakeFile, 0, routines.FakeHash)
	outMsg := hasher.GetOutMsgChannel()

	start := time.Now()
	hasher.Start()

	var msg messages.Message
MainLoop:
	for msg == nil || msg.GetMessageID() != messages.EndMessageID {
		select {
		case msg = <-outMsg:
			if msg.GetMessageID() == messages.ErrorMessageID {
				t.Error("error returned from hasher: ", msg.(*messages.ErrorMessage).Err)
				break MainLoop
			}
		}
	}
	t.Log("Last message is ", msg)

	//The data has been read from RAMdisk, cause data is saved into DataBlockMessage structures, this doubles the RAM
	//accesses; the effective max read speed may be a number between the estimated value and the double.
	duration := time.Since(start)
	dataPayloadMB := float64(size) / float64(utils.MB)
	mbSec := dataPayloadMB / duration.Seconds()
	t.Logf("Data has been read from a ramdisk, Hasher is able to read data with a speed in [%.2f,%.2f] MB/s", mbSec, 2*mbSec)
	err := hasher.Stop()
	if err != nil {
		t.Error(err)
	}
}
