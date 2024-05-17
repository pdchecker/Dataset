package main

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/zhj0811/dbzl/chaincode/handler"
	"github.com/zhj0811/dbzl/common/define"
)

var logger = flogging.MustGetLogger("chaincode")

type handlerFunc func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error)

var funcHandler = map[string]handlerFunc{
	define.SaveService: handler.SaveService,
	define.SavePolicy:  handler.SavePolicy,
	define.SaveCompany: handler.SaveCompany,
	define.QueryByTxID: handler.QueryByTxID,
	define.QueryByKey:  handler.QueryByKey,
}

//Chaincode cc对象用于调研各类fabric接口函数
type Chaincode struct {
}

func init() {
}

//Init 初始化或升级cc
func (t *Chaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

//Invoke 执行invoke或者query操作
func (t *Chaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	logger.Debugf("Invoke function=%v,args=%v\n", function, args)

	if len(args) < 1 || len(args[0]) == 0 {
		logger.Error("the invoke args not exist or arg[0] is empty")
		return shim.Error("the invoke args not exist  or arg[0] is empty")
	}

	currentFunc := funcHandler[function]
	if currentFunc == nil {
		logger.Error("the function name not exist!!")
		return shim.Error("the function name not exist!!")
	}

	payload, err := currentFunc(stub, function, args)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(payload)

}

func main() {
	err := shim.Start(new(Chaincode))
	fmt.Println("err333:", err)
	if err != nil {
		logger.Errorf("Error starting chaincode: %s", err)
	}
}
