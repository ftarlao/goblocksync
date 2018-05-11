package test

import (
	"errors"
	"github.com/ftarlao/goblocksync/controller/routines"
	"github.com/ftarlao/goblocksync/data/configuration"
	"github.com/ftarlao/goblocksync/data/messages"
	"github.com/ftarlao/goblocksync/data/messages/messageutils"
	"github.com/ftarlao/goblocksync/utils"
	"io"
	"math/rand"
	"reflect"
	"testing"
	"time"
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
	netManager := routines.NewNetworkManager(confOut.EstimateNetworkChannelSize(), pipeIn, pipeOut)
	inMsgChan := netManager.GetInMsgChannel()

	netManager.Start()

	helloOut := messages.NewHelloInfo()
	res := CheckMsgRoundtrip(helloOut, netManager, t)
	if !res {
		return
	}

	res = CheckMsgRoundtrip(confOut, netManager, t)
	if !res {
		return
	}

	//Send a bunch of them
	for i := 0; i < 6; i++ {
		hashGroupMsg := messageutils.RandomHashGroupMessage(rGen, confOut.BlockSize)
		res = CheckMsgRoundtrip(hashGroupMsg, netManager, t)
		if !res {
			return
		}
	}

	endMessage := messages.NewEndMessage()
	res = CheckMsgRoundtrip(endMessage, netManager, t)
	if !res {
		return
	}

	errorMessage := messages.NewErrorMessage(errors.New("boom"))
	res = CheckMsgRoundtrip(errorMessage, netManager, t)
	if !res {
		return
	}

	//Send a bunch of them
	for i := 0; i < 6; i++ {
		dataBlockMsg := messageutils.RandomDataBlockMessage(rGen, confOut.BlockSize)
		CheckMsgRoundtrip(dataBlockMsg, netManager, t)
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

func CheckMsgRoundtrip(msgOut messages.Message, netManager routines.NetworkManager, t *testing.T) bool {
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

	t.Log("***NetworkManager***\nCalculate throughput on message roundtrip")

	//dummy configuration
	confOut := &configuration.Configuration{
		StartLoc:        0,
		IsSource:        true,
		IsMaster:        true,
		DestinationFile: configuration.FileDetails{FileName: "pippo"},
		SourceFile:      configuration.FileDetails{FileName: "topolino"},
		BlockSize:       128 * utils.KB}

	//setup roundtrip NetworkManager, output returns as input
	pipeIn, pipeOut := io.Pipe()
	netManager := routines.NewNetworkManager(confOut.EstimateNetworkChannelSize(), pipeIn, pipeOut)
	inMsgChan := netManager.GetInMsgChannel()
	outMsgChan := netManager.GetOutMsgChannel()

	netManager.Start()

	go func() {
		for in := range inMsgChan {
			_ = in
		}
	}()

	t.Log("**HashGroupMessage throughput on message roundtrip**")
	hashMsg, msgPayloadSizeBytes := messageutils.DumbHashGroupMessage(111, confOut.BlockSize)
	benchmarkMsgRoundtrip(hashMsg, msgPayloadSizeBytes, testDataBytes, outMsgChan, t)

	t.Log("**DataBlockMessage throughput on message roundtrip**")
	dataMsg, msgPayloadSizeBytes := messageutils.DumbDataBlockMessage(confOut.BlockSize)
	benchmarkMsgRoundtrip(dataMsg, msgPayloadSizeBytes, testDataBytes, outMsgChan, t)

	err := netManager.Stop()
	if err != nil {
		t.Error(err)
	}
}

func benchmarkMsgRoundtrip(msg messages.Message, msgSize int64, testDataBytes int64, outMsgChan chan messages.Message, t *testing.T) {
	start := time.Now()
	var i int64 = 0
	for ; i < testDataBytes; i += msgSize {
		outMsgChan <- msg
	}

	duration := time.Since(start)
	dataPayloadMB := float64(i) / float64(utils.MB)
	mbSec := dataPayloadMB / duration.Seconds()
	t.Logf("Size of data payload: %.3f MB, Duration [sec]: %.3f  Serialization speed: %.3f MB/s", dataPayloadMB, duration.Seconds(), mbSec)
}
