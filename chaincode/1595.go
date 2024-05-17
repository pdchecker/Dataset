    
package main
import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
    "strconv"
    "strings"
    
    "github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type DRLDocsChaincode struct {
}

type DRLDocs struct {
    Id string `json:"Id"`
    CreatedOn time.Time `json:"CreatedOn"`
    CreatedBy string `json:"CreatedBy"`
    IsDelete bool `json:"IsDelete"`
    DPRNo string `json:"DPRNo"`
    ActiveCCDR string `json:"ActiveCCDR"`
    PassiveCCDR string `json:"PassiveCCDR"`
    TaxInvoice string `json:"TaxInvoice"`
    LRcopy string `json:"LRcopy"`
    SealCodeBeforeDispatch string `json:"SealCodeBeforeDispatch"`
    SealCodeAfterDelivery string `json:"SealCodeAfterDelivery"`
    SignedLRcopy string `json:"SignedLRcopy"`
    Notes string `json:"Notes"`
}

func (cc *DRLDocsChaincode) create(stub shim.ChaincodeStubInterface, arg []string) peer.Response {
 
    args := strings.Split(arg[0], "^^")

    if len(args) != 13 {
        return shim.Error("Incorrect number arguments. Expecting 13")
    }
	dateValue1, err1 := time.Parse(time.RFC3339, args[1])

	if err1 != nil {
		return shim.Error("Error converting string to date: " + err1.Error())
	}
	boolValue3, err3 := strconv.ParseBool(args[3])

	if err3  != nil {
		return shim.Error("Error converting string to bool: " + err3.Error())
	}
    data := DRLDocs{ Id: args[0], CreatedOn: dateValue1, CreatedBy: args[2], IsDelete: boolValue3, DPRNo: args[4], ActiveCCDR: args[5], PassiveCCDR: args[6], TaxInvoice: args[7], LRcopy: args[8], SealCodeBeforeDispatch: args[9], SealCodeAfterDelivery: args[10], SignedLRcopy: args[11], Notes: args[12] }

    dataBytes, errMarshal := json.Marshal(data)

    if errMarshal != nil {
        return shim.Error("Error converting data as bytes: " + errMarshal.Error())
    }

    errPut := stub.PutState(args[0], dataBytes)

    if errPut != nil {
        return shim.Error("Error putting the state: " + errPut.Error())
    }

    return shim.Success(nil)
}

func (cc *DRLDocsChaincode) get(stub shim.ChaincodeStubInterface, args []string) peer.Response {

    if len(args) != 1 {
        return shim.Error("Incorrect number arguments. Expecting 1")
    }

    stateBytes, err := stub.GetState(args[0])

    if err != nil {
        return shim.Error("Error getting the state: " + err.Error())
    }

    return shim.Success(stateBytes)
}
func (cc *DRLDocsChaincode) delete(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number arguments. Expecting 1")
	}

	dataBytes, err := stub.GetState(args[0])

	if err != nil {
		return shim.Error("Error getting the state: " + err.Error())
	}

	data := DRLDocs{}

	json.Unmarshal(dataBytes, &data)
    
    data.IsDelete = true

	updateDataBytes, err1 := json.Marshal(data)

	if err1 != nil {
		return shim.Error("Error converting data as bytes: " + err1.Error())
	}

	err2 := stub.PutState(args[0], updateDataBytes)

	if err2 != nil {
		return shim.Error("Error putting the data state: " + err2.Error())
	}

	return shim.Success(nil)
}
func (cc *DRLDocsChaincode) update(stub shim.ChaincodeStubInterface, arg []string) peer.Response {
    
    args := strings.Split(arg[0], "^^")     

	if len(args) != 13 {
		return shim.Error("Incorrect number arguments. Expecting 13")
	}
	dateValue1, err1 := time.Parse(time.RFC3339, args[1])

	if err1 != nil {
		return shim.Error("Error converting string to date: " + err1.Error())
	}
	boolValue3, err3 := strconv.ParseBool(args[3])

	if err3  != nil {
		return shim.Error("Error converting string to bool: " + err3.Error())
	}
    data := DRLDocs{ Id: args[0], CreatedOn: dateValue1, CreatedBy: args[2], IsDelete: boolValue3, DPRNo: args[4], ActiveCCDR: args[5], PassiveCCDR: args[6], TaxInvoice: args[7], LRcopy: args[8], SealCodeBeforeDispatch: args[9], SealCodeAfterDelivery: args[10], SignedLRcopy: args[11], Notes: args[12] }

    dataBytes, errMarshal := json.Marshal(data)

    if errMarshal != nil {
        return shim.Error("Error converting data as bytes: " + errMarshal.Error())
    }

    errPut := stub.PutState(args[0], dataBytes)

    if errPut != nil {
        return shim.Error("Error putting the data state: " + errPut.Error())
    }

	return shim.Success(nil)
}

func (cc *DRLDocsChaincode) history(stub shim.ChaincodeStubInterface, args []string) peer.Response {

    if len(args) != 1 {
        return shim.Error("Incorrect number of arguments. Expecting 1")
    }

    queryResult, err := stub.GetHistoryForKey(args[0])

    if err != nil {
        return shim.Error("Error getting history results: " + err.Error())
    }

    var buffer bytes.Buffer
    buffer.WriteString("[")

    isDataAdded := false
    for queryResult.HasNext() {
        queryResponse, err1 := queryResult.Next()
        if err1 != nil {
            return shim.Error(err1.Error())
        }

        if isDataAdded == true {
            buffer.WriteString(",")
        }

        buffer.WriteString(string(queryResponse.Value))

        isDataAdded = true
    }
    buffer.WriteString("]")

    return shim.Success(buffer.Bytes())
}

func (cc *DRLDocsChaincode) querystring(stub shim.ChaincodeStubInterface, args []string) peer.Response {

    if len(args) != 1 {
        return shim.Error("Incorrect number of arguments. Expecting 1")
    }

    queryResult, err := stub.GetQueryResult(args[0])

    if err != nil {
        return shim.Error("Error getting query string results: " + err.Error())
    }

    var buffer bytes.Buffer
    buffer.WriteString("[")

    isDataAdded := false
    for queryResult.HasNext() {
        queryResponse, err1 := queryResult.Next()
        if err1 != nil {
            return shim.Error(err1.Error())
        }

        if isDataAdded == true {
            buffer.WriteString(",")
        }

        buffer.WriteString(string(queryResponse.Value))

        isDataAdded = true
    }
    buffer.WriteString("]")

    return shim.Success(buffer.Bytes())
}
func (cc *DRLDocsChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (cc *DRLDocsChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	function, args := stub.GetFunctionAndParameters()

	if function == "create" { return cc.create(stub, args)
    } else if function == "get" { return cc.get(stub, args)
    } else if function == "delete" { return cc.delete(stub, args)
    } else if function == "update" { return cc.update(stub, args)
    } else if function == "history" { return cc.history(stub, args)
    } else if function == "querystring" { return cc.querystring(stub, args)
    }

	return shim.Error("Invalid invoke function name")
}

func main() {
    var _ = strconv.FormatInt(1234, 10)
    var _ = time.Now()
    var _ = strings.ToUpper("test")
    var _ = bytes.ToUpper([]byte("test"))

	err := shim.Start(new(DRLDocsChaincode))
	if err != nil {
		fmt.Printf("Error starting BioMetric chaincode: %s", err)
	}
}
