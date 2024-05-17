package main

import (
	"encoding/json"
	"fmt"

	//  "byte"

	"github.com/google/uuid"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a car
type SmartContract struct {
	contractapi.Contract
}
type QueryBuy struct {
	Key    string `json:"Key"`
	Record *Buy
}
type Buy struct {
	DSGId          string `json:"id"`
	OrderId        string `json:"orderId"`
	Amount         string `json:"amount"`
	AmountWithFees string `json:"amountWithFees"`
	TotalKgs       string `json:"totalKgs"`
	DwrReceiptId   string `json:"dwrReceiptId"`
	UserId         string `json:"userId"`
	AccountNo      string `json:"accountNo"`
}

func GetUId() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return id.String(), err
}

// InitLedger adds a base set of cars to the ledger
func (s *SmartContract) Init(ctx contractapi.TransactionContextInterface) error {
	fmt.Printf("Hello\n")
	return nil
}
func (s *SmartContract) CreateBuy(ctx contractapi.TransactionContextInterface, OrderId string, Amount string, AmountWithFees string, TotalKgs string, DwrReceiptId string, AccountNo string, UserId string) error {

	fmt.Printf("Adding Buy to the ledger ...\n")
	// if len(args) != 8 {
	// 	return fmt.Errorf("InvalidArgumentError: Incorrect number of arguments. Expecting 8")
	// }

	//Prepare key for the new Org
	uid, err := GetUId()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	id := "DSG-" + uid
	fmt.Printf("Validating Buy data\n")
	//Validate the Org data
	var buy = Buy{DSGId: id,
		OrderId:        OrderId,
		Amount:         Amount,
		AmountWithFees: AmountWithFees,
		TotalKgs:       TotalKgs,
		DwrReceiptId:   DwrReceiptId,
		AccountNo:      AccountNo,
		UserId:         UserId,
	}

	//Encrypt and Marshal Org data in order to put in world state
	buyAsBytes, _ := json.Marshal(buy)

	return ctx.GetStub().PutState(id, buyAsBytes)

}
func (s *SmartContract) QueryBuy(ctx contractapi.TransactionContextInterface, OrderId string) (*Buy, error) {
	buyAsBytes, err := ctx.GetStub().GetState(OrderId)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if buyAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", OrderId)
	}

	buy := new(Buy)
	_ = json.Unmarshal(buyAsBytes, buy)

	return buy, nil
}
func (s *SmartContract) GetBuy(ctx contractapi.TransactionContextInterface, OrderId string) ([]QueryBuy, error) {
	query := "{\"selector\": {\"_id\": {\"$regex\": \"^DSG-\"} } }"
	resultsIterator, err := ctx.GetStub().GetQueryResult(query)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	result := []QueryBuy{}
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		buy := new(Buy)
		_ = json.Unmarshal(queryResponse.Value, buy)
		if buy.OrderId == OrderId {

			queryResult := QueryBuy{Key: queryResponse.Key, Record: buy}
			result = append(result, queryResult)
		}
	}
	return result, nil
}
func (s *SmartContract) GetBuyList(ctx contractapi.TransactionContextInterface, OrderId string) ([]QueryBuy, error) {
	query := "{\"selector\": {\"_id\": {\"$regex\": \"^DSG-\"} } }"
	resultsIterator, err := ctx.GetStub().GetQueryResult(query)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	results := []QueryBuy{}
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		buy := new(Buy)
		_ = json.Unmarshal(queryResponse.Value, buy)
		if buy.OrderId == OrderId {
			queryResult := QueryBuy{Key: queryResponse.Key, Record: buy}
			results = append(results, queryResult)
		}
	}
	return results, nil
}
func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create  chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting  chaincode: %s", err.Error())
	}
}
