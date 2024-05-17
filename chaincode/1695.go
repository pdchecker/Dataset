package main

import (
	//"bytes"
	//"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	logger "github.com/sirupsen/logrus"
	//"strconv"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
}

// openIDLCC is a chaincode component that supports the main operations for the openIDL network
type openIDLCC struct {
	carriers map[string]Carrier
}

var crossInvocationChannels Channels

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	logger.Info("Initledger: enter")
	defer logger.Debug("InitLedger: exit")

	var logLevelConfig = os.Getenv(LOGGING_LEVEL)
	if logLevelConfig != "" {
		// logger.SetLevel(logLevelConfig)
		/*temporary*/
		logger.SetLevel(logger.DebugLevel)
	} else {
		logger.SetLevel(logger.DebugLevel)
	}
	return nil
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data, so be careful to avoid a scenario where you
// inadvertently clobber your ledger's data!
func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("Init: enter")
	defer logger.Debug("Init: exit")

	var logLevelConfig = os.Getenv(LOGGING_LEVEL)
	if logLevelConfig != "" {
		// logger.SetLevel(logLevelConfig)
		/*temporary*/
		logger.SetLevel(logger.DebugLevel)
	} else {
		logger.SetLevel(logger.DebugLevel)
	}

	/*init_args := stub.GetStringArgs()
	logger.Debug("Init args > ", init_args)
	if len(init_args) == 2 {
		initChannelsJson := init_args[1]
		logger.Debug("Channels Argument > ", initChannelsJson)
		json.Unmarshal([]byte(initChannelsJson), &crossInvocationChannels)
		logger.Debug("Channels List Marshalled Successfully > ", crossInvocationChannels)
	}*/

	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The 'set'
// method may create a new asset by specifying a new key-value pair.
//func (s *SmartContract) Invoke2(ctx contractapi.TransactionContextInterface, id string) error {
func (this *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("Invoke: enter")
	defer logger.Debug("Invoke: exit")
	//function and parameters
	function, args := stub.GetFunctionAndParameters()

	hasAccess, err := checkAccessForOrg(stub, function)

	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}
	if !hasAccess {
		errStr := fmt.Sprintf("checkAccessForOrg: The organisation doesn't have access for function %v", function)
		return shim.Error(errors.New(errStr).Error())
	}

	logger.Info("Invoke: function: ", function)

	switch function {
	case "Ping":
		return this.Ping(stub)
	case "ListDataCallsByCriteria":
		return this.ListDataCallsByCriteria(stub, args[0])
	case "ListMatureDataCalls":
		return this.ListMatureDataCalls(stub, args[0])
	case "CreateDataCall":
		return this.CreateDataCall(stub, args[0])
	case "SaveNewDraft":
		return this.SaveNewDraft(stub, args[0])
	case "UpdateDataCall":
		return this.UpdateDataCall(stub, args[0])
	case "IssueDataCall":
		return this.IssueDataCall(stub, args[0])
	case "GetDataCallVersionsById":
		return this.GetDataCallVersionsById(stub, args[0])
	case "GetDataCallByIdAndVersion":
		return this.GetDataCallByIdAndVersion(stub, args[0])
	case "ToggleLike":
		return this.ToggleLike(stub, args[0])
	case "CreateLikeCountEntry":
		return this.CreateLikeCountEntry(stub, args[0])
	case "CountLikes":
		return this.CountLikes(stub, args[0])
	case "CreateConsent":
		return this.CreateConsent(stub, args[0])
	case "CreateReconsent":
		return this.CreateReconsent(stub, args[0])
	case "CreateConsentCountEntry":
		return this.CreateConsentCountEntry(stub, args[0])
	case "CountConsents":
		return this.CountConsents(stub, args[0])
	case "ListConsentsByDataCall":
		return this.ListConsentsByDataCall(stub, args[0])
	case "GetConsentsByDataCall":
		return this.GetConsentsByDataCall(stub, args[0])
	case "GetHashById":
		return this.GetHashById(stub, args[0])
	case "GetConsentByDataCallAndOrganization":
		return this.GetConsentByDataCallAndOrganization(stub, args)
	case "ListLikesByDataCall":
		return this.ListLikesByDataCall(stub, args[0])
	case "GetLikesByDataCall":
		return this.GetLikesByDataCall(stub, args[0])
	case "GetLikeByDataCallAndOrganization":
		return this.GetLikeByDataCallAndOrganization(stub, args)
	case "SaveAndIssueDataCall":
		return this.SaveAndIssueDataCall(stub, args[0])
	case "CreateReport":
		return this.CreateReport(stub, args[0])
	case "UpdateReport":
		return this.UpdateReport(stub, args[0])
	case "ListReportsByCriteria":
		return this.ListReportsByCriteria(stub, args[0])
	case "ResetWorldState":
		return this.ResetWorldState(stub)
	case "LogDataCallTransaction":
		return this.LogDataCallTransaction(stub, args[0])
	case "GetDataCallTransactionHistory":
		return this.GetDataCallTransactionHistory(stub, args[0])
	case "GetReportById":
		return this.GetReportById(stub, args[0])
	case "ListDataCallTransactionHistory":
		return this.ListDataCallTransactionHistory(stub, args[0])
	case "SaveInsuranceDataHash":
		return this.SaveInsuranceDataHash(stub, args[0])
	case "CheckExtractionPatternIsSet":
		return this.CheckExtractionPatternIsSet(stub, args[0])
	case "SaveInsuranceData":
		return this.SaveInsuranceData(stub, args)
	case "CheckInsuranceDataExists":
		return this.CheckInsuranceDataExists(stub, args[0])
	case "GetExtractionPatternByIds":
		return this.GetExtractionPatternByIds(stub, args[0])
	case "GetInsuranceData":
		return this.GetInsuranceData(stub, args[0])
	case "CreateExtractionPattern":
		return this.CreateExtractionPattern(stub, args[0])
	case "UpdateExtractionPattern":
		return this.UpdateExtractionPattern(stub, args[0])
	case "GetDataCallAndExtractionPattern":
		return this.GetDataCallAndExtractionPattern(stub, args[0])
	case "ListExtractionPatterns":
		return this.ListExtractionPatterns(stub)
	case "UpdateLikeCountForDataCall":
		return this.UpdateLikeCountForDataCall(stub, args[0])
	case "UpdateConsentCountForDataCall":
		return this.UpdateConsentCountForDataCall(stub, args[0])
	case "ToggleDataCallCount":
		return this.ToggleDataCallCount(stub, args[0])
	case "GetDataCallCount":
		return this.GetDataCallCount(stub, args[0])
	case "UpdateDataCallCount":
		return this.UpdateDataCallCount(stub, args[0])
	case "SearchDataCalls":
		return this.SearchDataCalls(stub, args[0])
	case "UpdateConsentStatus":
		return this.UpdateConsentStatus(stub, args[0])
	default:
		//error
		return shim.Error("Invalid Function: " + function)
	}

}

// Ping simply returns a string as a way to validate that the chaincode component is up and running.
func (s *SmartContract) Ping(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Ping: enter")
	defer logger.Debug("Ping: exit")
	return shim.Success([]byte("Ping OK"))
}

func main() {
	// err := shim.Start(new(openIDLCC))
	// if err != nil {
	// 	logger.Error("Error starting openIDLCC: %s", err)
	// }
	// assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	// if err != nil {
	// 	logger.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
	// }

	// if err := assetChaincode.Start(); err != nil {
	// 	logger.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
	// }
	if err := shim.Start(new(SmartContract)); err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
