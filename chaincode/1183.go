/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2019-12-14 13:05 
# @File : fabric_base_facaded_service.go
# @Description : 门面service的baseCC
# @Attention : 
*/
package main

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"myLibrary/go-library/common/blockchain"
	"myLibrary/go-library/common/blockchain/base"
	error2 "myLibrary/go-library/common/error"
	"myLibrary/go-library/go/base/service"
	"myLibrary/go-library/go/constants"
)

type ArgeChecker func(name base.MethodName) (interface{},  error2.IBaseError)

type BaseTypeGetter func(name base.MethodName) (base.TransBaseType, error2.IBaseError)

type IFacadedHandler interface {

	Handler()BasePeerResponse
}

type IConcreteFacadedService interface {
	HandleDetail(name base.MethodName, req blockchain.BaseFabricAfterValidModel) (ITxBaseResper, error2.IBaseError)
	SecurityCheckAndConvt(name base.MethodName, args []string) (blockchain.BaseFabricAfterValidModel, error2.IBaseError)
	// 获取管道ID,不同的channel 拥有着不同的账本,因此查询交易的时候,key也是不同的
	// 2020-01-05 update  需要修改返回值为[]string,存在一条链码部署在多个channel的可能
	GetChannelID() string
}

type BaseFabricFacadedCC struct {
	Stub shim.ChaincodeStubInterface
	*service.BaseServiceImpl
	ConcreteFacadedService IConcreteFacadedService
}

// 2019-12-17 暂时直接返回这个BasePeerResponse,如果后续需要修改,直接修改这里即可
func (b *BaseFabricFacadedCC) Handler() BasePeerResponse {
	b.BeforeStart("Handler")
	defer b.AfterEnd()

	funcs, args := b.Stub.GetFunctionAndParameters()
	methodName := base.MethodName(funcs)
	b.Debug("开始参数校验")
	reqModel, baseError := b.ConcreteFacadedService.SecurityCheckAndConvt(methodName, args)
	if nil != baseError {
		b.Error("校验参数错误:%s", baseError.Error())
		return b.ReturnResult(nil, baseError)
	}
	b.Debug("结束参数校验")

	b.Debug("开始处理请求详情")
	detail, e := b.ConcreteFacadedService.HandleDetail(methodName, reqModel)
	if nil != e {
		b.Error("执行methodName=[%s]发生错误:%s", methodName, e.Error())
		return b.ReturnResult(nil, e)
	}
	detail.SetTxBaseType(reqModel.BaseTransType)
	detail.SetTxDescription(reqModel.BaseTransDescription)
	detail.SetTransactionID(b.Stub.GetTxID())
	detail.SetChannelID(base.ChannelID(b.ConcreteFacadedService.GetChannelID()))
	detail.InfoFix()

	b.Debug("结束处理请求详情")

	// b.Debug("判断是否需要对交易记录")
	// // 1. 先判断是否是成功的业务
	// if detail.GetCode()&constants.SUCCESS > 0 {
	// 	b.Debug("业务执行成功,判断是否需要记录交易")
	// 	if reqModel.BaseTransType&base.TX_BASE_NEED_RECORD >= base.TX_BASE_NEED_RECORD {
	// 		// 记录交易
	// 		b.Debug("begin 记录本次交易")
	// 		// 可以交给子类实现,也可以本父类实现,图省时,直接父类统一实现
	//
	// 		b.Debug("end 记录本次交易")
	// 	}
	// } else {
	// 	b.Debug("业务执行失败,失败原因:{%s}", detail.GetMsg())
	// }
	// "from":"12jEeDUaMwkrefwcS6AaBzSsFt3xiATcP3","to":"","token":0,"Data":{"dna":"0a5fb72c-f99f-4897-a05d-336923073749","prvKey":"","coinAddress":"1NzvZtHuuDVnyX56j5HaPWBZzrYDTUvrka"},"Code":1,"Msg":"SUCCESS"}
	// 返回结果
	b.Debug("SUCCESS 成功调用[%s],返回值为:[%v]", methodName, detail)

	return b.ReturnResult(detail, nil)
}

func (b *BaseFabricFacadedCC) ReturnResult(res interface{}, err error2.IBaseError) BasePeerResponse {
	if nil != err {
		return Fail(err)
	}

	if nil == res {
		return SuccessPeerResponse(SuccessWithEmptyData())
	}

	switch res.(type) {
	case ITxBaseResper:
		t := res.(ITxBaseResper)
		// d := t.GetReturnData()

		var (
		// dataBytes []byte
		// logBytes  []byte
		// e         error
		)

		// 1. 先判断是否是成功的业务
		if t.GetCode()&constants.SUCCESS > 0 {
			b.Debug("业务执行成功,判断是否需要记录交易")
			// if d != nil {
			// 	dataBytes, e = json.Marshal(d)
			// 	if nil != e {
			// 		return base.Fail(error3.NewJSONSerializeError(e, "序列化失败"))
			// 	}
			// }
			// b.Debug("判断是否需要对交易记录")
			// logInfos := t.GetTXRecordInfoList()
			// logBytes, e = json.Marshal(logInfos)
			// if nil != e {
			// 	return base.Fail(error3.NewJSONSerializeError(e, "序列化日志记录失败"))
			// }

			// if logInfo.BaseType.Contains(base.TX_BASE_NEED_RECORD) {
			// 	fmt.Println("begin basetype")
			// 	for _, c := range logInfo.BaseType {
			// 		fmt.Println(c)
			// 	}
			// 	fmt.Println("end basetype")
			// 	b.Debug("该提案的tx为:{%s},需要进行记录数据", logInfo.BaseType.String())
			// 	logBytes, e = json.Marshal(logInfo)
			// 	if nil != e {
			// 		return base.Fail(error3.NewJSONSerializeError(e, "序列化失败"))
			// 	}
			// }
		} else {
			b.Debug("业务执行失败,失败原因:{%s}", t.GetMsg())
		}
		transfer := TempTransfer{
			Code:                    t.GetCode(),
			Msg:                     t.GetMsg(),
			TxRecords:               t.GetTXRecordInfoList(),
			ReturnData:              t.GetReturnData(),
			BaseRespCommonAttribute: t.GetCommAttribute(),
		}
		return SuccessWithDetailTransfer(transfer)
		// return base.SuccessWithDetail(dataBytes, logBytes, t.GetCode(), t.GetMsg())
	default:
		// 必须实现该接口,但是不会遇到,因为都有一个外层封装类
		return Fail(error2.NewConfigError(nil, "必须实现IVlinkTxBaseResper接口"))
	}
}

// func (b *BaseFabricFacadedCC) ConfigArgChecker(methods []MethodName, params []ArgsParameter) {
// 	l := len(methods)
// 	for i := 0; i < l; i++ {
// 		b.AddCheck(methods[i], &params[i])
// 	}
// }

// func (b *BaseFabricFacadedCC) ConfigLogicDesc(methods []MethodName, descs []constants.TransBaseType) {
// 	l := len(methods)
// 	for i := 0; i < l; i++ {
// 		b.AddLogicDesc(methods[i], descs[i])
// 	}
// }

// func NewBaseFabricFacadedCC(ac ArgeChecker,b BaseTypeGetter) *BaseFabricFacadedCC {
// 	c := new(BaseFabricFacadedCC)
// 	c.Log = log.NewVlinkLog()
// 	// c.ArgumentDecrypt = decrypt
// 	// decrypt.SetParent(c)
// 	// c.CheckerAndDecrypter=ac
// 	// c.BaseTyperGetter=b
//
// 	return c
// }
func NewBaseFabricFacadedCC(stub shim.ChaincodeStubInterface, ConcreteFacadedService IConcreteFacadedService) *BaseFabricFacadedCC {
	c := new(BaseFabricFacadedCC)
	c.BaseServiceImpl = service.NewBaseServiceImplWithLog4goLogger()
	c.Stub = stub
	c.ConcreteFacadedService = ConcreteFacadedService

	return c
}
