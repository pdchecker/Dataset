package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"strconv"
	"strings"
)

/*
 * 企业账户积分合约实现：
 * 1. 登录数据平台账户基本信息管理
 * 2. 账户积分管理
 */

type Account struct {
	Name     string `json:"name"`     /*账户名称*/
	Password string `json:"password"` /*账户基本信息*/
	Type     int    `json:"type"`     /*账户类别：企业、政府*/
	OrgName  string `json:"orgName"`  /*企业或组织名称*/
	Address  string `json:"address"`  /*账户地址*/
	Frozen   bool   `json:"-"`        /*账户停用标记*/
	Token    int64  `json:"token"`    /*账户积分*/
}

func (a *Account) toBytes() []byte {
	dataAsBytes, _ := json.Marshal(a)
	return dataAsBytes
}

func (a *Account) toString() string {
	dataAsBytes, _ := json.Marshal(a)
	return string(dataAsBytes)
}

func GetAccountCompositeKey(stub shim.ChaincodeStubInterface, name string) (string, string) {
	indexName := "account"
	indexKey, err := stub.CreateCompositeKey(indexName, []string{name})
	if err != nil {
		fmt.Printf("GetAccountCompositeKey error: %s \n", err.Error())
	}
	return indexKey, indexName
}

func NewAccount(info Account) Account {
	hashUtils := DefaultHashUtil()
	storePassword := hashUtils.secret(info.Password)
	return Account{
		Name:     info.Name,
		Password: storePassword,
		Type:     info.Type,
		OrgName:  info.OrgName,
		Address:  info.Address,
		Frozen:   false,
		Token:    0,
	}
}

func (a *Account) balance() int64 {
	return a.Token
}

func (a *Account) transfer(_to *Account, _value int64) ([]byte, bool) {

	if a.Frozen {
		msg := fmt.Sprintf("账户 %s 已冻结", a.Name)
		return []byte(msg), false
	}
	if _to.Frozen {
		msg := fmt.Sprintf("账户 %s 已冻结", _to.Name)
		return []byte(msg), false
	}

	// 支持积分透支，取消持有积分判断
	// if account.Token >= _value {
	a.Token -= _value
	_to.Token += _value
	msg := fmt.Sprintf("账户 %s 往账户 %s 转账 %d 成功", a.Name, _to.Name, _value)
	return []byte(msg), true
	// } else {
	// 	msg := fmt.Sprintf("账户 %s 余额不足, 当前 %d, 需要支付 %d", account.Name, account.BalanceOf, _value)
	// 	return []byte(msg), false
	// }
}

func GetAccount(stub shim.ChaincodeStubInterface, name string) (*Account, error) {
	accountKey, _ := GetAccountCompositeKey(stub, name)
	keyAsBytes, _ := stub.GetState(accountKey)
	if keyAsBytes != nil {
		account := Account{}
		if err := json.Unmarshal(keyAsBytes, &account); err != nil {
			fmt.Printf("[getAccount] Failed to Unmarshal json %s \n", string(keyAsBytes))
		}
		return &account, nil
	}
	return nil, fmt.Errorf("can't find ok by name %s", name)
}

type AccountContract struct {
}

func (s *AccountContract) createAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("[createAccount] Incorrect number of arguments. Expecting 1")
	}

	var reqAccounts []Account
	if err := json.Unmarshal([]byte(args[0]), &reqAccounts); err != nil {
		return shim.Error("[createAccount] Incorrect arguments. Expecting a json array string.")
	}

	for _, val := range reqAccounts {
		accountKey, _ := GetAccountCompositeKey(stub, val.Name)
		existAsBytes, err := stub.GetState(accountKey)
		if existAsBytes != nil {
			return shim.Error("Failed to create account, Duplicate key.")
		}

		account := NewAccount(val)
		if err = stub.PutState(accountKey, account.toBytes()); err != nil {
			return shim.Error(err.Error())
		} else {
			fmt.Printf("createAccount - end %s \n", account.toBytes())
		}
	}
	return shim.Success(nil)
}

func (s *AccountContract) showAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var accounts []Account
	if len(args) == 0 {
		// return shim.Error("[showAccount] Incorrect number of arguments. Expecting length more then 0")
		resultIterator, _ := stub.GetStateByPartialCompositeKey("account", args)
		defer resultIterator.Close()
		for resultIterator.HasNext() {
			item, _ := resultIterator.Next()

			account := Account{}
			_ = json.Unmarshal(item.Value, &account)
			accounts = append(accounts, account)
		}
	} else {
		for _, val := range args {
			account, err := GetAccount(stub, val)
			if err != nil {
				return shim.Error(err.Error())
			}
			account.Password = ""
			accounts = append(accounts, *account)
		}
	}

	dataBytes, _ := json.Marshal(accounts)
	return shim.Success(dataBytes)
}

func (s *AccountContract) frozenAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("[frozenAccount] Incorrect number of arguments. Expecting 2")
	}

	_account := args[0]
	_status := args[1]

	account, err := GetAccount(stub, _account)
	if err != nil {
		return shim.Error(err.Error())
	}

	if strings.ToLower(_status) == "true" || _status == "1" {
		account.Frozen = true
	} else {
		account.Frozen = false
	}

	accountAsBytes := account.toBytes()
	err = stub.PutState(_account, accountAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	} else {
		fmt.Printf("frozenAccount - end %s \n", string(accountAsBytes))
	}

	return shim.Success(nil)
}

func (s *AccountContract) deleteAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("[deleteAccount] Incorrect number of arguments. Expecting 1")
	}

	_account := args[0]
	accountKey, _ := GetAccountCompositeKey(stub, _account)

	err := stub.DelState(accountKey)
	if err != nil {
		return shim.Error(err.Error())
	} else {
		fmt.Printf("deleteAccount - end %s \n", _account)
	}

	return shim.Success(nil)
}

func (s *AccountContract) mintToken(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	_account := args[0]
	_amount, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer Value for mint token holding")
	}

	account, err := GetAccount(stub, _account)
	if err != nil {
		return shim.Error(err.Error())
	}
	account.Token += int64(_amount)

	err = stub.PutState(_account, account.toBytes())
	if err != nil {
		return shim.Error(err.Error())
	} else {
		fmt.Printf("Accounter mint token - end %s %d \n", account.Name, account.Token)
	}

	token := AccountTokenResponse{Name: account.Name, Token: account.Token}
	tokenAsBytes := token.toBytes()

	return shim.Success(tokenAsBytes)
}

func (s *AccountContract) changeSecret(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	_account := args[0]
	_secret := args[1]

	if account, err := GetAccount(stub, _account); err == nil {
		hashUtil := DefaultHashUtil()
		account.Password = hashUtil.secret(_secret)
		if err1 := stub.PutState(_account, account.toBytes()); err1 != nil {
			return shim.Error(err1.Error())
		}
	} else {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}
