package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"hyperlibrary/common"
	"log"
	"os"
	"strconv"
)

func (t *SmartContract) Invoke(ctx contractapi.TransactionContextInterface) (string, error) {
	log.Println("ex02 Invoke")
	if os.Getenv("DEVMODE_ENABLED") != "" {
		log.Println("invoking in devmode")
	}
	function, args := ctx.GetStub().GetFunctionAndParameters()
	name, _, _ := ctx.GetClientIdentity().GetAttributeValue("Name")
	clientId, _ := ctx.GetClientIdentity().GetID()
	t.Events = []common.Event{}
	log.Println(fmt.Sprintf("Client id: %s %s", clientId, name))

	switch args[0] {
	case "create":
		var book common.Book
		err := json.Unmarshal([]byte(args[1]), &book)

		if err != nil {
			return "", err
		}

		book.DocType = "book"
		book.Owned = 0
		book.Available = 0

		err = t.CreateBook(ctx, &book)

		if err != nil {
			return "", err
		}

		return "", nil
	case "list":
		// Deletes an entity from its state
		books, err := t.ListBooks(ctx)

		if err != nil {
			return "", err
		}

		ret, err := json.Marshal(books)

		if err != nil {
			return "", err
		}

		return string(ret), nil
	case "purchase":
		isbn := args[1]
		q, err := strconv.ParseUint(args[2], 10, 8)

		if err != nil {
			return "", err
		}
		quantity := uint16(q)

		c, err := strconv.ParseFloat(args[3], 32)

		if err != nil {
			return "", err
		}
		cost := float32(c)

		insts, err := t.PurchaseBook(ctx, isbn, quantity, cost)
		print("foo")

		if err != nil {
			return "", err
		}

		instBytes, err := json.Marshal(insts)

		if err != nil {
			return "", err
		}

		return string(instBytes), nil
	case "inspect":
		instId := args[1]
		cond := common.Condition(args[2])
		c, _ := strconv.ParseFloat(args[3], 32)
		amt := float64(c)
		avail, _ := strconv.ParseBool(args[4])
		fee, err := t.InspectReturnedBook(ctx, instId, cond, amt, avail)

		if err != nil {
			return "", err
		}

		feeBytes, err := json.Marshal(fee)

		if err != nil {
			return "", err
		}

		return string(feeBytes), nil
	case "pay":
		c, err := strconv.ParseFloat(args[1], 64)

		if err != nil {
			return "", err
		}

		amount := float64(c)

		var ids []string
		err = json.Unmarshal([]byte(args[2]), &ids)

		if err != nil {
			return "", err
		}

		log.Println("Paying with the ids", amount, ids)

		p, err := t.PayFee(ctx, amount, ids)

		if err != nil {
			return "", err
		}

		log.Println("Payment", p)

		paymentBytes, err := json.Marshal(p)

		if err != nil {
			return "", err
		}

		return string(paymentBytes), nil
	default:
		return "", errors.New(fmt.Sprintf(`Invalid invoke "%s" function name. Expecting "invoke", "delete", "query", "respond", "mspid", or "event"`, function))
	}
}
