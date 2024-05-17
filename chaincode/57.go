package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"hyperlibrary/common"
	"log"
	"math"
	"time"
)

func (t *SmartContract) GetFees(ctx contractapi.TransactionContextInterface, clientId string) ([]*common.Fee, error) {
	query := fmt.Sprintf(`{"selector":{"docType":"fee", "borrower": {"clientId": "%s"}}}`, clientId)
	res, err := GetQueryResultForQueryString(ctx, query)

	if err != nil {
		return nil, err
	}

	var fees []*common.Fee
	for i := range res {
		feeBytes := res[i]
		var fee common.Fee
		err = json.Unmarshal(feeBytes, &fee)
		fees = append(fees, &fee)
	}
	return fees, nil
}

func (t *SmartContract) GetUnpaidFees(ctx contractapi.TransactionContextInterface, clientId string) ([]*common.Fee, error) {
	query := fmt.Sprintf(`{"selector":{"docType":"fee", "fullyPaid": false, "borrower": {"clientId": "%s"}}}`, clientId)
	res, err := GetQueryResultForQueryString(ctx, query)

	if err != nil {
		return nil, err
	}

	var fees []*common.Fee
	for i := range res {
		feeBytes := res[i]
		var fee common.Fee
		err = json.Unmarshal(feeBytes, &fee)
		fees = append(fees, &fee)
	}
	return fees, nil
}

func (t *SmartContract) GetMyFees(ctx contractapi.TransactionContextInterface) ([]*common.Fee, error) {
	id, _ := ctx.GetClientIdentity().GetID()
	return t.GetFees(ctx, id)
}

func (t *SmartContract) GetMyUnpaidFees(ctx contractapi.TransactionContextInterface) ([]*common.Fee, error) {
	id, _ := ctx.GetClientIdentity().GetID()
	return t.GetUnpaidFees(ctx, id)
}

func (t *SmartContract) checkForLateFee(ctx contractapi.TransactionContextInterface, inst *common.BookInstance) (*common.Fee, error) {
	now := time.Now()

	if inst.DueDate.Before(now) {
		diff := now.Sub(inst.DueDate).Round(time.Hour)
		diffDays := math.RoundToEven(diff.Hours() / 24)

		if diffDays > 0 {
			id := ctx.GetStub().GetTxID()
			fee := t.LateFeePerDay * diffDays
			log.Println("A late fee is owed", t.LateFeePerDay, diffDays, fee)
			ts, _ := ctx.GetStub().GetTxTimestamp()
			date := common.GetApproxTime(ts)

			if fee > 0 {
				var amt float64 = fee
				if fee > float64(inst.Cost) {
					log.Println("Fee would cost more than the price of the book!")
					amt = float64(inst.Cost)
				}

				lateFee := common.Fee{"fee", id, inst.Borrower, amt, common.LATE_FEE, date, 0.0, false}

				_, err := t.storeFee(ctx, &lateFee)

				if err != nil {
					return &common.Fee{}, err
				}

				//err = ctx.GetStub().SetEvent("LateFee.Created", lateFeeBytes)
				t.AddEvent(ctx, "Fee.Created", lateFee)

				if err != nil {
					return &common.Fee{}, err
				}

				return &lateFee, nil
			}
		}
	}

	return nil, nil
}

func (t *SmartContract) GetFee(ctx contractapi.TransactionContextInterface, feeId string) (*common.Fee, error) {
	feeBytes, err := ctx.GetStub().GetState(fmt.Sprintf("fee.%s", feeId))

	if err != nil {
		return nil, err
	}

	var fee *common.Fee
	err = json.Unmarshal(feeBytes, &fee)

	if err != nil {
		return nil, err
	}

	return fee, nil
}

func (t *SmartContract) storeFee(ctx contractapi.TransactionContextInterface, fee *common.Fee) ([]byte, error) {
	feeBytes, err := json.Marshal(fee)

	if err != nil {
		return nil, err
	}

	err = ctx.GetStub().PutState(fmt.Sprintf("fee.%s", fee.Id), feeBytes)

	if err != nil {
		return nil, err
	}

	return feeBytes, nil
}

func (t *SmartContract) distributePayment(ctx contractapi.TransactionContextInterface, amount float64, feeIds []string) (map[string]float64, error) {
	remainingAmount := amount
	fees := make(map[string]float64, len(feeIds))

	for i := range feeIds {
		feeId := feeIds[i]
		fee, err := t.GetFee(ctx, feeId)

		if err != nil {
			return map[string]float64{}, err
		}

		if fee.FullyPaid {
			continue
		}

		remainingFee := fee.Fee - fee.AmountPaid

		if remainingAmount > 0 {
			if remainingAmount >= remainingFee {
				fee.AmountPaid += remainingFee
				remainingAmount -= remainingFee
				fees[feeId] = remainingFee
			} else {
				ableToPay := remainingFee - remainingAmount
				fee.AmountPaid += ableToPay
				remainingAmount -= ableToPay
				fees[feeId] = ableToPay
			}

			var feeEvent string
			if fee.AmountPaid == fee.Fee {
				fee.FullyPaid = true
				feeEvent = "LateFee.FullyPaid"
			} else {
				feeEvent = "LateFee.PartiallyPaid"
			}

			_, err := t.storeFee(ctx, fee)

			if err != nil {
				return map[string]float64{}, err
			}

			//ctx.GetStub().SetEvent(feeEvent, feeBytes)
			t.AddEvent(ctx, feeEvent, fee)
		} else {
			fees[feeId] = 0
			log.Println("For the fee there isn't enough money left in the payment.", fee)
		}
	}

	return fees, nil
}

func (t *SmartContract) storePayment(ctx contractapi.TransactionContextInterface, payment *common.Payment) (*common.Payment, error) {
	paymentBytes, err := json.Marshal(payment)

	if err != nil {
		log.Fatalf(err.Error())
		return nil, err
	}

	err = ctx.GetStub().PutState(fmt.Sprintf("payment.%s", payment.Id), paymentBytes)

	if err != nil {
		log.Fatalf(err.Error())
		return nil, err
	}

	//err = ctx.GetStub().SetEvent("Payment.Created", paymentBytes)
	t.AddEvent(ctx, "Payment.Created", payment)

	if err != nil {
		log.Fatalf(err.Error())
		return nil, err
	}

	return payment, nil
}

func (t *SmartContract) PayFee(ctx contractapi.TransactionContextInterface, amount float64, feeIds []string) (*common.Payment, error) {
	feesPaid, err := t.distributePayment(ctx, amount, feeIds)

	log.Println("Fees to be paid", feesPaid)

	if err != nil {
		return nil, err
	}

	if len(feesPaid) > 0 {
		txId := ctx.GetStub().GetTxID()
		ts, _ := ctx.GetStub().GetTxTimestamp()
		date := common.GetApproxTime(ts)

		payment := common.Payment{"payment", txId,
			t.GetUserByClientId(ctx), amount, date, feesPaid,
		}

		log.Println("Creating payment", payment)

		p, err := t.storePayment(ctx, &payment)
		t.SetEvents(ctx)
		return p, err
	}

	return nil, nil
}

func (t *SmartContract) GetFeeHistory(ctx contractapi.TransactionContextInterface, id string) ([]common.History, error) {
	return t.GetHistory(ctx, fmt.Sprintf("fee.%s", id))
}
