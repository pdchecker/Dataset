package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// Structure to implement chaincode interface
type Reigocode struct {
	ID             string `json:"ID"`
	CalculatedData string `json:"CalculatedData"`
}

// Initializes chaincode
func (ccode *Reigocode) Init(stub shim.ChaincodeStubInterface) peer.Response {

	return shim.Success(nil)
}

func calcAvailableFund(avaiFund1 float64, arr []float64) float64 {
	amtPaidSum := 0.00
	for i := 0; i < len(arr); i++ {
		amtPaidSum += arr[i]

	}
	result := avaiFund1 - amtPaidSum

	return result
}

func calcAmountPaid(avaiFund float64, amtOwed float64, adjustment string) float64 {
	var amtPaid float64
	if adjustment != "" {
		val, _ := strconv.ParseFloat(adjustment, 64)
		amtPaid = math.Min(avaiFund, math.Min(amtOwed, val))
	} else {
		amtPaid = math.Min(avaiFund, amtOwed)
	}
	return amtPaid
}

func calcAdjustmentType(adjustment string) string {
	if adjustment != "" {
		return "CAP"
	} else {
		return ""
	}
}

func calcColumnL(j float64, irate float64, interestOwed float64) float64 {
	if irate != 0.00 {
		return ((j / 100) - (irate / 100)) * interestOwed / (irate / 100)
	} else {
		return 0.00
	}
}
func Date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
func (ccode *Reigocode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()
	var result string

	var err error

	if fn == "Generatetable" {

		result, err = Generatetable(stub, args)
	}
	if fn == "GetTable" {

		result, err = GetTable(stub, args)
	}
	if err != nil {

		return shim.Error(err.Error())
	}

	return shim.Success([]byte(result))

}

//Servicer Data Sheet
type ServicerReport struct {
	SRID           string `json:"srId"`
	SRKey          string `json:"srKey"`
	SRValue        string `json:"srValue"`
	SRMonth        string `json:"srMonth"`
	SRYear         string `json:"srYear"`
	SRUpdatedBy    string `json:"srUpdatedBy"`
	SRUpdationDate string `json:"srUpdationDate"`
	SRSeqNum       string `json:"srSeqNum"`
	SRDealID       string `json:"srDealId"`
}

//additional tab
type Additional struct {
	Key    string `json:"Key"`
	Value1 string `json:"Value1"`
	Value2 string `json:"Value2"`
	Value3 string `json:"Value3"`
	Value4 string `json:"Value4"`
}

//push data to buffer for additional tab table
func Status(current string, Fcolumn string) string {
	if current == "" {
		return "N/A"
	} else {
		if Fcolumn == "TRUE" {
			return "PASS"
		} else {
			return "FAIL"
		}
	}
}
func pushbuffer(key string, current string, limit string, limittype string, status string) string {
	var buffer1 bytes.Buffer
	additional := Additional{Key: key, Value1: current, Value2: limit, Value3: limittype, Value4: status}
	add1, _ := json.Marshal(additional)
	buffer1.WriteString(string(add1))
	buffer1.WriteString(",")
	return buffer1.String()
}

//table
func Generatetable(stub shim.ChaincodeStubInterface, args []string) (string, error) {

	jsonstring := args[0] // monthly ip

	var m []ServicerReport
	if err := json.Unmarshal([]byte(jsonstring), &m); err != nil {
		panic(err)
	}
	//fmt.Println(m)
	var key string
	var value string
	var smap = make(map[string]string)
	var monthval, yearval string

	for loop := 0; loop < len(m); loop++ {
		ServicerReportStruct := ServicerReport{
			SRID:           m[loop].SRID,
			SRKey:          m[loop].SRKey,
			SRValue:        m[loop].SRValue,
			SRMonth:        m[loop].SRMonth,
			SRYear:         m[loop].SRYear,
			SRUpdatedBy:    m[loop].SRUpdatedBy,
			SRUpdationDate: m[loop].SRUpdationDate,
			SRSeqNum:       m[loop].SRSeqNum,
			SRDealID:       m[loop].SRDealID,
		}
		fmt.Println(ServicerReportStruct)

		key = ServicerReportStruct.SRKey
		value = ServicerReportStruct.SRValue

		smap[key] = value
		monthval = ServicerReportStruct.SRMonth
		yearval = ServicerReportStruct.SRYear

	}
	//actial collateral balance
	type CollateralBal struct {
		Key    string `json:"key"`
		Value1 string `json:"value1"`
		Value2 string `json:"value2"`
	}

	//PRINCIPAL REMITTANCE
	type PrincipalRemittance struct {
		PrincipalRemittance string `json:"PRINCIPAL REMITTANCE"`
		BeginningBalance    string `json:"BeginningBalance"`
		Deposits            string `json:"Deposits"`
		//	PaidInFull          string `json:"PaidInFull"`
		//	Sale                string `json:"Sale"`
		//	Liquidation         string `json:"Liquidation"`
		//	Curtailments        string `json:"Curtailments"`
		//	ScheduledPrincipal  string `json:"ScheduledPrincipal"`
		PrincipalPaid string `json:"PrincipalPaid"`
		Other         string `json:"Other(+)"`
		Total         string `json:"Total:"`
		Withdrawals   string `json:"Withdrawals"`
		//	Purchased           string `json:"Purchased"`
		//	Funded              string `json:"Funded"`
		//	CapitalizedInterest string `json:"CapitalizedInterest"`
		Other1 string `json:"Other(-)"`
		//	OtherMinusNoncash   string `json:"Other(-)(non-cash)"`
		ToAvaiFunds   string `json:"ToAvailableFunds (prior period)"`
		PTotal        string `json:"Total: "`
		EndingBalance string `json:"EndingBalance"`
	}

	//INTEREST REMITTANCE
	type INTERESTREMITTANCE struct {
		INTERESTREMITTANCE string `json:"INTEREST REMITTANCE"`
		BeginningBalance   string `json:"BeginningBalance"`
		Deposits           string `json:"Deposits"`
		InterestPaid       string `json:"InterestPaid"`
		FeesPaid           string `json:"FeesPaid"`
		Other              string `json:"Other(+)"`
		Total              string `json:"Total:"`
		Withdrawals        string `json:"Withdrawals"`
		ServicerFees       string `json:"RetainedServicerFees"`
		Other1             string `json:"Other(-)"`
		ToAvaiFunds        string `json:"ToAvailableFunds (prior period)"`
		ITotal             string `json:"Total: "`
		EndingBalance      string `json:"EndingBalance"`
	}

	//AVAILABLEFUNDS
	type AVAILABLEFUNDS struct {
		AVAILABLEFUNDS       string `json:"AVAILABLE FUNDS"`
		BeginningBalance     string `json:"BeginningBalance"`
		Deposits             string `json:"Deposits"`
		PrincipalRemittance  string `json:"PrincipalRemittance"`
		InterestRemittance   string `json:"InterestRemittance"`
		Acc1                 string `json:"From ACCOUNT1: REVOLVING PERIOD REINVESTMENT ACCOUNT"`
		Acc2                 string `json:"From ACCOUNT2: RESERVE ACCOUNT"`
		Acc3                 string `json:"From ACCOUNT3: ASSET MANAGEMENT ACCOUNT"`
		Acct4                string `json:"From ACCOUNT4:"`
		Total                string `json:"Total:"`
		Withdrawals          string `json:"Withdrawals"`
		ToPriorityofPayments string `json:"ToPriorityOfPayments (prior period)"`
		ATotal               string `json:"Total: "`
		EndingBalance        string `json:"EndingBalance"`
	}
	//From ACCOUNT1: REVOLVING PERIOD REINVESTMENT ACCOUNT
	type RevolvingPerReinvAccts struct {
		RevolvingPerReinvAcct     string `json:"REVOLVING PERIOD REINVESTMENT ACCOUNT"`
		BeginBln                  string `json:"BeginningBalance"`
		Deposits                  string `json:"Deposits"`
		FromPriorityofPayments    string `json:"FromPriorityOfPayments (prior period)"`
		DepositTotal              string `json:"Total:"`
		Withdrawals               string `json:"Withdrawals"`
		ToPurchaseAdditionalLoans string `json:"ToPurchaseAdditionalLoans"`
		ToAvaiFunds               string `json:"ToAvailableFunds (prior period)"`
		Total                     string `json:"Total: "`
		EndingBalance             string `json:"EndingBalance"`
		RAdj                      string `json:"AdjustmentsForActivitiesPostCollectionPeriod"`
		EndingBalance1            string `json:"AdjustedEndingBalance "`
	}

	type ReserveAccount struct {
		RESERVEACCOUNT         string `json:"RESERVE ACCOUNT"`
		BeginningBalance       string `json:"BeginningBalance"`
		Deposits               string `json:"Deposits"`
		FromPriorityofPayments string `json:"FromPriorityOfPayments (prior period)"`
		Total                  string `json:"Total:"`
		Withdrawals            string `json:"Withdrawals"`
		ToAvaiFunds            string `json:"ToAvailableFunds (prior period)"`
		RTotal                 string `json:"Total: "`
		EndingBalance          string `json:"EndingBalance"`
		ResAdj                 string `json:"AdjustmentsForActivitiesPostCollectionPeriod"`
		EndingBalance1         string `json:"AdjustedEndingBalance "`
	}

	type AssetManagementAccount struct {
		ASSETMANAGEMENTACCOUNT string `json:"ASSET MANAGEMENT ACCOUNT"`
		BeginningBalance       string `json:"BeginningBalance"`
		Deposits               string `json:"Deposits"`
		FromPriorityofPayments string `json:"FromPriorityOfPayments (prior period)"`
		Total                  string `json:"Total:"`
		Withdrawals            string `json:"Withdrawals"`
		ToAvaiFunds            string `json:"ToAvailableFunds (prior period)"`
		ATotal                 string `json:"Total: "`
		EndingBalance          string `json:"EndingBalance"`
		AssetAdj               string `json:"AdjustmentsForActivitiesPostCollectionPeriod"`
		EndingBalance1         string `json:"AdjustedEndingBalance "`
	}

	type PlaceHolder struct {
		PLACEHOLDER            string `json:"ACCOUNT4:"`
		BeginningBalance       string `json:"BeginningBalance"`
		Deposits               string `json:"Deposits"`
		FromPriorityofPayments string `json:"FromPriorityOfPayments (prior period)"`
		Total                  string `json:"Total:"`
		Withdrawals            string `json:"Withdrawals"`
		ToAvaiFunds            string `json:"ToAvailableFunds (prior period)"`
		PTotal                 string `json:"Total: "`
		EndingBalance          string `json:"EndingBalance"`
		PlaceAdj               string `json:"AdjustmentsForActivitiesPostCollectionPeriod"`
		EndingBalance1         string `json:"AdjustedEndingBalance "`
	}

	type DealFeesAndExpenses1 struct {
		Fee             string `json:"Fees"`
		Type            string `json:"Type"`
		BeginningUnpaid string `json:"BeginningUnpaid"`
		CurrentDue      string `json:"CurrentDue"`
		TotalDue        string `json:"TotalDue"`
		TotalPaid       string `json:"TotalPaid"`
		EndingUnpaid    string `json:"EndingUnpaid"`
	}

	type DealFeesAndExpenses struct {
		Fee             string `json:"Fees"`
		Type            string `json:"Type"`
		BeginningUnpaid string `json:"BeginningUnpaid"`
		CurrentDue      string `json:"CurrentDue"`
		TotalDue        string `json:"TotalDue"`
		TotalPaid       string `json:"TotalPaid"`
		EndingUnpaid    string `json:"EndingUnpaid"`
	}

	//DEal Events
	type DealEvents struct {
		Key    string `json:"Key"`
		Value1 string `json:"Value1"`
		Value2 string `json:"Value2"`
		Value3 string `json:"Value3"`
		Value4 string `json:"Value4"`
	}

	type DealEvents2 struct {
		Key    string `json:"Key"`
		Value1 string `json:"Value1"`
		Value2 string `json:"Value2"`
	}

	//priority of payments
	type PriorityOfPayments struct {
		Key            string `json:"Key"`
		AvailableFunds string `json:"AvailableFunds"`
		AmountOwed     string `json:"AmountOwed"`
		Adjustment     string `json:"Adjustment"`
		AdjustmentType string `json:"AdjustmentType"`
		AmountPaid     string `json:"AmountPaid"`
	}

	//priority of payments
	type PrincipalPayments struct {
		Class            string `json:"Class"`
		OriginalBalance  string `json:"OriginalBalance / NotionalBalance"`
		BeginningBalance string `json:"BeginningBalance"`
		PrincipalPaid    string `json:"PrincipalPaid"`
		WriteDownWriteUp string `json:"(Write-Down) / Write-Up"`
		EndingBalance    string `json:"EndingBalance"`
	}

	type InterestPayments struct {
		Class             string `json:"Class"`
		InterestRate      string `json:"InterestRate"`
		BeginningBalance  string `json:"BeginningBalance / NotionalBalance"`
		InterestOwed      string `json:"InterestOwed"`
		InterestShortfall string `json:"BeginningInterestShortfall"`
		InterestPaid      string `json:"InterestPaid"`
		InterestUnpaid    string `json:"EndingInterestShortfall"`
	}
	type AdditionalDetails struct {
		Class                        string `json:"Class"`
		BeginningCumulativeWriteDown string `json:"BeginningCumulativeWrite-Down"`
		CumulativeWriteDownPaid      string `json:"CumulativeWritedownPaid"`
		EndingCumulativeWriteDown    string `json:"EndingCumulativeWritedown"`
		TotalCumulativeWACShortfall  string `json:"TotalCumulative WAC Shortfall"`
		CumulativeWACShotfallPaid    string `json:"Cumulative WAC ShortfallPaid"`
		EndingCumulativeWACShortfall string `json:"EndingCumulative WAC Shortfall"`
	}
	type PaymentSummary struct {
		Class            string `json:"Class"`
		CUSIP            string `json:"CUSIP"`
		BeginningBalance string `json:"BeginningBalance"`
		InterestPaid     string `json:"InterestPaid"`
		PrincipalPaid    string `json:"PrincipalPaid"`
		TotalPaid        string `json:"TotalPaid"`
		EndingBalance    string `json:"EndingBalance"`
	}
	type PaymentSummary1 struct {
		Class            string `json:"Class"`
		CUSIP            string `json:"CUSIP"`
		BeginningBalance string `json:"BeginningBalance"`
		InterestPaid     string `json:"InterestPaid"`
		PrincipalPaid    string `json:"PrincipalPaid"`
		TotalPaid        string `json:"TotalPaid"`
		EndingBalance    string `json:"EndingBalance"`
	}
	type ClassFactors struct {
		Class            string `json:"Class"`
		BeginningBalance string `json:"BeginningBalance / NotionalBalance"`
		InterestPaid     string `json:"InterestPaid"`
		PrincipalPaid    string `json:"PrincipalPaid"`
		TotalPaid        string `json:"TotalPaid"`
		EndingBalance    string `json:"EndingBalance"`
	}
	type CollateralSummary struct {
		Activity                   string `json:"Activity"`
		Count                      string `json:"Count"`
		PrincipalBalance           string `json:"PrincipalBalance"`
		CumulativeCount            string `json:"CumulativeCount"`
		CumulativePrincipalBalance string `json:"CumulativePrincipalBalance"`
	}

	//Performance Details
	type PerformanceDetails struct {
		Status            string `json:"Status"`
		PrincipalBalanceD string `json:"Principal Balance ($)"`
		PrincipalBalanceP string `json:"Principal Balance (%)"`
	}

	//paid in full
	type PaidInFull struct {
		LoanID           string `json:"LoanID"`
		PrincipalBalance string `json:"PrincipalBalance"`
		// FICO             string `json:"FICO"`
		// LTV              string `json:"LTV"`
		// RemainingTerm    string `json:"RemainingTerm"`
	}

	//Modified
	type Modified struct {
		LoanID           string `json:"LoanID"`
		PrincipalBalance string `json:"PrincipalBalance"`
		// ModificationType string `json:"ModificationType"`
		// FICO             string `json:"FICO"`
		// LTV              string `json:"LTV"`
		// RemainingTerm    string `json:"RemainingTerm"`
	}

	//Purchased
	type Purchased struct {
		LoanID           string `json:"LoanID"`
		PrincipalBalance string `json:"PrincipalBalance"`
		// PurchasePrice    string `json:"PurchasePrice"`
		// FICO             string `json:"FICO"`
		// LTV              string `json:"LTV"`
		// RemainingTerm    string `json:"RemainingTerm"`
	}

	//Funded
	type FUnded struct {
		LoanID           string `json:"LoanID"`
		PrincipalBalance string `json:"PrincipalBalance"`
		// FundedAmount     string `json:"FundedAmount"`
		// FICO             string `json:"FICO"`
		// LTV              string `json:"LTV"`
		// RemainingTerm    string `json:"RemainingTerm"`
	}

	type Dummy struct {
		Current1        string `json:"Current1"`
		Current2        string `json:"Current2"`
		AmtPaid         string `json:"AmtPaid"`
		RevolvingPeriod string `json:"RevolvingPeriod"`
	}

	var buffer bytes.Buffer
	buffer.WriteString("{\"AccountStatement\":[")

	//COLLATERAL BAL CALCS
	beginningBal1, _ := strconv.ParseFloat(smap["Collateral Balance - Beginning Principal Balance"], 64)
	Purchased1, _ := strconv.ParseFloat(smap["Collateral Balance - Purchased"], 64)
	Funded1, _ := strconv.ParseFloat(smap["Collateral Balance - Funded"], 64)
	CapitalizedInterest1, _ := strconv.ParseFloat(smap["Collateral Balance - Capitalized Interest"], 64)
	OtherPlus1, _ := strconv.ParseFloat(smap["Collateral Balance - Other(+)"], 64)
	OtherPlusNonCash1, _ := strconv.ParseFloat(smap["Collateral Balance - Other(+)(non-cash)"], 64)

	CTotal1 := Purchased1 + Funded1 + OtherPlus1 + CapitalizedInterest1 + OtherPlusNonCash1

	PaidInFull1, _ := strconv.ParseFloat(smap["Collateral Balance - Paid In Full"], 64)
	Sale1, _ := strconv.ParseFloat(smap["Collateral Balance - Sale"], 64)
	Liquidation1, _ := strconv.ParseFloat(smap["Collateral Balance - Liquidation"], 64)
	Curtailments1, _ := strconv.ParseFloat(smap["Collateral Balance - Curtailments"], 64)
	RealizedLosses1, _ := strconv.ParseFloat(smap["Collateral Balance - Realized Losses"], 64)
	ScheduldedPrin1, _ := strconv.ParseFloat(smap["Collateral Balance - Scheduled Principal"], 64)
	OtherMinus1, _ := strconv.ParseFloat(smap["Collateral Balance - Other(-)"], 64)
	OtherMinusNoncash1, _ := strconv.ParseFloat(smap["Collateral Balance - Other(-)(non-cash)"], 64)

	CTotal11 := PaidInFull1 + Sale1 + Liquidation1 + Curtailments1 + RealizedLosses1 + OtherMinus1 + ScheduldedPrin1 + OtherMinusNoncash1

	endingbalance := beginningBal1 + CTotal1 - CTotal11

	CoPur := 1
	CoFun := 0
	CoCap := 0
	CoOthPlus := 0
	CoOthPlusNonCah := 0
	CoPaid := -1
	CoSale := -1
	CoLiqui := -1
	CoCurt := 0
	CoReal := 0
	CoSchPrin := 0
	CoOthMinus := 0
	CoOthMinusNonCash := 0

	hashofloans, _ := strconv.ParseFloat(smap["Collateral Balance - Beginning Principal Balance "], 64)

	var Purchased2, Funded2, CapitalizedInterest2, OtherPlus2, OtherPlusNonCash2, PaidInFull2, Sale2, Liquidation2, Curtailments2, RealizedLosses2, ScheduldedPrin2, OtherMinus2, OtherMinusNoncash2 float64
	if CoPur != 0 {
		Purchased2, _ = strconv.ParseFloat(smap["Collateral Balance - Purchased "], 64)
	} else {
		Purchased2 = 0
	}
	if CoFun != 0 {
		Funded2, _ = strconv.ParseFloat(smap["Collateral Balance - Funded "], 64)
	} else {
		Funded2 = 0
	}
	if CoCap != 0 {
		CapitalizedInterest2, _ = strconv.ParseFloat(smap["Collateral Balance - Capitalized Interest "], 64)
	} else {
		CapitalizedInterest2 = 0
	}
	if CoOthPlus != 0 {
		OtherPlus2, _ = strconv.ParseFloat(smap["Collateral Balance - Other(+) "], 64)
	} else {
		OtherPlus2 = 0
	}
	if CoOthPlusNonCah != 0 {
		OtherPlusNonCash2, _ = strconv.ParseFloat(smap["Collateral Balance - Other(+)(non-cash) "], 64)
	} else {
		OtherPlusNonCash2 = 0
	}

	CTotal2 := Purchased2 + Funded2 + OtherPlus2 + CapitalizedInterest2 + OtherPlusNonCash2

	if CoPaid != 0 {
		PaidInFull2, _ = strconv.ParseFloat(smap["Collateral Balance - Paid In Full "], 64)
	} else {
		PaidInFull2 = 0
	}
	if CoSale != 0 {
		Sale2, _ = strconv.ParseFloat(smap["Collateral Balance - Sale "], 64)
	} else {
		Sale2 = 0
	}
	if CoLiqui != 0 {
		Liquidation2, _ = strconv.ParseFloat(smap["Collateral Balance - Liquidation "], 64)
	} else {
		Liquidation2 = 0
	}
	if CoCurt != 0 {
		Curtailments2, _ = strconv.ParseFloat(smap["Collateral Balance - Curtailments "], 64)
	} else {
		Curtailments2 = 0
	}
	if CoReal != 0 {
		RealizedLosses2, _ = strconv.ParseFloat(smap["Collateral Balance - Realized Losses "], 64)
	} else {
		RealizedLosses2 = 0
	}
	if CoSchPrin != 0 {
		ScheduldedPrin2, _ = strconv.ParseFloat(smap["Collateral Balance - Scheduled Principal "], 64)
	} else {
		ScheduldedPrin2 = 0
	}
	if CoOthMinus != 0 {
		OtherMinus2, _ = strconv.ParseFloat(smap["Collateral Balance - Other(-) "], 64)
	} else {
		OtherMinus2 = 0
	}
	if CoOthMinusNonCash != 0 {
		OtherMinusNoncash2, _ = strconv.ParseFloat(smap["Collateral Balance - Other(-)(non-cash) "], 64)
	} else {
		OtherMinusNoncash2 = 0
	}

	//cpop2 := 0.00
	//fmt.Println(cpop2)

	CTotal22 := PaidInFull2 + Sale2 + Liquidation2 + Curtailments2 + RealizedLosses2 + OtherMinus2 + ScheduldedPrin2 + OtherMinusNoncash2

	endingbalancehashofloans := hashofloans + CTotal2 - CTotal22

	cAdj, _ := strconv.ParseFloat(args[1], 64)

	FinalCEndingBal := endingbalance + cAdj

	fmt.Println(FinalCEndingBal)
	buffer.WriteString("[")
	collateralbal0 := CollateralBal{Key: "COLLATERAL BALANCE", Value1: "", Value2: ""}
	co0, _ := json.Marshal(collateralbal0)
	buffer.WriteString(string(co0))
	buffer.WriteString(",")

	collateralbal1 := CollateralBal{Key: "", Value1: "Number of Loans", Value2: "Unpaid Principal Balance"}
	co1, _ := json.Marshal(collateralbal1)
	buffer.WriteString(string(co1))
	buffer.WriteString(",")

	collateralbal2 := CollateralBal{Key: "Beginning Balance", Value1: strconv.FormatFloat(hashofloans, 'f', 0, 64), Value2: strconv.FormatFloat(beginningBal1, 'E', -1, 64)}
	co2, _ := json.Marshal(collateralbal2)
	buffer.WriteString(string(co2))
	buffer.WriteString(",")

	collateralbal3 := CollateralBal{Key: "Deposits", Value1: "", Value2: ""}
	co3, _ := json.Marshal(collateralbal3)
	buffer.WriteString(string(co3))
	buffer.WriteString(",")

	collateralbal4 := CollateralBal{Key: "Purchased", Value1: strconv.FormatFloat(Purchased2, 'f', 0, 64), Value2: strconv.FormatFloat(Purchased1, 'E', -1, 64)}
	co4, _ := json.Marshal(collateralbal4)
	buffer.WriteString(string(co4))
	buffer.WriteString(",")

	collateralbal5 := CollateralBal{Key: "Funded", Value1: strconv.FormatFloat(Funded2, 'f', 0, 64), Value2: strconv.FormatFloat(Funded1, 'E', -1, 64)}
	co5, _ := json.Marshal(collateralbal5)
	buffer.WriteString(string(co5))
	buffer.WriteString(",")

	collateralbal18 := CollateralBal{Key: "Capitalized Interest", Value1: strconv.FormatFloat(CapitalizedInterest2, 'f', 0, 64), Value2: strconv.FormatFloat(CapitalizedInterest1, 'E', -1, 64)}
	co18, _ := json.Marshal(collateralbal18)
	buffer.WriteString(string(co18))
	buffer.WriteString(",")

	collateralbal6 := CollateralBal{Key: "Other(+)", Value1: strconv.FormatFloat(OtherPlus2, 'f', 0, 64), Value2: strconv.FormatFloat(OtherPlus1, 'E', -1, 64)}
	co6, _ := json.Marshal(collateralbal6)
	buffer.WriteString(string(co6))
	buffer.WriteString(",")

	collateralbal06 := CollateralBal{Key: "Other(+)(non-cash)", Value1: strconv.FormatFloat(OtherPlusNonCash2, 'f', 0, 64), Value2: strconv.FormatFloat(OtherPlusNonCash1, 'E', -1, 64)}
	co06, _ := json.Marshal(collateralbal06)
	buffer.WriteString(string(co06))
	buffer.WriteString(",")

	collateralbal7 := CollateralBal{Key: "Total:", Value1: strconv.FormatFloat(CTotal2, 'f', 0, 64), Value2: strconv.FormatFloat(CTotal1, 'E', -1, 64)}
	co7, _ := json.Marshal(collateralbal7)
	buffer.WriteString(string(co7))
	buffer.WriteString(",")

	collateralbal8 := CollateralBal{Key: "Withdrawals", Value1: "", Value2: ""}
	co8, _ := json.Marshal(collateralbal8)
	buffer.WriteString(string(co8))
	buffer.WriteString(",")

	collateralbal9 := CollateralBal{Key: "Paid In Full", Value1: strconv.FormatFloat(PaidInFull2, 'f', 0, 64), Value2: strconv.FormatFloat(PaidInFull1, 'E', -1, 64)}
	co9, _ := json.Marshal(collateralbal9)
	buffer.WriteString(string(co9))
	buffer.WriteString(",")

	collateralbal10 := CollateralBal{Key: "Sale", Value1: strconv.FormatFloat(Sale2, 'f', 0, 64), Value2: strconv.FormatFloat(Sale1, 'E', -1, 64)}
	co10, _ := json.Marshal(collateralbal10)
	buffer.WriteString(string(co10))
	buffer.WriteString(",")

	collateralbal11 := CollateralBal{Key: "Liquidation", Value1: strconv.FormatFloat(Liquidation2, 'f', 0, 64), Value2: strconv.FormatFloat(Liquidation1, 'E', -1, 64)}
	co11, _ := json.Marshal(collateralbal11)
	buffer.WriteString(string(co11))
	buffer.WriteString(",")

	collateralbal12 := CollateralBal{Key: "Curtailments", Value1: strconv.FormatFloat(Curtailments2, 'f', 0, 64), Value2: strconv.FormatFloat(Curtailments1, 'E', -1, 64)}
	co12, _ := json.Marshal(collateralbal12)
	buffer.WriteString(string(co12))
	buffer.WriteString(",")

	collateralbal13 := CollateralBal{Key: "Realized Losses", Value1: strconv.FormatFloat(RealizedLosses2, 'f', 0, 64), Value2: strconv.FormatFloat(RealizedLosses1, 'E', -1, 64)}
	co13, _ := json.Marshal(collateralbal13)
	buffer.WriteString(string(co13))
	buffer.WriteString(",")

	collateralbal17 := CollateralBal{Key: "Scheduled Principal", Value1: strconv.FormatFloat(ScheduldedPrin2, 'f', 0, 64), Value2: strconv.FormatFloat(ScheduldedPrin1, 'E', -1, 64)}
	co17, _ := json.Marshal(collateralbal17)
	buffer.WriteString(string(co17))
	buffer.WriteString(",")

	collateralbal14 := CollateralBal{Key: "Other(-)", Value1: strconv.FormatFloat(OtherMinus2, 'f', 0, 64), Value2: strconv.FormatFloat(OtherMinus1, 'E', -1, 64)}
	co14, _ := json.Marshal(collateralbal14)
	buffer.WriteString(string(co14))
	buffer.WriteString(",")

	collateralbal19 := CollateralBal{Key: "Other(-)(non-cash)", Value1: strconv.FormatFloat(OtherMinusNoncash2, 'f', 0, 64), Value2: strconv.FormatFloat(OtherMinusNoncash1, 'E', -1, 64)}
	co19, _ := json.Marshal(collateralbal19)
	buffer.WriteString(string(co19))
	buffer.WriteString(",")

	collateralbal15 := CollateralBal{Key: "Total:", Value1: strconv.FormatFloat(CTotal22, 'f', 0, 64), Value2: strconv.FormatFloat(CTotal11, 'E', -1, 64)}
	co15, _ := json.Marshal(collateralbal15)
	buffer.WriteString(string(co15))
	buffer.WriteString(",")

	collateralbal16 := CollateralBal{Key: "Ending Balance", Value1: strconv.FormatFloat(endingbalancehashofloans, 'f', 0, 64), Value2: strconv.FormatFloat(endingbalance, 'E', -1, 64)}
	co16, _ := json.Marshal(collateralbal16)
	buffer.WriteString(string(co16))
	buffer.WriteString(",")

	collateralbal20 := CollateralBal{Key: "Adjustments for activities post collection period", Value1: "", Value2: strconv.FormatFloat(cAdj, 'E', -1, 64)}
	co20, _ := json.Marshal(collateralbal20)
	buffer.WriteString(string(co20))
	buffer.WriteString(",")

	collateralbal21 := CollateralBal{Key: "Adjusted Ending Balance", Value1: "", Value2: strconv.FormatFloat(FinalCEndingBal, 'E', -1, 64)}
	co21, _ := json.Marshal(collateralbal21)
	buffer.WriteString(string(co21))
	buffer.WriteString("],")

	//principal remittance calcs
	pbeginningbal, _ := strconv.ParseFloat(args[2], 64) //node app ip
	pPrincPaid, _ := strconv.ParseFloat(smap["Principal Remittance - Principal Paid"], 64)
	pOtherPlus, _ := strconv.ParseFloat(smap["Principal Remittance - Other(+)"], 64)

	ptotal := pPrincPaid + pOtherPlus

	pOtherMinus, _ := strconv.ParseFloat(smap["Principal Remittance - Other(-)"], 64)
	pToAvailableFunds := pbeginningbal

	pwithdrawTotal := pOtherMinus + pToAvailableFunds

	pEndingbalance := pbeginningbal + ptotal - pwithdrawTotal

	pAdj, _ := strconv.ParseFloat(args[3], 64)
	FinalPEndingBal := pEndingbalance + pAdj

	fmt.Println("FinalPEndingBal", FinalPEndingBal)

	princremmit := PrincipalRemittance{PrincipalRemittance: "", BeginningBalance: strconv.FormatFloat(pbeginningbal, 'E', -1, 64), Deposits: "", PrincipalPaid: strconv.FormatFloat(pPrincPaid, 'E', -1, 64), Other: strconv.FormatFloat(pOtherPlus, 'E', -1, 64), Total: strconv.FormatFloat(ptotal, 'E', -1, 64), Withdrawals: "", Other1: strconv.FormatFloat(pOtherMinus, 'E', -1, 64), ToAvaiFunds: strconv.FormatFloat(pToAvailableFunds, 'E', -1, 64), PTotal: strconv.FormatFloat(pwithdrawTotal, 'E', -1, 64), EndingBalance: strconv.FormatFloat(pEndingbalance, 'E', -1, 64)}
	pc, _ := json.Marshal(princremmit)
	buffer.WriteString(string(pc))
	buffer.WriteString(",")

	//Interest remittance calcs
	ibeginningbal, _ := strconv.ParseFloat(args[4], 64) //node ip - mapping not yet given
	iInterestpaid, _ := strconv.ParseFloat(smap["Interest Remittance - Interest Paid"], 64)
	iOtherplus, _ := strconv.ParseFloat(smap["Interest Remittance - Other(+)"], 64)
	iFeespaid, _ := strconv.ParseFloat(smap["Interest Remittance - Fees Paid"], 64)

	itotal := iInterestpaid + iOtherplus + iFeespaid

	iServicerFees, _ := strconv.ParseFloat(smap["Interest Remittance - Servicing Fees"], 64)
	iOtherMinus, _ := strconv.ParseFloat(smap["Interest Remittance - Other(-)"], 64)
	iToAvailableFunds := ibeginningbal

	iwithdrawTotal := iOtherMinus + iToAvailableFunds + iServicerFees

	iEndingbalance := ibeginningbal + itotal - iwithdrawTotal

	iAdj, _ := strconv.ParseFloat(args[5], 64)
	FinalIEndingBal := iEndingbalance + iAdj

	fmt.Println("FinalIEndingBal", FinalIEndingBal)

	intremmit := INTERESTREMITTANCE{INTERESTREMITTANCE: "", BeginningBalance: strconv.FormatFloat(ibeginningbal, 'E', -1, 64), Deposits: "", InterestPaid: strconv.FormatFloat(iInterestpaid, 'E', -1, 64), FeesPaid: strconv.FormatFloat(iFeespaid, 'E', -1, 64), Other: strconv.FormatFloat(iOtherplus, 'E', -1, 64), Total: strconv.FormatFloat(itotal, 'E', -1, 64), Withdrawals: "", ServicerFees: strconv.FormatFloat(iServicerFees, 'E', -1, 64), Other1: strconv.FormatFloat(iOtherMinus, 'E', -1, 64), ToAvaiFunds: strconv.FormatFloat(iToAvailableFunds, 'E', -1, 64), ITotal: strconv.FormatFloat(iwithdrawTotal, 'E', -1, 64), EndingBalance: strconv.FormatFloat(iEndingbalance, 'E', -1, 64)}
	intrem, _ := json.Marshal(intremmit)
	buffer.WriteString(string(intrem))
	buffer.WriteString(",")

	//Available fund calcs
	aBeginningBalance, _ := strconv.ParseFloat(args[6], 64) //node app ip
	aPrincipalRemittance := pEndingbalance
	aInterestRemittance := iEndingbalance
	acc1, _ := strconv.ParseFloat(args[7], 64)
	acc2, _ := strconv.ParseFloat(args[8], 64)
	acc3, _ := strconv.ParseFloat(args[9], 64)
	acc4, _ := strconv.ParseFloat(args[10], 64)

	aTotal := aPrincipalRemittance + aInterestRemittance + acc1 + acc2 + acc3 + acc4

	aToPoP := aBeginningBalance

	aWithdrawTotal := aToPoP

	aEndingbalance := aBeginningBalance + aTotal - aWithdrawTotal

	aAdj, _ := strconv.ParseFloat(args[11], 64)
	FinalAEndingBal := aEndingbalance + aAdj

	fmt.Println("FinalAEndingBal", FinalAEndingBal)

	avaiFund := AVAILABLEFUNDS{AVAILABLEFUNDS: "", BeginningBalance: strconv.FormatFloat(aBeginningBalance, 'E', -1, 64), Deposits: "", PrincipalRemittance: strconv.FormatFloat(aPrincipalRemittance, 'E', -1, 64), InterestRemittance: strconv.FormatFloat(aInterestRemittance, 'E', -1, 64), Acc1: strconv.FormatFloat(acc1, 'E', -1, 64), Acc2: strconv.FormatFloat(acc2, 'E', -1, 64), Acc3: strconv.FormatFloat(acc3, 'E', -1, 64), Acct4: strconv.FormatFloat(acc4, 'E', -1, 64), Total: strconv.FormatFloat(aTotal, 'E', -1, 64), Withdrawals: "", ToPriorityofPayments: strconv.FormatFloat(aToPoP, 'E', -1, 64), ATotal: strconv.FormatFloat(aWithdrawTotal, 'E', -1, 64), EndingBalance: strconv.FormatFloat(aEndingbalance, 'E', -1, 64)}
	avai, _ := json.Marshal(avaiFund)
	buffer.WriteString(string(avai))
	buffer.WriteString(",")

	///////////////////////////
	//ACCOUNT1: REVOLVING PERIOD REINV ACCT
	RevolvAccBeginBln, _ := strconv.ParseFloat(args[12], 64)

	RevolvFromPriorityofPayments, _ := strconv.ParseFloat(args[119], 64) //Hardcoded for now - from pop
	DepositTotal := RevolvFromPriorityofPayments

	RevolvToPurchaseOfAddiLoans, _ := strconv.ParseFloat(args[130], 64)
	RevolvToAvaiFund := 0.00 //formula not yet given
	WithdrawTotal := RevolvToAvaiFund + RevolvToPurchaseOfAddiLoans

	RevolvEndingBalance := RevolvAccBeginBln + DepositTotal - WithdrawTotal

	RevolvAdj, _ := strconv.ParseFloat(args[13], 64)
	FinalRevolvEndingBal := RevolvEndingBalance + RevolvAdj
	fmt.Println(FinalRevolvEndingBal)

	rev := RevolvingPerReinvAccts{RevolvingPerReinvAcct: "", BeginBln: strconv.FormatFloat(RevolvAccBeginBln, 'E', -1, 64), Deposits: "", FromPriorityofPayments: strconv.FormatFloat(RevolvFromPriorityofPayments, 'E', -1, 64), DepositTotal: strconv.FormatFloat(DepositTotal, 'E', -1, 64), Withdrawals: "", ToPurchaseAdditionalLoans: strconv.FormatFloat(RevolvToPurchaseOfAddiLoans, 'E', -1, 64), ToAvaiFunds: strconv.FormatFloat(RevolvToAvaiFund, 'E', -1, 64), Total: strconv.FormatFloat(WithdrawTotal, 'E', -1, 64), EndingBalance: strconv.FormatFloat(RevolvEndingBalance, 'E', -1, 64), RAdj: strconv.FormatFloat(RevolvAdj, 'E', -1, 64), EndingBalance1: strconv.FormatFloat(FinalRevolvEndingBal, 'E', -1, 64)}
	a1, _ := json.Marshal(rev)
	buffer.WriteString(string(a1))
	buffer.WriteString(",")

	//Reserver account calcs
	ResBeginBln, _ := strconv.ParseFloat(args[14], 64)
	RFromPriorityofPayments, _ := strconv.ParseFloat(args[120], 64) //From pop - hardcoded for now
	ResDtotal := RFromPriorityofPayments

	RToAvaiFunds := 0.00 //formula not yet given
	ResWtotal := RToAvaiFunds
	ResEndingBln := ResBeginBln + ResDtotal - ResWtotal

	ResAdj, _ := strconv.ParseFloat(args[15], 64)
	FinalResEndingBal := ResEndingBln + ResAdj

	fmt.Println("FinalResEndingBal", FinalResEndingBal)

	res := ReserveAccount{RESERVEACCOUNT: "", Deposits: "", BeginningBalance: strconv.FormatFloat(ResBeginBln, 'E', -1, 64), FromPriorityofPayments: strconv.FormatFloat(RFromPriorityofPayments, 'E', -1, 64), Total: strconv.FormatFloat(ResDtotal, 'E', -1, 64), Withdrawals: "", ToAvaiFunds: strconv.FormatFloat(RToAvaiFunds, 'E', -1, 64), RTotal: strconv.FormatFloat(ResWtotal, 'E', -1, 64), EndingBalance: strconv.FormatFloat(ResEndingBln, 'E', -1, 64), ResAdj: strconv.FormatFloat(ResAdj, 'E', -1, 64), EndingBalance1: strconv.FormatFloat(FinalResEndingBal, 'E', -1, 64)}
	re1, _ := json.Marshal(res)
	buffer.WriteString(string(re1))
	buffer.WriteString(",")

	//Asset management account calcs
	AssetBeginBln, _ := strconv.ParseFloat(args[16], 64)
	AssetFromPriorityofPayments, _ := strconv.ParseFloat(args[121], 64) //From pop - hardcoded for now
	AssetDtotal := AssetFromPriorityofPayments

	AssetToAvaiFunds := 0.00 //formula not yet given
	AssetWtotal := AssetToAvaiFunds
	AssetEndingBln := AssetBeginBln + AssetDtotal - AssetWtotal

	AssetAdj, _ := strconv.ParseFloat(args[17], 64)
	FinalAssetEndingBal := AssetEndingBln + AssetAdj

	fmt.Println("FinalAssetEndingBal", FinalAssetEndingBal)

	Asse := AssetManagementAccount{ASSETMANAGEMENTACCOUNT: "", Deposits: "", BeginningBalance: strconv.FormatFloat(AssetBeginBln, 'E', -1, 64), FromPriorityofPayments: strconv.FormatFloat(AssetFromPriorityofPayments, 'E', -1, 64), Total: strconv.FormatFloat(AssetDtotal, 'E', -1, 64), Withdrawals: "", ToAvaiFunds: strconv.FormatFloat(AssetToAvaiFunds, 'E', -1, 64), ATotal: strconv.FormatFloat(AssetWtotal, 'E', -1, 64), EndingBalance: strconv.FormatFloat(AssetEndingBln, 'E', -1, 64), AssetAdj: strconv.FormatFloat(AssetAdj, 'E', -1, 64), EndingBalance1: strconv.FormatFloat(FinalAssetEndingBal, 'E', -1, 64)}
	As1, _ := json.Marshal(Asse)
	buffer.WriteString(string(As1))
	//buffer.WriteString(",")

	//Place holder calcs
	PlBeginBln, _ := strconv.ParseFloat(args[18], 64)
	PlFromPriorityofPayments, _ := strconv.ParseFloat(args[122], 64) //From pop - hardcoded for now
	PlDtotal := PlFromPriorityofPayments

	PlToAvaiFunds := 0.00 //formula not yet given
	PlWtotal := PlToAvaiFunds
	PlEndingBln := PlBeginBln + PlDtotal - PlWtotal

	PlAdj, _ := strconv.ParseFloat(args[19], 64)
	FinalPlEndingBal := PlEndingBln + PlAdj

	fmt.Println("FinalPlEndingBal", FinalPlEndingBal)

	place := PlaceHolder{"", strconv.FormatFloat(PlBeginBln, 'E', -1, 64), "", strconv.FormatFloat(PlFromPriorityofPayments, 'E', -1, 64), strconv.FormatFloat(PlDtotal, 'E', -1, 64), "", strconv.FormatFloat(PlToAvaiFunds, 'E', -1, 64), strconv.FormatFloat(PlWtotal, 'E', -1, 64), strconv.FormatFloat(PlEndingBln, 'E', -1, 64), strconv.FormatFloat(PlAdj, 'E', -1, 64), strconv.FormatFloat(FinalPlEndingBal, 'E', -1, 64)}
	pl1, _ := json.Marshal(place)
	fmt.Println(string(pl1))
	//buffer.WriteString(string(pl1))
	buffer.WriteString("]")

	//Fee and expenses calculations

	BUnpaid1, _ := strconv.ParseFloat(args[20], 64) //prior data for subseq months
	BUnpaid2, _ := strconv.ParseFloat(args[21], 64) //prior data for subseq months
	BUnpaid3, _ := strconv.ParseFloat(args[22], 64) //prior data for subseq months
	BUnpaid4, _ := strconv.ParseFloat(args[23], 64) //prior data for subseq months
	BUnpaid5, _ := strconv.ParseFloat(args[24], 64) //prior data for subseq months
	BUnpaid6, _ := strconv.ParseFloat(args[25], 64) //prior data for subseq months
	BUnpaid7, _ := strconv.ParseFloat(args[26], 64) //prior data for subseq months

	BUnpaid8, _ := strconv.ParseFloat(args[27], 64)  //prior data for subseq months
	BUnpaid9, _ := strconv.ParseFloat(args[28], 64)  //prior data for subseq months
	BUnpaid10, _ := strconv.ParseFloat(args[29], 64) //prior data for subseq months
	BUnpaid11, _ := strconv.ParseFloat(args[30], 64) //prior data for subseq months
	BUnpaid12, _ := strconv.ParseFloat(args[31], 64) //prior data for subseq months
	BUnpaid13, _ := strconv.ParseFloat(args[32], 64) //prior data for subseq months
	BUnpaid14, _ := strconv.ParseFloat(args[33], 64) //prior data for subseq months

	var CDue1, CDue2, CDue3 float64

	NextPayDate := args[34] //input from node app
	c7split := strings.SplitN(NextPayDate, "/", -1)
	m1, _ := strconv.ParseInt(strings.TrimSpace(c7split[0]), 10, 64)
	d1, _ := strconv.ParseInt(strings.TrimSpace(c7split[1]), 10, 64)
	y1, _ := strconv.ParseInt(strings.TrimSpace(c7split[2]), 10, 64)

	if m1 == 7 && y1 >= 2022 {
		CDue1 = 10000
		CDue2 = 12000
	} else {
		CDue1 = 0
		CDue2 = 0
	}
	CDue3 = 4800 //formula not yet given
	CDue4 := 1000.00

	CDue5 := (0.04 / 100) / 12 * (RevolvAccBeginBln + AssetBeginBln + beginningBal1)
	CDue6 := (0.5 / 100) / 12 * (RevolvAccBeginBln + AssetBeginBln + beginningBal1)
	CDue7 := 0.00 //formula not yet given

	CDue8, _ := strconv.ParseFloat(args[123], 64)  //formula not yet given
	CDue9, _ := strconv.ParseFloat(args[124], 64)  //formula not yet given
	CDue10, _ := strconv.ParseFloat(args[125], 64) //formula not yet given
	CDue11, _ := strconv.ParseFloat(args[126], 64) //formula not yet given
	CDue12, _ := strconv.ParseFloat(args[127], 64) //formula not yet given
	CDue13, _ := strconv.ParseFloat(args[128], 64) //formula not yet given
	CDue14, _ := strconv.ParseFloat(args[129], 64) //formula not yet given

	TDue1 := BUnpaid1 + CDue1
	TDue2 := BUnpaid2 + CDue2
	TDue3 := BUnpaid3 + CDue3
	TDue4 := BUnpaid4 + CDue4
	TDue5 := BUnpaid5 + CDue5
	TDue6 := BUnpaid6 + CDue6
	TDue7 := BUnpaid7 + CDue7
	TDue8 := BUnpaid8 + CDue8
	TDue9 := BUnpaid9 + CDue9
	TDue10 := BUnpaid10 + CDue10
	TDue11 := BUnpaid11 + CDue11
	TDue12 := BUnpaid12 + CDue12
	TDue13 := BUnpaid13 + CDue13
	TDue14 := BUnpaid14 + CDue14

	// deal events calcs
	isEventDefault := strings.ToUpper(args[35])
	isServicerDefault := strings.ToUpper(args[36])
	isEarlyRedemption := strings.ToUpper(args[37])

	var isAmortEvent string
	if isEventDefault == "TRUE" || isServicerDefault == "TRUE" || isEarlyRedemption == "TRUE" {
		isAmortEvent = "TRUE"
	} else {
		isAmortEvent = "FALSE"
	}

	var current1, current2 float64

	K280, _ := strconv.ParseFloat(args[38], 64)
	L280, _ := strconv.ParseFloat(args[39], 64)
	M280, _ := strconv.ParseFloat(args[40], 64)

	K281, _ := strconv.ParseFloat(args[41], 64)
	L281, _ := strconv.ParseFloat(args[42], 64)
	M281, _ := strconv.ParseFloat(args[43], 64)

	if monthval == "6" && yearval == "2021" {
		current1 = K280
		current2 = K281
	} else if monthval == "7" && yearval == "2021" {
		current1 = ((K280 + L280) / 2)
		current2 = ((K281 + L281) / 2)
	} else {
		current1 = ((K280 + L280 + M280) / 3)
		current2 = ((K281 + L281 + M281) / 3)
	}

	limit1 := 15.00
	limit2 := 3.00
	limit3 := 3.0

	limittype1 := "MAX"
	limittype2 := "MAX"
	limittype3 := "MIN"

	var status1 string
	var status2 string
	var status3 string

	K307 := RevolvAccBeginBln + AssetBeginBln + beginningBal1
	origBln1, _ := strconv.ParseFloat(args[44], 64)
	origBln2, _ := strconv.ParseFloat(args[45], 64)
	origBln3, _ := strconv.ParseFloat(args[46], 64)
	origBln4, _ := strconv.ParseFloat(args[47], 64)

	var BeginBln1, BeginBln2, BeginBln3, BeginBln4 float64
	//to check
	bvalue1, _ := strconv.ParseFloat(args[48], 64) //prior data
	bvalue2, _ := strconv.ParseFloat(args[49], 64) //prior data
	bvalue3, _ := strconv.ParseFloat(args[50], 64) //prior data
	bvalue4, _ := strconv.ParseFloat(args[51], 64) //prior data

	fmt.Println(bvalue1)
	fmt.Println(bvalue2)
	fmt.Println(bvalue3)
	fmt.Println(bvalue4)

	if monthval == "6" && yearval == "2021" {
		BeginBln1 = origBln1
		BeginBln2 = origBln2
		BeginBln3 = origBln3
		BeginBln4 = origBln4
	} else {
		BeginBln1 = bvalue1
		BeginBln2 = bvalue2
		BeginBln3 = bvalue3
		BeginBln4 = bvalue4
	}

	test1 := K307 - (BeginBln1 + BeginBln2)
	sumBeginningBalance := BeginBln1 + BeginBln2 + BeginBln3 + BeginBln4
	current3 := test1 / sumBeginningBalance

	if limittype1 == "MAX" {
		if current1*100 > limit1 {
			status1 = "TRUE"
		} else {
			status1 = "FALSE"
		}
	} else {
		if current1*100 <= limit1 {
			status1 = "TRUE"
		} else {
			status1 = "FALSE"
		}
	}

	if limittype2 == "MAX" {
		if current2*100 > limit2 {
			status2 = "TRUE"
		} else {
			status2 = "FALSE"
		}
	} else {
		if current2*100 <= limit2 {
			status2 = "TRUE"
		} else {
			status2 = "FALSE"
		}
	}

	if limittype3 == "MAX" {
		if current3*100 > limit3 {
			status3 = "TRUE"
		} else {
			status3 = "FALSE"
		}
	} else {
		if current3*100 <= limit3 {
			status3 = "TRUE"
		} else {
			status3 = "FALSE"
		}
	}
	var isTriggerEvent string

	if status1 == "TRUE" || status2 == "TRUE" || status3 == "TRUE" {
		isTriggerEvent = "TRUE"
	} else {
		isTriggerEvent = "FALSE"
	}

	date1 := "12/25/2023"
	date2 := "6/25/2023"

	K303 := RevolvEndingBalance + AssetEndingBln + aEndingbalance + endingbalance

	K311 := 20.00
	K312 := 17.00
	//	K313 := 1.00
	K304 := FinalCEndingBal + FinalAssetEndingBal
	K305 := RevolvAccBeginBln + beginningBal1
	K306 := FinalAssetEndingBal + endingbalance
	test2 := 1 - BeginBln1/K303
	test3 := math.Max(0, K303*((K311/100)-1)+BeginBln1)
	var test4 string

	if test2*100 <= K312 {
		test4 = "TRUE"
	} else {
		test4 = "FALSE"
	}

	c260split := strings.SplitN(date2, "/", -1)
	m3, _ := strconv.ParseInt(strings.TrimSpace(c260split[0]), 10, 64)
	d3, _ := strconv.ParseInt(strings.TrimSpace(c260split[1]), 10, 64)
	y3, _ := strconv.ParseInt(strings.TrimSpace(c260split[2]), 10, 64)

	var flag2 float64 = 0
	var valu1 bool

	if y1 < y3 {
		flag2 = 1
	} else if y1 == y3 {
		if m1 < m3 {
			flag2 = 1
		} else if m1 == m3 {
			if d1 <= d3 {
				flag2 = 1
			}
		}
	}
	if flag2 == 1 {
		valu1 = true
	} else {
		valu1 = false
	}

	isAmortEventbool, _ := strconv.ParseBool(isAmortEvent)
	isRevolvingPeriod := strings.ToUpper(strconv.FormatBool(valu1 && !isAmortEventbool))

	fmt.Println(args[52])
	var test5 float64
	if isRevolvingPeriod == "TRUE" {
		test5 = 0.00
	} else {
		test5 = math.Max(0, K305-K306)
	}

	var notC290 string
	if isAmortEvent == "TRUE" {
		notC290 = "FALSE"
	} else {
		notC290 = "TRUE"
	}

	var account1 float64
	if notC290 == "TRUE" && isRevolvingPeriod == "TRUE" {
		account1 = 100000000 - K304
	} else {
		account1 = 0.00
	}

	account2, _ := strconv.ParseFloat(args[53], 64)

	fmt.Println("BeginBln1", BeginBln1)
	fmt.Println("BeginBln2", BeginBln2)
	fmt.Println("BeginBln3", BeginBln3)
	fmt.Println("BeginBln4", BeginBln4)

	//Deal events remaining calcs

	buffer.WriteString(",\"DealEvents\":[")

	dealevent := DealEvents{Key: "", Value1: "Current", Value2: "Level", Value3: "Limit Type", Value4: "Status"}
	de, _ := json.Marshal(dealevent)
	buffer.WriteString(string(de))
	buffer.WriteString(",")
	///////
	dealevent1 := DealEvents{Key: "the three-month trailing average ratio of the weighted average Mortgage Interest Rate of the Whole Mortgage Loans to the weighted average Note Rate of the Class A Notes, in each case as of the related Calculation Date, is less than 1.25:1;", Value1: strconv.FormatFloat(current1*100, 'E', -1, 64) + "%", Value2: strconv.FormatFloat(limit1, 'E', -1, 64) + "%", Value3: limittype1, Value4: status1}
	de1, _ := json.Marshal(dealevent1)
	buffer.WriteString(string(de1))
	buffer.WriteString(",")
	///////
	dealevent2 := DealEvents{Key: "the three-month trailing average 60+ Day Delinquency Rate for such Calculation Date and the Calculation Date for each of the two preceding Collection Periods is greater than 12.50%; or	", Value1: strconv.FormatFloat(current2*100, 'E', -1, 64) + "%", Value2: strconv.FormatFloat(limit2, 'E', -1, 64) + "%", Value3: limittype2, Value4: status2}
	de2, _ := json.Marshal(dealevent2)
	buffer.WriteString(string(de2))
	buffer.WriteString(",")
	////////
	dealevent3 := DealEvents{Key: "the Overcollateralization Amount for the related Payment Date is less than 3.0% of the aggregate Note Amount of the Notes", Value1: strconv.FormatFloat(current3*100, 'E', -1, 64) + "%", Value2: strconv.FormatFloat(limit3, 'E', -1, 64) + "%", Value3: limittype3, Value4: status3}
	de3, _ := json.Marshal(dealevent3)
	buffer.WriteString(string(de3))
	buffer.WriteString("]")
	///////
	buffer.WriteString(",\"DealEvents2\":[")
	dealevent001 := DealEvents2{Key: "", Value1: "", Value2: "Notes(s)"}
	de001, _ := json.Marshal(dealevent001)
	buffer.WriteString(string(de001))
	buffer.WriteString(",")

	dealevent01 := DealEvents2{Key: "EVENT - Is Event of Default", Value1: isEventDefault, Value2: ""}
	de01, _ := json.Marshal(dealevent01)
	buffer.WriteString(string(de01))
	buffer.WriteString(",")

	dealevent02 := DealEvents2{Key: "EVENT - Is Servicer Default", Value1: isServicerDefault, Value2: ""}
	de02, _ := json.Marshal(dealevent02)
	buffer.WriteString(string(de02))
	buffer.WriteString(",")

	dealevent03 := DealEvents2{Key: "EVENT - Is Early Redemption", Value1: isEarlyRedemption, Value2: ""}
	de03, _ := json.Marshal(dealevent03)
	buffer.WriteString(string(de03))
	buffer.WriteString(",")

	dealevent444 := DealEvents2{Key: "EVENT - Is Amortization Event", Value1: isAmortEvent, Value2: ""}
	de444, _ := json.Marshal(dealevent444)
	buffer.WriteString(string(de444))
	buffer.WriteString(",")

	dealevent4 := DealEvents2{Key: "EVENT - Is Trigger Event", Value1: isTriggerEvent, Value2: ""}
	de4, _ := json.Marshal(dealevent4)
	buffer.WriteString(string(de4))
	buffer.WriteString(",")
	///////
	dealevent5 := DealEvents2{Key: "DATE - Expected Redemption Date", Value1: date1, Value2: ""}
	de5, _ := json.Marshal(dealevent5)
	buffer.WriteString(string(de5))
	buffer.WriteString(",")
	///////
	dealevent6 := DealEvents2{Key: "DATE - Revolving Period End Date", Value1: date2, Value2: ""}
	de6, _ := json.Marshal(dealevent6)
	buffer.WriteString(string(de6))
	buffer.WriteString(",")
	///////
	dealevent7 := DealEvents2{Key: "TEST - Overcollateralization Amount", Value1: strconv.FormatFloat(test1, 'E', -1, 64), Value2: ""}
	de7, _ := json.Marshal(dealevent7)
	buffer.WriteString(string(de7))
	buffer.WriteString(",")
	///////
	dealevent8 := DealEvents2{Key: "TEST - Class A-1 Credit Enhancement Percentage", Value1: strconv.FormatFloat(test2*100, 'E', -1, 64) + "%", Value2: ""}
	de8, _ := json.Marshal(dealevent8)
	buffer.WriteString(string(de8))
	buffer.WriteString(",")
	///////
	dealevent10 := DealEvents2{Key: "TEST - Class A Target Amount", Value1: strconv.FormatFloat(test3, 'E', -1, 64), Value2: ""}
	de10, _ := json.Marshal(dealevent10)
	buffer.WriteString(string(de10))
	buffer.WriteString(",")
	///////
	dealevent1000 := DealEvents2{Key: "TEST - Is Class A Sequential Pay Trigger Event", Value1: test4, Value2: ""}
	de1000, _ := json.Marshal(dealevent1000)
	buffer.WriteString(string(de1000))
	buffer.WriteString(",")
	/////////////
	dealevent11 := DealEvents2{Key: "TEST - Optimal Principal Distribution Amount", Value1: strconv.FormatFloat(test5, 'E', -1, 64), Value2: ""}
	de11, _ := json.Marshal(dealevent11)
	buffer.WriteString(string(de11))
	buffer.WriteString(",")
	/////////////
	dealevent21 := DealEvents2{Key: "ACCOUNT - Optimal Revolving Period Reinvestment Account Balance", Value1: strconv.FormatFloat(account1, 'E', -1, 64), Value2: ""}
	de21, _ := json.Marshal(dealevent21)
	buffer.WriteString(string(de21))
	buffer.WriteString(",")
	/////////////
	dealevent22 := DealEvents2{Key: "ACCOUNT - Required Reserve Account Balance", Value1: strconv.FormatFloat(account2, 'E', -1, 64), Value2: ""}
	de22, _ := json.Marshal(dealevent22)
	buffer.WriteString(string(de22))
	buffer.WriteString("]")

	//Priority of payments calcs.
	a := aPrincipalRemittance
	b := a
	c := math.Min(a, b)

	aa1 := aInterestRemittance
	bb1 := aa1
	cc1 := math.Min(aa1, bb1)

	aa2 := acc1
	bb2 := aa2
	cc2 := math.Min(aa2, bb2)

	fmt.Println(cc2)

	aa3 := acc2
	bb3 := aa3
	cc3 := math.Min(aa3, bb3)

	fmt.Println(cc3)

	aa4 := acc3
	bb4 := aa4
	cc4 := math.Min(aa4, bb4)

	aa5 := acc4
	bb5 := aa5
	cc5 := math.Min(aa5, bb5)
	fmt.Println(cc5)

	buffer.WriteString(",\"PriorityOfPayments\":[")
	var arr []float64
	avaiFund1 := c + cc1 + cc2 + cc3 + cc4 + cc5
	amtOwed1 := 0.00 //to be fixed
	adjustment1 := ""
	adjustmentType1 := calcAdjustmentType(adjustment1)
	amtPaid1 := calcAmountPaid(avaiFund1, amtOwed1, adjustment1)
	arr = append(arr, amtPaid1)

	pop1 := PriorityOfPayments{Key: "Beginning Balance", AvailableFunds: strconv.FormatFloat(avaiFund1, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed1, 'E', -1, 64), Adjustment: adjustment1, AdjustmentType: adjustmentType1, AmountPaid: strconv.FormatFloat(amtPaid1, 'E', -1, 64)}
	p1, _ := json.Marshal(pop1)
	buffer.WriteString(string(p1))
	buffer.WriteString(",")
	///////////////////////
	avaiFund2 := calcAvailableFund(avaiFund1, arr)
	amtOwed2 := TDue1 + TDue2 + TDue3 + TDue4 + TDue5 + TDue6 + TDue7
	adjustment2 := ""
	adjustmentType2 := calcAdjustmentType(adjustment2)
	amtPaid2 := calcAmountPaid(avaiFund2, amtOwed2, adjustment2)
	arr = append(arr, amtPaid2)

	pop2 := PriorityOfPayments{Key: "first, pro rata, to the Owner Trustee, the Indenture Trustee, the Paying Agent, the Custodian, the Administrator and the Asset Manager, the Owner Trustee Fee, the Indenture Trustee Fee, the Custodial Fee, the Paying Agent Fee, the Administrator Fee and the Asset Manager Fee, respectively, and any related expenses and indemnification amounts due and owing to the Owner Trustee, the Indenture Trustee, the Paying Agent, the Note Registrar, the Custodian, the Administrator and the Asset Manager (the Transaction Party Expenses) up to the Annual Cap;	Fees", AvailableFunds: strconv.FormatFloat(avaiFund2, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed2, 'E', -1, 64), Adjustment: adjustment2, AdjustmentType: adjustmentType2, AmountPaid: strconv.FormatFloat(amtPaid2, 'E', -1, 64)}
	p2, _ := json.Marshal(pop2)
	buffer.WriteString(string(p2))
	buffer.WriteString(",")
	//////////////////////////
	avaiFund3 := calcAvailableFund(avaiFund1, arr)
	amtOwed3 := TDue8 + TDue9 + TDue10 + TDue11 + TDue12 + TDue13 + TDue14
	var adjustment3 string
	sum, _ := strconv.ParseFloat(args[54], 64)
	M272 := math.Max(0, 400000.00-sum)
	if isEventDefault == "TRUE" || isEarlyRedemption == "TRUE" {
		adjustment3 = ""
	} else {
		adjustment3 = strconv.FormatFloat(M272, 'E', -1, 64)
	}
	adjustmentType3 := calcAdjustmentType(adjustment3)
	amtPaid3 := calcAmountPaid(avaiFund3, amtOwed3, adjustment3)
	arr = append(arr, amtPaid3)

	pop3 := PriorityOfPayments{Key: "Expenses", AvailableFunds: strconv.FormatFloat(avaiFund3, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed3, 'E', -1, 64), Adjustment: adjustment3, AdjustmentType: adjustmentType3, AmountPaid: strconv.FormatFloat(amtPaid3, 'E', -1, 64)}
	p3, _ := json.Marshal(pop3)
	buffer.WriteString(string(p3))
	buffer.WriteString(",")
	////////////////////////
	avaiFund4 := calcAvailableFund(avaiFund1, arr)
	amtOwed4 := amtOwed2 + amtOwed3
	adjustment4 := ""
	adjustmentType4 := calcAdjustmentType(adjustment4)
	amtPaid4 := amtPaid2 + amtPaid3 //calcAmountPaid(avaiFund4, amtOwed4, adjustment4)
	//arr = append(arr, amtPaid4)

	pop4 := PriorityOfPayments{Key: "Total:", AvailableFunds: strconv.FormatFloat(avaiFund4, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed4, 'E', -1, 64), Adjustment: adjustment4, AdjustmentType: adjustmentType4, AmountPaid: strconv.FormatFloat(amtPaid4, 'E', -1, 64)}
	p4, _ := json.Marshal(pop4)
	buffer.WriteString(string(p4))
	buffer.WriteString(",")
	////////////////////////
	avaiFund5 := calcAvailableFund(avaiFund1, arr)
	amtOwed5 := 0.00 // WIP
	adjustment5 := ""
	adjustmentType5 := calcAdjustmentType(adjustment5)
	amtPaid5 := calcAmountPaid(avaiFund5, amtOwed5, adjustment5)
	arr = append(arr, amtPaid5)

	pop5 := PriorityOfPayments{Key: "second, pro rata, to the Servicer and any Additional Servicers, any unreimbursed Servicing Advances, Servicing Expenses, costs and liabilities by and reimbursable to the Servicer or such Additional Servicer pursuant to the Servicing Agreement or related additional servicing agreement, in each case, to the extent that the Servicer or related Additional Servicer has not already reimbursed itself or paid itself for such amounts from Collections;	Expenses", AvailableFunds: strconv.FormatFloat(avaiFund5, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed5, 'E', -1, 64), Adjustment: adjustment5, AdjustmentType: adjustmentType5, AmountPaid: strconv.FormatFloat(amtPaid5, 'E', -1, 64)}
	p5, _ := json.Marshal(pop5)
	buffer.WriteString(string(p5))
	buffer.WriteString(",")
	//////////////////////////////////
	avaiFund6 := calcAvailableFund(avaiFund1, arr)
	amtOwed6 := 0.00 //WIP
	adjustment6 := ""
	adjustmentType6 := calcAdjustmentType(adjustment6)
	amtPaid6 := calcAmountPaid(avaiFund6, amtOwed6, adjustment6)
	arr = append(arr, amtPaid6)

	pop6 := PriorityOfPayments{Key: "third, pro rata, to the Servicer and any Additional Servicers, the Servicing Fee and related additional servicing fee, to the extent not otherwise retained by the Servicer pursuant to the Servicing Agreement or the related Additional Servicer pursuant to the applicable additional servicing agreement; Fees", AvailableFunds: strconv.FormatFloat(avaiFund6, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed6, 'E', -1, 64), Adjustment: adjustment6, AdjustmentType: adjustmentType6, AmountPaid: strconv.FormatFloat(amtPaid6, 'E', -1, 64)}
	p6, _ := json.Marshal(pop6)
	buffer.WriteString(string(p6))
	buffer.WriteString(",")
	///////////////////////////
	K50 := "TRUE"

	k51, _ := strconv.ParseFloat(args[55], 64) //node

	t := Date(int(y1), int(m1), 0)
	K47 := float64(t.Day())
	K48 := float64(365)
	K51 := k51 * K47 / K48 * 12

	TotalCDue := CDue1 + CDue2 + CDue3 + CDue4 + CDue5 + CDue6 + CDue7 + CDue8 + CDue9 + CDue10 + CDue11 + CDue12 + CDue13 + CDue14

	K52 := CDue1 + CDue2 + CDue3 + CDue4 + CDue5 + iServicerFees
	K53 := K51 - K52*12

	denoo1, _ := strconv.ParseFloat(args[56], 64) //node
	L53 := (K53 / denoo1) * 100

	J56 := 3.220
	J57 := 5.470
	J58 := 0.00
	J65 := 0.100

	var irate1, irate2, irate3, irate4 float64
	if K50 == "TRUE" {
		irate1 = math.Min(J56, L53)
		irate2 = math.Min(J57, L53)
		irate3 = math.Min(J58, L53)
		irate4 = math.Min(J65, L53)
	} else {
		irate1 = J56
		irate2 = J57
		irate3 = J58
		irate4 = J65
	}

	c8, _ := strconv.ParseFloat(args[57], 64)
	ipBeginBln4 := beginningBal1

	interestOwed1 := BeginBln1 * (c8 / 360) * (irate1 / 100)
	interestOwed2 := BeginBln2 * (c8 / 360) * (irate2 / 100)
	interestOwed3 := BeginBln3 * (c8 / 360) * (irate3 / 100)
	interestOwed4 := ipBeginBln4 * (c8 / 360) * (irate4 / 100)

	var ifall1, ifall2, ifall3, ifall4 float64

	ifall1, _ = strconv.ParseFloat(args[58], 64)
	ifall2, _ = strconv.ParseFloat(args[59], 64)
	ifall3, _ = strconv.ParseFloat(args[60], 64)
	ifall4, _ = strconv.ParseFloat(args[61], 64)

	avaiFund7 := calcAvailableFund(avaiFund1, arr)
	amtOwed7 := interestOwed4 + ifall4
	adjustment7 := ""
	adjustmentType7 := calcAdjustmentType(adjustment7)
	amtPaid7 := calcAmountPaid(avaiFund7, amtOwed7, adjustment7)
	arr = append(arr, amtPaid7)

	pop7 := PriorityOfPayments{Key: "fourth, to the Class AIOS Notes, to pay the Interest Payment Amount thereon;", AvailableFunds: strconv.FormatFloat(avaiFund7, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed7, 'E', -1, 64), Adjustment: adjustment7, AdjustmentType: adjustmentType7, AmountPaid: strconv.FormatFloat(amtPaid7, 'E', -1, 64)}
	p7, _ := json.Marshal(pop7)
	buffer.WriteString(string(p7))
	buffer.WriteString(",")
	////////////////////////
	avaiFund8 := calcAvailableFund(avaiFund1, arr)
	amtOwed8 := interestOwed1 + ifall1
	adjustment8 := ""
	adjustmentType8 := calcAdjustmentType(adjustment8)
	amtPaid8 := calcAmountPaid(avaiFund8, amtOwed8, adjustment8)
	arr = append(arr, amtPaid8)

	pop8 := PriorityOfPayments{Key: "fifth, to the Class A1 Notes, to pay the Interest Payment Amount thereon;", AvailableFunds: strconv.FormatFloat(avaiFund8, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed8, 'E', -1, 64), Adjustment: adjustment8, AdjustmentType: adjustmentType8, AmountPaid: strconv.FormatFloat(amtPaid8, 'E', -1, 64)}
	p8, _ := json.Marshal(pop8)
	buffer.WriteString(string(p8))
	buffer.WriteString(",")
	///////////////////
	avaiFund9 := calcAvailableFund(avaiFund1, arr)
	amtOwed9 := interestOwed2 + ifall2
	adjustment9 := ""
	adjustmentType9 := calcAdjustmentType(adjustment9)
	amtPaid9 := calcAmountPaid(avaiFund9, amtOwed9, adjustment9)
	arr = append(arr, amtPaid9)

	pop9 := PriorityOfPayments{Key: "sixth, to the Class A2 Notes, to pay the Interest Payment Amount thereon;", AvailableFunds: strconv.FormatFloat(avaiFund9, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed9, 'E', -1, 64), Adjustment: adjustment9, AdjustmentType: adjustmentType9, AmountPaid: strconv.FormatFloat(amtPaid9, 'E', -1, 64)}
	p9, _ := json.Marshal(pop9)
	buffer.WriteString(string(p9))
	buffer.WriteString(",")
	/////////////////////////////
	G345 := test5
	avaiFund10 := calcAvailableFund(avaiFund1, arr)
	amtOwed10 := BeginBln1
	amtOwed11 := BeginBln2
	amtOwed12 := amtOwed10 + amtOwed11
	adjustment12 := G345
	var adjustment10, adjustment11 float64
	if amtOwed12 == 0.00 {
		adjustment10 = 0.00
		adjustment11 = 0.00
	} else {
		adjustment10 = adjustment12 * amtOwed10 / amtOwed12
		adjustment11 = adjustment12 * amtOwed11 / amtOwed12
	}

	adjustmentType10 := calcAdjustmentType(strconv.FormatFloat(adjustment10, 'E', -1, 64))
	adjustmentType11 := calcAdjustmentType(strconv.FormatFloat(adjustment11, 'E', -1, 64))

	amtPaid10 := calcAmountPaid(avaiFund10, amtOwed10, strconv.FormatFloat(adjustment10, 'E', -1, 64))
	arr = append(arr, amtPaid10)

	pop10 := PriorityOfPayments{Key: "seventh, (a) if a Class A Sequential Pay Trigger Event is not in effect, to pay principal on the Class A1	Notes and the Class A2 Notes, pro rata, based on the Note Amounts thereof outstanding as of the related	Payment Date, up to the Optimal Principal Payment Amount for such Payment Date, until the Note Amounts thereof have been reduced to zero, or (b) if a Class A Sequential Pay Trigger Event is in effect, in an amount up to the Class A Target Amount, first, to pay principal on the Class A1 Notes, until the	Note Amount thereof has been reduced to zero, and second, to pay principal on the Class A2 Notes, until	the Note Amount thereof has been reduced to zero;	A1 Notes", AvailableFunds: strconv.FormatFloat(avaiFund10, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed10, 'E', -1, 64), Adjustment: strconv.FormatFloat(adjustment10, 'E', -1, 64), AdjustmentType: adjustmentType10, AmountPaid: strconv.FormatFloat(amtPaid10, 'E', -1, 64)}
	p10, _ := json.Marshal(pop10)
	buffer.WriteString(string(p10))
	buffer.WriteString(",")
	/////////////////////////////
	avaiFund11 := calcAvailableFund(avaiFund1, arr)
	amtPaid11 := calcAmountPaid(avaiFund11, amtOwed11, strconv.FormatFloat(adjustment11, 'E', -1, 64))
	arr = append(arr, amtPaid11)

	pop11 := PriorityOfPayments{Key: "A2 Notes", AvailableFunds: strconv.FormatFloat(avaiFund11, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed11, 'E', -1, 64), Adjustment: strconv.FormatFloat(adjustment11, 'E', -1, 64), AdjustmentType: adjustmentType11, AmountPaid: strconv.FormatFloat(amtPaid11, 'E', -1, 64)}
	p11, _ := json.Marshal(pop11)
	buffer.WriteString(string(p11))
	buffer.WriteString(",")
	/////////////////////////
	avaiFund12 := calcAvailableFund(avaiFund1, arr)
	adjustmentType12 := calcAdjustmentType(strconv.FormatFloat(adjustment12, 'E', -1, 64))
	amtPaid12 := amtPaid10 + amtPaid11

	pop12 := PriorityOfPayments{Key: "Total:", AvailableFunds: strconv.FormatFloat(avaiFund12, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed12, 'E', -1, 64), Adjustment: strconv.FormatFloat(adjustment12, 'E', -1, 64), AdjustmentType: adjustmentType12, AmountPaid: strconv.FormatFloat(amtPaid12, 'E', -1, 64)}
	p12, _ := json.Marshal(pop12)
	buffer.WriteString(string(p12))
	buffer.WriteString(",")
	/////////////////////////
	avaiFund13 := calcAvailableFund(avaiFund1, arr)
	amtOwed13 := amtOwed10 - amtPaid10
	G348 := test3
	adjustment13 := math.Min(amtOwed13, G348)
	adjustmentType13 := calcAdjustmentType(strconv.FormatFloat(adjustment13, 'E', -1, 64))

	amtPaid13 := calcAmountPaid(avaiFund13, amtOwed13, strconv.FormatFloat(adjustment13, 'E', -1, 64))
	arr = append(arr, amtPaid13)
	pop13 := PriorityOfPayments{Key: "Pay Trigger = True	A1 Notes", AvailableFunds: strconv.FormatFloat(avaiFund13, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed13, 'E', -1, 64), Adjustment: strconv.FormatFloat(adjustment13, 'E', -1, 64), AdjustmentType: adjustmentType13, AmountPaid: strconv.FormatFloat(amtPaid13, 'E', -1, 64)}
	p13, _ := json.Marshal(pop13)
	buffer.WriteString(string(p13))
	buffer.WriteString(",")
	////////////////////////////////////
	avaiFund14 := calcAvailableFund(avaiFund1, arr)
	amtOwed14 := amtOwed11 - amtPaid11
	adjustment14 := math.Min(amtOwed14, G348-amtPaid13)
	adjustmentType14 := calcAdjustmentType(strconv.FormatFloat(adjustment14, 'E', -1, 64))
	amtPaid14 := calcAmountPaid(avaiFund14, amtOwed14, strconv.FormatFloat(adjustment14, 'E', -1, 64))

	arr = append(arr, amtPaid14)
	pop14 := PriorityOfPayments{Key: "A2 Notes", AvailableFunds: strconv.FormatFloat(avaiFund14, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed14, 'E', -1, 64), Adjustment: strconv.FormatFloat(adjustment14, 'E', -1, 64), AdjustmentType: adjustmentType14, AmountPaid: strconv.FormatFloat(amtPaid14, 'E', -1, 64)}
	p14, _ := json.Marshal(pop14)
	buffer.WriteString(string(p14))
	buffer.WriteString(",")
	////////////////////////////////
	avaiFund15 := calcAvailableFund(avaiFund1, arr)
	amtOwed15 := amtOwed13 + amtOwed14
	adjustment15 := G348
	adjustmentType15 := calcAdjustmentType(strconv.FormatFloat(adjustment15, 'E', -1, 64))
	amtPaid15 := amtPaid13 + amtPaid14

	//arr = append(arr, amtPaid15)
	pop15 := PriorityOfPayments{Key: "Total:", AvailableFunds: strconv.FormatFloat(avaiFund15, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed15, 'E', -1, 64), Adjustment: strconv.FormatFloat(adjustment15, 'E', -1, 64), AdjustmentType: adjustmentType15, AmountPaid: strconv.FormatFloat(amtPaid15, 'E', -1, 64)}
	p15, _ := json.Marshal(pop15)
	buffer.WriteString(string(p15))
	buffer.WriteString(",")
	////////////////////////////////
	avaiFund16 := calcAvailableFund(avaiFund1, arr)
	amtOwed16 := amtOwed3 - amtPaid3
	adjustment16 := math.Max(0, account1-FinalRevolvEndingBal)
	adjustmentType16 := calcAdjustmentType(strconv.FormatFloat(adjustment16, 'E', -1, 64))
	amtPaid16 := calcAmountPaid(avaiFund16, amtOwed16, strconv.FormatFloat(adjustment16, 'E', -1, 64))

	arr = append(arr, amtPaid16)
	pop16 := PriorityOfPayments{Key: "eighth, pro rata, to the Indenture Trustee, the Owner Trustee, the Paying Agent, the Note Registrar, the Servicer, the Custodian and the Asset Manager, any amounts not paid to such parties as a result of the Annual Cap;", AvailableFunds: strconv.FormatFloat(avaiFund16, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed16, 'E', -1, 64), Adjustment: strconv.FormatFloat(adjustment16, 'E', -1, 64), AdjustmentType: adjustmentType16, AmountPaid: strconv.FormatFloat(amtPaid16, 'E', -1, 64)}
	p16, _ := json.Marshal(pop16)
	buffer.WriteString(string(p16))
	buffer.WriteString(",")
	///////////////////////////////
	avaiFund17 := calcAvailableFund(avaiFund1, arr)
	adjustment17 := math.Max(0, account2-FinalResEndingBal)
	amtOwed17 := adjustment17
	adjustmentType17 := calcAdjustmentType(strconv.FormatFloat(adjustment17, 'E', -1, 64))
	amtPaid17 := calcAmountPaid(avaiFund17, amtOwed17, strconv.FormatFloat(adjustment17, 'E', -1, 64))

	arr = append(arr, amtPaid17)
	pop17 := PriorityOfPayments{Key: "ninth, to the Reserve Account, up to an amount equal to the Required Reserve Account Balance;", AvailableFunds: strconv.FormatFloat(avaiFund17, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed17, 'E', -1, 64), Adjustment: strconv.FormatFloat(adjustment17, 'E', -1, 64), AdjustmentType: adjustmentType17, AmountPaid: strconv.FormatFloat(amtPaid17, 'E', -1, 64)}
	p17, _ := json.Marshal(pop17)
	buffer.WriteString(string(p17))
	buffer.WriteString(",")
	/////////////////////////////
	avaiFund18 := calcAvailableFund(avaiFund1, arr)
	adjustment18 := adjustment16
	amtOwed18 := adjustment18
	adjustmentType18 := calcAdjustmentType(strconv.FormatFloat(adjustment18, 'E', -1, 64))
	amtPaid18 := calcAmountPaid(avaiFund18, amtOwed18, strconv.FormatFloat(adjustment18, 'E', -1, 64))

	arr = append(arr, amtPaid18)
	pop18 := PriorityOfPayments{Key: "tenth, to the Revolving Period Reinvestment Account, up to an amount equal to the Optimal Revolving Period Reinvestment Account Balance; and", AvailableFunds: strconv.FormatFloat(avaiFund18, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed18, 'E', -1, 64), Adjustment: strconv.FormatFloat(adjustment18, 'E', -1, 64), AdjustmentType: adjustmentType18, AmountPaid: strconv.FormatFloat(amtPaid18, 'E', -1, 64)}
	p18, _ := json.Marshal(pop18)
	buffer.WriteString(string(p18))
	buffer.WriteString(",")
	///////////////////////////////
	avaiFund19 := calcAvailableFund(avaiFund1, arr)
	amtOwed19 := avaiFund19
	adjustment19 := ""
	adjustmentType19 := calcAdjustmentType(adjustment19)
	amtPaid19 := calcAmountPaid(avaiFund19, amtOwed19, adjustment19)

	arr = append(arr, amtPaid19)
	pop19 := PriorityOfPayments{Key: "eleventh, to the Class C Notes, any remaining amounts.", AvailableFunds: strconv.FormatFloat(avaiFund19, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed19, 'E', -1, 64), Adjustment: adjustment19, AdjustmentType: adjustmentType19, AmountPaid: strconv.FormatFloat(amtPaid19, 'E', -1, 64)}
	p19, _ := json.Marshal(pop19)
	buffer.WriteString(string(p19))
	buffer.WriteString(",")
	////////////////////////////////
	avaiFund20 := calcAvailableFund(avaiFund1, arr)
	if strconv.FormatFloat(avaiFund20, 'f', 6, 64) == "-0.000000" {
		avaiFund20 = 0.00
	}
	amtOwed20 := 0.00
	adjustment20 := ""
	adjustmentType20 := calcAdjustmentType(adjustment20)
	amtPaid20 := calcAmountPaid(avaiFund20, amtOwed20, adjustment20)

	pop20 := PriorityOfPayments{Key: "Ending Balance", AvailableFunds: strconv.FormatFloat(avaiFund20, 'E', -1, 64), AmountOwed: strconv.FormatFloat(amtOwed20, 'E', -1, 64), Adjustment: adjustment20, AdjustmentType: adjustmentType20, AmountPaid: strconv.FormatFloat(amtPaid20, 'E', -1, 64)}
	p20, _ := json.Marshal(pop20)
	buffer.WriteString(string(p20))
	buffer.WriteString("]")
	/////////////////////////
	//Fee and exoense remaining

	I335 := amtPaid2
	F335 := amtOwed2

	I336 := amtPaid3
	F336 := amtOwed3

	var TPaid1, TPaid2, TPaid3, TPaid4, TPaid5, TPaid6, TPaid7, TPaid8, TPaid9, TPaid10, TPaid11, TPaid12, TPaid13, TPaid14 float64

	if F335*TDue1 == 0.00 {
		TPaid1 = 0.00
	} else {
		TPaid1 = I335 / F335 * TDue1
	}
	if F335*TDue2 == 0.00 {
		TPaid2 = 0.00
	} else {
		TPaid2 = I335 / F335 * TDue2
	}
	if F335*TDue3 == 0.00 {
		TPaid3 = 0.00
	} else {
		TPaid3 = I335 / F335 * TDue3
	}
	if F335*TDue4 == 0.00 {
		TPaid4 = 0.00
	} else {
		TPaid4 = I335 / F335 * TDue4
	}
	if F335*TDue5 == 0.00 {
		TPaid5 = 0.00
	} else {
		TPaid5 = I335 / F335 * TDue5
	}
	if F335*TDue6 == 0.00 {
		TPaid6 = 0.00
	} else {
		TPaid6 = I335 / F335 * TDue6
	}
	if F335*TDue7 == 0.00 {
		TPaid7 = 0.00
	} else {
		TPaid7 = I335 / F335 * TDue7
	}

	if F336*TDue8 == 0.00 {
		TPaid8 = 0.00
	} else {
		TPaid8 = I336 / F336 * TDue8
	}
	if F336*TDue9 == 0.00 {
		TPaid9 = 0.00
	} else {
		TPaid9 = I336 / F336 * TDue9
	}
	if F336*TDue10 == 0.00 {
		TPaid10 = 0.00
	} else {
		TPaid10 = I336 / F336 * TDue10
	}
	if F336*TDue11 == 0.00 {
		TPaid11 = 0.00
	} else {
		TPaid11 = I336 / F336 * TDue11
	}
	if F336*TDue12 == 0.00 {
		TPaid12 = 0.00
	} else {
		TPaid12 = I336 / F336 * TDue12
	}
	if F336*TDue13 == 0.00 {
		TPaid13 = 0.00
	} else {
		TPaid13 = I336 / F336 * TDue13
	}
	if F336*TDue14 == 0.00 {
		TPaid14 = 0.00
	} else {
		TPaid14 = I336 / F336 * TDue14
	}

	EUnpaid1 := TDue1 - TPaid1
	EUnpaid2 := TDue2 - TPaid2
	EUnpaid3 := TDue3 - TPaid3
	EUnpaid4 := TDue4 - TPaid4
	EUnpaid5 := TDue5 - TPaid5
	EUnpaid6 := TDue6 - TPaid6
	EUnpaid7 := TDue7 - TPaid7
	EUnpaid8 := TDue8 - TPaid8
	EUnpaid9 := TDue9 - TPaid9
	EUnpaid10 := TDue10 - TPaid10
	EUnpaid11 := TDue11 - TPaid11
	EUnpaid12 := TDue12 - TPaid12
	EUnpaid13 := TDue13 - TPaid13
	EUnpaid14 := TDue14 - TPaid14

	TotalTDue := TDue1 + TDue2 + TDue3 + TDue4 + TDue5 + TDue6 + TDue7 + TDue8 + TDue9 + TDue10 + TDue11 + TDue12 + TDue13 + TDue14

	TotalTPaid := TPaid1 + TPaid2 + TPaid3 + TPaid4 + TPaid5 + TPaid6 + TPaid7 + TPaid9 + TPaid9 + TPaid10 + TPaid11 + TPaid12 + TPaid13 + TPaid14
	TotalEUnpaid := EUnpaid1 + EUnpaid2 + EUnpaid3 + EUnpaid4 + EUnpaid5 + EUnpaid6 + EUnpaid7 + EUnpaid8 + EUnpaid9 + EUnpaid10 + EUnpaid11 + EUnpaid12 + EUnpaid13 + EUnpaid14
	TotalBegUnpaid := BUnpaid1 + BUnpaid2 + BUnpaid3 + BUnpaid4 + BUnpaid5 + BUnpaid6 + BUnpaid7 + BUnpaid8 + BUnpaid9 + BUnpaid10 + BUnpaid11 + BUnpaid12 + BUnpaid13 + BUnpaid14
	buffer.WriteString(",\"DealFeesAndExpenses\":[")

	dfe1 := DealFeesAndExpenses{Fee: "Owner Trustee", Type: "Fee", BeginningUnpaid: strconv.FormatFloat(BUnpaid1, 'E', -1, 64), CurrentDue: strconv.FormatFloat(CDue1, 'E', -1, 64), TotalDue: strconv.FormatFloat(TDue1, 'E', -1, 64), TotalPaid: strconv.FormatFloat(TPaid1, 'E', -1, 64), EndingUnpaid: strconv.FormatFloat(EUnpaid1, 'E', -1, 64)}
	dd1, _ := json.Marshal(dfe1)
	buffer.WriteString(string(dd1))
	buffer.WriteString(",")

	dfe2 := DealFeesAndExpenses{Fee: "Indenture Trustee Fee", Type: "Fee", BeginningUnpaid: strconv.FormatFloat(BUnpaid2, 'E', -1, 64), CurrentDue: strconv.FormatFloat(CDue2, 'E', -1, 64), TotalDue: strconv.FormatFloat(TDue2, 'E', -1, 64), TotalPaid: strconv.FormatFloat(TPaid2, 'E', -1, 64), EndingUnpaid: strconv.FormatFloat(EUnpaid2, 'E', -1, 64)}
	dd2, _ := json.Marshal(dfe2)
	buffer.WriteString(string(dd2))
	buffer.WriteString(",")

	dfe3 := DealFeesAndExpenses{Fee: "Paying Agent", Type: "Fee", BeginningUnpaid: strconv.FormatFloat(BUnpaid3, 'E', -1, 64), CurrentDue: strconv.FormatFloat(CDue3, 'E', -1, 64), TotalDue: strconv.FormatFloat(TDue3, 'E', -1, 64), TotalPaid: strconv.FormatFloat(TPaid3, 'E', -1, 64), EndingUnpaid: strconv.FormatFloat(EUnpaid3, 'E', -1, 64)}
	dd3, _ := json.Marshal(dfe3)
	buffer.WriteString(string(dd3))
	buffer.WriteString(",")

	dfe4 := DealFeesAndExpenses{Fee: "Custodian", Type: "Fee", BeginningUnpaid: strconv.FormatFloat(BUnpaid4, 'E', -1, 64), CurrentDue: strconv.FormatFloat(CDue4, 'E', -1, 64), TotalDue: strconv.FormatFloat(TDue4, 'E', -1, 64), TotalPaid: strconv.FormatFloat(TPaid4, 'E', -1, 64), EndingUnpaid: strconv.FormatFloat(EUnpaid4, 'E', -1, 64)}
	dd4, _ := json.Marshal(dfe4)
	buffer.WriteString(string(dd4))
	buffer.WriteString(",")

	dfe5 := DealFeesAndExpenses{Fee: "Administrator", Type: "Fee", BeginningUnpaid: strconv.FormatFloat(BUnpaid5, 'E', -1, 64), CurrentDue: strconv.FormatFloat(CDue5, 'E', -1, 64), TotalDue: strconv.FormatFloat(TDue5, 'E', -1, 64), TotalPaid: strconv.FormatFloat(TPaid5, 'E', -1, 64), EndingUnpaid: strconv.FormatFloat(EUnpaid5, 'E', -1, 64)}
	dd5, _ := json.Marshal(dfe5)
	buffer.WriteString(string(dd5))
	buffer.WriteString(",")

	dfe6 := DealFeesAndExpenses{Fee: "Asset Manager", Type: "Fee", BeginningUnpaid: strconv.FormatFloat(BUnpaid6, 'E', -1, 64), CurrentDue: strconv.FormatFloat(CDue6, 'E', -1, 64), TotalDue: strconv.FormatFloat(TDue6, 'E', -1, 64), TotalPaid: strconv.FormatFloat(TPaid6, 'E', -1, 64), EndingUnpaid: strconv.FormatFloat(EUnpaid6, 'E', -1, 64)}
	dd6, _ := json.Marshal(dfe6)
	buffer.WriteString(string(dd6))
	buffer.WriteString(",")

	dfe7 := DealFeesAndExpenses{Fee: "Servicer", Type: "Fee", BeginningUnpaid: strconv.FormatFloat(BUnpaid7, 'E', -1, 64), CurrentDue: strconv.FormatFloat(CDue7, 'E', -1, 64), TotalDue: strconv.FormatFloat(TDue7, 'E', -1, 64), TotalPaid: strconv.FormatFloat(TPaid7, 'E', -1, 64), EndingUnpaid: strconv.FormatFloat(EUnpaid7, 'E', -1, 64)}
	dd7, _ := json.Marshal(dfe7)
	buffer.WriteString(string(dd7))
	buffer.WriteString(",")

	dfe8 := DealFeesAndExpenses{Fee: "Owner Trustee", Type: "Expenses", BeginningUnpaid: strconv.FormatFloat(BUnpaid8, 'E', -1, 64), CurrentDue: strconv.FormatFloat(CDue8, 'E', -1, 64), TotalDue: strconv.FormatFloat(TDue8, 'E', -1, 64), TotalPaid: strconv.FormatFloat(TPaid8, 'E', -1, 64), EndingUnpaid: strconv.FormatFloat(EUnpaid8, 'E', -1, 64)}
	dd8, _ := json.Marshal(dfe8)
	buffer.WriteString(string(dd8))
	buffer.WriteString(",")

	dfe9 := DealFeesAndExpenses{Fee: "Indenture Trustee", Type: "Expenses", BeginningUnpaid: strconv.FormatFloat(BUnpaid9, 'E', -1, 64), CurrentDue: strconv.FormatFloat(CDue9, 'E', -1, 64), TotalDue: strconv.FormatFloat(TDue9, 'E', -1, 64), TotalPaid: strconv.FormatFloat(TPaid9, 'E', -1, 64), EndingUnpaid: strconv.FormatFloat(EUnpaid9, 'E', -1, 64)}
	dd9, _ := json.Marshal(dfe9)
	buffer.WriteString(string(dd9))
	buffer.WriteString(",")

	dfe10 := DealFeesAndExpenses{Fee: "Paying Agent", Type: "Expenses", BeginningUnpaid: strconv.FormatFloat(BUnpaid10, 'E', -1, 64), CurrentDue: strconv.FormatFloat(CDue10, 'E', -1, 64), TotalDue: strconv.FormatFloat(TDue10, 'E', -1, 64), TotalPaid: strconv.FormatFloat(TPaid10, 'E', -1, 64), EndingUnpaid: strconv.FormatFloat(EUnpaid10, 'E', -1, 64)}
	dd10, _ := json.Marshal(dfe10)
	buffer.WriteString(string(dd10))
	buffer.WriteString(",")

	dfe14 := DealFeesAndExpenses{Fee: "Custodian", Type: "Expenses", BeginningUnpaid: strconv.FormatFloat(BUnpaid11, 'E', -1, 64), CurrentDue: strconv.FormatFloat(CDue11, 'E', -1, 64), TotalDue: strconv.FormatFloat(TDue11, 'E', -1, 64), TotalPaid: strconv.FormatFloat(TPaid11, 'E', -1, 64), EndingUnpaid: strconv.FormatFloat(EUnpaid11, 'E', -1, 64)}
	dd14, _ := json.Marshal(dfe14)
	buffer.WriteString(string(dd14))
	buffer.WriteString(",")

	dfe15 := DealFeesAndExpenses{Fee: "Administrator", Type: "Expenses", BeginningUnpaid: strconv.FormatFloat(BUnpaid12, 'E', -1, 64), CurrentDue: strconv.FormatFloat(CDue12, 'E', -1, 64), TotalDue: strconv.FormatFloat(TDue12, 'E', -1, 64), TotalPaid: strconv.FormatFloat(TPaid12, 'E', -1, 64), EndingUnpaid: strconv.FormatFloat(EUnpaid12, 'E', -1, 64)}
	dd15, _ := json.Marshal(dfe15)
	buffer.WriteString(string(dd15))
	buffer.WriteString(",")

	dfe16 := DealFeesAndExpenses{Fee: "Asset Manager", Type: "Expenses", BeginningUnpaid: strconv.FormatFloat(BUnpaid13, 'E', -1, 64), CurrentDue: strconv.FormatFloat(CDue13, 'E', -1, 64), TotalDue: strconv.FormatFloat(TDue13, 'E', -1, 64), TotalPaid: strconv.FormatFloat(TPaid13, 'E', -1, 64), EndingUnpaid: strconv.FormatFloat(EUnpaid13, 'E', -1, 64)}
	dd16, _ := json.Marshal(dfe16)
	buffer.WriteString(string(dd16))
	buffer.WriteString(",")

	dfe17 := DealFeesAndExpenses{Fee: "Servicer", Type: "Expenses", BeginningUnpaid: strconv.FormatFloat(BUnpaid14, 'E', -1, 64), CurrentDue: strconv.FormatFloat(CDue14, 'E', -1, 64), TotalDue: strconv.FormatFloat(TDue14, 'E', -1, 64), TotalPaid: strconv.FormatFloat(TPaid14, 'E', -1, 64), EndingUnpaid: strconv.FormatFloat(EUnpaid14, 'E', -1, 64)}
	dd17, _ := json.Marshal(dfe17)
	buffer.WriteString(string(dd17))
	buffer.WriteString(",")

	dfe13 := DealFeesAndExpenses{Fee: "Total:", Type: "", BeginningUnpaid: strconv.FormatFloat(TotalBegUnpaid, 'E', -1, 64), CurrentDue: strconv.FormatFloat(TotalCDue, 'E', -1, 64), TotalDue: strconv.FormatFloat(TotalTDue, 'E', -1, 64), TotalPaid: strconv.FormatFloat(TotalTPaid, 'E', -1, 64), EndingUnpaid: strconv.FormatFloat(TotalEUnpaid, 'E', -1, 64)}
	dd13, _ := json.Marshal(dfe13)
	buffer.WriteString(string(dd13))
	buffer.WriteString("]")

	/////////////////////////////
	class1 := args[62]
	class2 := args[63]
	class3 := args[64]
	class4 := args[65]
	//Interest payment calcs
	buffer.WriteString(",\"InterestPayments\":[")
	interestowedTotal := interestOwed1 + interestOwed2 + interestOwed3 + interestOwed4
	ifallTotal := ifall1 + ifall2 + ifall3 + ifall4

	ipaid1 := amtPaid8
	ipaid2 := amtPaid9
	ipaid3 := amtPaid19
	ipaid4 := amtPaid7

	iunpaid1 := math.Max(0, interestOwed1+ifall1-ipaid1)
	iunpaid2 := math.Max(0, interestOwed2+ifall2-ipaid2)
	iunpaid3 := math.Max(0, interestOwed3+ifall3-ipaid3)
	iunpaid4 := math.Max(0, interestOwed4+ifall4-ipaid4)

	ipaidTotal := ipaid1 + ipaid2 + ipaid3 + ipaid4
	iunpaidTotal := iunpaid1 + iunpaid2 + iunpaid3 + iunpaid4
	ipsumBeginningBalance := BeginBln1 + BeginBln2 + BeginBln3 + ipBeginBln4

	ipay1 := InterestPayments{Class: class1, InterestRate: strconv.FormatFloat(irate1, 'E', -1, 64) + "%", BeginningBalance: strconv.FormatFloat(BeginBln1, 'E', -1, 64), InterestOwed: strconv.FormatFloat(interestOwed1, 'E', -1, 64), InterestShortfall: strconv.FormatFloat(ifall1, 'E', -1, 64), InterestPaid: strconv.FormatFloat(ipaid1, 'E', -1, 64), InterestUnpaid: strconv.FormatFloat(iunpaid1, 'E', -1, 64)}
	inp1, _ := json.Marshal(ipay1)
	buffer.WriteString(string(inp1))
	buffer.WriteString(",")

	ipay2 := InterestPayments{Class: class2, InterestRate: strconv.FormatFloat(irate2, 'E', -1, 64) + "%", BeginningBalance: strconv.FormatFloat(BeginBln2, 'E', -1, 64), InterestOwed: strconv.FormatFloat(interestOwed2, 'E', -1, 64), InterestShortfall: strconv.FormatFloat(ifall2, 'E', -1, 64), InterestPaid: strconv.FormatFloat(ipaid2, 'E', -1, 64), InterestUnpaid: strconv.FormatFloat(iunpaid2, 'E', -1, 64)}
	inp2, _ := json.Marshal(ipay2)
	buffer.WriteString(string(inp2))
	buffer.WriteString(",")

	ipay3 := InterestPayments{Class: class3, InterestRate: strconv.FormatFloat(irate3, 'E', -1, 64) + "%", BeginningBalance: strconv.FormatFloat(BeginBln3, 'E', -1, 64), InterestOwed: strconv.FormatFloat(interestOwed3, 'E', -1, 64), InterestShortfall: strconv.FormatFloat(ifall3, 'E', -1, 64), InterestPaid: strconv.FormatFloat(ipaid3, 'E', -1, 64), InterestUnpaid: strconv.FormatFloat(iunpaid3, 'E', -1, 64)}
	inp3, _ := json.Marshal(ipay3)
	buffer.WriteString(string(inp3))
	buffer.WriteString(",")

	ipay4 := InterestPayments{Class: class4, InterestRate: strconv.FormatFloat(irate4, 'E', -1, 64) + "%", BeginningBalance: strconv.FormatFloat(ipBeginBln4, 'E', -1, 64), InterestOwed: strconv.FormatFloat(interestOwed4, 'E', -1, 64), InterestShortfall: strconv.FormatFloat(ifall4, 'E', -1, 64), InterestPaid: strconv.FormatFloat(ipaid4, 'E', -1, 64), InterestUnpaid: strconv.FormatFloat(iunpaid4, 'E', -1, 64)}
	inp4, _ := json.Marshal(ipay4)
	buffer.WriteString(string(inp4))
	buffer.WriteString(",")

	ipay10 := InterestPayments{Class: "Total:", InterestRate: "", BeginningBalance: strconv.FormatFloat(ipsumBeginningBalance, 'E', -1, 64), InterestOwed: strconv.FormatFloat(interestowedTotal, 'E', -1, 64), InterestShortfall: strconv.FormatFloat(ifallTotal, 'E', -1, 64), InterestPaid: strconv.FormatFloat(ipaidTotal, 'E', -1, 64), InterestUnpaid: strconv.FormatFloat(iunpaidTotal, 'E', -1, 64)}
	inp10, _ := json.Marshal(ipay10)
	buffer.WriteString(string(inp10))
	buffer.WriteString("]")
	////////////////////////////

	M14, _ := strconv.ParseFloat(args[66], 64) //wrire down - prior data
	M15, _ := strconv.ParseFloat(args[67], 64)
	M16, _ := strconv.ParseFloat(args[68], 64)
	M23, _ := strconv.ParseFloat(args[69], 64)

	princpaid1 := amtPaid13 + amtPaid10
	princpaid2 := amtPaid11 + amtPaid14
	princpaid3 := 0.00
	princpaid4 := 0.00

	K14 := BeginBln1 - princpaid1
	K15 := BeginBln2 - princpaid2
	K16 := BeginBln3 - princpaid3
	K23 := BeginBln4 - princpaid4

	SumK := K14 + K15 + K16 + K23

	K10 := "FALSE"
	K302 := 0.00 //Hardcoded for now
	var K11 float64
	if K10 == "TRUE" {
		K11 = K302 - SumK
	} else {
		K11 = 0
	}

	L25 := K11
	var L14, L15, L16, L23 float64
	if L25 < 0 {
		L23 = math.Max(-K23, L25)
		L16 = math.Max(-K16, L25-(L23))
		L15 = math.Max(-K15, L25-(L16+L23))
		L14 = math.Max(-K14, L25-(L15+L16+L23))
	} else {
		L23 = math.Min(M23, L25)
		L16 = math.Min(M16, L25-(L23))
		L15 = math.Min(M15, L25-(L16+L23))
		L14 = math.Min(M14, L25-(L15+L16+L23))
	}

	//principal payment calcs
	buffer.WriteString(",\"PrincipalPayments\":[")

	endbal1 := BeginBln1 - princpaid1 + L14
	endbal2 := BeginBln2 - princpaid2 + L15
	endbal3 := BeginBln3 - princpaid3 + L16
	endbal4 := BeginBln4 - princpaid4 + L23

	OrigBalTotal := origBln1 + origBln2 + origBln3 + origBln4
	PrincPaidTotal := princpaid1 + princpaid2 + princpaid3 + princpaid4
	EndingbalTotal := endbal1 + endbal2 + endbal3 + endbal4
	SumL := L14 + L15 + L16 + L23

	prin1 := PrincipalPayments{Class: class1, OriginalBalance: strconv.FormatFloat(origBln1, 'E', -1, 64), BeginningBalance: strconv.FormatFloat(BeginBln1, 'E', -1, 64), PrincipalPaid: strconv.FormatFloat(princpaid1, 'E', -1, 64), WriteDownWriteUp: strconv.FormatFloat(L14, 'E', -1, 64), EndingBalance: strconv.FormatFloat(endbal1, 'E', -1, 64)}
	pr1, _ := json.Marshal(prin1)
	buffer.WriteString(string(pr1))
	buffer.WriteString(",")

	prin2 := PrincipalPayments{Class: class2, OriginalBalance: strconv.FormatFloat(origBln2, 'E', -1, 64), BeginningBalance: strconv.FormatFloat(BeginBln2, 'E', -1, 64), PrincipalPaid: strconv.FormatFloat(princpaid2, 'E', -1, 64), WriteDownWriteUp: strconv.FormatFloat(L15, 'E', -1, 64), EndingBalance: strconv.FormatFloat(endbal2, 'E', -1, 64)}
	pr2, _ := json.Marshal(prin2)
	buffer.WriteString(string(pr2))
	buffer.WriteString(",")

	prin3 := PrincipalPayments{Class: class3, OriginalBalance: strconv.FormatFloat(origBln3, 'E', -1, 64), BeginningBalance: strconv.FormatFloat(BeginBln3, 'E', -1, 64), PrincipalPaid: strconv.FormatFloat(princpaid3, 'E', -1, 64), WriteDownWriteUp: strconv.FormatFloat(L16, 'E', -1, 64), EndingBalance: strconv.FormatFloat(endbal3, 'E', -1, 64)}
	pr3, _ := json.Marshal(prin3)
	buffer.WriteString(string(pr3))
	buffer.WriteString(",")

	prin4 := PrincipalPayments{Class: class4, OriginalBalance: strconv.FormatFloat(origBln4, 'E', -1, 64), BeginningBalance: strconv.FormatFloat(BeginBln4, 'E', -1, 64), PrincipalPaid: strconv.FormatFloat(princpaid4, 'E', -1, 64), WriteDownWriteUp: strconv.FormatFloat(L23, 'E', -1, 64), EndingBalance: strconv.FormatFloat(endbal4, 'E', -1, 64)}
	pr4, _ := json.Marshal(prin4)
	buffer.WriteString(string(pr4))
	buffer.WriteString(",")

	prin10 := PrincipalPayments{Class: "Total:", OriginalBalance: strconv.FormatFloat(OrigBalTotal, 'E', -1, 64), BeginningBalance: strconv.FormatFloat(sumBeginningBalance, 'E', -1, 64), PrincipalPaid: strconv.FormatFloat(PrincPaidTotal, 'E', -1, 64), WriteDownWriteUp: strconv.FormatFloat(SumL, 'E', -1, 64), EndingBalance: strconv.FormatFloat(EndingbalTotal, 'E', -1, 64)}
	pr10, _ := json.Marshal(prin10)
	buffer.WriteString(string(pr10))
	buffer.WriteString("]")

	//Below are additional details calcs

	N14 := 0.00
	N15 := 0.00
	N16 := 0.00
	N23 := 0.00

	O14 := math.Min(0, L14) + M14 - N14
	O15 := math.Min(0, L15) + M15 - N15
	O16 := math.Min(0, L16) + M16 - N16
	O23 := math.Min(0, L23) + M23 - N23

	var L56, L57, L58, L65 float64

	L56 = calcColumnL(J56, irate1, interestOwed1)
	L57 = calcColumnL(J57, irate2, interestOwed2)
	L58 = calcColumnL(J58, irate3, interestOwed3)
	L65 = calcColumnL(J65, irate4, interestOwed4)

	M56, _ := strconv.ParseFloat(args[70], 64) //prior data
	M57, _ := strconv.ParseFloat(args[71], 64) //prior data
	M58, _ := strconv.ParseFloat(args[72], 64) //prior data
	M65, _ := strconv.ParseFloat(args[73], 64) //prior data

	N56 := L56 + M56
	N57 := L57 + M57
	N58 := L58 + M58
	N65 := L65 + M65

	O56 := 0.00
	O57 := 0.00
	O58 := 0.00
	O65 := 0.00

	P56 := N56 - O56
	P57 := N57 - O57
	P58 := N58 - O58
	P65 := N65 - O65

	//Additional details calcs
	buffer.WriteString(",\"AdditionalDetails\":[")

	specadj1 := AdditionalDetails{Class: class1, BeginningCumulativeWriteDown: strconv.FormatFloat(M14, 'E', -1, 64), CumulativeWriteDownPaid: strconv.FormatFloat(N14, 'E', -1, 64), EndingCumulativeWriteDown: strconv.FormatFloat(O14, 'E', -1, 64), TotalCumulativeWACShortfall: strconv.FormatFloat(N56, 'E', -1, 64), CumulativeWACShotfallPaid: strconv.FormatFloat(O56, 'E', -1, 64), EndingCumulativeWACShortfall: strconv.FormatFloat(P56, 'E', -1, 64)}
	sa1, _ := json.Marshal(specadj1)
	buffer.WriteString(string(sa1))
	buffer.WriteString(",")

	specadj2 := AdditionalDetails{Class: class2, BeginningCumulativeWriteDown: strconv.FormatFloat(M15, 'E', -1, 64), CumulativeWriteDownPaid: strconv.FormatFloat(N15, 'E', -1, 64), EndingCumulativeWriteDown: strconv.FormatFloat(O15, 'E', -1, 64), TotalCumulativeWACShortfall: strconv.FormatFloat(N57, 'E', -1, 64), CumulativeWACShotfallPaid: strconv.FormatFloat(O57, 'E', -1, 64), EndingCumulativeWACShortfall: strconv.FormatFloat(P57, 'E', -1, 64)}
	sa2, _ := json.Marshal(specadj2)
	buffer.WriteString(string(sa2))
	buffer.WriteString(",")

	specadj3 := AdditionalDetails{Class: class3, BeginningCumulativeWriteDown: strconv.FormatFloat(M16, 'E', -1, 64), CumulativeWriteDownPaid: strconv.FormatFloat(N16, 'E', -1, 64), EndingCumulativeWriteDown: strconv.FormatFloat(O16, 'E', -1, 64), TotalCumulativeWACShortfall: strconv.FormatFloat(N58, 'E', -1, 64), CumulativeWACShotfallPaid: strconv.FormatFloat(O58, 'E', -1, 64), EndingCumulativeWACShortfall: strconv.FormatFloat(P58, 'E', -1, 64)}
	sa3, _ := json.Marshal(specadj3)
	buffer.WriteString(string(sa3))
	buffer.WriteString(",")

	specadj4 := AdditionalDetails{Class: class4, BeginningCumulativeWriteDown: strconv.FormatFloat(M23, 'E', -1, 64), CumulativeWriteDownPaid: strconv.FormatFloat(N23, 'E', -1, 64), EndingCumulativeWriteDown: strconv.FormatFloat(O23, 'E', -1, 64), TotalCumulativeWACShortfall: strconv.FormatFloat(N65, 'E', -1, 64), CumulativeWACShotfallPaid: strconv.FormatFloat(O65, 'E', -1, 64), EndingCumulativeWACShortfall: strconv.FormatFloat(P65, 'E', -1, 64)}
	sa4, _ := json.Marshal(specadj4)
	buffer.WriteString(string(sa4))
	buffer.WriteString(",")

	TotalBC := M14 + M15 + M16 + M23
	TotalCWD := N14 + N15 + N16 + N23
	TotalECWD := O14 + O15 + O16 + O23
	TotalCWS := N56 + N57 + N58 + N65
	TotalCWSP := O56 + O57 + O58 + O65
	TotalEndWS := P56 + P57 + P58 + P65

	specadj10 := AdditionalDetails{Class: "Total:", BeginningCumulativeWriteDown: strconv.FormatFloat(TotalBC, 'E', -1, 64), CumulativeWriteDownPaid: strconv.FormatFloat(TotalCWD, 'E', -1, 64), EndingCumulativeWriteDown: strconv.FormatFloat(TotalECWD, 'E', -1, 64), TotalCumulativeWACShortfall: strconv.FormatFloat(TotalCWS, 'E', -1, 64), CumulativeWACShotfallPaid: strconv.FormatFloat(TotalCWSP, 'E', -1, 64), EndingCumulativeWACShortfall: strconv.FormatFloat(TotalEndWS, 'E', -1, 64)}
	sa10, _ := json.Marshal(specadj10)
	buffer.WriteString(string(sa10))
	buffer.WriteString("]")

	//payment summary calcs
	buffer.WriteString(",\"PaymentSummary\":[")

	cusip1 := args[74]
	cusip2 := args[75]
	cusip3 := args[76]
	cusip4 := args[77]

	pipaid1 := ipaid1 + O56
	pipaid2 := ipaid2 + O57
	pipaid3 := ipaid3 + O58
	pipaid4 := ipaid4 + O65

	ppaid1 := princpaid1 + N14
	ppaid2 := princpaid2 + N15
	ppaid3 := princpaid3 + N16
	ppaid4 := princpaid4 + N23

	ptpaid1 := pipaid1 + ppaid1
	ptpaid2 := pipaid2 + ppaid2
	ptpaid3 := pipaid3 + ppaid3
	ptpaid4 := pipaid4 + ppaid4

	ptotalipaid := pipaid1 + pipaid2 + pipaid3 + pipaid4
	ptotalppaid := ppaid1 + ppaid2 + ppaid3 + ppaid4
	ptotaltpaid := ptpaid1 + ptpaid2 + ptpaid3 + ptpaid4

	PaySu1 := PaymentSummary{Class: class1, CUSIP: cusip1, BeginningBalance: strconv.FormatFloat(BeginBln1, 'E', -1, 64), InterestPaid: strconv.FormatFloat(pipaid1, 'E', -1, 64), PrincipalPaid: strconv.FormatFloat(ppaid1, 'E', -1, 64), TotalPaid: strconv.FormatFloat(ptpaid1, 'E', -1, 64), EndingBalance: strconv.FormatFloat(endbal1, 'E', -1, 64)}
	ps1, _ := json.Marshal(PaySu1)
	buffer.WriteString(string(ps1))
	buffer.WriteString(",")

	PaySu2 := PaymentSummary{Class: class2, CUSIP: cusip2, BeginningBalance: strconv.FormatFloat(BeginBln2, 'E', -1, 64), InterestPaid: strconv.FormatFloat(pipaid2, 'E', -1, 64), PrincipalPaid: strconv.FormatFloat(ppaid2, 'E', -1, 64), TotalPaid: strconv.FormatFloat(ptpaid2, 'E', -1, 64), EndingBalance: strconv.FormatFloat(endbal2, 'E', -1, 64)}
	ps2, _ := json.Marshal(PaySu2)
	buffer.WriteString(string(ps2))
	buffer.WriteString(",")

	PaySu3 := PaymentSummary{Class: class3, CUSIP: cusip3, BeginningBalance: strconv.FormatFloat(BeginBln3, 'E', -1, 64), InterestPaid: strconv.FormatFloat(pipaid3, 'E', -1, 64), PrincipalPaid: strconv.FormatFloat(ppaid3, 'E', -1, 64), TotalPaid: strconv.FormatFloat(ptpaid3, 'E', -1, 64), EndingBalance: strconv.FormatFloat(endbal3, 'E', -1, 64)}
	ps3, _ := json.Marshal(PaySu3)
	buffer.WriteString(string(ps3))
	buffer.WriteString(",")

	PaySu4 := PaymentSummary{Class: class4, CUSIP: cusip4, BeginningBalance: strconv.FormatFloat(BeginBln4, 'E', -1, 64), InterestPaid: strconv.FormatFloat(pipaid4, 'E', -1, 64), PrincipalPaid: strconv.FormatFloat(ppaid4, 'E', -1, 64), TotalPaid: strconv.FormatFloat(ptpaid4, 'E', -1, 64), EndingBalance: strconv.FormatFloat(endbal4, 'E', -1, 64)}
	ps4, _ := json.Marshal(PaySu4)
	buffer.WriteString(string(ps4))
	buffer.WriteString(",")

	PaySu11 := PaymentSummary{Class: "Total:", CUSIP: "", BeginningBalance: strconv.FormatFloat(sumBeginningBalance, 'E', -1, 64), InterestPaid: strconv.FormatFloat(ptotalipaid, 'E', -1, 64), PrincipalPaid: strconv.FormatFloat(ptotalppaid, 'E', -1, 64), TotalPaid: strconv.FormatFloat(ptotaltpaid, 'E', -1, 64), EndingBalance: strconv.FormatFloat(EndingbalTotal, 'E', -1, 64)}
	ps11, _ := json.Marshal(PaySu11)
	buffer.WriteString(string(ps11))
	buffer.WriteString("]")

	////////////////////////
	/*buffer.WriteString(",\"EXCHANGE/EXCHANGEABLE NOTES\":[")

	payBegBln1 := (Notes1 - Notes2) / Notes1 * BeginBln4
	payBegBln2 := (Notes1 - Notes2) / Notes1 * BeginBln5
	payBegBln3 := (Notes1 - Notes2) / Notes1 * BeginBln6
	payBegBln4 := (BeginBln4 + BeginBln5 + BeginBln6) - (payBegBln1 + payBegBln2 + payBegBln3)

	payIpaid1 := (Notes1 - Notes2) / Notes1 * pipaid4
	payIpaid2 := (Notes1 - Notes2) / Notes1 * pipaid5
	payIpaid3 := (Notes1 - Notes2) / Notes1 * pipaid6
	payIpaid4 := (pipaid4 + pipaid5 + pipaid6) - (payIpaid1 + payIpaid2 + payIpaid3)

	payppaid1 := (Notes1 - Notes2) / Notes1 * ppaid4
	payppaid2 := (Notes1 - Notes2) / Notes1 * ppaid5
	payppaid3 := (Notes1 - Notes2) / Notes1 * ppaid6
	payppaid4 := (ppaid4 + ppaid5 + ppaid6) - (payppaid1 + payppaid2 + payppaid3)

	paytpaid1 := (Notes1 - Notes2) / Notes1 * ptpaid4
	paytpaid2 := (Notes1 - Notes2) / Notes1 * ptpaid5
	paytpaid3 := (Notes1 - Notes2) / Notes1 * ptpaid6
	paytpaid4 := (ptpaid4 + ptpaid5 + ptpaid6) - (paytpaid1 + paytpaid2 + paytpaid3)

	payendbal1 := (Notes1 - Notes2) / Notes1 * endbal4
	payendbal2 := (Notes1 - Notes2) / Notes1 * endbal5
	payendbal3 := (Notes1 - Notes2) / Notes1 * endbal6
	payendbal4 := (endbal4 + endbal5 + endbal6) - (payendbal1 + payendbal2 + payendbal3)

	payTotal1 := payBegBln1 + payBegBln2 + payBegBln3 + payBegBln4
	payTotal2 := payIpaid1 + payIpaid2 + payIpaid3 + payIpaid4
	payTotal3 := payppaid1 + payppaid2 + payppaid3 + payppaid4
	payTotal4 := paytpaid1 + paytpaid2 + paytpaid3 + paytpaid4
	payTotal5 := payendbal1 + payendbal2 + payendbal3 + payendbal4

	PaySum4 := PaymentSummary1{Class: class4, CUSIP: cusip4, BeginningBalance: strconv.FormatFloat(payBegBln1, 'E', -1, 64), InterestPaid: strconv.FormatFloat(payIpaid1, 'E', -1, 64), PrincipalPaid: strconv.FormatFloat(payppaid1, 'E', -1, 64), TotalPaid: strconv.FormatFloat(paytpaid1, 'E', -1, 64), EndingBalance: strconv.FormatFloat(payendbal1, 'E', -1, 64)}
	psu4, _ := json.Marshal(PaySum4)
	buffer.WriteString(string(psu4))
	buffer.WriteString(",")

	PaySum5 := PaymentSummary1{Class: class5, CUSIP: cusip5, BeginningBalance: strconv.FormatFloat(payBegBln2, 'E', -1, 64), InterestPaid: strconv.FormatFloat(payIpaid2, 'E', -1, 64), PrincipalPaid: strconv.FormatFloat(payppaid2, 'E', -1, 64), TotalPaid: strconv.FormatFloat(paytpaid2, 'E', -1, 64), EndingBalance: strconv.FormatFloat(payendbal2, 'E', -1, 64)}
	psu5, _ := json.Marshal(PaySum5)
	buffer.WriteString(string(psu5))
	buffer.WriteString(",")

	PaySum6 := PaymentSummary1{Class: class6, CUSIP: cusip6, BeginningBalance: strconv.FormatFloat(payBegBln3, 'E', -1, 64), InterestPaid: strconv.FormatFloat(payIpaid3, 'E', -1, 64), PrincipalPaid: strconv.FormatFloat(payppaid3, 'E', -1, 64), TotalPaid: strconv.FormatFloat(paytpaid3, 'E', -1, 64), EndingBalance: strconv.FormatFloat(payendbal3, 'E', -1, 64)}
	psu6, _ := json.Marshal(PaySum6)
	buffer.WriteString(string(psu6))
	buffer.WriteString(",")

	PaySum8 := PaymentSummary1{Class: class8, CUSIP: cusip8, BeginningBalance: strconv.FormatFloat(payBegBln4, 'E', -1, 64), InterestPaid: strconv.FormatFloat(payIpaid4, 'E', -1, 64), PrincipalPaid: strconv.FormatFloat(payppaid4, 'E', -1, 64), TotalPaid: strconv.FormatFloat(paytpaid4, 'E', -1, 64), EndingBalance: strconv.FormatFloat(payendbal4, 'E', -1, 64)}
	psu8, _ := json.Marshal(PaySum8)
	buffer.WriteString(string(psu8))
	buffer.WriteString(",")

	PaySum11 := PaymentSummary1{Class: "Total:", CUSIP: "", BeginningBalance: strconv.FormatFloat(payTotal1, 'E', -1, 64), InterestPaid: strconv.FormatFloat(payTotal2, 'E', -1, 64), PrincipalPaid: strconv.FormatFloat(payTotal3, 'E', -1, 64), TotalPaid: strconv.FormatFloat(payTotal4, 'E', -1, 64), EndingBalance: strconv.FormatFloat(payTotal5, 'E', -1, 64)}
	psu11, _ := json.Marshal(PaySum11)
	buffer.WriteString(string(psu11))
	buffer.WriteString("]")
	*/
	////////////////////////
	buffer.WriteString(",\"ClassFactorsPer1000\":[")

	var FBegBal1, FBegBal2, FBegBal3, FBegBal4 string
	var FIpaid1, FIpaid2, FIpaid3, FIpaid4 string
	var FPpaid1, FPpaid2, FPpaid3, FPpaid4 string
	var FTpaid1, FTpaid2, FTpaid3, FTpaid4 string
	var FEBal1, FEBal2, FEBal3, FEBal4 string

	deno1 := origBln1
	deno2 := origBln2
	deno3 := origBln3
	deno4 := ipBeginBln4

	if deno1 == 0.00 {
		FBegBal1 = ""
		FIpaid1 = ""
		FPpaid1 = ""
		FTpaid1 = ""
		FEBal1 = ""
	} else {
		FBegBal1 = strconv.FormatFloat(BeginBln1/deno1*1000, 'E', -1, 64)
		FIpaid1 = strconv.FormatFloat(pipaid1/deno1*1000, 'E', -1, 64)
		FPpaid1 = strconv.FormatFloat(ppaid1/deno1*1000, 'E', -1, 64)
		FTpaid1 = strconv.FormatFloat(ptpaid1/deno1*1000, 'E', -1, 64)
		FEBal1 = strconv.FormatFloat(endbal1/deno1*1000, 'E', -1, 64)

	}

	if deno2 == 0.00 {
		FBegBal2 = ""
		FIpaid2 = ""
		FPpaid2 = ""
		FTpaid2 = ""
		FEBal2 = ""
	} else {
		FBegBal2 = strconv.FormatFloat(BeginBln2/deno2*1000, 'E', -1, 64)
		FIpaid2 = strconv.FormatFloat(pipaid2/deno2*1000, 'E', -1, 64)
		FPpaid2 = strconv.FormatFloat(ppaid2/deno2*1000, 'E', -1, 64)
		FTpaid2 = strconv.FormatFloat(ptpaid2/deno2*1000, 'E', -1, 64)
		FEBal2 = strconv.FormatFloat(endbal2/deno2*1000, 'E', -1, 64)

	}

	if deno3 == 0.00 {
		FBegBal3 = ""
		FIpaid3 = ""
		FPpaid3 = ""
		FTpaid3 = ""
		FEBal3 = ""
	} else {
		FBegBal3 = strconv.FormatFloat(BeginBln3/deno3*1000, 'E', -1, 64)
		FIpaid3 = strconv.FormatFloat(pipaid3/deno3*1000, 'E', -1, 64)
		FPpaid3 = strconv.FormatFloat(ppaid3/deno3*1000, 'E', -1, 64)
		FTpaid3 = strconv.FormatFloat(ptpaid3/deno3*1000, 'E', -1, 64)
		FEBal3 = strconv.FormatFloat(endbal3/deno3*1000, 'E', -1, 64)

	}

	if deno4 == 0.00 {
		FBegBal4 = ""
		FIpaid4 = ""
		FPpaid4 = ""
		FTpaid4 = ""
		FEBal4 = ""
	} else {
		FBegBal4 = strconv.FormatFloat(ipBeginBln4/deno4*1000, 'E', -1, 64)
		FIpaid4 = strconv.FormatFloat(pipaid4/deno4*1000, 'E', -1, 64)
		FPpaid4 = strconv.FormatFloat(ppaid4/deno4*1000, 'E', -1, 64)
		FTpaid4 = strconv.FormatFloat(ptpaid4/deno4*1000, 'E', -1, 64)
		FEBal4 = strconv.FormatFloat(ipBeginBln4/deno4*1000, 'E', -1, 64)

	}

	classFact1 := ClassFactors{Class: class1, BeginningBalance: FBegBal1, InterestPaid: FIpaid1, PrincipalPaid: FPpaid1, TotalPaid: FTpaid1, EndingBalance: FEBal1}
	cf1, _ := json.Marshal(classFact1)
	buffer.WriteString(string(cf1))
	buffer.WriteString(",")

	classFact2 := ClassFactors{Class: class2, BeginningBalance: FBegBal2, InterestPaid: FIpaid2, PrincipalPaid: FPpaid2, TotalPaid: FTpaid2, EndingBalance: FEBal2}
	cf2, _ := json.Marshal(classFact2)
	buffer.WriteString(string(cf2))
	buffer.WriteString(",")

	classFact3 := ClassFactors{Class: class3, BeginningBalance: FBegBal3, InterestPaid: FIpaid3, PrincipalPaid: FPpaid3, TotalPaid: FTpaid3, EndingBalance: FEBal3}
	cf3, _ := json.Marshal(classFact3)
	buffer.WriteString(string(cf3))
	buffer.WriteString(",")

	classFact4 := ClassFactors{Class: class4, BeginningBalance: FBegBal4, InterestPaid: FIpaid4, PrincipalPaid: FPpaid4, TotalPaid: FTpaid4, EndingBalance: FEBal4}
	cf4, _ := json.Marshal(classFact4)
	buffer.WriteString(string(cf4))
	buffer.WriteString("]")

	//Collateral summary calcs
	activity1 := "Purchased"
	activity2 := "Funded"
	activity3 := "Capitalized Interest"
	activity4 := "Other(+)"
	activity5 := "Other(+)(non-cash)"
	activity6 := "Paid In Full"
	activity7 := "Sale"
	activity8 := "Liquidation"
	activity9 := "Curtailments"
	activity10 := "Realized Losses"
	activity11 := "Scheduled Principal"
	activity12 := "Other(-)"
	activity13 := "Other(-)(non-cash)"

	count1, _ := strconv.ParseFloat(smap["Collateral Balance - Purchased "], 64)
	count2, _ := strconv.ParseFloat(smap["Collateral Balance - Funded "], 64)
	count3, _ := strconv.ParseFloat(smap["Collateral Balance - Capitalized Interest "], 64)
	count4, _ := strconv.ParseFloat(smap["Collateral Balance - Other(+) "], 64)
	count5, _ := strconv.ParseFloat(smap["Collateral Balance - Other(+)(non-cash) "], 64)
	count6, _ := strconv.ParseFloat(smap["Collateral Balance - Paid In Full "], 64)
	count7, _ := strconv.ParseFloat(smap["Collateral Balance - Sale "], 64)
	count8, _ := strconv.ParseFloat(smap["Collateral Balance - Liquidation "], 64)
	count9, _ := strconv.ParseFloat(smap["Collateral Balance - Curtailments "], 64)
	count10, _ := strconv.ParseFloat(smap["Collateral Balance - Realized Losses "], 64)
	count11, _ := strconv.ParseFloat(smap["Collateral Balance - Scheduled Principal "], 64)
	count12, _ := strconv.ParseFloat(smap["Collateral Balance - Other(-) "], 64)
	count13, _ := strconv.ParseFloat(smap["Collateral Balance - Other(-)(non-cash) "], 64)

	prinbal1, _ := strconv.ParseFloat(smap["Collateral Balance - Purchased"], 64)
	prinbal2, _ := strconv.ParseFloat(smap["Collateral Balance - Funded"], 64)
	prinbal3, _ := strconv.ParseFloat(smap["Collateral Balance - Capitalized Interest"], 64)
	prinbal4, _ := strconv.ParseFloat(smap["Collateral Balance - Other(+)"], 64)
	prinbal5, _ := strconv.ParseFloat(smap["Collateral Balance - Other(+)(non-cash)"], 64)
	prinbal6, _ := strconv.ParseFloat(smap["Collateral Balance - Paid In Full"], 64)
	prinbal7, _ := strconv.ParseFloat(smap["Collateral Balance - Sale"], 64)
	prinbal8, _ := strconv.ParseFloat(smap["Collateral Balance - Liquidation"], 64)
	prinbal9, _ := strconv.ParseFloat(smap["Collateral Balance - Curtailments"], 64)
	prinbal10, _ := strconv.ParseFloat(smap["Collateral Balance - Realized Losses"], 64)
	prinbal11, _ := strconv.ParseFloat(smap["Collateral Balance - Scheduled Principal"], 64)
	prinbal12, _ := strconv.ParseFloat(smap["Collateral Balance - Other(-)"], 64)
	prinbal13, _ := strconv.ParseFloat(smap["Collateral Balance - Other(-)(non-cash)"], 64)

	//prior datas
	cumc1, _ := strconv.ParseFloat(args[78], 64)
	cumc2, _ := strconv.ParseFloat(args[79], 64)
	cumc3, _ := strconv.ParseFloat(args[80], 64)
	cumc4, _ := strconv.ParseFloat(args[81], 64)
	cumc5, _ := strconv.ParseFloat(args[82], 64)
	cumc6, _ := strconv.ParseFloat(args[83], 64)
	cumc7, _ := strconv.ParseFloat(args[84], 64)
	cumc8, _ := strconv.ParseFloat(args[85], 64)
	cumc9, _ := strconv.ParseFloat(args[86], 64)
	cumc10, _ := strconv.ParseFloat(args[87], 64)
	cumc11, _ := strconv.ParseFloat(args[88], 64)
	cumc12, _ := strconv.ParseFloat(args[89], 64)
	cumc13, _ := strconv.ParseFloat(args[90], 64)

	//prior datas
	cump1, _ := strconv.ParseFloat(args[91], 64)
	cump2, _ := strconv.ParseFloat(args[92], 64)
	cump3, _ := strconv.ParseFloat(args[93], 64)
	cump4, _ := strconv.ParseFloat(args[94], 64)
	cump5, _ := strconv.ParseFloat(args[95], 64)
	cump6, _ := strconv.ParseFloat(args[96], 64)
	cump7, _ := strconv.ParseFloat(args[97], 64)
	cump8, _ := strconv.ParseFloat(args[98], 64)
	cump9, _ := strconv.ParseFloat(args[99], 64)
	cump10, _ := strconv.ParseFloat(args[100], 64)
	cump11, _ := strconv.ParseFloat(args[101], 64)
	cump12, _ := strconv.ParseFloat(args[102], 64)
	cump13, _ := strconv.ParseFloat(args[103], 64)

	cumcount1 := count1 + cumc1
	cumcount2 := count2 + cumc2
	cumcount3 := count3 + cumc3
	cumcount4 := count4 + cumc4
	cumcount5 := count5 + cumc5
	cumcount6 := count6 + cumc6
	cumcount7 := count7 + cumc7
	cumcount8 := count8 + cumc8
	cumcount9 := count9 + cumc9
	cumcount10 := count10 + cumc10
	cumcount11 := count11 + cumc11
	cumcount12 := count12 + cumc12
	cumcount13 := count13 + cumc13

	cumprin1 := prinbal1 + cump1
	cumprin2 := prinbal2 + cump2
	cumprin3 := prinbal3 + cump3
	cumprin4 := prinbal4 + cump4
	cumprin5 := prinbal5 + cump5
	cumprin6 := prinbal6 + cump6
	cumprin7 := prinbal7 + cump7
	cumprin8 := prinbal8 + cump8
	cumprin9 := prinbal9 + cump9
	cumprin10 := prinbal10 + cump10
	cumprin11 := prinbal11 + cump11
	cumprin12 := prinbal12 + cump12
	cumprin13 := prinbal13 + cump13

	collateral1 := CollateralSummary{Activity: activity1, Count: strconv.FormatFloat(count1, 'f', 0, 64), PrincipalBalance: strconv.FormatFloat(prinbal1, 'E', -1, 64), CumulativeCount: strconv.FormatFloat(cumcount1, 'f', 0, 64), CumulativePrincipalBalance: strconv.FormatFloat(cumprin1, 'E', -1, 64)}
	collateral2 := CollateralSummary{Activity: activity2, Count: strconv.FormatFloat(count2, 'f', 0, 64), PrincipalBalance: strconv.FormatFloat(prinbal2, 'E', -1, 64), CumulativeCount: strconv.FormatFloat(cumcount2, 'f', 0, 64), CumulativePrincipalBalance: strconv.FormatFloat(cumprin2, 'E', -1, 64)}
	collateral3 := CollateralSummary{Activity: activity3, Count: strconv.FormatFloat(count3, 'f', 0, 64), PrincipalBalance: strconv.FormatFloat(prinbal3, 'E', -1, 64), CumulativeCount: strconv.FormatFloat(cumcount3, 'f', 0, 64), CumulativePrincipalBalance: strconv.FormatFloat(cumprin3, 'E', -1, 64)}
	collateral4 := CollateralSummary{Activity: activity4, Count: strconv.FormatFloat(count4, 'f', 0, 64), PrincipalBalance: strconv.FormatFloat(prinbal4, 'E', -1, 64), CumulativeCount: strconv.FormatFloat(cumcount4, 'f', 0, 64), CumulativePrincipalBalance: strconv.FormatFloat(cumprin4, 'E', -1, 64)}
	collateral5 := CollateralSummary{Activity: activity5, Count: strconv.FormatFloat(count5, 'f', 0, 64), PrincipalBalance: strconv.FormatFloat(prinbal5, 'E', -1, 64), CumulativeCount: strconv.FormatFloat(cumcount5, 'f', 0, 64), CumulativePrincipalBalance: strconv.FormatFloat(cumprin5, 'E', -1, 64)}
	collateral6 := CollateralSummary{Activity: activity6, Count: strconv.FormatFloat(count6, 'f', 0, 64), PrincipalBalance: strconv.FormatFloat(prinbal6, 'E', -1, 64), CumulativeCount: strconv.FormatFloat(cumcount6, 'f', 0, 64), CumulativePrincipalBalance: strconv.FormatFloat(cumprin6, 'E', -1, 64)}
	collateral7 := CollateralSummary{Activity: activity7, Count: strconv.FormatFloat(count7, 'f', 0, 64), PrincipalBalance: strconv.FormatFloat(prinbal7, 'E', -1, 64), CumulativeCount: strconv.FormatFloat(cumcount7, 'f', 0, 64), CumulativePrincipalBalance: strconv.FormatFloat(cumprin7, 'E', -1, 64)}
	collateral8 := CollateralSummary{Activity: activity8, Count: strconv.FormatFloat(count8, 'f', 0, 64), PrincipalBalance: strconv.FormatFloat(prinbal8, 'E', -1, 64), CumulativeCount: strconv.FormatFloat(cumcount8, 'f', 0, 64), CumulativePrincipalBalance: strconv.FormatFloat(cumprin8, 'E', -1, 64)}
	collateral9 := CollateralSummary{Activity: activity9, Count: strconv.FormatFloat(count9, 'f', 0, 64), PrincipalBalance: strconv.FormatFloat(prinbal9, 'E', -1, 64), CumulativeCount: strconv.FormatFloat(cumcount9, 'f', 0, 64), CumulativePrincipalBalance: strconv.FormatFloat(cumprin9, 'E', -1, 64)}
	collateral10 := CollateralSummary{Activity: activity10, Count: strconv.FormatFloat(count10, 'f', 0, 64), PrincipalBalance: strconv.FormatFloat(prinbal10, 'E', -1, 64), CumulativeCount: strconv.FormatFloat(cumcount10, 'f', 0, 64), CumulativePrincipalBalance: strconv.FormatFloat(cumprin10, 'E', -1, 64)}
	collateral11 := CollateralSummary{Activity: activity11, Count: strconv.FormatFloat(count11, 'f', 0, 64), PrincipalBalance: strconv.FormatFloat(prinbal11, 'E', -1, 64), CumulativeCount: strconv.FormatFloat(cumcount11, 'f', 0, 64), CumulativePrincipalBalance: strconv.FormatFloat(cumprin11, 'E', -1, 64)}
	collateral12 := CollateralSummary{Activity: activity12, Count: strconv.FormatFloat(count12, 'f', 0, 64), PrincipalBalance: strconv.FormatFloat(prinbal12, 'E', -1, 64), CumulativeCount: strconv.FormatFloat(cumcount12, 'f', 0, 64), CumulativePrincipalBalance: strconv.FormatFloat(cumprin12, 'E', -1, 64)}
	collateral13 := CollateralSummary{Activity: activity13, Count: strconv.FormatFloat(count13, 'f', 0, 64), PrincipalBalance: strconv.FormatFloat(prinbal13, 'E', -1, 64), CumulativeCount: strconv.FormatFloat(cumcount13, 'f', 0, 64), CumulativePrincipalBalance: strconv.FormatFloat(cumprin13, 'E', -1, 64)}

	c1, _ := json.Marshal(collateral1)
	c2, _ := json.Marshal(collateral2)
	c3, _ := json.Marshal(collateral3)
	c4, _ := json.Marshal(collateral4)
	c5, _ := json.Marshal(collateral5)
	c6, _ := json.Marshal(collateral6)
	c7, _ := json.Marshal(collateral7)
	cc8, _ := json.Marshal(collateral8)
	c9, _ := json.Marshal(collateral9)
	c10, _ := json.Marshal(collateral10)
	c11, _ := json.Marshal(collateral11)
	c12, _ := json.Marshal(collateral12)
	c13, _ := json.Marshal(collateral13)

	buffer.WriteString(",\"CollateralSummary\":[")
	buffer.WriteString(string(c1))
	buffer.WriteString(",")
	buffer.WriteString(string(c2))
	buffer.WriteString(",")
	buffer.WriteString(string(c3))
	buffer.WriteString(",")
	buffer.WriteString(string(c4))
	buffer.WriteString(",")
	buffer.WriteString(string(c5))
	buffer.WriteString(",")
	buffer.WriteString(string(c6))
	buffer.WriteString(",")
	buffer.WriteString(string(c7))
	buffer.WriteString(",")
	buffer.WriteString(string(cc8))
	buffer.WriteString(",")
	buffer.WriteString(string(c9))
	buffer.WriteString(",")
	buffer.WriteString(string(c10))
	buffer.WriteString(",")
	buffer.WriteString(string(c11))
	buffer.WriteString(",")
	buffer.WriteString(string(c12))
	buffer.WriteString(",")
	buffer.WriteString(string(c13))
	buffer.WriteString("]")

	// performance details calcs
	buffer.WriteString(",\"CollateralPerformance\":[")

	pval1, _ := strconv.ParseFloat(args[104], 64)
	pval2, _ := strconv.ParseFloat(args[105], 64)
	pval3, _ := strconv.ParseFloat(args[106], 64)
	pval4, _ := strconv.ParseFloat(args[107], 64)
	pval5, _ := strconv.ParseFloat(args[108], 64)
	pval6, _ := strconv.ParseFloat(args[109], 64)
	pval7, _ := strconv.ParseFloat(args[110], 64)
	pval8, _ := strconv.ParseFloat(args[111], 64)
	pval9, _ := strconv.ParseFloat(args[112], 64)
	pval10, _ := strconv.ParseFloat(args[113], 64)
	pval11, _ := strconv.ParseFloat(args[114], 64)

	pvaltotal := pval1 + pval2 + pval3 + pval4 + pval5 + pval6 + pval7 + pval8 + pval9 + pval10 + pval11

	pbal1 := (pval1 / pvaltotal) * 100
	pbal2 := (pval2 / pvaltotal) * 100
	pbal3 := (pval3 / pvaltotal) * 100
	pbal4 := (pval4 / pvaltotal) * 100
	pbal5 := (pval5 / pvaltotal) * 100
	pbal6 := (pval6 / pvaltotal) * 100
	pbal7 := (pval7 / pvaltotal) * 100
	pbal8 := (pval8 / pvaltotal) * 100
	pbal9 := (pval9 / pvaltotal) * 100
	pbal10 := (pval10 / pvaltotal) * 100
	pbal11 := (pval11 / pvaltotal) * 100

	fmt.Println(pbal7)

	pbaltotal := (pvaltotal / pvaltotal) * 100 // pbal0 + pbal1 + pbal2 + pbal3 + pbal4 + pbal5 + pbal6 + pbal7 + pbal8 + pbal9 + pbal10 + pbal11

	perf1 := PerformanceDetails{Status: "Current", PrincipalBalanceD: strconv.FormatFloat(pval1, 'E', -1, 64), PrincipalBalanceP: strconv.FormatFloat(pbal1, 'E', -1, 64) + "%"}
	pe1, _ := json.Marshal(perf1)
	buffer.WriteString(string(pe1))
	buffer.WriteString(",")

	perf2 := PerformanceDetails{Status: "30-59_days_dq", PrincipalBalanceD: strconv.FormatFloat(pval2, 'E', -1, 64), PrincipalBalanceP: strconv.FormatFloat(pbal2, 'E', -1, 64) + "%"}
	pe2, _ := json.Marshal(perf2)
	buffer.WriteString(string(pe2))
	buffer.WriteString(",")

	perf3 := PerformanceDetails{Status: "60-89_days_dq", PrincipalBalanceD: strconv.FormatFloat(pval3, 'E', -1, 64), PrincipalBalanceP: strconv.FormatFloat(pbal3, 'E', -1, 64) + "%"}
	pe3, _ := json.Marshal(perf3)
	buffer.WriteString(string(pe3))
	buffer.WriteString(",")

	perf4 := PerformanceDetails{Status: "90-119_days_dq", PrincipalBalanceD: strconv.FormatFloat(pval4, 'E', -1, 64), PrincipalBalanceP: strconv.FormatFloat(pbal4, 'E', -1, 64) + "%"}
	pe4, _ := json.Marshal(perf4)
	buffer.WriteString(string(pe4))
	buffer.WriteString(",")

	perf5 := PerformanceDetails{Status: "120-149_days_dq", PrincipalBalanceD: strconv.FormatFloat(pval5, 'E', -1, 64), PrincipalBalanceP: strconv.FormatFloat(pbal5, 'E', -1, 64) + "%"}
	pe5, _ := json.Marshal(perf5)
	buffer.WriteString(string(pe5))
	buffer.WriteString(",")

	perf6 := PerformanceDetails{Status: "150-179_days_dq", PrincipalBalanceD: strconv.FormatFloat(pval6, 'E', -1, 64), PrincipalBalanceP: strconv.FormatFloat(pbal6, 'E', -1, 64) + "%"}
	pe6, _ := json.Marshal(perf6)
	buffer.WriteString(string(pe6))
	buffer.WriteString(",")

	perf7 := PerformanceDetails{Status: "90+_days_dq", PrincipalBalanceD: strconv.FormatFloat(pval7, 'E', -1, 64), PrincipalBalanceP: strconv.FormatFloat(pbal7, 'E', -1, 64) + "%"}
	pe7, _ := json.Marshal(perf7)
	buffer.WriteString(string(pe7))
	buffer.WriteString(",")

	perf8 := PerformanceDetails{Status: "180+_days_dq", PrincipalBalanceD: strconv.FormatFloat(pval8, 'E', -1, 64), PrincipalBalanceP: strconv.FormatFloat(pbal8, 'E', -1, 64) + "%"}
	pe8, _ := json.Marshal(perf8)
	buffer.WriteString(string(pe8))
	buffer.WriteString(",")

	perf9 := PerformanceDetails{Status: "reo", PrincipalBalanceD: strconv.FormatFloat(pval9, 'E', -1, 64), PrincipalBalanceP: strconv.FormatFloat(pbal9, 'E', -1, 64) + "%"}
	pe9, _ := json.Marshal(perf9)
	buffer.WriteString(string(pe9))
	buffer.WriteString(",")

	perf10 := PerformanceDetails{Status: "foreclosure", PrincipalBalanceD: strconv.FormatFloat(pval10, 'E', -1, 64), PrincipalBalanceP: strconv.FormatFloat(pbal10, 'E', -1, 64) + "%"}
	pe10, _ := json.Marshal(perf10)
	buffer.WriteString(string(pe10))
	buffer.WriteString(",")

	perf11 := PerformanceDetails{Status: "forebearance", PrincipalBalanceD: strconv.FormatFloat(pval11, 'E', -1, 64), PrincipalBalanceP: strconv.FormatFloat(pbal11, 'E', -1, 64) + "%"}
	pe11, _ := json.Marshal(perf11)
	buffer.WriteString(string(pe11))
	buffer.WriteString(",")

	perf12 := PerformanceDetails{Status: "Total:", PrincipalBalanceD: strconv.FormatFloat(pvaltotal, 'E', -1, 64), PrincipalBalanceP: strconv.FormatFloat(pbaltotal, 'E', -1, 64) + "%"}
	pe12, _ := json.Marshal(perf12)
	buffer.WriteString(string(pe12))
	buffer.WriteString("]")

	jsonstring1 := args[115]
	jsonstring2 := args[116]
	jsonstring3 := args[117]
	jsonstring4 := args[118]

	var mm1 []PaidInFull
	if err := json.Unmarshal([]byte(jsonstring1), &mm1); err != nil {
		panic(err)
	}
	var endPriBal1 float64 = 0.00

	var mm2 []Modified
	if err := json.Unmarshal([]byte(jsonstring2), &mm2); err != nil {
		panic(err)
	}
	var endPriBal2 float64 = 0.00

	var mm3 []Purchased
	if err := json.Unmarshal([]byte(jsonstring3), &mm3); err != nil {
		panic(err)
	}
	var endPriBal3 float64 = 0.00

	var mm4 []FUnded
	if err := json.Unmarshal([]byte(jsonstring4), &mm4); err != nil {
		panic(err)
	}
	var endPriBal4 float64 = 0.00

	buffer.WriteString(",\"PaidInFull\":[")

	for _, details1 := range mm1 {
		endingprincipalbal, _ := strconv.ParseFloat(details1.PrincipalBalance, 64)
		pif := PaidInFull{LoanID: details1.LoanID, PrincipalBalance: strconv.FormatFloat(endingprincipalbal, 'E', -1, 64)}
		p, _ := json.Marshal(pif)
		endPriBal1 = endPriBal1 + endingprincipalbal
		buffer.WriteString(string(p))
		buffer.WriteString(",")
	}

	if len(mm1) != 0 {
		pif1 := PaidInFull{LoanID: "Total:", PrincipalBalance: strconv.FormatFloat(endPriBal1, 'E', -1, 64)}
		pp1, _ := json.Marshal(pif1)
		buffer.WriteString(string(pp1))
	} else if len(mm1) == 0 {
		pif1 := PaidInFull{LoanID: "", PrincipalBalance: ""}
		pp1, _ := json.Marshal(pif1)
		buffer.WriteString(string(pp1))
	}
	buffer.WriteString("]")

	buffer.WriteString(",\"Modified\":[")

	for _, details1 := range mm2 {
		endingprincipalbal, _ := strconv.ParseFloat(details1.PrincipalBalance, 64)
		pif := Modified{LoanID: details1.LoanID, PrincipalBalance: strconv.FormatFloat(endingprincipalbal, 'E', -1, 64)}
		p, _ := json.Marshal(pif)
		endPriBal2 = endPriBal2 + endingprincipalbal
		buffer.WriteString(string(p))
		buffer.WriteString(",")
	}

	if len(mm2) != 0 {
		pif1 := Modified{LoanID: "Total:", PrincipalBalance: strconv.FormatFloat(endPriBal2, 'E', -1, 64)}
		pp1, _ := json.Marshal(pif1)
		buffer.WriteString(string(pp1))
	} else if len(mm2) == 0 {
		pif1 := Modified{LoanID: "", PrincipalBalance: ""}
		pp1, _ := json.Marshal(pif1)
		buffer.WriteString(string(pp1))
	}
	buffer.WriteString("]")
	buffer.WriteString(",\"Purchased\":[")

	for _, details1 := range mm3 {
		endingprincipalbal, _ := strconv.ParseFloat(details1.PrincipalBalance, 64)
		pif := Purchased{LoanID: details1.LoanID, PrincipalBalance: strconv.FormatFloat(endingprincipalbal, 'E', -1, 64)}
		p, _ := json.Marshal(pif)
		endPriBal3 = endPriBal3 + endingprincipalbal
		buffer.WriteString(string(p))
		buffer.WriteString(",")
	}

	if len(mm3) != 0 {
		pif1 := Purchased{LoanID: "Total:", PrincipalBalance: strconv.FormatFloat(endPriBal3, 'E', -1, 64)}
		pp1, _ := json.Marshal(pif1)
		buffer.WriteString(string(pp1))
	} else if len(mm3) == 0 {
		pif1 := Purchased{LoanID: "", PrincipalBalance: ""}
		pp1, _ := json.Marshal(pif1)
		buffer.WriteString(string(pp1))
	}
	buffer.WriteString("]")

	buffer.WriteString(",\"Funded\":[")

	for _, details1 := range mm4 {
		endingprincipalbal, _ := strconv.ParseFloat(details1.PrincipalBalance, 64)
		pif := FUnded{LoanID: details1.LoanID, PrincipalBalance: strconv.FormatFloat(endingprincipalbal, 'E', -1, 64)}
		p, _ := json.Marshal(pif)
		endPriBal4 = endPriBal4 + endingprincipalbal
		buffer.WriteString(string(p))
		buffer.WriteString(",")
	}

	if len(mm4) != 0 {
		pif1 := FUnded{LoanID: "Total:", PrincipalBalance: strconv.FormatFloat(endPriBal4, 'E', -1, 64)}
		pp1, _ := json.Marshal(pif1)
		buffer.WriteString(string(pp1))
	} else if len(mm4) == 0 {
		pif1 := FUnded{LoanID: "", PrincipalBalance: ""}
		pp1, _ := json.Marshal(pif1)
		buffer.WriteString(string(pp1))
	}
	buffer.WriteString("]")

	buffer.WriteString(",\"Dummy\":[")
	dum := Dummy{Current1: strconv.FormatFloat(current1*100, 'E', -1, 64), Current2: strconv.FormatFloat(current2*100, 'E', -1, 64), AmtPaid: strconv.FormatFloat(amtPaid2, 'E', -1, 64), RevolvingPeriod: isRevolvingPeriod}
	du, _ := json.Marshal(dum)
	buffer.WriteString(string(du))
	buffer.WriteString("]")

	//additional tab calcs -- extra tab in reigo intain model sheet

	buffer.WriteString(",\"Additional\":[")
	buffer.WriteString(pushbuffer("", "Current", "Limit", "Limit Type", "Status"))
	buffer.WriteString(pushbuffer("Eligibility Criteria for the Mortgage Loans", "", "", "", ""))
	additioncurrent1 := "$" + args[131]
	buffer.WriteString(pushbuffer("Maximum Original Principal Balance", additioncurrent1, "$10000000", "MAX", Status(additioncurrent1, "TRUE")))
	additioncurrent2 := "$" + args[132]
	buffer.WriteString(pushbuffer("Maximum Loan Participation (either partial or whole participation) / Check Size per loan", additioncurrent2, "$3500000", "MAX", Status(additioncurrent2, "TRUE")))
	additioncurrent3, _ := strconv.ParseFloat(args[133], 64)
	buffer.WriteString(pushbuffer("Maximum Loan-to-Cost Ratio (at origination)", strconv.FormatFloat(additioncurrent3*100, 'f', 2, 64)+"%", "85.00%", "MAX", Status(strconv.FormatFloat(additioncurrent3*100, 'f', 2, 64)+"%", "TRUE")))
	additioncurrent4, _ := strconv.ParseFloat(args[134], 64)
	buffer.WriteString(pushbuffer("Maximum Loan-to-ARV Ratio", strconv.FormatFloat(additioncurrent4*100, 'f', 2, 64)+"%", "75.00%", "MAX", Status(strconv.FormatFloat(additioncurrent4*100, 'f', 2, 64)+"%", "TRUE")))
	buffer.WriteString(pushbuffer("Maximum Original Term to Maturity", args[135], "24", "MAX", Status(args[135], "TRUE")))
	buffer.WriteString(pushbuffer("Minimum Borrower Credit Score", args[136], "600", "MIN", Status(args[136], "TRUE")))
	additioncurrent7, _ := strconv.ParseFloat(args[137], 64)
	buffer.WriteString(pushbuffer("Minimum Mortgage Loans with Business Purpose", strconv.FormatFloat(additioncurrent7*100, 'f', 2, 64)+"%", "100.00%", "MIN", Status(strconv.FormatFloat(additioncurrent7*100, 'f', 2, 64)+"%", "TRUE")))
	buffer.WriteString(pushbuffer("Mortgage Loans 30+ Days Delinquent", "", "Not Permitted", "", Status("", "FALSE")))
	buffer.WriteString(pushbuffer("Mortgage Loans Secured by Multi-family (5+ units) or Mixed-use Properties", "", "Permitted", "", Status("", "FALSE")))
	buffer.WriteString(pushbuffer("Mortgage Loans for Ground-up or New Construction", "", "Permitted", "", Status("", "FALSE")))
	buffer.WriteString(pushbuffer("Reigo Score", args[138], "12", "MAX", Status(args[138], "TRUE")))
	buffer.WriteString(pushbuffer("Concentration Limits for Whole Mortgage Loans", "", "", "", ""))
	buffer.WriteString(pushbuffer("Loans Secured by First Lien", "", "100.00%", "MAX", Status("", "TRUE")))
	buffer.WriteString(pushbuffer("Maximum Weighted Average Loan-to Cost Ratio", "", "82.00%", "MAX", Status("", "TRUE")))
	buffer.WriteString(pushbuffer("Maximum Non-Zero Weighted Average Loan-to-ARV Ratio", "", "65.00%", "MAX", Status("", "TRUE")))
	buffer.WriteString(pushbuffer("Minimum Mortgage Loans Made to Experienced Borrowers", "", "90.00%", "MIN", Status("", "FALSE")))
	buffer.WriteString(pushbuffer("Maximum Exposures to a Single Guarantor", "", "7.50%", "MAX", Status("", "TRUE")))
	buffer.WriteString(pushbuffer("Maximum Mortgaged Properties in Any Single State", "", "30.00%", "MAX", Status("", "TRUE")))
	buffer.WriteString(pushbuffer("Minimum Weighted Average Borrower Credit Score", "", "690", "MIN", Status("", "FALSE")))
	buffer.WriteString(pushbuffer("Maximum Mortgage Loans Secured by Mixed-use Properties", "", "15.00%", "MAX", Status("", "TRUE")))
	buffer.WriteString(pushbuffer("Maximum Mortgage Loans Secured by Multi-family (5+ units) Properties", "", "25.00%", "MAX", Status("", "TRUE")))
	buffer.WriteString(pushbuffer("Maximum Mortgage Loans Secured by Ground-up or New Construction", "", "25.00%", "MAX", Status("", "TRUE")))

	addition := Additional{"Maximum Mortgage Loans Secured by Loans with Original Principal Balance > $3,500,000", "", "20.00%", "MAX", Status("", "TRUE")}
	add, _ := json.Marshal(addition)
	buffer.WriteString(string(add))
	buffer.WriteString("]}")

	id := args[139]
	fmt.Println("id:::::::::::", id)
	bb, _ := json.Marshal(buffer.String())

	//	fmt.Println("b::::::::::", bb)

	var calculatedData = string(bb)

	fmt.Println(buffer.String())
	//fmt.Println("calculated data::::::::::", calculatedData)

	fig2struct := Reigocode{
		ID:             id,
		CalculatedData: calculatedData,
	}
	fig2structBytes, err := json.Marshal(fig2struct)
	if err != nil {
		fmt.Println("Couldn't marshal data from struct", err)
		return "", fmt.Errorf("Couldn't marshal data from struct")
	}
	fig2structErr := stub.PutState(id, []byte(fig2structBytes))
	if fig2structErr != nil {
		fmt.Println("Couldn't save reigo Characterestic data to ledger", fig2structErr)
		return "", fmt.Errorf("Couldn't save reigo Characterestic data to ledger")
	}
	return fig2struct.CalculatedData, nil

}

//Querying table details by deal id
func GetTable(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	fmt.Println("Entering get table data")

	if len(args) < 1 {
		fmt.Println("Invalid number of arguments")
		return "", errors.New("Missing ID")
	}

	var ID = args[0]
	fmt.Println("id:::::", ID)
	value, err := stub.GetState(ID)
	if err != nil {
		fmt.Println("Couldn't get id "+ID+" from ledger", err)
		return "", errors.New("Missing ID")
	}

	fmt.Println("value:::::: ", string(value))
	return string(value), nil
}

func main() {
	server := &shim.ChaincodeServer{
		CCID:    os.Getenv("CHAINCODE_CCID"),
		Address: os.Getenv("CHAINCODE_ADDRESS"),
		CC:      new(Reigocode),
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	// Start the chaincode external server
	err := server.Start()

	if err != nil {
		fmt.Printf("Error starting Reigo chaincode: %s", err)
	}
	// err := shim.Start(new(Reigocode))
	// {
	// 	fmt.Printf("Error starting reigo chaincode: %s", err)
	// }
}
