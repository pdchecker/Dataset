package main

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

type Order struct {
	L_ORDERKEY      int     `json:"L_ORDERKEY"`
	L_PARTKEY       int     `json:"L_PARTKEY"`
	L_SUPPKEY       int     `json:"L_SUPPKEY"`
	L_LINENUMBER    int     `json:"L_LINENUMBER"`
	L_QUANTITY      int     `json:"L_QUANTITY"`
	L_EXTENDEDPRICE float64 `json:"L_EXTENDEDPRICE"`
	L_DISCOUNT      float64 `json:"L_DISCOUNT"`
	L_TAX           float64 `json:"L_TAX"`
	L_RETURNFLAG    string  `json:"L_RETURNFLAG"`
	L_LINESTATUS    string  `json:"L_LINESTATUS"`
	L_SHIPDATE      string  `json:"L_SHIPDATE"`
	L_COMMITDATE    string  `json:"L_COMMITDATE"`
	L_RECEIPTDATE   string  `json:"L_RECEIPTDATE"`
	L_SHIPINSTRUCT  string  `json:"L_SHIPINSTRUCT"`
	L_SHIPMODE      string  `json:"L_SHIPMODE"`
	L_COMMENT       string  `json:"L_COMMENT"`
}

type QueryResult struct {
	Key       string `json:"Key"`
	Record    *Order `json:"record"`
	Timestamp string `json:"timestamp"`
}

// SimpleContract contract for handling writing and reading from the world state
type SmartContract struct {
}

func (sc *SmartContract) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (sc *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := stub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	switch function {
	case "InitLedger":
		return sc.InitLedger(stub)
	case "CreateBulk":
		return sc.CreateBulk(stub, args)
	case "CreateBulkParallel":
		return sc.CreateBulkParallel(stub, args)
	case "Create":
		return sc.Create(stub, args)
	case "getHistoryForAsset":
		return sc.getHistoryForAsset(stub, args)
	case "getHistoryForAssets":
		return sc.getHistoryForAssets(stub, args)
	case "getVersionsForAsset":
		return sc.getVersionsForAsset(stub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}
}

func (sc *SmartContract) InitLedger(stub shim.ChaincodeStubInterface) sc.Response {
	log.Println("'============= Initialized Ledger ==========='")
	return shim.Success(nil)

}

func (sc *SmartContract) Create(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var order Order
	json.Unmarshal([]byte(args[0]), &order)

	orderBytes, err := json.Marshal(order)
	if err != nil {
		return shim.Error("Failed to marshal order JSON: " + err.Error())
	}

	orderKey := strconv.FormatInt(int64(order.L_ORDERKEY), 10)
	log.Printf("Appending order: %s\n", orderKey)

	err = stub.PutState(orderKey, orderBytes)
	if err != nil {
		return shim.Error("failed to put order on ledger: " + err.Error())
	}

	return shim.Success(nil)

}

// Create a new key-value pair and send to state database
func (sc *SmartContract) CreateBulk(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	buffer := args[0]

	var orders []Order
	json.Unmarshal([]byte(buffer), &orders)

	for _, order := range orders {

		orderBytes, err := json.Marshal(order)
		if err != nil {
			return shim.Error("failed to marshal order JSON: " + err.Error())
		}

		orderKey := strconv.FormatInt(int64(order.L_ORDERKEY), 10)

		// Fabric key must be a string
		//fmt.Sprintf("%d", order.L_ORDERKEY)
		log.Printf("Appending order %s with part %d\n", orderKey, order.L_PARTKEY)
		err = stub.PutState(orderKey, orderBytes)
		if err != nil {
			return shim.Error("failed to put order on ledger: " + err.Error())
		}
	}

	return shim.Success(nil)

}

func (sc *SmartContract) CreateBulkParallel(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var orders []Order
	json.Unmarshal([]byte(args[0]), &orders)

	for _, order := range orders {
		orderBytes, err := json.Marshal(order)
		if err != nil {
			return shim.Error("Error marshaling order object: " + err.Error())
		}

		err = stub.PutState(strconv.Itoa(order.L_ORDERKEY), orderBytes)
		if err != nil {
			return shim.Error("Failed to create order: " + err.Error())
		}
	}
	return shim.Success(nil)
}

// getHistoryForAsset calls built in GetHistoryForKey() API
func (sc *SmartContract) getHistoryForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	historyIer, err := stub.GetHistoryForKey(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	var history []QueryResult
	for historyIer.HasNext() {
		historyData, err := historyIer.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var order Order
		json.Unmarshal(historyData.Value, &order)

		//Convert google.protobuf.Timestamp to string
		timestamp := time.Unix(historyData.Timestamp.Seconds, int64(historyData.Timestamp.Nanos)).String()

		history = append(history, QueryResult{Key: historyData.TxId, Record: &order, Timestamp: timestamp})
	}

	historyAsBytes, _ := json.Marshal(history)
	return shim.Success(historyAsBytes)
}

// getHistoryForAssets calls custom GetHistoryForKeys() API
// OLD
// func (sc *SmartContract) getHistoryForAssets(stub shim.ChaincodeStubInterface, args []string) sc.Response {
// 	if len(args) < 1 {
// 		return shim.Error("Incorrect number of arguments. Expecting 1 or more")
// 	}

// 	// calling the GetHistoryForKeys() API with keys as args
// 	historyIers, err := stub.GetHistoryForKeys(args)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	var histories [][]QueryResult
// 	for _, historyIer := range historyIers {
// 		var history []QueryResult
// 		for historyIer.HasNext() {
// 			historyData, err := historyIer.Next()
// 			if err != nil {
// 				return shim.Error(err.Error())
// 			}

// 			var order Order
// 			json.Unmarshal(historyData.Value, &order)

// 			history = append(history, QueryResult{Key: historyData.TxId, Record: &order})
// 		}
// 		histories = append(histories, history)
// 	}

// 	historiesAsBytes, _ := json.Marshal(histories)
// 	return shim.Success(historiesAsBytes)
// }

// getHistoryForAssets calls custom GetHistoryForKeys() API
// NEW
func (sc *SmartContract) getHistoryForAssets(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1 or more")
	}

	// calling the GetHistoryForKeys() API with keys as args
	historyIer, err := stub.GetHistoryForKeys(args) // historyIters in old version
	if err != nil {
		return shim.Error(err.Error())
	}

	var history []QueryResult
	for historyIer.HasNext() {
		historyData, err := historyIer.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var order Order
		json.Unmarshal(historyData.Value, &order)

		//Convert google.protobuf.Timestamp to string
		timestamp := time.Unix(historyData.Timestamp.Seconds, int64(historyData.Timestamp.Nanos)).String()

		history = append(history, QueryResult{Key: historyData.TxId, Record: &order, Timestamp: timestamp})
	}

	historyAsBytes, _ := json.Marshal(history)
	return shim.Success(historyAsBytes)
}

func (sc *SmartContract) getVersionsForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	start, _ := strconv.ParseUint(args[1], 10, 64)
	end, _ := strconv.ParseUint(args[2], 10, 64)

	versionIter, err := stub.GetVersionsForKey(args[0], start, end)
	if err != nil {
		return shim.Error(err.Error())
	}

	var versions []QueryResult
	for versionIter.HasNext() {
		versionData, err := versionIter.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var order Order
		json.Unmarshal(versionData.Value, &order) // .Value?

		timestamp := time.Unix(versionData.Timestamp.Seconds, int64(versionData.Timestamp.Nanos)).String()

		versions = append(versions, QueryResult{Key: versionData.TxId, Record: &order, Timestamp: timestamp})
	}

	versionAsBytes, _ := json.Marshal(versions)
	return shim.Success(versionAsBytes)
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		log.Printf("Error starting chaincode: %v \n", err)
	}
}
