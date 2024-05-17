package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/drive/base"
	"github.com/drive/conos"
	Crypto "github.com/drive/crypto"
	"github.com/drive/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/shopspring/decimal"
)

// initliaze contract
type SmartContract struct {
	contractapi.Contract
}

const allowancePrefix = "allowance~ccid~user"
const likePrefix = "like~ccid~user~txId"
const dislikePrefix = "dislike~ccid~user~txId"
const downloadCount = "ccid~user~txId"

/**
Create Content
@param string ccid
@param string author
@param string approveTo
@returns
@memeberof Drive
*/
func (s *SmartContract) CreateFile(ctx contractapi.TransactionContextInterface, data string) (interface{}, error) {

	var cd base.Content
	var err error
	// check for file existance
	if err = json.Unmarshal([]byte(data), &cd); err != nil {
		return nil, fmt.Errorf(base.JSONParseError)
	}
	hashSha1 := Crypto.EncodeToSha256(cd.IpfsHash)
	if exists, err := s.FileExists(ctx, hashSha1); err != nil {
		return nil, err
	} else if exists {
		return nil, fmt.Errorf("%s %s", base.FileExistsError, hashSha1)
	}
	txTime, _ := ctx.GetStub().GetTxTimestamp()
	err = ctx.GetStub().PutState(hashSha1, []byte(data))
	if err != nil {
		return nil, fmt.Errorf(base.PutStateError)
	}

	details := &base.DetailsTx{
		From:   cd.Author,
		To:     "Drive",
		Action: "Create",
		Value:  hashSha1,
	}

	dtl, err := json.Marshal(details)
	err = ctx.GetStub().PutState(ctx.GetStub().GetTxID(), dtl)
	if err != nil {
		return nil, fmt.Errorf(base.PutStateError)
	}
	res := &base.Response{
		Success:   true,
		Fcn:       "CreateFile",
		TxID:      ctx.GetStub().GetTxID(),
		Value:     hashSha1,
		Timestamp: txTime,
	}
	content, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf(base.JSONParseError)
	}

	return string(content), nil

}

/**
Approve Content
@param string ccid
@param string author
@param string approveTo
@returns
@memeberof Drive
*/
func (s *SmartContract) Approve(ctx contractapi.TransactionContextInterface, ccidcode, author, spenderAdr string) (interface{}, error) {
	var cd base.Content
	var err error
	ccidByte, err := ctx.GetStub().GetState(ccidcode)
	if err != nil {
		return nil, fmt.Errorf(base.GetstateError)
	}
	if err = json.Unmarshal(ccidByte, &cd); err != nil {
		return nil, fmt.Errorf(base.JSONParseError)
	}

	if cd.Author != author {
		return nil, fmt.Errorf("%s: %s, %s", base.OwnerError, cd.Author, author)
	}
	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{ccidcode, spenderAdr})
	if err != nil {
		return nil, fmt.Errorf(base.KeyCreationError)
	}
	err = ctx.GetStub().PutState(allowanceKey, []byte(spenderAdr))
	if err != nil {
		return nil, fmt.Errorf(base.PutStateError)
	}
	txTime, _ := ctx.GetStub().GetTxTimestamp()
	res := &base.Response{
		Success:   true,
		Fcn:       "Approve",
		TxID:      ctx.GetStub().GetTxID(),
		Timestamp: txTime,
	}
	content, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf(base.JSONParseError)
	}
	details := &base.DetailsTx{
		From:   author,
		To:     "Drive",
		Action: "Approve",
		Value:  fmt.Sprintf("%s to %s ", ccidcode, spenderAdr),
	}

	dtl, err := json.Marshal(details)
	err = ctx.GetStub().PutState(ctx.GetStub().GetTxID(), dtl)
	if err != nil {
		return nil, fmt.Errorf(base.PutStateError)
	}
	return string(content), nil

}

/**
Like Content Counter
@param string ccid
@param string walletid
@param []int args [contentID, userID]
@returns
@memeberof Drive
*/
func (s *SmartContract) Allowance(ctx contractapi.TransactionContextInterface, ccidcode, spender string) (bool, error) {

	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{ccidcode, spender})
	if err != nil {
		return false, fmt.Errorf(base.KeyCreationError)
	}
	allowanceBytes, err := ctx.GetStub().GetState(allowanceKey)
	if err != nil {
		return false, fmt.Errorf(base.GetstateError)
	}

	if allowanceBytes == nil {
		return false, fmt.Errorf(base.EmptyAllowance)
	}

	return true, nil
}

/**
Like Content Counter
@param string ccid
@param string walletid
@returns
@memeberof Drive
*/
func (s *SmartContract) LikeContent(ctx contractapi.TransactionContextInterface, action string) (interface{}, error) {
	var act base.Action
	var err error

	if err = json.Unmarshal([]byte(action), &act); err != nil {
		return nil, fmt.Errorf(base.JSONParseError)
	}

	exists, err := s.FileExists(ctx, act.Ccid)
	if err != nil {
		return nil, fmt.Errorf(base.CheckFileError)
	} else if !exists {
		return nil, fmt.Errorf(base.EmptyFile)
	}
	deltaResultIterator, deltaErr := ctx.GetStub().GetStateByPartialCompositeKey(likePrefix, []string{act.Ccid, act.Wallet})
	if deltaErr != nil {
		return nil, fmt.Errorf("error occured while getting file")
	}

	defer deltaResultIterator.Close()

	if deltaResultIterator.HasNext() {
		return nil, fmt.Errorf("User Already Liked")
	}
	contentLikeKey, err := ctx.GetStub().CreateCompositeKey(likePrefix, []string{act.Ccid, act.Wallet})
	if err != nil {
		return nil, fmt.Errorf(base.KeyCreationError)
	}
	err = ctx.GetStub().PutState(contentLikeKey, []byte{0x00})
	if err != nil {
		return nil, fmt.Errorf(base.PutStateError)
	}
	txTime, _ := ctx.GetStub().GetTxTimestamp()
	res := &base.Response{
		Success:   true,
		Fcn:       "LikeContent",
		TxID:      ctx.GetStub().GetTxID(),
		Timestamp: txTime,
	}
	content, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	details := &base.DetailsTx{
		From:   act.Wallet,
		To:     "Drive",
		Action: "Like",
		Value:  act.Ccid,
	}
	// set event
	likeEvent := &base.Event{UserID: act.UserID, ContentID: act.ContentID, Timestamp: txTime}
	likeEventJSON, err := json.Marshal(likeEvent)
	if err != nil {
		return nil, fmt.Errorf(base.JSONParseError)
	}
	err = ctx.GetStub().SetEvent("UserLikes", likeEventJSON)
	if err != nil {
		return nil, fmt.Errorf(base.EventError)
	}

	dtl, err := json.Marshal(details)
	err = ctx.GetStub().PutState(ctx.GetStub().GetTxID(), dtl)
	if err != nil {
		return nil, fmt.Errorf(base.PutStateError)
	}
	return string(content), nil

}

/**
Download Content Counter
@param string ccid
@param string walletid
@param []int args [contentID, userID]
@returns
@memeberof Drive
*/
func (s *SmartContract) CountDownloads(ctx contractapi.TransactionContextInterface, action string) (interface{}, error) {
	var act base.Action
	var err error

	if err = json.Unmarshal([]byte(action), &act); err != nil {
		return nil, fmt.Errorf(base.JSONParseError)
	}

	// check for file existance
	exists, err := s.FileExists(ctx, act.Ccid)
	if err != nil {
		return nil, fmt.Errorf(base.CheckFileError)
	} else if !exists {
		return nil, fmt.Errorf(base.EmptyFile)
	}
	// get txID
	txID := ctx.GetStub().GetTxID()
	download, err := ctx.GetStub().CreateCompositeKey(downloadCount, []string{act.Ccid, act.Wallet, txID})
	if err != nil {
		return nil, fmt.Errorf(base.KeyCreationError)
	}

	err = ctx.GetStub().PutState(download, []byte{0x00})
	if err != nil {
		return nil, fmt.Errorf(base.PutStateError)
	}
	// set response
	txTime, _ := ctx.GetStub().GetTxTimestamp()
	res := &base.Response{
		Success:   true,
		Fcn:       "CountDownloads",
		TxID:      ctx.GetStub().GetTxID(),
		Timestamp: txTime,
	}
	content, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf(base.JSONParseError)
	}
	details := &base.DetailsTx{
		From:   act.Wallet,
		To:     "Drive",
		Action: "Download",
		Value:  act.Ccid,
	}

	// set event
	downloadEvent := &base.Event{UserID: act.UserID, ContentID: act.ContentID, Timestamp: txTime}
	downloadEventJSON, err := json.Marshal(downloadEvent)
	if err != nil {
		return nil, fmt.Errorf(base.JSONParseError)
	}
	// emit event
	err = ctx.GetStub().SetEvent("UserDownloads", downloadEventJSON)
	if err != nil {
		return nil, fmt.Errorf(base.EventError)
	}
	// marshal json data
	dtl, err := json.Marshal(details)
	err = ctx.GetStub().PutState(ctx.GetStub().GetTxID(), dtl)
	if err != nil {
		return nil, fmt.Errorf(base.PutStateError)
	}
	return string(content), nil
}

/* check file exists
 *this function strictly called inside chaincode
 */
func (s *SmartContract) FileExists(ctx contractapi.TransactionContextInterface, ccid string) (bool, error) {
	cdata, err := ctx.GetStub().GetState(ccid)

	if err != nil {
		return false, fmt.Errorf(base.GetstateError)
	}

	return cdata != nil, nil
}

func (s *SmartContract) GetFile(ctx contractapi.TransactionContextInterface, ccid, spender string) (interface{}, error) {
	var err error
	var cd base.Content
	var exists bool
	if exists, err = s.FileExists(ctx, ccid); err != nil {
		return nil, fmt.Errorf(base.CheckFileError)
	} else if !exists {
		return nil, fmt.Errorf(base.EmptyFile)
	}

	//check allowance
	if ok, _ := s.Allowance(ctx, ccid, spender); ok {
		data, err := ctx.GetStub().GetState(ccid)
		if err != nil {
			return nil, fmt.Errorf(base.GetstateError)
		}
		if data == nil {
			return nil, fmt.Errorf("Ipfs hash is empty")

		}
		if err = json.Unmarshal(data, &cd); err != nil {
			return nil, fmt.Errorf(base.JSONParseError)
		}

		res := &base.Response{
			Success: true,
			Fcn:     "GetFile",
			Value:   cd.IpfsHash,
		}
		content, err := json.Marshal(res)
		if err != nil {
			return nil, err
		}

		return string(content), nil

	}
	return nil, fmt.Errorf("You do not have allowance for this file")
}

func (s *SmartContract) GetTotalLikes(ctx contractapi.TransactionContextInterface, ccid string) (interface{}, error) {

	//get all deltas for the variable

	deltaResultIterator, deltaErr := ctx.GetStub().GetStateByPartialCompositeKey(likePrefix, []string{ccid})
	if deltaErr != nil {
		return nil, fmt.Errorf("error occured while getting file")
	}

	defer deltaResultIterator.Close()

	if !deltaResultIterator.HasNext() {
		return nil, fmt.Errorf("error getting file empty")
	}

	var finalVal int
	var i int

	for i = 0; deltaResultIterator.HasNext(); i++ {
		//get the next row
		_, nextErr := deltaResultIterator.Next()
		if nextErr != nil {
			return nil, fmt.Errorf(nextErr.Error())
		}

		finalVal += 1
	}
	res := &base.Response{
		Success: true,
		Fcn:     "GetTotalLikes",
		Value:   strconv.Itoa(finalVal),
	}
	content, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	return string(content), nil
}

func (s *SmartContract) GetTotalDownloads(ctx contractapi.TransactionContextInterface, ccid string) (interface{}, error) {

	//get all deltas for the variable

	deltaResultIterator, deltaErr := ctx.GetStub().GetStateByPartialCompositeKey(downloadCount, []string{ccid})
	if deltaErr != nil {
		return nil, fmt.Errorf("error occured while getting file")
	}

	defer deltaResultIterator.Close()

	if !deltaResultIterator.HasNext() {
		return nil, fmt.Errorf("error getting file empty")
	}

	var finalVal int
	var i int

	for i = 0; deltaResultIterator.HasNext(); i++ {
		//get the next row
		_, nextErr := deltaResultIterator.Next()
		if nextErr != nil {
			return nil, fmt.Errorf(nextErr.Error())
		}

		finalVal += 1
	}

	res := &base.Response{
		Success: true,
		Fcn:     "GetTotalDownloads",
		Value:   strconv.Itoa(finalVal),
	}
	content, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	return string(content), nil
}

/**
Purchase Content
*/
func (s *SmartContract) PuchaseContent(ctx contractapi.TransactionContextInterface, ccid, wallet, price string) (interface{}, error) {
	var err error
	var cd base.Content
	var p, a decimal.Decimal
	cdByte, err := ctx.GetStub().GetState(ccid)
	if err != nil {
		return nil, fmt.Errorf(base.GetstateError)
	}
	if err = json.Unmarshal(cdByte, &cd); err != nil {
		return nil, fmt.Errorf(base.JSONParseError)
	}
	p, err = util.ParsePositive(price)
	if err != nil {
		return nil, err
	}
	if a, err = util.ParsePositive(cd.Price); err != nil {
		return nil, err
	}
	if p.Cmp(a) < 0 {
		return nil, fmt.Errorf(base.WrongAmount)
	}
	ok, err := conos.SendToken(ctx, wallet, cd.Author, cd.Price)
	if ok == false {
		return nil, fmt.Errorf("Purchase failed %s", err)
	}

	return nil, err

}
