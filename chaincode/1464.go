package main

import (
	"encoding/base64"
	"github.com/gogo/protobuf/proto"
	fashim "github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
	"hyperledger.abchain.org/adapter/hyfabric/rlp"
	yashim "hyperledger.abchain.org/chaincode/shim"
	"time"
)

const MultipleEvents = "multiple_events"

type faChainCode struct {
	yaCc yashim.Chaincode
	//used for some special rpc, in which the arguments has to be transferred
	//in plaintext rather than bytes, additional decoding of base64 is required
	argEncoded bool
}

// AdaptChainCode convert abchain chaincode to fabric chaincode
func AdaptChainCode(yaCc yashim.Chaincode) fashim.Chaincode {
	return &faChainCode{yaCc: yaCc}
}

// AdaptChainCodeWithArgEncoded act as AdaptChainCode, with input arguments is
// encoded by base64
func AdaptChainCodeWithArgEncoded(yaCc yashim.Chaincode) fashim.Chaincode {
	return &faChainCode{yaCc: yaCc, argEncoded: true}
}

func (f *faChainCode) Init(stub fashim.ChaincodeStubInterface) peer.Response {
	return fashim.Success(nil)
}

func (f *faChainCode) Invoke(stub fashim.ChaincodeStubInterface) peer.Response {
	function := ""
	args := stub.GetArgs()
	if len(args) > 0 {
		function = string(args[0])
		args = args[1:]
	}

	if f.argEncoded {
		for i, arg := range args {
			bt, err := base64.RawStdEncoding.DecodeString(string(arg))
			if err != nil {
				return fashim.Error(err.Error())
			}
			args[i] = bt
		}
	}

	res, err := f.yaCc.Invoke(&adaptChainCodeStub{ChaincodeStubInterface: stub}, function, args, false)
	if err != nil {
		return fashim.Error(err.Error())
	}
	return fashim.Success(res)
}

// helper
type adaptChainCodeStub struct {
	fashim.ChaincodeStubInterface
	chaincodeEvents [][]byte
}

func (s *adaptChainCodeStub) GetTxTime() (time.Time, error) {
	tmp, err := s.GetTxTimestamp()
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(tmp.Seconds, int64(tmp.Nanos)), nil
}

func (s *adaptChainCodeStub) SetEvent(name string, payload []byte) error {
	var err error
	event := &peer.ChaincodeEvent{EventName: name, Payload: payload}
	data, err := proto.Marshal(event)
	if err != nil {
		return err
	}
	s.chaincodeEvents = append(s.chaincodeEvents, data)
	if len(s.chaincodeEvents) == 1 {
		err = s.ChaincodeStubInterface.SetEvent(name, payload)
	} else {
		datas, err := rlp.EncodeToBytes(s.chaincodeEvents)
		if err != nil {
			return err
		}
		err = s.ChaincodeStubInterface.SetEvent(MultipleEvents, datas)
	}
	return err
}

func (s *adaptChainCodeStub) RangeQueryState(startKey, endKey string) (yashim.StateRangeQueryIteratorInterface, error) {
	state, err := s.GetStateByRange(startKey, endKey)
	return &adaptState{state: state}, err
}

func (s *adaptChainCodeStub) GetRawStub() interface{} {
	return s.ChaincodeStubInterface
}

type adaptState struct {
	state fashim.StateQueryIteratorInterface
}

func (s *adaptState) HasNext() bool {
	return s.state.HasNext()
}

func (s *adaptState) Close() error {
	return s.state.Close()
}

func (s *adaptState) Next() (string, []byte, error) {
	next, err := s.state.Next()
	return next.Key, next.Value, err
}
