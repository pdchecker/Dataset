package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"topic/chaincode"
	"topic/chaincode/mocks"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	// _ "github.com/maxbrunsfeld/counterfeiter/v6"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/stretchr/testify/require"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o mocks/transaction.go -fake-name TransactionContext . transactionContext
type transactionContext interface {
	contractapi.TransactionContextInterface
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o mocks/chaincodestub.go -fake-name ChaincodeStub . chaincodeStub
type chaincodeStub interface {
	shim.ChaincodeStubInterface
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o mocks/statequeryiterator.go -fake-name StateQueryIterator . stateQueryIterator
type stateQueryIterator interface {
	shim.StateQueryIteratorInterface
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o mocks/clientIdentity.go -fake-name ClientIdentity . clientIdentity
type clientIdentity interface {
	cid.ClientIdentity
}

const myOrg1Msp = "Org1Testmsp"
const myOrg1Clientid = "myOrg1Userid"
const myOrg1PrivCollection = "Org1TestmspPrivateCollection"
const myOrg2Msp = "Org2Testmsp"
const myOrg2Clientid = "myOrg2Userid"
const myOrg2PrivCollection = "Org2TestmspPrivateCollection"

var sampleTopic = &chaincode.Topic{
	Hash:      "1",
	Title:     "1",
	Creator:   "1",
	CID:       "1",
	Category:  "1",
	Tags:      []string{"1"},
	Images:    []string{"1"},
	Upvotes:   []string{"1"},
	Downvotes: []string{"1"},
}

var sampleInput, _ = json.Marshal(sampleTopic)

func TestCreateTopic(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	err := topic.CreateTopic(transactionContext, string(sampleInput))
	require.NoError(t, err)

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = topic.CreateTopic(transactionContext, string(sampleInput))
	require.EqualError(t, err, "failed to read from world state: failure")

	expectedTopic := &chaincode.Topic{Hash: "1"}
	bytes, err := json.Marshal(expectedTopic)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	err = topic.CreateTopic(transactionContext, string(sampleInput))
	require.EqualError(t, err, "the topic 1 already exists")

	chaincodeStub.GetStateReturns(nil, nil)
	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = topic.CreateTopic(transactionContext, string(sampleInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

func TestDeleteTopic(t *testing.T) {
	deleteRequest := &chaincode.Delete{Hash: "1", Creator: "1"}
	deleteInput, _ := json.Marshal(deleteRequest)

	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err := topic.DeleteTopic(transactionContext, string(deleteInput))
	require.EqualError(t, err, "failed to read from world state: failure")

	expectedTopic := &chaincode.Topic{Hash: "1", Creator: "1"}
	bytes, err := json.Marshal(expectedTopic)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(bytes, nil)
	err = topic.DeleteTopic(transactionContext, string(deleteInput))
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(nil, nil)
	err = topic.DeleteTopic(transactionContext, string(deleteInput))
	require.EqualError(t, err, "the topic 1 does not exist")

	expectedTopic = &chaincode.Topic{Hash: "1", Creator: "2"}
	bytes, err = json.Marshal(expectedTopic)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(bytes, nil)
	err = topic.DeleteTopic(transactionContext, string(deleteInput))
	require.EqualError(t, err, "the topic 1 is not created by 1")

	expectedTopic = &chaincode.Topic{Hash: "1", Creator: "1"}
	bytes, err = json.Marshal(expectedTopic)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(bytes, nil)
	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = topic.DeleteTopic(transactionContext, string(deleteInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")

	err = topic.DeleteTopic(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")
}

func TestReadTopic(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	tmpTopic := &chaincode.Topic{Hash: "1"}
	bytes, _ := json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)
	_, err := topic.ReadTopic(transactionContext, "1")
	require.NoError(t, err)

	err = topic.CreateTopic(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	_, err = topic.ReadTopic(transactionContext, "1")
	require.EqualError(t, err, "failed to read from world state: failure")

	chaincodeStub.GetStateReturns(nil, nil)
	_, err = topic.ReadTopic(transactionContext, "1")
	require.EqualError(t, err, "the topic 1 does not exist")
}

func TestUpdateTopic(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	err := topic.UpdateTopic(transactionContext, string(sampleInput))
	require.EqualError(t, err, "the topic 1 does not exist")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = topic.UpdateTopic(transactionContext, string(sampleInput))
	require.EqualError(t, err, "failed to read from world state: failure")

	tmpTopic := &chaincode.Topic{Hash: "1", Creator: "1"}
	bytes, _ := json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = topic.UpdateTopic(transactionContext, string(sampleInput))
	require.NoError(t, err)

	err = topic.UpdateTopic(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	tmpTopic = &chaincode.Topic{Hash: "1", Creator: myOrg2Clientid}
	bytes, _ = json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)

	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = topic.UpdateTopic(transactionContext, string(sampleInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

var upvoteInput, _ = json.Marshal(&chaincode.Upvote{Hash: "1", Creator: "1"})

func TestUpvoteTopic(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	err := topic.UpvoteTopic(transactionContext, string(upvoteInput))
	require.EqualError(t, err, "the topic 1 does not exist")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = topic.UpvoteTopic(transactionContext, string(upvoteInput))
	require.EqualError(t, err, "failed to read from world state: failure")

	tmpTopic := &chaincode.Topic{Hash: "1", Creator: "1"}
	bytes, _ := json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = topic.UpvoteTopic(transactionContext, string(upvoteInput))
	require.NoError(t, err)

	tmpTopic = &chaincode.Topic{Hash: "1", Creator: "1", Upvotes: []string{"1"}}
	bytes, _ = json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)
	err = topic.UpvoteTopic(transactionContext, string(upvoteInput))
	require.NoError(t, err)

	tmpTopic = &chaincode.Topic{Hash: "1", Creator: "1", Downvotes: []string{"1"}}
	bytes, _ = json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)
	err = topic.UpvoteTopic(transactionContext, string(upvoteInput))
	require.NoError(t, err)

	err = topic.UpvoteTopic(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	tmpTopic = &chaincode.Topic{Hash: "1", Creator: myOrg2Clientid}
	bytes, _ = json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)

	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = topic.UpvoteTopic(transactionContext, string(upvoteInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

func TestDownvoteTopic(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	err := topic.DownvoteTopic(transactionContext, string(upvoteInput))
	require.EqualError(t, err, "the topic 1 does not exist")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = topic.DownvoteTopic(transactionContext, string(upvoteInput))
	require.EqualError(t, err, "failed to read from world state: failure")

	tmpTopic := &chaincode.Topic{Hash: "1", Creator: "1"}
	bytes, _ := json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = topic.DownvoteTopic(transactionContext, string(upvoteInput))
	require.NoError(t, err)

	tmpTopic = &chaincode.Topic{Hash: "1", Creator: "1", Downvotes: []string{"1"}}
	bytes, _ = json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)
	err = topic.DownvoteTopic(transactionContext, string(upvoteInput))
	require.NoError(t, err)

	tmpTopic = &chaincode.Topic{Hash: "1", Creator: "1", Upvotes: []string{"1"}}
	bytes, _ = json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)
	err = topic.DownvoteTopic(transactionContext, string(upvoteInput))
	require.NoError(t, err)

	err = topic.DownvoteTopic(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	tmpTopic = &chaincode.Topic{Hash: "1", Creator: myOrg2Clientid}
	bytes, _ = json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)

	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = topic.DownvoteTopic(transactionContext, string(upvoteInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

var emojiInput, _ = json.Marshal(&chaincode.Emoji{Hash: "1", Creator: "1", Code: "1"})

func TestAddEmojiTopic(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	err := topic.AddEmojiTopic(transactionContext, string(emojiInput))
	require.EqualError(t, err, "the topic 1 does not exist")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = topic.AddEmojiTopic(transactionContext, string(emojiInput))
	require.EqualError(t, err, "failed to read from world state: failure")

	tmpTopic := &chaincode.Topic{Hash: "1", Creator: "1"}
	bytes, _ := json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = topic.AddEmojiTopic(transactionContext, string(emojiInput))
	require.NoError(t, err)

	err = topic.AddEmojiTopic(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	tmpTopic = &chaincode.Topic{Hash: "1", Creator: myOrg2Clientid}
	bytes, _ = json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)

	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = topic.AddEmojiTopic(transactionContext, string(emojiInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

func TestRemoveEmojiTopic(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	err := topic.RemoveEmojiTopic(transactionContext, string(emojiInput))
	require.EqualError(t, err, "the topic 1 does not exist")

	chaincodeStub.GetStateReturns([]byte{}, fmt.Errorf("failure"))
	err = topic.RemoveEmojiTopic(transactionContext, string(emojiInput))
	require.EqualError(t, err, "failed to read from world state: failure")

	tmpTopic := &chaincode.Topic{Hash: "1", Creator: "1"}
	bytes, _ := json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)

	err = topic.RemoveEmojiTopic(transactionContext, string(emojiInput))
	require.NoError(t, err)

	err = topic.RemoveEmojiTopic(transactionContext, "sad")
	require.EqualError(t, err, "invalid character 's' looking for beginning of value")

	tmpTopic = &chaincode.Topic{Hash: "1", Creator: myOrg2Clientid}
	bytes, _ = json.Marshal(tmpTopic)
	chaincodeStub.GetStateReturns(bytes, nil)

	chaincodeStub.PutStateReturns(fmt.Errorf("failed inserting key"))
	err = topic.RemoveEmojiTopic(transactionContext, string(emojiInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")

	testRemoveEmojiTopic := &chaincode.Topic{Hash: "1", Creator: "1", Emojis: make(map[string][]string)}
	testRemoveEmojiTopic.Emojis["1"] = []string{"1"}
	bytes, _ = json.Marshal(testRemoveEmojiTopic)
	chaincodeStub.GetStateReturns(bytes, nil)
	err = topic.RemoveEmojiTopic(transactionContext, string(emojiInput))
	require.EqualError(t, err, "failed to put to world state: failed inserting key")
}

func TestGetAllTopics(t *testing.T) {
	asset := &chaincode.Topic{Hash: "user1"}
	bytes, err := json.Marshal(asset)
	require.NoError(t, err)

	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturns(&queryresult.KV{Value: bytes}, nil)

	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	chaincodeStub.GetStateByRangeReturns(iterator, nil)
	userprofile := &chaincode.SmartContract{}
	assets, err := userprofile.GetAllTopics(transactionContext)
	require.NoError(t, err)
	require.Equal(t, []*chaincode.Topic{asset}, assets)

	iterator.HasNextReturns(true)
	iterator.NextReturns(nil, fmt.Errorf("failed retrieving next item"))
	assets, err = userprofile.GetAllTopics(transactionContext)
	require.EqualError(t, err, "failed retrieving next item")
	require.Nil(t, assets)

	chaincodeStub.GetStateByRangeReturns(nil, fmt.Errorf("failed retrieving all assets"))
	assets, err = userprofile.GetAllTopics(transactionContext)
	require.EqualError(t, err, "failed retrieving all assets")
	require.Nil(t, assets)
}

func TestQueryTopicsByTitle(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	chaincodeStub.GetQueryResultReturns(nil, fmt.Errorf("failure"))
	_, err := topic.QueryTopicsByTitle(transactionContext, "1")
	require.EqualError(t, err, "failure")
}
func TestQueryTopicsByCreator(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	chaincodeStub.GetQueryResultReturns(nil, fmt.Errorf("failure"))
	_, err := topic.QueryTopicsByCreator(transactionContext, "1")
	require.EqualError(t, err, "failure")

	tmpTopic := &chaincode.Topic{Hash: "user1", Creator: myOrg1Clientid}
	bytes, _ := json.Marshal(tmpTopic)

	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturns(&queryresult.KV{Value: bytes}, nil)

	chaincodeStub.GetQueryResultReturns(iterator, nil)
	topics, err := topic.QueryTopicsByCreator(transactionContext, "1")
	require.NoError(t, err)
	require.Equal(t, []*chaincode.Topic{tmpTopic}, topics)

	iterator.NextReturns(nil, fmt.Errorf("failed retrieving next item"))
	iterator.HasNextReturnsOnCall(2, true)
	iterator.HasNextReturnsOnCall(3, true)
	chaincodeStub.GetQueryResultReturns(iterator, nil)
	topics, err = topic.QueryTopicsByCreator(transactionContext, "1")
	require.EqualError(t, err, "failed retrieving next item")
	require.Nil(t, topics)
}

func TestQueryTopicsByCategory(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	chaincodeStub.GetQueryResultReturns(nil, fmt.Errorf("failure"))
	_, err := topic.QueryTopicsByCategory(transactionContext, 1)
	require.EqualError(t, err, "failure")
}

func TestQueryTopicsByTag(t *testing.T) {
	transactionContext, chaincodeStub := prepMocksAsOrg1()
	topic := chaincode.SmartContract{}

	chaincodeStub.GetQueryResultReturns(nil, fmt.Errorf("failure"))
	_, err := topic.QueryTopicsByTag(transactionContext, "1")
	require.EqualError(t, err, "failure")
}

func prepMocksAsOrg1() (*mocks.TransactionContext, *mocks.ChaincodeStub) {
	return prepMocks(myOrg1Msp, myOrg1Clientid)
}
func prepMocksAsOrg2() (*mocks.TransactionContext, *mocks.ChaincodeStub) {
	return prepMocks(myOrg2Msp, myOrg2Clientid)
}
func prepMocks(orgMSP, clientId string) (*mocks.TransactionContext, *mocks.ChaincodeStub) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	clientIdentity := &mocks.ClientIdentity{}
	clientIdentity.GetMSPIDReturns(orgMSP, nil)
	clientIdentity.GetIDReturns(base64.StdEncoding.EncodeToString([]byte(clientId)), nil)
	// set matching msp ID using peer shim env variable
	os.Setenv("CORE_PEER_LOCALMSPID", orgMSP)
	transactionContext.GetClientIdentityReturns(clientIdentity)
	return transactionContext, chaincodeStub
}

func prepMocksIllegalId() (*mocks.TransactionContext, *mocks.ChaincodeStub) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	clientIdentity := &mocks.ClientIdentity{}
	clientIdentity.GetMSPIDReturns(myOrg1Msp, nil)
	// clientIdentity.GetIDReturns("illegal", nil)
	clientIdentity.GetIDReturns("", fmt.Errorf("failure"))
	// set matching msp ID using peer shim env variable
	os.Setenv("CORE_PEER_LOCALMSPID", myOrg1Msp)
	transactionContext.GetClientIdentityReturns(clientIdentity)
	return transactionContext, chaincodeStub
}
