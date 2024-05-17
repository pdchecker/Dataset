package main

import (
	"encoding/json"
	"fmt"
	"strings"
	//"strconv"
	"encoding/base64"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type SmartContract struct {
	contractapi.Contract
}

// 인증서 등록기관의 ID, 인증서의 주인 ID, 인증서의 종류, 만기일, 인증서 해쉬
type Certificate struct{
	OrgID string `json:"orgid"`
	UserID string `json:"userid"`
	CertificateName string `json:"certificatename"`
	ExpiredDate string `json:"expiredate"`
	CertHash string `json:"certhash"`
}

// Addcert 인증서 발급, Querycert 인증서 조회, Certdelete 인증서 삭제

// userID, certificateName, expiredDate, cerHash값 받아오고 tx생성한 peer정보 블록에 기록 
func (s *SmartContract) AddCert(ctx contractapi.TransactionContextInterface, userID string, certificateName string, expiredDate string, certHash string) (string, error){
	
	// 인증서 생성하려는 조직의 ID 가져오기
	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to read clientID: %v", err)
	}

	// 가져온 ID string으로 번역
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}
	stringID := string(decodeID)

	// 가져온 정보 중 조직 ID부분만 잘라내기, 12조각 중 1번째 값에 ID있다.
	// slice[0]에 붙어있는 해더 제거
	slice := make([]string, 12)
	slice = strings.Split(stringID, ",")
	headerAndID := slice[0]
	orgID := headerAndID[9:]

	// 구조체에 값 집어넣기
	certificate := Certificate{
		OrgID:   orgID,
		UserID:  userID,
		CertificateName: certificateName,
		ExpiredDate:  expiredDate,
		CertHash: certHash,
	}
	
	// 구조체 JSON형식으로 번역
	certAsBytes, _ := json.Marshal(certificate)

	// 블록에 저장하고 성공메세지 출력
	return "adding certificatione is successfully done", ctx.GetStub().PutState(certHash, certAsBytes)
}

//certHash값으로 기록된 cert 쿼리
func (s *SmartContract) QueryCert(ctx contractapi.TransactionContextInterface, certHash string) (*Certificate, error){
	
	certAsBytes, err := ctx.GetStub().GetState(certHash)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if certAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", certHash)
	}

	cert := new(Certificate)
	_ = json.Unmarshal(certAsBytes, cert)

	return cert, nil
}

//certHash값으로 기록된 cert 쿼리 후 삭제
func (s *SmartContract) DeleteCert(ctx contractapi.TransactionContextInterface, certHash string) error{
	
	exists, err := s.QueryCert(ctx, certHash)
	if err != nil {
		return err
	}
	if exists == nil {
		return fmt.Errorf("the asset %s does not exist", certHash)
	}

	return ctx.GetStub().DelState(certHash)
}



// 1. constructQueryResponseFromIterator constructs a slice of assets from the resultsIterator. getQueryResultForQueryString와 연계된다.
// 쿼리명령문으로 조회된 결과값 string으로 번역하여 append. getQueryResultForQueryString와 연계된다.
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Certificate, error) {
	var certs []*Certificate
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var cert Certificate
		err = json.Unmarshal(queryResult.Value, &cert)
		if err != nil {
			return nil, err
		}
		certs = append(certs, &cert)
	}

	return certs, nil
}

// 2. getQueryResultForQueryString executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results. 
// 쿼리 명령문으로 Value값으로 블록 조회. QueryCertByUserID와 연계된다.
func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Certificate, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// UserId로 cert 쿼리하기 위하여, JSON형식의 쿼리 명령문 작성 함수.
func (s *SmartContract) QueryCertByUserID(ctx contractapi.TransactionContextInterface, userID string) ([]*Certificate, error) {
	queryString := fmt.Sprintf(`{"selector":{"userid":"%s"}}`, userID)
	return getQueryResultForQueryString(ctx, queryString)
}


func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create teamate chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting teamate chaincode: %s", err.Error())
	}
}
