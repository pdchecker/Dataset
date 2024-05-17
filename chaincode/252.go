package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type CommonResponse struct {
	Key    string `json:"Key"`
	Record string `json:"Record"`
}

type CommonResponsePaginated struct {
	Results      []CommonResponse `json:"Results"`
	RecordsCount string           `json:"RecordsCount"`
	Bookmark     string           `json:"Bookmark"`
}

// Mango query: {"selector": {"type": "FundRequest"}}, PageSize: 10, Bookmark: 'sfdrr4wereaf'
func (s *SmartContract) CommonQuery(ctx contractapi.TransactionContextInterface, queryString string) ([]CommonResponse, error) {

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []CommonResponse{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		queryResult := CommonResponse{Key: queryResponse.Key, Record: string(queryResponse.Value)}
		results = append(results, queryResult)
	}

	return results, nil
}

func (s *SmartContract) CommonQueryPagination(ctx contractapi.TransactionContextInterface, arg string) (*CommonResponsePaginated, error) {

	var args []string

	err := json.Unmarshal([]byte(arg), &args)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	if len(args) < 3 {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 3")
	}

	queryString := args[0]
	pageSize, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil || pageSize <= 0 {
		return nil, fmt.Errorf("Invalid page size!")
	}
	bookmark := args[2]

	resultsIterator, responseMetadata, err := ctx.GetStub().GetQueryResultWithPagination(queryString, int32(pageSize), bookmark)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	commonResponsePaginated := new(CommonResponsePaginated)
	commonResponsePaginated.RecordsCount = fmt.Sprintf("%v", responseMetadata.FetchedRecordsCount)
	commonResponsePaginated.Bookmark = responseMetadata.Bookmark

	results := []CommonResponse{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		queryResult := CommonResponse{Key: queryResponse.Key, Record: string(queryResponse.Value)}
		results = append(results, queryResult)
	}

	commonResponsePaginated.Results = results

	return commonResponsePaginated, nil
}

//returns all transactions between 2 dates.
func (s *SmartContract) QueryForTransactionRange(ctx contractapi.TransactionContextInterface, arg string) ([]byte, error) {
	InfoLogger.Printf("*************** QueryForTransactionRange Started ***************")
	InfoLogger.Printf("args received:", arg)

	var args []string

	err := json.Unmarshal([]byte(arg), &args)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	if len(args) != 2 {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 2")
	} else if len(args[0]) <= 0 {
		return nil, fmt.Errorf("1st argument must be a non-empty string")
	} else if len(args[1]) <= 0 {
		return nil, fmt.Errorf("2nd argument must be a non-empty string")
	}

	fromDate, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("Error converting date " + err.Error())
	}

	toDate, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("Error converting date " + err.Error())
	}

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"Transaction\", \"date\":{ \"$and\":[{ \"$gt\":%d }, {\"$lt\":%d}]}}}", fromDate, toDate)

	queryResults, err := GetQueryResultForQueryString(ctx, queryString)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return queryResults, nil
}

func GetQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]byte, error) {

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)

	if err != nil {
		return nil, err
	}
	if resultsIterator != nil {
		defer resultsIterator.Close()
	}

	buffer, err := ConstructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func ConstructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false

	if resultsIterator != nil {
		for resultsIterator.HasNext() {

			queryResponse, err := resultsIterator.Next()
			if err != nil {
				return nil, err
			}
			// Add a comma before array members, suppress it for the first array member
			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}
			buffer.WriteString("{\"Key\":")
			buffer.WriteString("\"")
			buffer.WriteString(queryResponse.Key)
			buffer.WriteString("\"")

			buffer.WriteString(", \"Record\":")
			buffer.WriteString(string(queryResponse.Value))
			buffer.WriteString("}")
			bArrayMemberAlreadyWritten = true
		}
	}

	buffer.WriteString("]")
	return &buffer, nil
}

func TempGetQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]byte, error) {

	InfoLogger.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := TempConstructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	InfoLogger.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func TempConstructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		InfoLogger.Printf("queryResponse")
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return &buffer, nil
}

func (s *SmartContract) GeneralQueryFunction(ctx contractapi.TransactionContextInterface, arg string) ([]byte, error) {
	InfoLogger.Printf("*************** generalQueryFunction Started ***************")

	queryString := arg

	queryResults, err := GetQueryResultForQueryString(ctx, queryString)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	InfoLogger.Printf("*************** generalQueryFunction Successfull ***************")
	return queryResults, nil
}

//General pagination query
func (s *SmartContract) GeneralQueryFunctionPagination(ctx contractapi.TransactionContextInterface, arg string) ([]byte, error) {
	InfoLogger.Printf("*************** generalQueryFunctionPagination Started ***************")

	var args []string

	err := json.Unmarshal([]byte(arg), &args)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	if len(args) < 3 {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 3")
	}

	queryString := args[0]
	pageSize, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	bookmark := args[2]

	queryResults, err := GetQueryResultForQueryStringWithPagination(ctx, queryString, int32(pageSize), bookmark)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	InfoLogger.Printf("*************** generalQueryFunctionPagination Successfull ***************")
	return queryResults, nil
}

// =========================================================================================
// getQueryResultForQueryStringWithPagination executes the passed in query string with
// pagination info. Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func GetQueryResultForQueryStringWithPagination(ctx contractapi.TransactionContextInterface, queryString string, pageSize int32, bookmark string) ([]byte, error) {

	InfoLogger.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, responseMetadata, err := ctx.GetStub().GetQueryResultWithPagination(queryString, pageSize, bookmark)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := ConstructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	bufferWithPaginationInfo := AddPaginationMetadataToQueryResults(buffer, responseMetadata)

	InfoLogger.Printf("- getQueryResultForQueryString queryResult:\n%s\n", bufferWithPaginationInfo.String())

	return buffer.Bytes(), nil
}

// ===========================================================================================
// addPaginationMetadataToQueryResults adds QueryResponseMetadata, which contains pagination
// info, to the constructed query results
// ===========================================================================================
func AddPaginationMetadataToQueryResults(buffer *bytes.Buffer, responseMetadata *pb.QueryResponseMetadata) *bytes.Buffer {

	buffer.WriteString("#[{\"ResponseMetadata\":{\"RecordsCount\":")
	buffer.WriteString("\"")
	buffer.WriteString(fmt.Sprintf("%v", responseMetadata.FetchedRecordsCount))
	buffer.WriteString("\"")
	buffer.WriteString(", \"Bookmark\":")
	buffer.WriteString("\"")
	buffer.WriteString(responseMetadata.Bookmark)
	buffer.WriteString("\"}}]")

	return buffer
}

//get all the transactions
func (s *SmartContract) GetTransaction(ctx contractapi.TransactionContextInterface) ([]byte, error) {
	InfoLogger.Printf("*************** getTransaction Started ***************")

	//getusercontext to populate the required data
	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return nil, fmt.Errorf("Error getting transaction creator: " + err.Error())
	}
	mspId, commonName, _ := getTxCreatorInfo(ctx, creator)

	queryString := gqs([]string{"docType", "Transaction", "from", commonName})

	InfoLogger.Printf("current logged in user:", commonName, "with mspId:", mspId)

	queryResults, err := GetQueryResultForQueryString(ctx, queryString)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	InfoLogger.Printf("*************** getTransaction Successful ***************")
	return queryResults, nil
}

// query callback representing the query of a chaincode
func (s *SmartContract) QueryByKey(ctx contractapi.TransactionContextInterface, arg string) ([]byte, error) {

	var args []string

	err := json.Unmarshal([]byte(arg), &args)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 1")
	}

	requestAsBytes, _ := ctx.GetStub().GetState(args[0])
	if requestAsBytes == nil {
		return nil, fmt.Errorf("No data exists for the key: " + args[0])
	}
	return requestAsBytes, nil
}

///to get balence of any org
func (s *SmartContract) GetBalance(ctx contractapi.TransactionContextInterface) ([]byte, error) {

	//getusercontext to populate the required data
	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return nil, fmt.Errorf("Error getting transaction creator: " + err.Error())
	}
	_, commonName, _ := getTxCreatorInfo(ctx, creator)

	allBalances := make(map[string]float64)
	amount := 0.0

	tokenBalanceAsBytes, _ := ctx.GetStub().GetState(commonName)

	if tokenBalanceAsBytes != nil {
		amount, _ = strconv.ParseFloat(string(tokenBalanceAsBytes), 64)
		allBalances["balance"] = amount
	}

	balJSON, _ := json.Marshal(allBalances)
	jsonStr := string(balJSON)
	return []byte(jsonStr), nil
}

//list of all the transaction on perticular project
func (s *SmartContract) GetProjectTransactions(ctx contractapi.TransactionContextInterface, arg string) ([]byte, error) {
	InfoLogger.Printf("*************** getProjectTransactions Started ***************")

	var args []string

	err := json.Unmarshal([]byte(arg), &args)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	queryString := ""

	//getusercontext to populate the required data
	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return nil, fmt.Errorf("Error getting transaction creator: " + err.Error())
	}
	mspId, commanName, _ := getTxCreatorInfo(ctx, creator)

	InfoLogger.Printf("current logged in user:", commanName, "with mspId:", mspId)
	queryString = gqs([]string{"docType", "Transaction", "objRef", args[0]})

	queryResults, err := GetQueryResultForQueryString(ctx, queryString)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	InfoLogger.Printf("*************** getProjectTransactions Successfull ***************")
	return queryResults, nil
}

type Record struct {
	Qty             float64 `json:qty`
	Balance         float64 `json:balance`
	Corporate       string  `json:corporate`
	Id              string  `json:_id`
	ProjectCount    float64 `json:projectCount`
}

//getCorporateDetails
func (s *SmartContract) GetCorporateDetails(ctx contractapi.TransactionContextInterface) ([]byte, error) {
	InfoLogger.Printf("*************** getCorporateDetails Started ***************")

	organisations := getCorporates(ctx)
	recordObj := Record{}
	var endResult []Record
	var sum float64

	for _, org := range organisations {

		//get value of how much amount that csr has issued to corporate
		queryString := "{\"selector\":{\"docType\":\"Transaction\",\"txType\":\"AssignToken\",\"to\":\"" + org + "\"},\"fields\":[\"qty\"]}"
		queryResults, err := TempGetQueryResultForQueryString(ctx, queryString)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		var testObj []Record
		err = json.Unmarshal(queryResults, &testObj)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		sum = 0.0
		for i := 0; i < len(testObj); i++ {
			sum = math.Round((sum+testObj[i].Qty)*100) / 100
		}

		args := org
		//get balance of corresponding corporates
		result1, err := s.GetBalanceCorporate(ctx, args)
		result2 := []byte(result1)
		InfoLogger.Printf(string(result1))
		json.Unmarshal(result2, &recordObj)
		balance := recordObj.Balance

		//get the no of ongoing projects they are working
		list1 := strings.Split(org, ".")
		res := list1[0] + "\\\\." + list1[1] + "\\\\." + list1[2] + "\\\\." + list1[3]
		queryString = "{\"selector\":{\"docType\":\"Project\",\"contributors." + res + "\":{\"$exists\":true}},\"fields\":[\"_id\"]}"
		queryResults, err = GetQueryResultForQueryString(ctx, queryString)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}

		err = json.Unmarshal(queryResults, &testObj)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		projectCount := len(testObj)

		//create new object to send
		recordObj = Record{}
		recordObj.Corporate = org
		recordObj.Qty = sum
		recordObj.Balance = balance
		recordObj.ProjectCount = float64(projectCount)

		endResult = append(endResult, recordObj)
	}

	bytAr, err := json.Marshal(endResult)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	InfoLogger.Printf("*************** getCorporateDetails Successfull ***************")
	return bytAr, nil
}

//get balance of only corporate
func (s *SmartContract) GetBalanceCorporate(ctx contractapi.TransactionContextInterface, arg string) (string, error) {
	InfoLogger.Printf("*************** getBalanceCorporate Started ***************")
	InfoLogger.Printf("args received:", arg)

	allBalances := make(map[string]float64)
	amount := 0.0
	orgName := arg

	tokenBalanceAsBytes, _ := ctx.GetStub().GetState(orgName)

	if tokenBalanceAsBytes != nil {
		amount, _ = strconv.ParseFloat(string(tokenBalanceAsBytes), 64)
		allBalances["balance"] = amount
	}

	balJSON, _ := json.Marshal(allBalances)
	jsonStr := string(balJSON)
	InfoLogger.Printf("*************** getBalanceCorporate Successfull ***************")
	return jsonStr, nil
}

func (s *SmartContract) GetAllCorporates(ctx contractapi.TransactionContextInterface) ([]string, error) {

	corporatesBytes, _ := ctx.GetStub().GetState("corporates")

	var result []string

	if corporatesBytes == nil || len(string(corporatesBytes)) <= 2 {
		return nil, nil
	}

	err := json.Unmarshal(corporatesBytes, &result)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	
	return result, nil
}
