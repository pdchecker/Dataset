package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	putils "internal/promark_utils"
	"io/ioutil"
	"log"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	"net"
	"net/http"

	"strings"
	"sync"
)

var LOG_MODE = "test"

type CampaignSmartContract struct {
	contractapi.Contract
}

var (
	cryptoServiceSocketURL     = "external.promark.com:5000"
	cryptoServiceURL           = "http://external.promark.com:5000"
	cryptoParamsRequestURL     = cryptoServiceURL + "/camp"
	userCryptoParamsRequestURL = cryptoServiceURL + "/usercamp"
	logURL                     = "http://logs.promark.com:5003/log"
	// H                          = eutils.
)

type ChaincodeData struct {
	Id        string `json:"id"`
	MSPId     string `json:"mspID"`
	Cert      string `json:"cert"`
	Issuer    string `json:"issuer"`
	Subject   string `json:"subject"`
	Signature string `json:"signature"`
	// Attributes *attrmgr.Attributes `json:"attributes"`
	CN string `json:"cn"`
	OU string `json:"ou"`
}

func (s *CampaignSmartContract) GetChaincodeData(ctx contractapi.TransactionContextInterface) (*ChaincodeData, error) {
	// stub := ctx.GetStub()
	sinfo, err := cid.New(ctx.GetStub())

	if err != nil {
		return nil, fmt.Errorf("failed to get transaction invoker's identity from the chaincode stub: %s", err)
	}

	mspId, err := sinfo.GetMSPID()

	if err != nil {
		return nil, fmt.Errorf("error %s", err)
	}

	id, err := sinfo.GetID()

	if err != nil {
		return nil, fmt.Errorf("error %s", err)
	}

	cert, err := sinfo.GetX509Certificate()

	if err != nil {
		return nil, fmt.Errorf("error %s", err)
	}

	// cn, found, err := sinfo.GetAttributeValue("CN")

	// if err != nil {
	// 	return nil, fmt.Errorf("error %s", err)
	// }

	// if !found {
	// 	return nil, fmt.Errorf("found %s", found)
	// }

	// ou, found, err := sinfo.GetAttributeValue("OU")

	// if err != nil {
	// 	return nil, fmt.Errorf("error %s", err)
	// }

	// if !found {
	// 	return nil, fmt.Errorf("found %s", found)
	// }

	// fmt.Println("OU:" + ou)
	// attrs, err := attrmgr.New().GetAttributesFromCert(cert)

	// if err != nil {
	// 	return nil, fmt.Errorf("error %s", err)
	// }

	// sinfo.getAttributesFromIdemix()

	chaincodeData := ChaincodeData{
		Id:        id,
		MSPId:     mspId,
		Issuer:    cert.Issuer.String(),
		Subject:   cert.Subject.String(),
		Signature: string(cert.Signature),
		// CN:        cn,
		// OU: ou,

		// PublicKey: string(cert.PublicKey),
	}

	// attrVal, found, err = sinfo.GetAttributeValue("role")
	// cert, err := cid.GetX509Certificate(stub)

	// err = proto.Unmarshal(creator, sid)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to unmarshal transaction invoker's identity: %s", err)
	// }
	return &chaincodeData, nil

}

func (s *CampaignSmartContract) GetAllCampaigns(ctx contractapi.TransactionContextInterface) ([]*putils.Campaign, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	// close the resultsIterator when this function is finished
	defer resultsIterator.Close()

	var campaigns []*putils.Campaign
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		// putils.SendLog("queryResponse.Value", string(queryResponse.Value))
		var campaign putils.Campaign
		err = json.Unmarshal(queryResponse.Value, &campaign)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, &campaign)
	}

	return campaigns, nil
}

func (s *CampaignSmartContract) CreateCampaign(ctx contractapi.TransactionContextInterface, camId string, name string, advertiser string, publisher string, startTimeStr string, endTimeStr string, verifierURLStr string, deviceIdsStr string) (*putils.Campaign, error) {
	putils.SendLog("campaignId", camId, LOG_MODE)
	putils.SendLog("campaignName", name, LOG_MODE)
	putils.SendLog("advertiser", advertiser, LOG_MODE)
	putils.SendLog("publisher", publisher, LOG_MODE)

	putils.SendLog("camId", camId, LOG_MODE)

	existing, err := ctx.GetStub().GetState(camId)
	if err != nil {
		return nil, errors.New("Unable to read the world state")
	}

	if existing != nil {
		return nil, fmt.Errorf("Cannot create campaign since its id %s is existed", camId)
	}

	verifierURLs := strings.Split(verifierURLStr, ";")
	deviceIds := strings.Split(deviceIdsStr, ";")
	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		return nil, err
	}

	endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		return nil, err
	}

	campaign, err := CreateCampaign(camId, name, advertiser, publisher, startTime, endTime, verifierURLs, deviceIds)

	if err != nil {
		return nil, err
	}

	campaignJSON, err := json.Marshal(campaign)
	if err != nil {
		return nil, err
	}

	err = ctx.GetStub().PutState(camId, campaignJSON)

	if err != nil {
		return nil, err
	}

	fmt.Println("CampaignJSON:" + string(campaignJSON))

	return campaign, nil
}

func (s *CampaignSmartContract) CreateCampaignAsync(ctx contractapi.TransactionContextInterface, camId string, name string, advertiser string, publisher string, startTimeStr string, endTimeStr string, verifierURLStr string, deviceIdsStr string) (*putils.Campaign, error) {
	putils.SendLog("campaignId", camId, LOG_MODE)
	putils.SendLog("campaignName", name, LOG_MODE)
	putils.SendLog("advertiser", advertiser, LOG_MODE)
	putils.SendLog("publisher", publisher, LOG_MODE)

	putils.SendLog("camId", camId, LOG_MODE)

	existing, err := ctx.GetStub().GetState(camId)
	if err != nil {
		return nil, errors.New("Unable to read the world state")
	}

	if existing != nil {
		return nil, fmt.Errorf("Cannot create campaign since its id %s is existed", camId)
	}

	verifierURLs := strings.Split(verifierURLStr, ";")
	deviceIds := strings.Split(deviceIdsStr, ";")
	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		return nil, err
	}

	endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		return nil, err
	}

	campaign, err := CreateCampaignSocket(camId, name, advertiser, publisher, startTime, endTime, verifierURLs, deviceIds)

	if err != nil {
		return nil, err
	}

	campaignJSON, err := json.Marshal(campaign)
	if err != nil {
		return nil, err
	}

	err = ctx.GetStub().PutState(camId, campaignJSON)

	if err != nil {
		return nil, err
	}

	fmt.Println("CampaignJSON:" + string(campaignJSON))

	return campaign, nil
}

func CreateCampaign(camId string, name string, advertiser string, publisher string, startTime int64, endTime int64, verifierURLs []string, deviceIds []string) (*putils.Campaign, error) {
	fmt.Println("Call RequestCampaignCryptoParamsSocket")

	err := putils.InitializeCampaignCryptoParams(camId, verifierURLs)

	if err != nil {
		return nil, err
	}

	campaign := putils.Campaign{
		Id:           camId,
		Name:         name,
		Advertiser:   advertiser,
		Publisher:    publisher,
		StartTime:    startTime,
		EndTime:      endTime,
		VerifierURLs: verifierURLs,
		DeviceIds:    deviceIds,
	}

	// err = InitializeDeviceProofs(campaign, deviceIds)

	fmt.Println("Closing")
	return &campaign, nil
}

func CreateCampaignSocket(camId string, name string, advertiser string, publisher string, startTime int64, endTime int64, verifierURLs []string, deviceIds []string) (*putils.Campaign, error) {
	fmt.Println("Call RequestCampaignCryptoParamsSocket")

	err := InitializeCampaignCryptoParamsAsync(camId, verifierURLs)

	if err != nil {
		return nil, err
	}

	campaign := putils.Campaign{
		Id:           camId,
		Name:         name,
		Advertiser:   advertiser,
		Publisher:    publisher,
		StartTime:    startTime,
		EndTime:      endTime,
		VerifierURLs: verifierURLs,
		DeviceIds:    deviceIds,
	}

	// err = InitializeDeviceProofs(campaign, deviceIds)

	fmt.Println("Closing")
	return &campaign, nil
}

// func InitializeCampaignCryptoParams(camId string, verifierURLs []string) error {
// 	cryptoParams, err := RequestCampaignCryptoParamsSocket(camId)
// 	if err != nil {
// 		return err
// 	}
// 	// putils.CampaignCryptoParams
// 	// cryptoParams := putils.CampaignCryptoParams{
// 	// 	CamID: camId,
// 	// 	H:     cryptoParams.H,
// 	// }
// 	// putils.SendLog("cryptoParams.H", cryptoParams.H)
// 	fmt.Println("cryptoParams.H:" + cryptoParams.H)
// 	// numVerifiers := len(verifierURLs)

// 	// for i := 0; i < numVerifiers; i++ {
// 	// 	verifierURL := verifierURLs[i]
// 	// 	// requestCreateVerifierCryptoURL := verifierURL
// 	// 	// putils.SendLog("verifierURL", verifierURL)
// 	// 	// putils.SendLog("comURL", requestCreateVerifierCryptoURL)

// 	// 	fmt.Println("Call RequestToCreateVerifierCampaignCryptoParamsSocket: " + verifierURL)
// 	// 	_, err := RequestToCreateVerifierCampaignCryptoParamsSocket(camId, verifierURL, cryptoParams)

// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// }

// 	return nil
// }

func InitializeCampaignCryptoParamsAsync(camId string, verifierURLs []string) error {
	cryptoParams, err := RequestCampaignCryptoParamsSocket(camId)
	if err != nil {
		return err
	}

	// putils.SendLog("cryptoParams.H", cryptoParams.H)
	fmt.Println("cryptoParams.H:" + cryptoParams.H)
	numVerifiers := len(verifierURLs)
	vChannel := make(chan putils.VerifierCryptoChannelResult)
	// defer close(vChannel)

	wg := sync.WaitGroup{}
	wg.Add(numVerifiers)
	for i := 0; i < numVerifiers; i++ {
		verifierURL := verifierURLs[i]
		requestCreateVerifierCryptoURL := verifierURL
		// putils.SendLog("verifierURL", verifierURL)
		// putils.SendLog("comURL", requestCreateVerifierCryptoURL)

		fmt.Println("Call RequestToCreateVerifierCampaignCryptoParamsSocket: " + verifierURL)
		go ConcurrentRequestToCreateVerifierCampaignCryptoParamsSocket(camId, requestCreateVerifierCryptoURL, cryptoParams, vChannel, &wg)
	}

	wg.Wait()

	fmt.Println("Printing results")
	for i := 0; i < numVerifiers; i++ {
		result := <-vChannel
		fmt.Println(result.URL)
		fmt.Println(result.Error)
		fmt.Println(result.VerifierCryptoParams)

		if result.Error != nil {
			return result.Error
		}
	}

	close(vChannel)

	return nil
}

// func InitializeDeviceProofs(campaign *putils.Campaign) error {
// 	for _, deviceId := range campaign.DeviceIds {
// 		proof, err := putils.GenerateProofFromVerifiersSocket(&campaign, deviceId)
// 	}
// }

func RequestCampaignCryptoParamsSocket(camId string) (*putils.CampaignCryptoParams, error) {
	conn, err := net.Dial("tcp", cryptoServiceSocketURL)
	if err != nil {
		// sendLog("Error connecting:", err.Error())

		log.Println("Error connecting:", err.Error())
		return nil, errors.New("ERROR:" + err.Error())
	}

	putils.SendRequest(conn, "create", camId)

	// wait for response
	responseStr, err := bufio.NewReader(conn).ReadString('\n')

	if err != nil {
		// sendLog("Error connecting:", err.Error())
		log.Println("Error after creating:", err.Error())
		return nil, errors.New("Error  after creating:" + err.Error())
	}
	log.Println("Reiceived: " + responseStr)

	response, err := putils.ParseResponse(responseStr)

	if err != nil {
		return nil, errors.New("Error:" + err.Error())
	}
	var cryptoParams putils.CampaignCryptoParams
	err = json.Unmarshal([]byte(response.Data), &cryptoParams)

	return &cryptoParams, nil
}

func ConcurrentRequestToCreateVerifierCampaignCryptoParamsSocket(camId string, requestCreateVerifierCryptoURL string, cryptoParams *putils.CampaignCryptoParams, results chan putils.VerifierCryptoChannelResult, wg *sync.WaitGroup) {
	vCryptoParams, err := RequestToCreateVerifierCampaignCryptoParamsSocket(camId, requestCreateVerifierCryptoURL, cryptoParams)

	wg.Done()
	fmt.Println("Done with " + requestCreateVerifierCryptoURL)
	fmt.Println("vCryptoParams.CamId:" + vCryptoParams.CamId)
	fmt.Println("vCryptoParams.H:" + vCryptoParams.H)
	fmt.Println("vCryptoParams.S:" + vCryptoParams.S)

	results <- putils.VerifierCryptoChannelResult{
		URL:                  requestCreateVerifierCryptoURL,
		VerifierCryptoParams: *vCryptoParams,
		Error:                err,
	}
}

func RequestToCreateVerifierCampaignCryptoParamsSocket(camId string, url string, cryptoParams *putils.CampaignCryptoParams) (*putils.VerifierCryptoParams, error) {
	conn, err := net.Dial("tcp", url)
	if err != nil {
		// sendLog("Error connecting:", err.Error())

		fmt.Println("Error connecting:" + err.Error())
		return nil, errors.New("Error connecting:" + err.Error())
	}

	requestArgs := putils.VerifierCryptoParamsRequest{
		CamId: camId,
		H:     cryptoParams.H,
	}

	jsonArgs, err := json.Marshal(requestArgs)

	putils.SendRequest(conn, "create", string(jsonArgs))
	// wait for response
	// wait for response
	responseStr, err := bufio.NewReader(conn).ReadString('\n')

	if err != nil {
		// sendLog("Error connecting:", err.Error())
		log.Println("Error after creating:", err.Error())
		return nil, errors.New("Error  after creating:" + err.Error())
	}
	fmt.Println("Reiceived From: " + url + "-Response:" + responseStr)

	response, err := putils.ParseResponse(responseStr)

	if err != nil {
		return nil, errors.New("Error:" + err.Error())
	}

	var vCryptoParams putils.VerifierCryptoParams
	err = json.Unmarshal([]byte(response.Data), &vCryptoParams)

	if err != nil {
		fmt.Printf("http.NewRequest() error: %v\n", err)
		return nil, err
	}

	fmt.Println("Returned-vCryptoParams.CamId:" + vCryptoParams.CamId)
	return &vCryptoParams, nil
}

func RequestToCreateVerifierCampaignCryptoParamsHandler(camId string, url string, cryptoParams *putils.CampaignCryptoParams) (*putils.VerifierCryptoParams, error) {
	// func computeCommitment(campId string, url string, i int, cryptoParams CampaignCryptoParams) string {
	//connect to verifier: campId,  H , r
	putils.SendLog("request init verifier crypto at", url, LOG_MODE)

	client := &http.Client{}
	requestArgs := putils.VerifierCryptoParamsRequest{
		CamId: camId,
		H:     cryptoParams.H,
	}

	jsonArgs, err := json.Marshal(requestArgs)
	request := string(jsonArgs)
	reqData, err := http.NewRequest("POST", url, strings.NewReader(request))
	putils.SendLog("request", request, LOG_MODE)
	// putils.SendLog("err", err.Error())
	if err != nil {
		fmt.Printf("http.NewRequest() error: %v\n", err)
		return nil, err
	}

	respJSON, err := client.Do(reqData)
	if err != nil {
		fmt.Printf("http.Do() error: %v\n", err)
		return nil, err
	}
	defer respJSON.Body.Close()

	data, err := ioutil.ReadAll(respJSON.Body)
	putils.SendLog("data", string(data), LOG_MODE)
	if err != nil {
		fmt.Printf("ioutil.ReadAll() error: %v\n", err)
		return nil, err
	}

	fmt.Println("return data all:", string(data))
	var vCryptoParams putils.VerifierCryptoParams
	err = json.Unmarshal([]byte(data), &vCryptoParams)

	if err != nil {
		fmt.Printf("http.NewRequest() error: %v\n", err)
		return nil, err
	}

	return &vCryptoParams, nil
}

func (s *CampaignSmartContract) DeleteCampaignById(ctx contractapi.TransactionContextInterface, camId string) (bool, error) {
	campaignJSON, err := ctx.GetStub().GetState(camId)
	// backupJSON, err := ctx.GetStub().GetState(backupId)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	if campaignJSON == nil {
		return false, fmt.Errorf("the campaign raw id %s does not exist", camId)
	}

	var campaign putils.Campaign
	err = json.Unmarshal(campaignJSON, &campaign)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().DelState(camId)
	if err != nil {
		return false, fmt.Errorf("Failed to delete state:" + err.Error())
	}

	err = s.DeleteVerifiersData(ctx, camId)
	if err != nil {
		return false, err
	}

	return true, err
}

func (s *CampaignSmartContract) GetCampaignById(ctx contractapi.TransactionContextInterface, camId string) (*putils.Campaign, error) {
	campaignJSON, err := ctx.GetStub().GetState(camId)
	// backupJSON, err := ctx.GetStub().GetState(backupId)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if campaignJSON == nil {
		return nil, fmt.Errorf("the campaign raw id %s does not exist", camId)
	}

	var campaign putils.Campaign
	err = json.Unmarshal(campaignJSON, &campaign)
	if err != nil {
		return nil, err
	}

	return &campaign, nil
}

func (s *CampaignSmartContract) DeleteAllCampaigns(ctx contractapi.TransactionContextInterface) (int, error) {
	campaigns, err := s.GetAllCampaigns(ctx)

	if err != nil {
		return -1, err
	}

	var result bool
	count := 1

	for _, cam := range campaigns {
		result, err = s.DeleteCampaignById(ctx, cam.Id)

		if err != nil {
			return -1, err
		}

		if !result {
			return -1, fmt.Errorf("cannot remove campaign %s", cam.Id)
		}
		count += 1
	}

	return count, nil
}

func (s *CampaignSmartContract) DeleteVerifiersData(ctx contractapi.TransactionContextInterface, camId string) error {
	campaign, err := s.GetCampaignById(ctx, camId)

	if err != nil {
		return err
	}

	numVerifiers := len(campaign.VerifierURLs)

	for i := 0; i < numVerifiers; i++ {
		// request clearing data
		err := putils.RequestClearVerifierData(campaign.VerifierURLs[i], camId)
		if err != nil {
			return err
		}
	}

	return nil
}
