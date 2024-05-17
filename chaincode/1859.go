    
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

type passiveCCDRChaincode struct {
}

type passiveCCDR struct {
    Id string `json:"Id"`
    CreatedOn time.Time `json:"CreatedOn"`
    CreatedBy string `json:"CreatedBy"`
    IsDelete bool `json:"IsDelete"`
    DPRNo string `json:"DPRNo"`
    Department string `json:"Department"`
    PickListNumber string `json:"PickListNumber"`
    Unit string `json:"Unit"`
    GeneralInstruction string `json:"GeneralInstruction"`
    PackagingOperation string `json:"PackagingOperation"`
    InnerBoxPacking string `json:"InnerBoxPacking"`
    OuterBoxPacking string `json:"OuterBoxPacking"`
    ShipmentTracking string `json:"ShipmentTracking"`
    DocumentVerification string `json:"DocumentVerification"`
    EffectiveDate string `json:"EffectiveDate"`
    Destination string `json:"Destination"`
    Misc string `json:"Misc"`
    Notes string `json:"Notes"`
}

func (cc *passiveCCDRChaincode) create(stub shim.ChaincodeStubInterface, arg []string) peer.Response {
 
    args := strings.Split(arg[0], "^^")

    if len(args) != 18 {
        return shim.Error("Incorrect number arguments. Expecting 18")
    }
	dateValue1, err1 := time.Parse(time.RFC3339, args[1])

	if err1 != nil {
		return shim.Error("Error converting string to date: " + err1.Error())
	}
	boolValue3, err3 := strconv.ParseBool(args[3])

	if err3  != nil {
		return shim.Error("Error converting string to bool: " + err3.Error())
	}
    data := passiveCCDR{ Id: args[0], CreatedOn: dateValue1, CreatedBy: args[2], IsDelete: boolValue3, DPRNo: args[4], Department: args[5], PickListNumber: args[6], Unit: args[7], GeneralInstruction: args[8], PackagingOperation: args[9], InnerBoxPacking: args[10], OuterBoxPacking: args[11], ShipmentTracking: args[12], DocumentVerification: args[13], EffectiveDate: args[14], Destination: args[15], Misc: args[16], Notes: args[17] }

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

func (cc *passiveCCDRChaincode) get(stub shim.ChaincodeStubInterface, args []string) peer.Response {

    if len(args) != 1 {
        return shim.Error("Incorrect number arguments. Expecting 1")
    }

    stateBytes, err := stub.GetState(args[0])

    if err != nil {
        return shim.Error("Error getting the state: " + err.Error())
    }

    return shim.Success(stateBytes)
}
func (cc *passiveCCDRChaincode) delete(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number arguments. Expecting 1")
	}

	dataBytes, err := stub.GetState(args[0])

	if err != nil {
		return shim.Error("Error getting the state: " + err.Error())
	}

	data := passiveCCDR{}

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
func (cc *passiveCCDRChaincode) update(stub shim.ChaincodeStubInterface, arg []string) peer.Response {
    
    args := strings.Split(arg[0], "^^")     

	if len(args) != 18 {
		return shim.Error("Incorrect number arguments. Expecting 18")
	}
	dateValue1, err1 := time.Parse(time.RFC3339, args[1])

	if err1 != nil {
		return shim.Error("Error converting string to date: " + err1.Error())
	}
	boolValue3, err3 := strconv.ParseBool(args[3])

	if err3  != nil {
		return shim.Error("Error converting string to bool: " + err3.Error())
	}
    data := passiveCCDR{ Id: args[0], CreatedOn: dateValue1, CreatedBy: args[2], IsDelete: boolValue3, DPRNo: args[4], Department: args[5], PickListNumber: args[6], Unit: args[7], GeneralInstruction: args[8], PackagingOperation: args[9], InnerBoxPacking: args[10], OuterBoxPacking: args[11], ShipmentTracking: args[12], DocumentVerification: args[13], EffectiveDate: args[14], Destination: args[15], Misc: args[16], Notes: args[17] }

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

func (cc *passiveCCDRChaincode) history(stub shim.ChaincodeStubInterface, args []string) peer.Response {

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

func (cc *passiveCCDRChaincode) querystring(stub shim.ChaincodeStubInterface, args []string) peer.Response {

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
func (cc *passiveCCDRChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (cc *passiveCCDRChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

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

	err := shim.Start(new(passiveCCDRChaincode))
	if err != nil {
		fmt.Printf("Error starting BioMetric chaincode: %s", err)
	}
}
