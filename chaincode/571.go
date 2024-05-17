package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/common"
	mspprotos "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func getAccountStructFromID(
	ctx contractapi.TransactionContextInterface,
	accountID string) (*Account, error) {

	var errMsg string
	account := &Account{}
	accountAsBytes, err := ctx.GetStub().GetState(accountID)
	if err != nil {
		return account, err
	} else if accountAsBytes == nil {
		errMsg = fmt.Sprintf("Error: Account does not exist (%s)", accountID)
		return account, fmt.Errorf(errMsg)
	}
	err = json.Unmarshal([]byte(accountAsBytes), account)
	if err != nil {
		return account, err
	}
	return account, nil
}

func getAccountArrayFromQuery(
	ctx contractapi.TransactionContextInterface,
	queryString string) ([]Account, error) {

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	defer resultsIterator.Close()

	accountArr := []Account{}
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		accountAsBytes := queryResponse.Value
		account := Account{}
		json.Unmarshal(accountAsBytes, &account)
		accountArr = append(accountArr, account)
	}
	return accountArr, nil
}

func getQueueStructFromID(
	ctx contractapi.TransactionContextInterface,
	queueID string) (*QueuedTransaction, error) {

	var errMsg string
	queue := &QueuedTransaction{}
	queueAsBytes, err := ctx.GetStub().GetState(queueID)
	if err != nil {
		return queue, fmt.Errorf(err.Error())
	} else if queueAsBytes == nil {
		errMsg = fmt.Sprintf("Error: QueuedTransaction ID does not exist: %s", queueID)
		return queue, fmt.Errorf(errMsg)
	}
	err = json.Unmarshal([]byte(queueAsBytes), queue)
	if err != nil {
		return queue, fmt.Errorf(err.Error())
	}
	return queue, nil
}

func getQueueArrayFromQuery(
	ctx contractapi.TransactionContextInterface,
	queryString string) ([]QueuedTransaction, error) {

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	defer resultsIterator.Close()

	queueArr := []QueuedTransaction{}
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		jsonByteObj := queryResponse.Value
		queue := QueuedTransaction{}
		json.Unmarshal(jsonByteObj, &queue)
		queueArr = append(queueArr, queue)
	}
	return queueArr, nil
}

func getSortedQueues(
	ctx contractapi.TransactionContextInterface,
	queryString string) ([]QueuedTransaction, error) {
	queryResults, err := getQueueArrayFromQuery(ctx, queryString)
	if err != nil {
		return nil, err
	}
	queryResults = sortQueues(queryResults)
	return queryResults, nil
}

func getCompletedTxArrFromQuery(ctx contractapi.TransactionContextInterface,
	queryString string) ([]CompletedTransaction, error) {

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	defer resultsIterator.Close()

	completedTransactionArr := []CompletedTransaction{}
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		completedTransactionAsBytes := queryResponse.Value
		completedTransaction := CompletedTransaction{}
		json.Unmarshal(completedTransactionAsBytes, &completedTransaction)
		completedTransactionArr = append(completedTransactionArr, completedTransaction)
	}

	return completedTransactionArr, nil
}

func getMoveOutInFundStructFromID(ctx contractapi.TransactionContextInterface,
	moveOutInFundID string) (*MoveOutInFund, error) {

	moveOutInFund := &MoveOutInFund{}
	moveOutInFundAsBytes, err := ctx.GetStub().GetState(moveOutInFundID)
	if err != nil {
		return moveOutInFund, err
	} else if moveOutInFundAsBytes == nil {
		errMsg := fmt.Sprintf("Error: MoveOutInFund (%s) does not exist", moveOutInFundID)
		return moveOutInFund, fmt.Errorf(errMsg)
	}
	err = json.Unmarshal([]byte(moveOutInFundAsBytes), moveOutInFund)
	if err != nil {
		return moveOutInFund, err
	}
	return moveOutInFund, nil
}

func getMoveOutInFundArrayFromQuery(ctx contractapi.TransactionContextInterface,
	queryString string) ([]MoveOutInFund, error) {

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	defer resultsIterator.Close()

	moveOutInFundArr := []MoveOutInFund{}
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		jsonByteObj := queryResponse.Value
		moveOutInFund := MoveOutInFund{}
		json.Unmarshal(jsonByteObj, &moveOutInFund)
		moveOutInFundArr = append(moveOutInFundArr, moveOutInFund)
	}
	return moveOutInFundArr, nil
}

func getPledgeRedeemFundArrFromQuery(ctx contractapi.TransactionContextInterface,
	queryString string) ([]PledgeRedeemFund, error) {

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	defer resultsIterator.Close()

	pledgeRedeemFundArr := []PledgeRedeemFund{}
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		jsonByteObj := queryResponse.Value
		pledgeRedeemFund := PledgeRedeemFund{}
		json.Unmarshal(jsonByteObj, &pledgeRedeemFund)
		pledgeRedeemFundArr = append(pledgeRedeemFundArr, pledgeRedeemFund)
	}
	return pledgeRedeemFundArr, nil
}

func resetAllAccounts(ctx contractapi.TransactionContextInterface) error {

	queryString := fmt.Sprintf(`{
		"selector":{"docType":"%s"}}`,
		accountObjectType)
	accountArr, err := getAccountArrayFromQuery(ctx, queryString)
	if err != nil {
		return err
	}
	for _, account := range accountArr {
		account.Status = "NORMAL"
		account.Amount = 0
		UpdatedAcctAsBytes, err := json.Marshal(account)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(account.AccountID, UpdatedAcctAsBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func resetAllQueues(
	ctx contractapi.TransactionContextInterface) error {

	queryString := fmt.Sprintf(
		`{"selector":{"docType":"%s"}}`,
		queuedTxObjectType)
	queueArr, err := getQueueArrayFromQuery(ctx, queryString)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	for _, queueElement := range queueArr {
		err = ctx.GetStub().DelState(queueElement.RefID)
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}
	return nil
}

func resetAllCompletedTx(ctx contractapi.TransactionContextInterface) error {

	queryString := fmt.Sprintf(
		`{"selector":{"docType":"%s"}}`,
		completedTxObjectType)
	completedTxArr, err := getCompletedTxArrFromQuery(ctx, queryString)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	for _, completedTx := range completedTxArr {
		err = ctx.GetStub().DelState(completedTx.RefID)
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}
	return nil
}

func resetAllPledgeRedeem(ctx contractapi.TransactionContextInterface) error {

	queryString := fmt.Sprintf(
		`{"selector":{
			"$or":[
				{"docType":"%s"},
				{"docType":"%s"},
				{"docType":"%s"},
				{"docType":"%s"}
			]
		}}`,
		pledgeObjectType,
		redeemObjectType,
		nettingSubtractObjectType,
		nettingAddObjectType)

	pledgeRedeemArr, err := getPledgeRedeemFundArrFromQuery(ctx, queryString)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	for _, pledgeRedeem := range pledgeRedeemArr {
		err = ctx.GetStub().DelState(pledgeRedeem.RefID)
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}
	return nil
}

func resetAllMoveOutInFund(ctx contractapi.TransactionContextInterface) error {

	queryString := fmt.Sprintf(
		`{"selector":{
			"$or":[
				{"docType":"%s"},
				{"docType":"%s"}
			]
		}}`,
		moveOutObjectType,
		moveInObjectType)

	moveOutInFundArr, err := getMoveOutInFundArrayFromQuery(ctx, queryString)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	for _, moveOutInFund := range moveOutInFundArr {
		err = ctx.GetStub().DelState(moveOutInFund.RefID)
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}
	return nil
}

func checkArgArrayLength(args []string, expectedArgLength int) error {

	argArrayLength := len(args)
	if argArrayLength != expectedArgLength {
		errMsg := fmt.Sprintf(
			"Incorrect number of arguments: Received %d, expecting %d",
			argArrayLength,
			expectedArgLength)
		return fmt.Errorf(errMsg)
	}
	return nil
}

func getSigner(ctx contractapi.TransactionContextInterface) (string, error) {

	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	id := &mspprotos.SerializedIdentity{}
	err = proto.Unmarshal(creator, id)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	block, _ := pem.Decode(id.GetIdBytes())
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	// mspID := id.GetMspid() // if you need the mspID
	signer := cert.Subject.CommonName
	return signer, nil
}

func verifyIdentity(ctx contractapi.TransactionContextInterface,
	identities ...string) error {

	creatorString, err := getSigner(ctx)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	isVerified := false
	for _, identity := range identities {
		if creatorString == identity {
			isVerified = true
		}
	}

	if !isVerified {
		identitiesString := strings.Join(identities, " or ")
		errMsg := fmt.Sprintf(
			"Error: Identity of creator (%s) does not match %s",
			creatorString,
			identitiesString)
		return fmt.Errorf(errMsg)
	}
	return nil
}

func getChannelName(ctx contractapi.TransactionContextInterface) (string, error) {

	signedProp, _ := ctx.GetStub().GetSignedProposal()

	proposal := &peer.Proposal{}
	err := proto.Unmarshal(signedProp.ProposalBytes, proposal)
	if err != nil {
		return "", err
	}

	header := &common.Header{}
	err = proto.Unmarshal(proposal.Header, header)
	if err != nil {
		return "", err
	}

	chHeader := &common.ChannelHeader{}
	err = proto.Unmarshal(header.ChannelHeader, chHeader)
	if err != nil {
		return "", err
	}

	channelId := chHeader.ChannelId
	return channelId, nil
}

func getTxTimeStampAsTime(ctx contractapi.TransactionContextInterface) (time.Time, error) {

	timestampTime := time.Time{}
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return timestampTime, err
	}
	timestampTime, err = ptypes.Timestamp(timestamp)
	if err != nil {
		return timestampTime, err
	}

	return timestampTime, nil
}

func crossChannelQuery(ctx contractapi.TransactionContextInterface,
	queryArgs [][]byte,
	targetChannel string,
	targetChaincode string) ([]byte, error) {

	response := ctx.GetStub().InvokeChaincode(targetChaincode, queryArgs, targetChannel)

	if response.Status != 200 {
		errStr := fmt.Sprintf(
			"Failed to invoke chaincode. Got error: %s",
			string(response.Payload))
		return nil, fmt.Errorf(errStr)
	}

	responseAsBytes := response.Payload

	return responseAsBytes, nil
}

func getTotalQueuedAmount(queueArr []QueuedTransaction) (float64, error) {

	var totalAmount float64
	totalAmount = 0
	for _, queueElement := range queueArr {
		totalAmount += queueElement.Amount
	}
	return totalAmount, nil
}

func sortQueues(queueArr []QueuedTransaction) []QueuedTransaction {

	priority := func(c1, c2 *QueuedTransaction) bool {
		return c1.Priority > c2.Priority
	}
	createtime := func(c1, c2 *QueuedTransaction) bool {
		return c1.CreateTime.Before(c2.CreateTime)
	}

	OrderedBy(priority, createtime).Sort(queueArr)
	return queueArr
}

func (ms *multiSorter) Sort(changes []QueuedTransaction) {

	ms.Changes = changes
	sort.Sort(ms)
}

func OrderedBy(less ...lessFunc) *multiSorter {
	return &multiSorter{
		less: less,
	}
}

func (ms *multiSorter) Len() int {
	return len(ms.Changes)
}

func (ms *multiSorter) Swap(i, j int) {
	ms.Changes[i], ms.Changes[j] = ms.Changes[j], ms.Changes[i]
}

func (ms *multiSorter) Less(i, j int) bool {
	p, q := &ms.Changes[i], &ms.Changes[j]
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}

	return ms.less[k](p, q)
}
