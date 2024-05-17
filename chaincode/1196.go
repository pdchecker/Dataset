package main

import (
	//"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	logger "github.com/sirupsen/logrus"
)

func (this *SmartContract) ResetWorldState(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("ResetWorldState: enter")
	defer logger.Debug("ResetWorldState: exit")

	dataCallsDeleteCount, _ := DeleteStateByKey(stub, DATA_CALL_PREFIX)
	reportsDeleteCount, _ := DeleteStateByKey(stub, REPORT_PREFIX)
	likesDeleteCount, _ := DeleteStateByKey(stub, LIKE_PREFIX)
	consentsDeleteCount, _ := DeleteStateByKey(stub, CONSENT_PREFIX)
	datacallLogDeleteCount, _ := DeleteStateByKey(stub, DATACALL_LOG_PREFIX)
	// logger.Debugf("%s DataCalls, %s Reports, %s Likes, %s Consents Deleted, %s DataCallsLog", dataCallsDeleteCount, reportsDeleteCount, likesDeleteCount, consentsDeleteCount, datacallLogDeleteCount)
	totalRecordsDeleted := dataCallsDeleteCount + reportsDeleteCount + likesDeleteCount + consentsDeleteCount + datacallLogDeleteCount
	logger.Debug("Total Records Deleted: ", totalRecordsDeleted)
	return shim.Success([]byte(strconv.Itoa(totalRecordsDeleted)))
}

//helper function to reset world state
func DeleteStateByKey(stub shim.ChaincodeStubInterface, key string) (int, error) {
	logger.Debug("DeleteStateByKey: enter")
	defer logger.Debug("DeleteStateByKey: exit")
	var document_prefix string
	var document_type string

	switch key {
	case DATA_CALL_PREFIX:
		document_prefix = DATA_CALL_PREFIX
		document_type = DOCUMENT_TYPE
	case REPORT_PREFIX:
		document_prefix = REPORT_PREFIX
		document_type = REPORT_DOCUMENT_TYPE
	case LIKE_PREFIX:
		document_prefix = LIKE_PREFIX
		document_type = LIKE_DOCUMENT_TYPE
	case CONSENT_PREFIX:
		document_prefix = CONSENT_PREFIX
		document_type = CONSENT_DOCUMENT_TYPE
	case DATACALL_LOG_PREFIX:
		document_prefix = DATACALL_LOG_PREFIX
		document_type = DATACALL_LOG_DOCUMENT

	}

	var recordsDeletedCount = 0
	var pks []string = []string{document_prefix}
	iterator, err := stub.GetStateByPartialCompositeKey(document_type, pks)
	//iterator, err := stub.GetQueryResult(queryStr)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get iterator for partial composite key: Error: %s", err.Error())
		logger.Error(errorMsg)
		return 0, err
	}
	// Once we are done with the iterator, we must close it
	defer iterator.Close()
	logger.Debugf("Starting to delete all records with prefix %s", document_prefix)
	for iterator.HasNext() {
		responseRange, err := iterator.Next()
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to get next record from iterator: %s", err.Error())
			logger.Error(errorMsg)
		}
		recordKey := responseRange.GetKey()
		logger.Debugf("About to delete record with key %s", recordKey)
		err = stub.DelState(recordKey)

		if err != nil {
			errorMsg := fmt.Sprintf("Failed to delete record '%d' with key %s: %s", recordsDeletedCount, recordKey, err.Error())
			logger.Error(errorMsg)
			return 0, err
		}

		recordsDeletedCount++
		// logger.Debugf("%s - Successfully deleted record '%d' ", recordsDeletedCount)
	}
	logger.Debug("Finished deleting all records for prefix: ", document_prefix)
	return recordsDeletedCount, nil

}
