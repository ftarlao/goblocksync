package test

import (
	"github.com/ftarlao/goblocksync/controller/routines"
	"github.com/ftarlao/goblocksync/data/configuration"
	"github.com/ftarlao/goblocksync/data/messages"
	"io"
	"math/rand"
	"reflect"
	"testing"
	"github.com/ftarlao/goblocksync/data/messages/messageutils"
	"time"
	"errors"
	"github.com/ftarlao/goblocksync/utils"
)

func TestUnitNetworkManagerRoundtrip(t *testing.T) {
	t.Log("***NetworkManager***\nCheck Roundtrip for different message types")

	//Base configuration
	confOut := &configuration.Configuration{
		StartLoc:        0,
		IsSource:        true,
		IsMaster:        true,
		DestinationFile: configuration.FileDetails{FileName: "pippo"},
		SourceFile:      configuration.FileDetails{FileName: "topolino"},
		BlockSize:       utils.KB}

	//random generator
	source := rand.NewSource(0)
	rGen := rand.New(source)

	pipeIn, pipeOut := io.Pipe()
	netManager := routines.NewNetworkManager(confOut.BlockSize, pipeIn, pipeOut)
	inMsgChan := netManager.GetInMsgChannel()

	netManager.Start()

	helloOut := messages.NewHelloInfo()
	res := CheckMsgRoundtrip(helloOut, netManager, rGen, t)
	if !res {
		return
	}

	res = CheckMsgRoundtrip(confOut, netManager, rGen, t)
	if !res {
		return
	}

	for i := 0; i < 5; i++ {
		hashGroupMsg := messageutils.RandomHashGroupMessage(rGen, confOut.BlockSize)
		res = CheckMsgRoundtrip(hashGroupMsg, netManager, rGen, t)
		if !res {
			return
		}
	}

	endMessage := messages.NewEndMessage()
	res = CheckMsgRoundtrip(endMessage, netManager, rGen, t)
	if !res {
		return
	}

	errorMessage := messages.NewErrorMessage(errors.New("boom"))
	res = CheckMsgRoundtrip(errorMessage, netManager, rGen, t)
	if !res {
		return
	}

	for i := 0; i < 6; i++ {
		dataBlockMsg := messageutils.RandomDataBlockMessage(rGen, confOut.BlockSize)
		CheckMsgRoundtrip(dataBlockMsg, netManager, rGen, t)
		if !res {
			return
		}
	}

	for i := true; i; {
		select {
		case m := <-inMsgChan:
			t.Error("found one message more than expected, message : ", m)
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
		rType := reflect.TypeOf(msgOut)
		t.Error("Message of type: ", rType, ", FAIL")
		return false
	}
}


func TestUnitNetworkManagerThroughput(t *testing.T) {
	const testDataBytes = utils.GB
	const blockSizeBytes = 128*utils.KB
	t.Log("***NetworkManager***\nCalculate throughput on message roundtrip")


	pipeIn, pipeOut := io.Pipe()
	netManager := routines.NewNetworkManager(blockSizeBytes, pipeIn, pipeOut)
	inMsgChan := netManager.GetInMsgChannel()
	outMsgChan := netManager.GetOutMsgChannel()


	netManager.Start()

	go func() {
		for in := range inMsgChan {
			_ = in
		}
	}()

	t.Log("**HashGroupMessage throughput on message roundtrip**")
	hashMsg, msgPayloadSizeBytes := messageutils.DumbHashGroupMessage(111, blockSizeBytes)
	bechmarkMsgRoundtrip(hashMsg, msgPayloadSizeBytes,testDataBytes, outMsgChan, t)

	t.Log("**DataBlockMessage throughput on message roundtrip**")
	dataMsg, msgPayloadSizeBytes := messageutils.DumbDataBlockMessage(blockSizeBytes)
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
	dataPayloadMB := float64(i)/float64(utils.MB)
	mbSec := dataPayloadMB/duration.Seconds()
	t.Logf("Size of data payload: %.3f MB, Duration [sec]: %.3f  Serialization speed: %.3f MB/s",dataPayloadMB, duration.Seconds(),mbSec)
}
