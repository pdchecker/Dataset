package main

import (
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const accountObjectType string = "account"
const queuedTxObjectType string = "queue"
const completedTxObjectType string = "completedtx"
const pledgeObjectType string = "pledgefund"
const redeemObjectType string = "redeemfund"
const nettingAddObjectType string = "nettingadd"
const nettingSubtractObjectType string = "nettingsubtract"
const moveOutObjectType string = "moveoutfund"
const moveInObjectType string = "moveinfund"

type multiSorter struct {
	Changes []QueuedTransaction
	less    []lessFunc
}

type Account struct {
	ObjectType string  `json:"docType"`
	AccountID  string  `json:"accountID"`
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
	Status     string  `json:"status"`
}

type QueuedTransaction struct {
	ObjectType string    `json:"docType"`
	RefID      string    `json:"refID"`
	Sender     string    `json:"sender"`
	Receiver   string    `json:"receiver"`
	Priority   int       `json:"priority"`
	Nettable   bool      `json:"nettable"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
	IsFrozen   bool      `json:"isFrozen"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

type CompletedTransaction struct {
	ObjectType string    `json:"docType"`
	RefID      string    `json:"refID"`
	Sender     string    `json:"sender"`
	Receiver   string    `json:"receiver"`
	Priority   int       `json:"priority"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

type PledgeRedeemFund struct {
	ObjectType string    `json:"docType"`
	RefID      string    `json:"refID"`
	AccountID  string    `json:"accountID"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	CreateTime time.Time `json:"createTime"`
}

type MoveOutInFund struct {
	ObjectType  string    `json:"docType"`
	RefID       string    `json:"refID"`
	AccountID   string    `json:"accountID"`
	ChannelFrom string    `json:"channelFrom"`
	ChannelTo   string    `json:"channelTo"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	CreateTime  time.Time `json:"createTime"`
}

type TransactionHistory struct {
	CompletedTx  []CompletedTransaction `json:"completedtx"`
	PledgeRedeem []PledgeRedeemFund     `json:"pledgeredeem"`
	MoveInOut    []MoveOutInFund        `json:"moveinout"`
}

type SettleMLNettingResp struct {
	NettableTxList []string `json:"nettableTxList"`
	WentToLoop     bool     `json:"wentToLoop"`
}

type PingChaincode struct {
	ObjectType string `json:"docType"`
	Number     int    `json:"number"`
}
type SimpleChaincode struct {
	contractapi.Contract
}
