package test

import (
	"github.com/ftarlao/goblocksync/controller/routines"
	"github.com/ftarlao/goblocksync/data/configuration"
	"github.com/ftarlao/goblocksync/data/messages"
	"io"
	"math/rand"
	"reflect"
	"testing"
	//"errors"
	"time"
	"errors"
)

func TestUnitNetworkManagerRoundtrip(t *testing.T) {
	t.Log("***NetworkManager***\nCheck Roundtrip for different message types")

	//random generator
	source := rand.NewSource(0)
	rgen := rand.New(source)

	pipeIn, pipeOut := io.Pipe()
	netManager := routines.NewNetworkManager(pipeIn, pipeOut)
	inMsgChan := netManager.GetInMsgChannel()

	netManager.Start()

	helloOut := messages.NewHelloInfo()
	res := CheckMsgRoundtrip(helloOut, netManager, rgen, t)
	if !res {
		return
	}

	confOut := &configuration.Configuration{
		StartLoc:        0,
		IsSource:        true,
		IsMaster:        true,
		DestinationFile: configuration.FileDetails{FileName: "pippo"},
		SourceFile:      configuration.FileDetails{FileName: "topolino"},
		BlockSize:       1024}
	res = CheckMsgRoundtrip(confOut, netManager, rgen, t)
	if !res {
		return
	}

	for i := 0; i < 5; i++ {
		hashGroupMsg := randomHashGroupMessage(rgen)
		res = CheckMsgRoundtrip(hashGroupMsg, netManager, rgen, t)
		if !res {
			return
		}
	}

	endMessage := messages.NewEndMessage()
	res = CheckMsgRoundtrip(endMessage, netManager, rgen, t)
	if !res {
		return
	}

	errorMessage := messages.NewErrorMessage(errors.New("boom"))
	res = CheckMsgRoundtrip(errorMessage, netManager, rgen, t)
	if !res {
		return
	}

	for i := 0; i < 6; i++ {
		dataBlockMsg := RandomDataBlockMessage(rgen, confOut.BlockSize)
		CheckMsgRoundtrip(dataBlockMsg, netManager, rgen, t)
		if !res {
			return
		}
	}

	for i := true; i; {
		select {
		case m := <-inMsgChan:
			t.Error("found message more than expected, message ID: ", m.GetMessageID())
			return
		case <-time.After(1 * time.Second):
			{
				t.Log("No other messages")
				i = false
			}
		}
	}

	err := netManager.Stop()
	if err != nil {
		t.Error(err)
	}
}

func randomHashGroupMessage(rgen *rand.Rand) *messages.HashGroupMessage {
	hashGroupMsg := messages.NewHashGroupMessage(11)
	numHash := rgen.Intn(configuration.HashGroupMessageSize)

	for i := 0; i < numHash; i++ {
		hash := make([]byte, configuration.HashSize)
		rgen.Read(hash)
		hashGroupMsg.AddHash(hash)
	}

	return hashGroupMsg
}

func RandomDataBlockMessage(rgen *rand.Rand, blockSize int64) *messages.DataBlockMessage {
	data := make([]byte, blockSize)
	rgen.Read(data)
	dataBlockMsg := messages.NewDataBlockMessage(12, data)
	return dataBlockMsg
}

func CheckMsgRoundtrip(msgOut messages.Message, netManager routines.NetworkManager, rgen *rand.Rand, t *testing.T) bool {
	outMsgChan := netManager.GetOutMsgChannel()
	inMsgChan := netManager.GetInMsgChannel()

	outMsgChan <- msgOut
	msgIn := <-inMsgChan
	return checkMsgEquality(msgOut, msgIn, t)
}

func checkMsgEquality(msgOut messages.Message, msgIn messages.Message, t *testing.T) bool {
	if reflect.DeepEqual(msgOut, msgIn) {
		t.Log("Message of type: ", reflect.TypeOf(msgOut), ", OK")
		return true
	} else {
		rtype := reflect.TypeOf(msgOut)
		t.Error("Message of type: ", rtype, ", FAIL")
		return false
	}
}


func TestUnitNetworkManagerThroughput(t *testing.T) {
	const testDataBytes = 1024 * 1024 *1024

	t.Log("***NetworkManager***\nCalculate throughput on message roundtrip")


	pipeIn, pipeOut := io.Pipe()
	netManager := routines.NewNetworkManager(pipeIn, pipeOut)
	inMsgChan := netManager.GetInMsgChannel()
	outMsgChan := netManager.GetOutMsgChannel()


	netManager.Start()

	go func() {
		for in := range inMsgChan {
			_ = in
		}
	}()

	t.Log("***HashGroupMessage throughput on message roundtrip")
	hashMsg, msgPayloadSizeBytes := dumbHashGroupMessage(111)
	bechmarkMsgRoundtrip(hashMsg, msgPayloadSizeBytes,testDataBytes, outMsgChan, t)

	t.Log("***DataBlockMessage throughput on message roundtrip")
	dataMsg, msgPayloadSizeBytes := dumbDataBlockMessage(128*1024)
	bechmarkMsgRoundtrip(dataMsg, msgPayloadSizeBytes,testDataBytes, outMsgChan, t)

	close(outMsgChan)
}

func bechmarkMsgRoundtrip(msg messages.Message, msgSize int64, testDataBytes int64, outMsgChan chan messages.Message, t *testing.T){
	start := time.Now()
	var i int64 = 0
	for ; i < testDataBytes; i+=msgSize {
		outMsgChan <- msg
	}

	duration := time.Since(start)
	dataPayloadMB := float64(i)/float64(1024*1024)
	mbSec := dataPayloadMB/duration.Seconds()
	t.Logf("Size of data payload: %.3f MB, Duration [sec]: %.3f  Serialization speed: %.3f MB/s",dataPayloadMB, duration.Seconds(),mbSec)
}

func dumbHashGroupMessage(seed int64) (*messages.HashGroupMessage, int64) {
	hashGroupMsg := messages.NewHashGroupMessage(seed)

	for i := 0; i < configuration.HashGroupMessageSize; i++ {
		hash := make([]byte, configuration.HashSize)
		for j := range hash {
			hash[j] = byte(i % 256)
		}
		hashGroupMsg.AddHash(hash)
	}

	return hashGroupMsg, configuration.HashGroupMessageSize*configuration.HashSize+8+2
}

func dumbDataBlockMessage(blockSizeBytes int64) (*messages.DataBlockMessage, int64) {
	data := make([]byte, blockSizeBytes)
	for j := range data {
		data[j] = byte(j % 256)
	}
	dataBlockMsg := messages.NewDataBlockMessage(12, data)
	return dataBlockMsg, blockSizeBytes + 8
}
