/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2020-06-16 14:05 
# @File : args.go
# @Description : 
# @Attention : 
*/
package main

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	error4 "myLibrary/go-library/chaincode/error"
	"myLibrary/go-library/common/blockchain/base"
	error3 "myLibrary/go-library/common/error"
	"myLibrary/go-library/go/authentication"
	"vlink.com/v2/vlink-common/base/fabric"
)

var (
	COMMON_STRING_OR_STRING_ARRAY_KEY_GENERATOR = func(stub shim.ChaincodeStubInterface, objectType base.ObjectType, param ...interface{}) (string, error3.IBaseError) {
		switch param[0].(type) {
		case string:
			return COMMON_STRING_KEY_GENERATOR(stub, objectType, param[0])
		case []string:
			return COMMON_STRING_ARRAY_KEY_GENERATOR(stub, objectType, param[0])
		}
		return "", error4.NewChainCodeError(nil, "找不到匹配的处理")
	}
	COMMON_STRING_ARRAY_KEY_GENERATOR = func(stub shim.ChaincodeStubInterface, objectType base.ObjectType, param ...interface{}) (string, error3.IBaseError) {
		strings := param[0].([]string)
		s, e := stub.CreateCompositeKey(string(objectType), strings)
		if nil != e {
			return "", error4.NewChainCodeError(e, "创建组合键失败")
		}
		return s, nil
	}

	COMMON_STRING_KEY_GENERATOR = func(stub shim.ChaincodeStubInterface, objectType base.ObjectType, param ...interface{}) (string, error3.IBaseError) {
		strings := param[0].(string)
		s, e := stub.CreateCompositeKey(string(objectType), []string{strings})
		if nil != e {
			return "", error4.NewChainCodeError(e, "创建组合键失败")
		}
		return s, nil
	}
)

type ArgsChecker = func(args []string) error3.IBaseError
type ArgsConverter = func(args []string) (interface{}, error3.IBaseError)
type ArgsDecrypter = func(data interface{}, version string) (interface{}, error3.IBaseError)

type ArgsParameter struct {
	ArgsChecker   ArgsChecker
	ArgsConverter ArgsConverter
}

type TransBaseDescription struct {
	TransBaseType base.TransBaseTypeV2
	Description   string
}

func (this TransBaseDescription) String() string {
	return "{[ baseType=" + this.TransBaseType.String() + " ],[Description=" + this.Description + " ]"
}

func NewNeedRecordTransBaseDescription(baseValue base.TransBaseTypeV2Value, desc string) TransBaseDescription {
	description := TransBaseDescription{
		Description: desc,
	}
	description.TransBaseType = base.CreateNeedRecordBaseType(baseValue)
	return description
}

func NewUnRecordTransBaseDescription(baseValue base.TransBaseTypeV2Value, desc string) TransBaseDescription {
	description := TransBaseDescription{
		Description: desc,
	}
	description.TransBaseType = base.CreateUnNeedRecordBaseType(baseValue)
	return description
}

func ConvBytes2TransBaseTypeV2(bytes []byte) cc.TransBaseTypeV2 {
	authorities, _ := authentication.BigEndianConvtBytes2Authority(bytes)
	return cc.TransBaseTypeV2(authorities)
}

var (
	DefaultNumberChecker = func(args []string) error3.IBaseError {
		if len(args) < 1 {
			return error3.NewArguError(nil, "参数长度必须大于1")
		}
		return nil
	}
)

func NewDefaultCheckerParameter(ArgsConverter ArgsConverter) ArgsParameter {
	a := ArgsParameter{}
	a.ArgsChecker = DefaultNumberChecker
	a.ArgsConverter = ArgsConverter

	return a

}
