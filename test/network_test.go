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
	res := checkMsgRoundtrip(helloOut, netManager, rgen, t)
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
	res = checkMsgRoundtrip(confOut, netManager, rgen, t)
	if !res {
		return
	}

	for i := 0; i < 5; i++ {
		hashGroupMsg := randomHashGroupMessage(rgen)
		res = checkMsgRoundtrip(hashGroupMsg, netManager, rgen, t)
		if !res {
			return
		}
	}

	endMessage := messages.NewEndMessage()
	res = checkMsgRoundtrip(endMessage, netManager, rgen, t)
	if !res {
		return
	}

	errorMessage := messages.NewErrorMessage(errors.New("boom"))
	res = checkMsgRoundtrip(errorMessage, netManager, rgen, t)
	if !res {
		return
	}

	for i := 0; i < 6; i++ {
		dataBlockMsg := randomDataBloclMessage(rgen)
		checkMsgRoundtrip(dataBlockMsg, netManager, rgen, t)
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

func randomDataBloclMessage(rgen *rand.Rand) *messages.DataBlockMessage {
	data := make([]byte, 1200)
	rgen.Read(data)
	dataBlockMsg := messages.NewDataBlockMessage(12, data)
	return dataBlockMsg
}

func checkMsgRoundtrip(msgOut messages.Message, netManager routines.NetworkManager, rgen *rand.Rand, t *testing.T) bool {
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
