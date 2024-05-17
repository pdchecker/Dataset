/*
 * CdlChainCode.go
 *
 * CDLイベント情報を管理するチェーンコード
 *
 * COPYRIGHT 2021 Fujitsu Limited
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

//--------------------------------
// 定数定義
//--------------------------------

// CdlChainCode チェーンコード本体の構造体定義
type CdlChainCode struct {
	contractapi.Contract
}

//------------------------------------
// チェーンコード本体の実装
//------------------------------------

// Init チェーンコード配備時に実行される初期化関数
func (cc *CdlChainCode) Init(ctx contractapi.TransactionContextInterface) {
	log.Print("init: called...")

	// 初期化処理無し

	log.Print("init: done.")
}

//---------------------------------
// invoke系の関数
//---------------------------------

// regist CDL Event to the Block-Chain
//
// @param ctx the transaction context
// @param key the key for the CDL Event
// @param jsonString the make of the new event
func (cc *CdlChainCode) RegistCDLEvent(ctx contractapi.TransactionContextInterface, key string, jsonString string) (error) {
	myFunc := "RegistCDLEvent"
	stub := ctx.GetStub()

	log.Print("cdl-chaincode : RegistCDLEvent() called key=" + key)

	// イベントが未登録であることをチェック
	state, err := stub.GetState(key)
	if err != nil {
		msg := fmt.Sprintf("["+myFunc+"] "+
			"GetState(key) %v, Error: "+ err.Error(), key)
		log.Print(msg)
		return fmt.Errorf(msg)
	}
	if state != nil {
		msg := fmt.Sprintf("cdleventid '%s' already exists", key)
		log.Print(msg)
		return fmt.Errorf(msg)
	}
	
	// WorldStateにイベント情報を登録
	err = stub.PutState(key, []byte(jsonString))
	if err != nil {
		msg := "[" + myFunc + "] " +
			"PutState(key) fail, Error: " + err.Error()
		log.Print(msg)
		return fmt.Errorf(msg)
	}

    log.Print("cdl-chaincode : RegistCDLEvent() end key=" + key)

	// 異常が無ければ正常復帰する
	return nil
}

// regist and update CDL Event to the Block-Chain
//
// @param ctx the transaction context
// @param key the key for the CDL Event
// @param jsonString CDL Event
// @param updates events to update
func (cc *CdlChainCode) RegistUpdateCDLEvent(ctx contractapi.TransactionContextInterface, key string, jsonString string, updates string) (error) {
	myFunc := "RegistUpdateCDLEvent"
	stub := ctx.GetStub()

	log.Print("cdl-chaincode : RegistUpdateCDLEvent() called key=" + key)

	// イベントが未登録であることをチェック
	state, err := stub.GetState(key)
	if err != nil {
		msg := fmt.Sprintf("["+myFunc+"] "+
			"GetState(key) %v, Error: "+ err.Error(), key)
		log.Print(msg)
		return fmt.Errorf(msg)
	}
	if state != nil {
		msg := fmt.Sprintf("cdleventid '%s' already exists", key)
		log.Print(msg)
		return fmt.Errorf(msg)
	}

	// イベント登録
	err = stub.PutState(key, []byte(jsonString))
	if err != nil {
		msg := "[" + myFunc + "] " +
			"PutState(key) fail, Error: " + err.Error()
		log.Print(msg)
		return fmt.Errorf(msg)
	}

	updateMap := make(map[string]string)
	err = json.Unmarshal([]byte(updates), &updateMap)
	if err != nil {
		return fmt.Errorf("[" + myFunc + "] json.Unmarshal(updates) Error: " + err.Error())
	}
	for k, v := range updateMap {
		// 更新する前イベントが登録済みかをチェック
		state, err = stub.GetState(k)
		if err != nil {
			msg := fmt.Sprintf("["+myFunc+"] "+
				"GetState(key) %v, Error: "+ err.Error(), k)
			log.Print(msg)
			return fmt.Errorf(msg)
		}
		if state == nil {
			msg := fmt.Sprintf("cdleventid '%s' not found", k)
			log.Print(msg)
			return fmt.Errorf(msg)
		}

		// 前イベント更新
		err = stub.PutState(k, []byte(v))
		if err != nil {
			msg := "[" + myFunc + "] " +
				"PutState(key) fail, Error: " + err.Error()
			log.Print(msg)
			return fmt.Errorf(msg)
		}
	}

	log.Print("cdl-chaincode : RegistUpdateCDLEvent() end key=" + key)

	// 異常が無ければ正常復帰する
	return nil
}

//---------------------------------
// query系関数
//---------------------------------

// Check if the key exists on the Block-Chain
//
// @param ctx the transaction context
// @param key the key
// @return bool
func (cc *CdlChainCode) KeyExists(ctx contractapi.TransactionContextInterface, key string) (bool, error) {
	myFunc := "KeyExists"
	stub := ctx.GetStub()

	log.Print("cdl-chaincode : KeyExists() called key=" + key)

	// イベントの登録有無をチェック
	result := false
	state, err := stub.GetState(key)
	if err != nil {
		msg := fmt.Sprintf("["+myFunc+"] "+
			"GetState(key) %v, Error: "+ err.Error(), key)
		log.Print(msg)
		return result, fmt.Errorf(msg)
	}

	if state != nil {
		result = true
	}

    log.Print("cdl-chaincode : KeyExists() end key=" + key)

	// 異常が無ければ正常復帰する
	return result, nil
}

// query CDL Event from the Block-Chain
//
// @param ctx the transaction context
// @param key the key
// @return CDL Event (Json String)
func (cc *CdlChainCode) QueryCDLEvent(ctx contractapi.TransactionContextInterface, key string) (string, error) {
	myFunc := "QueryCDLEvent"
	stub := ctx.GetStub()

	log.Print("cdl-chaincode : QueryCDLEvent() called key=" + key)

	// イベントが未登録であることをチェック
	state, err := stub.GetState(key)
	if err != nil {
		msg := fmt.Sprintf("["+myFunc+"] "+
			"GetState(key) %v, Error: "+ err.Error(), key)
		log.Print(msg)
		return "", fmt.Errorf(msg)
	}
	if state == nil {
		msg := fmt.Sprintf("cdleventid '%s' does not exist", key)
		log.Print(msg)
		return "", fmt.Errorf(msg)
	}

    log.Print("cdl-chaincode : QueryCDLEvent() end key=" + key)

	// 異常が無ければ正常復帰する
	return string(state), nil
}

// rich query CDL Event from the Block-Chain
//
// @param ctx the transaction context
// @param query the query
// @return CDL Event (Json String)
func (cc *CdlChainCode) QueryCDLEventByRichQuery(ctx contractapi.TransactionContextInterface, query string) (string, error) {
	myFunc := "QueryCDLEventByRichQuery"
	stub := ctx.GetStub()

	log.Print("cdl-chaincode : QueryCDLEventByRichQuery() called query=" + query)

	// リッチクエリ
	resultsIterator, err := stub.GetQueryResult(query)
	if err != nil {
		msg := fmt.Sprintf("["+myFunc+"] "+
			"GetQueryResult(query) %v, Error: "+ err.Error(), query)
		log.Print(msg)
		return "", fmt.Errorf(msg)
	}
	defer resultsIterator.Close()

	// 応答レスポンスを構築
	var responseBuf bytes.Buffer
	responseBuf.WriteString("[\n")
	
	first := true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			msg := fmt.Sprintf("["+myFunc+"] "+
			"Next() %v, Error: "+ err.Error(), queryResponse.Key)
			log.Print(msg)
			return "", fmt.Errorf(msg)
		}
		if first == false {
			responseBuf.WriteString(",\n")
		} else {
			first = false
		}
		responseBuf.WriteString(string(queryResponse.Value))
	}
	responseBuf.WriteString("\n]")

	log.Print("cdl-chaincode : QueryCDLEventByRichQuery() end query=" + query)

	// 異常が無ければ正常復帰する
	return responseBuf.String(), nil
}

// query CDL Events by eventid's array from the Block-Chain
//
// @param ctx the transaction context
// @param eventidarray a string representing an array of eventids
// @return CDL Events (Json String)
func (cc *CdlChainCode) QueryCDLEventByArray(ctx contractapi.TransactionContextInterface, eventidarray string) (string, error) {
	myFunc := "QueryCDLEventByArray"
	stub := ctx.GetStub()

	log.Print("cdl-chaincode : QueryCDLEventByArray() called eventidarray=" + eventidarray)

	// イベントID配列の解析
	var eventidarrayStr []string
	err := json.Unmarshal([]byte(eventidarray), &eventidarrayStr)
	if err != nil {
	    msg := "[" + myFunc + "] json.Unmarshal(eventidarray) Error: " + err.Error()
	    log.Print(msg)
		return "", fmt.Errorf(msg)
	}
	// イベントID配列が空配列の場合、エラー
	if len(eventidarrayStr) == 0 {
		msg := "[" + myFunc + "] The specified eventid's size is zero."
		log.Print(msg)
		return "", fmt.Errorf(msg)
	}

	// 応答レスポンスを構築
    var responseBuf bytes.Buffer
    responseBuf.WriteString("[\n")

	first := true
	for _, key := range eventidarrayStr {
        // イベントが未登録である場合はエラー
        state, err := stub.GetState(key)
        if err != nil {
        	msg := fmt.Sprintf("["+myFunc+"] "+
        		"GetState(key) %v, Error: "+ err.Error(), key)
        	log.Print(msg)
        	return "", fmt.Errorf(msg)
        }
        if state == nil {
        	msg := fmt.Sprintf("cdleventid '%s' does not exist", key)
        	log.Print(msg)
        	return "", fmt.Errorf(msg)
        }

        if first == false {
        	responseBuf.WriteString(",\n")
        } else {
        	first = false
        }
        responseBuf.WriteString(string(state))
    }
    responseBuf.WriteString("\n]")

    log.Print("cdl-chaincode : QueryCDLEventByArray() end eventidarray=" + eventidarray)

	// 異常が無ければ正常復帰する
	return responseBuf.String(), nil
}

// query CDL Event from the Block-Chain by range
//
// (現状、CDLから呼び出すメソッドではなく、
//  直接 peer chaincode コマンドで呼び出し、ブロックチェーンの中身を確認するためのデバッグ用メソッド)
//
// @param ctx the transaction context
// @param search start key
// @param search end key
// @return CDL Event (Json Strings)
func (cc *CdlChainCode) QueryCDLEventByRange(ctx contractapi.TransactionContextInterface, startKey string, endKey string) (string, error) {
	myFunc := "QueryCDLEventByRange"
	stub := ctx.GetStub()

	log.Print("cdl-chaincode : QueryCDLEventByRange() called")

	// リッチクエリ
	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		msg := fmt.Sprintf("["+myFunc+"] "+
			"GetStateByRange(startKey, endKey) %v %v, Error: "+ err.Error(), startKey, endKey)
		log.Print(msg)
		return "", fmt.Errorf(msg)
	}
	defer resultsIterator.Close()

	// 応答レスポンスを構築
	var responseBuf bytes.Buffer
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			msg := fmt.Sprintf("["+myFunc+"] "+
			"Next() %v, Error: "+ err.Error(), queryResponse.Key)
			log.Print(msg)
			return "", fmt.Errorf(msg)
		}
		responseBuf.WriteString("'")
		responseBuf.WriteString(queryResponse.Key)
		responseBuf.WriteString("' = ")
		responseBuf.WriteString(string(queryResponse.Value))
		responseBuf.WriteString("\n")
	}

	log.Print("cdl-chaincode : QueryCDLEventByRange() end")

	// 異常が無ければ正常復帰する
	return responseBuf.String(), nil
}

// main関数
//
// チェーンコード起動時のエントリポイント
func main() {
	smartContract, err := contractapi.NewChaincode(new(CdlChainCode))

	if err != nil {
		fmt.Printf("Error creating chaincode: %s", err.Error())
		return
	}

	if err := smartContract.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %s", err.Error())
		return
	}
}