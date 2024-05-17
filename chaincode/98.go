package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	localUtils "github.com/newHouseSale/utils"
)

type Order struct {
	Id       string    `json:"id"`
	Customer *Customer `json:"customer"`
	House    *House    `json:"house"`
}

func (order *Order) insert(ctx contractapi.TransactionContextInterface) (*Order, error) {

	orderAsBytes, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal json array: %s ", err)
	}

	err = ctx.GetStub().PutState(localUtils.MakeOrderKey(order.Id, order.Customer.Id, order.Customer.Name, order.Customer.PhoneNum), orderAsBytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to put into world state: %s ", err)
	}

	return order, nil
}

func (order *Order) Insert(ctx contractapi.TransactionContextInterface) (*Order, error) {

	var err error
	order.Customer, err = order.Customer.Insert(ctx)
	if err != nil {
		return nil, err
	}
	return order.insert(ctx)
}


func (order *Order) Get(ctx contractapi.TransactionContextInterface) (*Order, error) {

	orderAsBytes, err := ctx.GetStub().GetState(localUtils.MakeOrderKey(order.Id, order.Customer.Id, order.Customer.Name, order.Customer.PhoneNum))
	if err != nil {
		return nil, fmt.Errorf("Failed to query the world state: %s ", err.Error())
	}

	if orderAsBytes == nil {
		return nil, nil
	}

	err = json.Unmarshal(orderAsBytes, order)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal order json data: %s ", err.Error())
	}
	return order, nil
}

func (order *Order) ParseArgsWithCustomer(ctx contractapi.TransactionContextInterface) (*Order, error) {
	args := ctx.GetStub().GetArgs()
	// 4 parameters: 1、funcName; 2、customerId; 3、customerName; 4、customerPhoneNumber
	if len(args) != 5 {
		return nil, fmt.Errorf("parameters do not match")
	}

	orderId := string(args[1])

	customerId := string(args[2])
	customerName := string(args[3])
	customerPhonenum := string(args[4])
	if orderId == "" || customerId == "" || customerName == "" || customerPhonenum == "" {
		return nil, fmt.Errorf("parameters not valid, missing customer or order info")
	}

	order.Id = orderId
	order.Customer = new(Customer)
	order.Customer.Id = customerId
	order.Customer.Name = customerName
	order.Customer.PhoneNum = customerPhonenum
	return order, nil
}

func (order *Order) ParseArgsWithCustomerAndHouse(ctx contractapi.TransactionContextInterface) (*Order, error) {
	args := ctx.GetStub().GetArgs()
	// 4 parameters: 1、funcName; 2、customerId; 3、customerName; 4、customerPhoneNumber
	if len(args) != 9 {
		return nil, fmt.Errorf("parameters do not match")
	}


	orderId := string(args[1])
	customerId := string(args[2])
	customerName := string(args[3])
	customerPhonenum := string(args[4])

	estateOrg := string(args[5])
	neighborhood := string(args[6])
	buildingId := string(args[7])
	roomId := string(args[8])


	if orderId == "" || customerId == "" || customerName == "" || customerPhonenum == "" {
		return nil, fmt.Errorf("parameters not valid, missing customer or order info")
	}

	if estateOrg == "" || neighborhood == "" || buildingId == "" || roomId == ""{
		return nil, fmt.Errorf("parameters not valid, missing house info")
	}

	order.Id = orderId
	order.Customer = new(Customer)
	order.Customer.Id = customerId
	order.Customer.Name = customerName
	order.Customer.PhoneNum = customerPhonenum

	order.House = new(House)
	order.House.EstateOrg = estateOrg
	order.House.Neighborhood = neighborhood
	order.House.BuildingId = buildingId
	order.House.RoomId = roomId
	return order, nil
}
