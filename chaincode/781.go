package main

import (
  "encoding/json"
  "fmt"
  "log"

  "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Transaction
type SmartContract struct {
  contractapi.Contract
}

// Transaction describes basic details of what makes up a simple transaction
type Transaction struct {
  TranId             string `json:"TranId"`
  ApprovalCode          string `json:"ApprovalCode"`
  BillDetails           string    `json:"BillDetails"`
  CustomerDetails          string `json:"CustomerDetails"`
  Status string    `json:"Status"`
}

// InitLedger adds a base set of transaction to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
  trans := []Transaction{
    {TranId: "transaction-0", ApprovalCode: "22310", BillDetails: "store:123,items:fooditems,datetime:06/06/2023", CustomerDetails: "Cname:Sam,Cnumber:,CardNumb:3345-1233", Status: "Successful"},
    {TranId: "transaction-1", ApprovalCode: "22311", BillDetails: "store:122,items:clothing,datetime:06/05/2023", CustomerDetails: "Cname:Ram,Cnumber:,CardNumb:3345-1235", Status: "failure"},
    {TranId: "transaction-2", ApprovalCode: "22312", BillDetails: "store:121,items:fooditems,datetime:06/03/2023", CustomerDetails: "Cname:Tom,Cnumber:,CardNumb:3345-1232", Status: "Successful"},
    {TranId: "transaction-3", ApprovalCode: "22313", BillDetails: "store:123,items:clothing,datetime:06/01/2023", CustomerDetails: "Cname:Jerry,Cnumber:,CardNumb:3345-1231", Status: "disputed"},
    {TranId: "transaction-4", ApprovalCode: "22314", BillDetails: "store:124,items:jellwery,datetime:06/07/2023", CustomerDetails: "Cname:Riki,Cnumber:,CardNumb:3345-1230", Status: "Successful"},
    {TranId: "transaction-5", ApprovalCode: "22315", BillDetails: "store:124,items:jellwery,datetime:06/07/2023", CustomerDetails: "Cname:Mike,Cnumber:,CardNumb:3345-1255", Status: "Successful"},
    {TranId: "transaction-6", ApprovalCode: "22316", BillDetails: "store:122,items:shoes,datetime:06/08/2023", CustomerDetails: "Cname:Mitch,Cnumber:,CardNumb:3345-1265", Status: "failure"},
  }

  for _, tran := range trans {
    tranJSON, err := json.Marshal(tran)
    if err != nil {
      return err
    }

    err = ctx.GetStub().PutState(tran.TranId, tranJSON)
    if err != nil {
      return fmt.Errorf("failed to put to world state. %v", err)
    }
  }

  return nil
}

// CreateTransaction issues a new transaction to the world state with given details.
func (s *SmartContract) CreateTransaction(ctx contractapi.TransactionContextInterface, tranid string, approvalcode string, billdetails string, customerdetails string, status string) error {
  exists, err := s.TranExists(ctx, tranid)
  if err != nil {
    return err
  }
  if exists {
    return fmt.Errorf("the transaction %s already exists", tranid)
  }

  trans := Transaction{
    TranId:             tranid,
    ApprovalCode:          approvalcode,
    BillDetails:           billdetails,
    CustomerDetails:          customerdetails,
    Status: status,
  }
  transJSON, err := json.Marshal(trans)
  if err != nil {
    return err
  }

  return ctx.GetStub().PutState(trans.TranId, transJSON)
}

// ReadTransaction returns the transaction stored in the world state with given tranid.
func (s *SmartContract) ReadTransaction(ctx contractapi.TransactionContextInterface, tranid string) (*Transaction, error) {
  tranJSON, err := ctx.GetStub().GetState(tranid)
  if err != nil {
    return nil, fmt.Errorf("failed to read from world state: %v", err)
  }
  if tranJSON == nil {
    return nil, fmt.Errorf("the transaction %s does not exist", tranid)
  }

  var tran Transaction
  err = json.Unmarshal(tranJSON, &tran)
  if err != nil {
    return nil, err
  }

  return &tran, nil
}

// UpdateTransaction updates an existing transaction in the world state with provided parameters.
func (s *SmartContract) UpdateTransaction(ctx contractapi.TransactionContextInterface,  tranid string, approvalcode string, billdetails string, customerdetails string, status string) error {
  exists, err := s.TranExists(ctx, tranid)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the transaction %s does not exist", tranid)
  }

  // overwriting original transaction with new transaction
  trans := Transaction{
    TranId:             tranid,
    ApprovalCode:          approvalcode,
    BillDetails:           billdetails,
    CustomerDetails:          customerdetails,
    Status: status,
  }
  transJSON, err := json.Marshal(trans)
  if err != nil {
    return err
  }

  return ctx.GetStub().PutState(trans.TranId, transJSON)
}

// TranExists returns true when transaction with given ID exists in world state
func (s *SmartContract) TranExists(ctx contractapi.TransactionContextInterface, tranid string) (bool, error) {
  tranJSON, err := ctx.GetStub().GetState(tranid)
  if err != nil {
    return false, fmt.Errorf("failed to read from world state: %v", err)
  }

  return tranJSON != nil, nil
}

func main() {
  tranChaincode, err := contractapi.NewChaincode(&SmartContract{})
  if err != nil {
    log.Panicf("Error creating transaction chaincode: %v", err)
  }

  if err := tranChaincode.Start(); err != nil {
    log.Panicf("Error starting transaction chaincode: %v", err)
  }
}
