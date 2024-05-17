/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2020-03-26 13:23 
# @File : demo_chaincode.go
# @Description : 
# @Attention : 
*/
package main

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
	"strings"
	"time"
)

type DemoChainCode struct {
}

func (this *DemoChainCode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("init")
	return shim.Success(nil)
}

func (this *DemoChainCode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	args := stub.GetStringArgs()
	s, args := stub.GetFunctionAndParameters()
	strs:=""
	for _,v:=range args{
	    strs+=v+","
	}
	fmt.Printf("参数为:%s \n",strs)
	s = strings.ToLower(s)
	key, value := "key", "joker"
	if len(args)==1{
	    key=args[0]
	}
	if len(args) >= 2 {
		key = args[0]
		value = args[1]
	}

	var result []byte

	var e error
	switch s {
	case "setvalue":
		fmt.Printf(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")
    	fmt.Printf("当前时间:%s ,方法= %s  key=%s,value=%s ,\n", time.Now().String(),s, key, value)
    	fmt.Printf("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\n\n")
		e = stub.PutState(key, []byte(value))
		result=[]byte("ok")
	case "getvalue":
	    fmt.Printf(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")
        fmt.Printf("当前时间:%s ,方法= %s  key=%s,\n", time.Now().String(),s, key)
        fmt.Printf("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\n\n")
		bytes, err := stub.GetState(key)
		if nil != err {
			e = err
		} else if len(bytes) != 0 {
			result = bytes
		}else{
		    result=[]byte("123")
		 }
	}
	if nil != e {
		return shim.Error(e.Error())
	}

	return shim.Success([]byte(result))
}

func main() {
	u := new(DemoChainCode)
	if err := shim.Start(u); nil != err {
		panic(err)
	}
}
