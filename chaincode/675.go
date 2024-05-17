package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type CreditRecord struct {
	RecordId   string `json:"record_id,omitempty"  bson:"record_id"  form:"record_id"  binding:"record_id"`
	Email string `json:"email,omitempty"  bson:"email"  form:"email"  binding:"email"` 
	UserUid    string `json:"user_uid,omitempty"  bson:"user_uid"  form:"user_uid"  binding:"user_uid"`
	Nik        string `json:"nik,omitempty"  bson:"nik"  form:"nik"  binding:"nik"`
	CreditType string `json:"credit_type,omitempty"  bson:"credit_type"  form:"credit_type"  binding:"credit_type"`
	BankName   string `json:"bank_name,omitempty"  bson:"bank_name"  form:"bank_name"  binding:"bank_name"`
	Amount     string `json:"amount,omitempty"  bson:"amount"  form:"amount"  binding:"amount"`
	Status     string `json:"status,omitempty"  bson:"status"  form:"status"  binding:"status"`
}

type PaginatedQueryResult struct {
	Records             []*CreditRecord `json:"records"`
	FetchedRecordsCount int32           `json:"fetchedRecordsCount"`
	Bookmark            string          `json:"bookmark"`
}

const index = "useruid~recordid"
const index2 = "bankname~recordid"
const index3 = "nik~recordid"

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Printf("Error creating creditcc chaincode: %v", err)
	}
	if err := assetChaincode.Start(); err != nil {
		log.Printf("Error starting creditcc chaincode: %v", err)
	}
}

func (s *SmartContract) CheckCreditFromSameBank(ctx contractapi.TransactionContextInterface, userUid string, bankName string) (bool, error) {
	// assetJson, err := ctx.GetStub().GetState(userUid)
	records, err := s.ReadAssetByUserUid(ctx, userUid)
	if err != nil {
		return false, fmt.Errorf("failed to read from state database: %v", err)
	}

	if len(records) > 0 {
		for _, val := range records {
			if val.BankName != bankName {
				return true, nil
			}
		}
	}

	// var creditRecord CreditRecord
	// err = json.Unmarshal(assetJson, &creditRecord)
	// if err != nil {
	// 	return false, err
	// }

	// if creditRecord.BankName != bankName {
	// 	return true, nil
	// }
	return false, nil
}

func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, recordId string, email string, userUid string,
	nik string, creditType string, bankName string, amount string, status string) error {
	// checkStatus, err := s.CheckCreditFromSameBank(ctx, userUid, bankName)
	// if err != nil {
	// 	fmt.Printf("A")
	// 	return fmt.Errorf("A : %s", err.Error())
	// }
	// if checkStatus {
	// 	return fmt.Errorf("this user (%s) has a credit record from another bank", userUid)
	// }

	creditRecord := CreditRecord{
		RecordId:   recordId,
		Email: email,
		UserUid:    userUid,
		Nik:        nik,
		CreditType: creditType,
		BankName:   bankName,
		Amount:     amount,
		Status:     status,
	}

	assetJson, err := json.Marshal(creditRecord)
	if err != nil {
		fmt.Printf("C")
		return fmt.Errorf("c : %s", err.Error())
	}
	err = ctx.GetStub().PutState(recordId, assetJson)
	if err != nil {
		fmt.Printf("D")
		return fmt.Errorf("f : %s", err.Error())
	}

	indexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{userUid, recordId})
	if err != nil {
		fmt.Printf("E")
		return fmt.Errorf("e Error Record ID : %s, %s", recordId, err.Error())
	}
	value := []byte{0x00}
	// fmt.Printf("Store index key")
	err = ctx.GetStub().PutState(indexKey, value)
	if err != nil {
		return fmt.Errorf("f :  %s", err.Error())
	}

	indexKey3, err := ctx.GetStub().CreateCompositeKey(index3, []string{nik, recordId})
	if err != nil {
		fmt.Printf("E")
		return fmt.Errorf("g Error Record ID :  %s, %s", recordId, err.Error())
	}
	err = ctx.GetStub().PutState(indexKey3, value)
	if err != nil {
		return fmt.Errorf("h :  %s", err.Error())
	}

	indexKey2, err := ctx.GetStub().CreateCompositeKey(index2, []string{bankName, recordId})
	if err != nil {
		fmt.Printf("E")
		return fmt.Errorf("i Error Record ID : %s, %s", recordId, err.Error())
	}
	return ctx.GetStub().PutState(indexKey2, value)
	// err = ctx.GetStub().PutState(indexKey, value)
	// if err != nil {
	// 	return fmt.Errorf("f")
	// }
	// return fmt.Errorf("nil")
}

func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, recordId string) (*CreditRecord, error) {
	assetBytes, err := ctx.GetStub().GetState(recordId)
	if err != nil {
		return nil, fmt.Errorf("failed to read record %s: %v", recordId, err)
	}
	if assetBytes == nil {
		return nil, fmt.Errorf("record %s does not exist", recordId)
	}

	var record CreditRecord
	err = json.Unmarshal(assetBytes, &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, recordId string) error {
	exists, err := s.AssetExists(ctx, recordId)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the user %s does not exist", recordId)
	}
		record, err := s.ReadAsset(ctx, recordId)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	indexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{record.UserUid, recordId})
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().DelState(indexKey)
	if err != nil {
		return fmt.Errorf(err.Error())
	}


	indexKey2, err := ctx.GetStub().CreateCompositeKey(index2, []string{record.BankName, recordId})
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().DelState(indexKey2)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	indexKey3, err := ctx.GetStub().CreateCompositeKey(index3, []string{record.Nik, recordId})
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().DelState(indexKey3)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	return ctx.GetStub().DelState(recordId)
}

func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*CreditRecord, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []*CreditRecord
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var record CreditRecord
		err = json.Unmarshal(queryResponse.Value, &record)
		if err != nil {
			return nil, err
		}
		records = append(records, &record)
	}

	return records, nil
}

func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, recordId string, email string, userUid string,
	nik string, creditType string, bankName string, amount string, status string) error {
	exists, err := s.AssetExists(ctx, recordId)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the record %s does not exist", recordId)
	}
	creditRecord := CreditRecord{
		RecordId:   recordId,
		Email: email,
		UserUid:    userUid,
		Nik:        nik,
		CreditType: creditType,
		BankName:   bankName,
		Amount:     amount,
		Status:     status,
	}
	assetJson, err := json.Marshal(creditRecord)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(recordId, assetJson)
}

func (s *SmartContract) ReadAssetByUserUid(ctx contractapi.TransactionContextInterface, userUid string) ([]*CreditRecord, error) {
	indexIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(index, []string{userUid})
	if err != nil {
		return nil, err
	}
	defer indexIterator.Close()

	var records []*CreditRecord
	// var parts []string
	for indexIterator.HasNext() {
		responseRange, err := indexIterator.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		if len(compositeKeyParts) > 1 {
			recordId := compositeKeyParts[1]
			record, err := s.ReadAsset(ctx, recordId)
			if err != nil {
				return nil, err
			}
			records = append(records, record)
		}

		// parts = append(parts, compositeKeyParts[0])
		// parts = append(parts, compositeKeyParts[1])

	}
	return records, nil
}

func (s *SmartContract) ReadAssetByBank(ctx contractapi.TransactionContextInterface, bankName string) ([]*CreditRecord, error) {
	indexIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(index2, []string{bankName})
	if err != nil {
		return nil, err
	}
	defer indexIterator.Close()

	var records []*CreditRecord
	// var parts []string
	for indexIterator.HasNext() {
		responseRange, err := indexIterator.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		if len(compositeKeyParts) > 1 {
			recordId := compositeKeyParts[1]
			record, err := s.ReadAsset(ctx, recordId)
			if err != nil {
				return nil, err
			}
			records = append(records, record)
		}

		// parts = append(parts, compositeKeyParts[0])
		// parts = append(parts, compositeKeyParts[1])

	}
	return records, nil
}

func (s *SmartContract) ReadAssetByNik(ctx contractapi.TransactionContextInterface, nik string) ([]*CreditRecord, error) {
	indexIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(index3, []string{nik})
	if err != nil {
		return nil, err
	}
	defer indexIterator.Close()

	var records []*CreditRecord
	// var parts []string
	for indexIterator.HasNext() {
		responseRange, err := indexIterator.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		if len(compositeKeyParts) > 1 {
			recordId := compositeKeyParts[1]
			record, err := s.ReadAsset(ctx, recordId)
			if err != nil {
				return nil, err
			}
			records = append(records, record)
		}

		// parts = append(parts, compositeKeyParts[0])
		// parts = append(parts, compositeKeyParts[1])

	}
	return records, nil
}

func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, recordId string) (bool, error) {
	assetJson, err := ctx.GetStub().GetState(recordId)
	if err != nil {
		return false, fmt.Errorf("failed to read from state database: %v", err)
	}
	return assetJson != nil, nil
}
