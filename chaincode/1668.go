package main

// atoi 메서드 util화 시키기
import (
	"encoding/json"
	"fmt"
	"hyperledger_dapp/model"
	"hyperledger_dapp/repository"
	"hyperledger_dapp/util"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

// Transfer is invoke fnc that moves amount token
// from the caller's address to recipient
// params - caller's address, recipient's address, amount of token
func (cc *Controller) Transfer(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	// check parameter
	if len(params) != 3 {
		shim.Error("Transfer only 3 params")
	}

	callerAddress, recipientAddress, transferAmount := params[0], params[1], params[2]
	transferAmountInt, err := util.ConvertToPositive("transfer amount", transferAmount)
	if err != nil {
		return shim.Error(err.Error())
	}

	// get caller amount
	callerAmount, err := repository.GetBalance(stub, callerAddress, false)
	if err != nil {
		return shim.Error(err.Error())
	}

	// get recipient amount
	recipientAmount, err := repository.GetBalance(stub, callerAddress, false)
	if err != nil {
		return shim.Error(err.Error())
	}

	// calculate amount
	callerResultAmount := *callerAmount - *transferAmountInt
	recipientResultAmount := *recipientAmount + *transferAmountInt

	// check calculate amount is positive
	if callerResultAmount < 0 {
		return shim.Error("caller's balance is not sufficient")
	}

	// save the caller & recipient amount
	err = repository.SaveBalance(stub, callerAddress, strconv.Itoa(callerResultAmount))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = repository.SaveBalance(stub, callerAddress, strconv.Itoa(recipientResultAmount))
	if err != nil {
		return shim.Error(err.Error())
	}

	// emit transfer event
	transferEvent := model.NewTransferEvent(callerAddress, recipientAddress, *transferAmountInt)

	transferEventBytes, err := json.Marshal(transferEvent)
	if err != nil {
		return shim.Error("failed to marshal transferEvent, error :" + err.Error())
	}

	err = stub.SetEvent("transferEvent", transferEventBytes)
	if err != nil {
		return shim.Error("failed to SetEvent of transferEvent, error :" + err.Error())
	}

	fmt.Println(callerAddress + " send " + transferAmount + " to " + recipientAddress)

	return shim.Success([]byte("transfer Success"))
}

// Approve is invoke fnc that sets amount as the allowance
// of spender over the owner tokens
// params - owner's address, spender's address, amount of token
func (cc *Controller) Approve(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	// check the number of params is 3
	if len(params) != 3 {
		return shim.Error("Approve only 3 params")
	}

	ownerAddress, spenderAddress, allowanceAmount := params[0], params[1], params[2]

	// check amount is integer & positive
	// allowanceAmountInt, err := strconv.Atoi(allowanceAmount)
	allowanceAmountInt, err := util.ConvertToPositive(" Amount int ", allowanceAmount)
	if err != nil {
		return shim.Error(err.Error())
	}

	// create composite key for allowance - approval/{owner}/{spender}
	approvalKey, err := stub.CreateCompositeKey("approval", []string{ownerAddress, spenderAddress})
	if err != nil {
		return shim.Error("failed to create CompositeKey for approval")
	}

	// save allowance amount
	err = stub.PutState(approvalKey, []byte(allowanceAmount))
	if err != nil {
		return shim.Error("failed to PutState for approvalKey")
	}

	// emit approval event

	approvalEvent := model.NewApproval(spenderAddress, ownerAddress, *allowanceAmountInt)
	approvalBytes, err := json.Marshal(approvalEvent)
	if err != nil {
		return shim.Error("failed to Marshal for approvalEvent")
	}

	err = stub.SetEvent("approvalEvent", approvalBytes)
	if err != nil {
		return shim.Error("failed to marshal for approval SetEvent")
	}

	return shim.Success([]byte("approve success"))
}

// TransferFrom is invoke fnc that moves amount of token from sender to recipient
// using allowance of sender
// params - owner's address, spender's address, recipient's address, amount of token
func (cc *Controller) TransferFrom(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 4 {
		return shim.Error("transferFrom only 4 params")
	}

	ownerAddress, spenderAddress, recipientAddress, transferAmount := params[0], params[1], params[2], params[3]

	// check amount is integer & positive
	transferAmountInt, err := util.ConvertToPositive(" Amount ", transferAmount)
	if err != nil {
		return shim.Error(err.Error())
	}

	// get allowance
	allowanceResponse := cc.Allowance(stub, []string{ownerAddress, spenderAddress})

	if allowanceResponse.GetStatus() >= 400 {
		return shim.Error("failed to get allowance, error : " + allowanceResponse.GetMessage())
	}

	// convert allowance response payload to allowance data

	allowanceInt, err := util.ConvertToPositive("Payload", string(allowanceResponse.GetPayload()))
	if err != nil {
		return shim.Error(err.Error())
	}

	// transfer from owner to recipient
	transferResponse := cc.Transfer(stub, []string{ownerAddress, recipientAddress, transferAmount})
	if transferResponse.GetStatus() >= 400 {
		return shim.Error("failed to get transfer error : " + err.Error())
	}

	// decrease allowance amount
	approveAmountInt := *allowanceInt - *transferAmountInt
	approveAmount := strconv.Itoa(approveAmountInt)

	// approve amount of tokens trasfered
	approveResponse := cc.Approve(stub, []string{ownerAddress, spenderAddress, approveAmount})
	if approveResponse.GetStatus() >= 400 {
		return shim.Error("failed to get transfer error : " + err.Error())
	}

	return shim.Success([]byte("transferFrom success"))
}

// IncreaseAllowance is invoke fnc that increases spender's allowance by owner
// params - owner's address, spender's addresss, amount of token
func (cc *Controller) IncreaseAllowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	// check params 3
	if len(params) != 3 {
		return shim.Error("increaseAllowance only 3 params")
	}

	ownerAddress, spenderAddress, increaseAmount := params[0], params[1], params[2]

	// check amount is integer & positive
	increaseAmountInt, err := util.ConvertToPositive("Amount", increaseAmount)
	if err != nil {
		return shim.Error(err.Error())
	}

	// get allowance
	allowanceResponse := cc.Allowance(stub, []string{ownerAddress, spenderAddress})
	if allowanceResponse.GetStatus() >= 400 {
		return shim.Error("failed to get allowance, error : " + allowanceResponse.GetMessage())
	}

	// convert allowance response payload to allowance data
	allowanceInt, err := util.ConvertToPositive("allowance", string(allowanceResponse.GetPayload()))
	if err != nil {
		return shim.Error(err.Error())
	}

	// increase allowance
	resultAmountInt := *allowanceInt + *increaseAmountInt
	resultAmount := strconv.Itoa(resultAmountInt)

	// call approve
	approveResponse := cc.Approve(stub, []string{ownerAddress, spenderAddress, resultAmount})
	if approveResponse.GetStatus() >= 400 {
		return shim.Error("failed to get allowance, error : " + approveResponse.GetMessage())
	}

	return shim.Success([]byte("increaseAllowance success"))
}

// DecreaseAllowance is invoke fnc that decreases spender's allowance by owner
// params - owner's address, spender's addresss, amount of token
func (cc *Controller) DecreaseAllowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	// check params 3
	if len(params) != 3 {
		return shim.Error("DecreaseAllowance only 3 params")
	}

	ownerAddress, spenderAddress, decreaseAmount := params[0], params[1], params[2]

	// check amount is integer & positive

	decreaseAmountInt, err := util.ConvertToPositive("Amount", decreaseAmount)
	if err != nil {
		return shim.Error(err.Error())
	}

	// get allowance
	allowanceResponse := cc.Allowance(stub, []string{ownerAddress, spenderAddress})
	if allowanceResponse.GetStatus() >= 400 {
		return shim.Error("failed to get allowance, error : " + allowanceResponse.GetMessage())
	}

	// convert allowance response payload to allowance data
	allowanceInt, err := util.ConvertToPositive("allowance", decreaseAmount)
	if err != nil {
		return shim.Error("allowance must be positive")
	}

	// decrease allowance
	resultAmountInt := *allowanceInt + *decreaseAmountInt
	resultAmount := strconv.Itoa(resultAmountInt)

	// call approve
	approveResponse := cc.Approve(stub, []string{ownerAddress, spenderAddress, resultAmount})
	if approveResponse.GetStatus() >= 400 {
		return shim.Error("failed to get allowance, error : " + approveResponse.GetMessage())
	}

	return shim.Success([]byte("decreaseAllowance success"))
}

// TransferOtherToken is invoke fnc that moves amount other chaincode tokens
// from the caller's addresss to recipient
// params - chaincode name, caller's addresss, recipient's address, amount
func (cc *Controller) TransferOtherToken(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 4 {
		return shim.Error("transferFrom only 4 params")
	}

	chaincodeName, callerAddress, recipientAddress, transferAmount := params[0], params[1], params[2], params[3]

	// make arguments
	args := [][]byte{[]byte("transfer"), []byte(callerAddress), []byte(recipientAddress), []byte(transferAmount)}

	// get channel
	channel := stub.GetChannelID()

	// transfer other chaincode token
	transferResponse := stub.InvokeChaincode(chaincodeName, args, channel)

	if transferResponse.GetStatus() >= 400 {
		errMsg := fmt.Sprintf("failed to transfer %s, error :%s ", chaincodeName, transferResponse.GetMessage())
		return shim.Error(errMsg)
	}

	return shim.Success([]byte("transfer other token success"))
}

// mint is invoke fnc that creates amount tokens and assign them to address, increasing the total supply
// param - token name, recipient address, amount token
func (cc *Controller) Mint(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	// chk parameter
	if len(params) != 3 {
		return shim.Error("Mint only 3 params")
	}
	tokenName, owner, mintAmount := params[0], params[1], params[2]

	// amount must be positive
	mintAmountInt, err := util.ConvertToPositive("mint amount", mintAmount)

	if err != nil {
		return shim.Error(err.Error())
	}

	// increase total supply
	erc20Metadata, err := repository.GetERC20Metadata(stub, tokenName)
	if err != nil {
		return shim.Error(err.Error())
	}

	resultTotalSupply := *erc20Metadata.GetTotalSupply() + uint64(*mintAmountInt)

	err = repository.SaveERC20Metadata(stub, erc20Metadata.Name, erc20Metadata.Symbol, erc20Metadata.Owner, uint(resultTotalSupply))
	if err != nil {
		return shim.Error(err.Error())
	}

	// increase owner balance
	curBalance, err := repository.GetBalance(stub, owner, false)
	if err != nil {
		return shim.Error(err.Error())
	}

	resultBalance := *curBalance + *mintAmountInt

	err = repository.SaveBalance(stub, owner, strconv.Itoa(resultBalance))
	if err != nil {
		return shim.Error(err.Error())
	}

	// emit transfer event
	err = repository.EmitTransferEvent(stub, "admin", tokenName, *mintAmountInt)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("mint success"))
}

func (cc *Controller) Burn(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	return shim.Success(nil)
}
