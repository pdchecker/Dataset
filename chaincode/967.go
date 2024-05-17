package main

import (
	"fmt"
	"strconv"
	"encoding/json"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (s *SmartContract) Feed(ctx contractapi.TransactionContextInterface, fungusId uint, FeedId uint) error {
			
	//get fungusDna : (사용자입력) 버섯id -> 버섯 Dna 확보
	fungusJSON, err := s._getState(ctx, strconv.Itoa(int(fungusId)))
	if err !=nil {
		return fmt.Errorf("failed to get fungus: %v", err)
	}

	var fungus Fungus
	err = json.Unmarshal(fungusJSON, &fungus)
	if err != nil {
		return err
	}

	// check readytime
	unixtime := time.Now().Unix()
	if uint32(unixtime) < fungus.ReadyTime {
		return fmt.Errorf("failed to get Multiply : Not Ready")
	}
	
	// make args for InvokeChaincode ( feedfactory, GetFeed() )
	params := []string{"GetFeed", strconv.Itoa(int(FeedId))}
	invokeargs := make ([][]byte, len(params))
	for i, arg := range params {
		invokeargs[i] = []byte(arg)
	}

	// Get feedDna
	response := ctx.GetStub().InvokeChaincode("feedfactory", invokeargs, "mychannel")
	if response.Status != 200 {
		return fmt.Errorf("failed to InvokeChaincode: %s", response.Message)
	}
	// Fungus Asset describes basic details
	var feed struct {		
		Dna			uint	`json:"dna"`
	}

	err = json.Unmarshal(response.Payload, &feed)
	if err != nil {
		return err
	}
	
	err = s._feedAndMultiply(ctx, fungus.Dna, feed.Dna)
	if err != nil {
		return fmt.Errorf("failed to get Multiply: %v", err)
	}

	return nil
}


// 버섯dna + 먹이dna => new 버섯 dna 생성 ( 유티크.. 14자리.. 마지막 두자리 01 ) + 원장에 저장
func (s *SmartContract) _feedAndMultiply(ctx contractapi.TransactionContextInterface, fungusDna uint, feedDna uint) error {

	// 14자리 + 14자리 = 5자리 이하 => 14자리수 (평균구하는 방식)
	var newDna uint = (fungusDna + feedDna) /2	
	// 14자리수 ( 마지막 두자리 = 01 )
	newDna = newDna - (newDna % 100) + 1

	err := s._createFungus(ctx, "noname" , newDna)
	if err != nil {
		return fmt.Errorf("failed to createFungus: %v", err)
	}
	return nil
}