package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"net"
	"sync"

	// "errors"
	"fmt"

	// "log"
	// "hash/fnv"
	"strconv"
	"strings"

	tecc_utils "github.com/tuhoag/elliptic-curve-cryptography-go/utils"
	putils "internal/promark_utils"

	ristretto "github.com/bwesterb/go-ristretto"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric/common/util"
	// eutils "github.com/tuhoag/elliptic-curve-cryptography-go/utils"
)

var LOG_MODE = "test"
var DEVICE_VALIDATION_MODE = "device"
var ALL_VALIDATION_MODE = "all"

type ProofSmartContract struct {
	contractapi.Contract
}

var (
	cryptoServiceURL           = "http://external.promark.com:5000"
	cryptoParamsRequestURL     = cryptoServiceURL + "/camp"
	userCryptoParamsRequestURL = cryptoServiceURL + "/usercamp"
	logURL                     = "http://logs.promark.com:5003/log"
)

func (s *ProofSmartContract) GetAllProofs(ctx contractapi.TransactionContextInterface) ([]*putils.CustomerCampaignTokenTransaction, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	// close the resultsIterator when this function is finished
	defer resultsIterator.Close()

	var proofs []*putils.CustomerCampaignTokenTransaction
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		putils.SendLog("queryResponse.Value", string(queryResponse.Value), LOG_MODE)
		var proof putils.CustomerCampaignTokenTransaction
		err = json.Unmarshal(queryResponse.Value, &proof)
		if err != nil {
			return nil, err
		}
		proofs = append(proofs, &proof)
	}

	return proofs, nil
}

func (s *ProofSmartContract) GetAllProofsInCampaignRange(ctx contractapi.TransactionContextInterface, camId string) ([]*putils.CustomerCampaignTokenTransaction, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	// close the resultsIterator when this function is finished
	defer resultsIterator.Close()

	var proofs []*putils.CustomerCampaignTokenTransaction
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		putils.SendLog("queryResponse.Value", string(queryResponse.Value), LOG_MODE)
		var proof putils.CustomerCampaignTokenTransaction
		err = json.Unmarshal(queryResponse.Value, &proof)
		if err != nil {
			return nil, err
		}
		proofs = append(proofs, &proof)
	}

	return proofs, nil
}

func CallGetCampaignById(ctx contractapi.TransactionContextInterface, camId string) (*putils.Campaign, error) {
	campaignChaincodeArgs := util.ToChaincodeArgs("GetCampaignById", camId)
	response := ctx.GetStub().InvokeChaincode("campaign", campaignChaincodeArgs, "mychannel")

	putils.SendLog("response.Payload", string(response.Payload), LOG_MODE)
	putils.SendLog("response.Status", string(response.Status), LOG_MODE)
	putils.SendLog("response.message", string(response.Message), LOG_MODE)

	if response.Message != "" {
		return nil, fmt.Errorf(response.Message)
	}

	var campaign putils.Campaign

	err := json.Unmarshal([]byte(response.Payload), &campaign)
	if err != nil {
		return nil, err
	}

	return &campaign, nil
}

func (s *ProofSmartContract) GetProofById(ctx contractapi.TransactionContextInterface, proofId string) (*putils.CollectedCustomerProof, error) {
	proofJSON, err := ctx.GetStub().GetState(proofId)
	// backupJSON, err := ctx.GetStub().GetState(backupId)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if proofJSON == nil {
		return nil, fmt.Errorf("the campaign raw id %s does not exist", proofId)
	}

	var proof putils.CollectedCustomerProof
	err = json.Unmarshal(proofJSON, &proof)
	if err != nil {
		return nil, err
	}

	return &proof, nil
}

func (s *ProofSmartContract) DeleteProofById(ctx contractapi.TransactionContextInterface, proofId string) (bool, error) {
	proofJSON, err := ctx.GetStub().GetState(proofId)
	// backupJSON, err := ctx.GetStub().GetState(backupId)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	if proofJSON == nil {
		return false, fmt.Errorf("the proof id %s does not exist", proofId)
	}

	var proof putils.CollectedCustomerProof
	err = json.Unmarshal(proofJSON, &proof)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().DelState(proofId)
	if err != nil {
		return false, fmt.Errorf("Failed to delete state:" + err.Error())
	}

	return true, err
}

func (s *ProofSmartContract) GeneratePoCProof2(ctx contractapi.TransactionContextInterface, camId string, userId string) (*putils.PoCProof, error) {
	putils.SendLog("GeneratePoCProof", "", LOG_MODE)
	putils.SendLog("camId:", camId, LOG_MODE)
	putils.SendLog("userId", userId, LOG_MODE)

	campaignChaincodeArgs := util.ToChaincodeArgs("GetCampaignById", camId)
	response := ctx.GetStub().InvokeChaincode("campaign", campaignChaincodeArgs, "mychannel")

	putils.SendLog("response.Payload", string(response.Payload), LOG_MODE)
	putils.SendLog("response.Status", string(response.Status), LOG_MODE)
	putils.SendLog("response.message", string(response.Message), LOG_MODE)
	// putils.SendLog("response.error", string(response.Error))
	// putils.SendLog("response.message is nil", strconv.FormatBool(response.Message == ""))

	if response.Message != "" {
		return nil, fmt.Errorf(response.Message)
	}

	var campaign putils.Campaign

	err := json.Unmarshal([]byte(response.Payload), &campaign)
	if err != nil {
		return nil, err
	}

	proof, err := putils.GeneratePoCProofFromVerifiers2(&campaign, userId)

	if err != nil {
		return nil, err
	}

	return proof, nil
}

func (s *ProofSmartContract) VerifyPoCProof(ctx contractapi.TransactionContextInterface, camId string, cStr string, rStr string) (bool, error) {
	putils.SendLog("VerifyPoCProof", "", LOG_MODE)
	putils.SendLog("camId", camId, LOG_MODE)

	// get campaign
	campaign, err := CallGetCampaignById(ctx, camId)
	if err != nil {
		return false, err
	}

	// if numVerifiers != len(campaign.VerifierURLs) {
	// 	return false, fmt.Errorf("Input number of verifiers (%s) is different with campaign num verifiers (%s)", numVerifiers, len(campaign.VerifierURLs))
	// }

	numVerifiers := len(campaign.VerifierURLs)
	putils.SendLog("campaign.Id", campaign.Id, LOG_MODE)
	putils.SendLog("campaign.Name", campaign.Name, LOG_MODE)
	putils.SendLog("campaign.VerifierURLs", string(numVerifiers), LOG_MODE)

	if err != nil {
		return false, err
	}

	proof := putils.PoCProof{
		Comm:         cStr,
		R:            rStr,
		NumVerifiers: numVerifiers,
	}

	verificationResult, err := putils.VerifyPoCSocket(campaign, &proof)

	return verificationResult, err
}

func (s *ProofSmartContract) VerifyTPoCProof(ctx contractapi.TransactionContextInterface, camId string, csStr string, rsStr string, hashesStr string, keyStr string) (bool, error) {
	putils.SendLog("VerifyTPoCProof", "", LOG_MODE)
	putils.SendLog("camId", camId, LOG_MODE)

	// get campaign
	campaign, err := CallGetCampaignById(ctx, camId)
	if err != nil {
		return false, err
	}

	putils.SendLog("campaign.Id", campaign.Id, LOG_MODE)
	putils.SendLog("campaign.Name", campaign.Name, LOG_MODE)
	putils.SendLog("campaign.VerifierURLs", string(len(campaign.VerifierURLs)), LOG_MODE)

	if err != nil {
		return false, err
	}

	rs := strings.Split(rsStr, ";")
	comms := strings.Split(csStr, ";")
	hashes := strings.Split(hashesStr, ";")

	if len(rs) != len(campaign.VerifierURLs) || len(comms) != len(campaign.VerifierURLs) || len(hashes) != len(campaign.VerifierURLs) {
		return false, fmt.Errorf("len rs (%d), cs (%d) or hashes (%d) is different to the number of verifiers (%d)", len(rs), len(comms), len(hashes), len(campaign.VerifierURLs))
	}

	proof := putils.TPoCProof{
		TComms: comms,
		TRs:    rs,
		Hashes: hashes,
		Key:    keyStr,
	}

	// return VerifyTPoCProof(campaign, &proof)
	verificationResult, err := putils.VerifyTPoCSocket(campaign, &proof)

	return verificationResult, err
}

func (s *ProofSmartContract) AddCampaignTokenTransaction(ctx contractapi.TransactionContextInterface, camId string, deviceId string, addedTimeStr int64, dCsStr string, dRsStr string, dHashesStr string, dKeyStr string, uCsStr string, uRsStr string, uHashesStr string, uKeyStr string) (*putils.CustomerCampaignTokenTransaction, error) {
	putils.SendLog("AddCampaignToken", "", LOG_MODE)
	putils.SendLog("camId", camId, LOG_MODE)

	// check trans id
	hash := putils.Hash(uCsStr)
	tranId := fmt.Sprintf("p:%s:%s:%s", camId, deviceId, hash)

	tranJSON, err := ctx.GetStub().GetState(tranId)
	if err != nil && tranJSON != nil {
		return nil, fmt.Errorf("transaction %s is existed or has error %s", tranId, err)
	}

	isValid, err := s.VerifyTPoCProof(ctx, camId, dCsStr, dRsStr, dHashesStr, dKeyStr)
	if err != nil {
		return nil, err
	}

	putils.SendLog("Validity", fmt.Sprintf("%t", isValid), LOG_MODE)
	if !isValid {
		return nil, fmt.Errorf("device TPoC does not belong to campaign %s", camId)
	}

	transaction := putils.CustomerCampaignTokenTransaction{
		Id: tranId,
		DeviceTPoC: putils.TPoCProof{
			TComms: strings.Split(dCsStr, ";"),
			TRs:    strings.Split(dRsStr, ";"),
			Hashes: strings.Split(dHashesStr, ";"),
			Key:    dKeyStr,
		},
		CustomerTPoC: putils.TPoCProof{
			TComms: strings.Split(uCsStr, ";"),
			TRs:    strings.Split(uRsStr, ";"),
			Hashes: strings.Split(uHashesStr, ";"),
			Key:    dKeyStr,
		},
		AddedTime: addedTimeStr,
	}

	tranJSON, err = json.Marshal(transaction)
	if err != nil {
		return nil, err
	}

	err = ctx.GetStub().PutState(tranId, tranJSON)
	if err != nil {
		return nil, err
	}

	fmt.Println("CollectedCustomerCampaignTokenTransactionJSON:" + string(tranJSON))

	return &transaction, err
}

func (s *ProofSmartContract) VerifyCampaignProof(ctx contractapi.TransactionContextInterface, camId string, proofId string) (bool, error) {
	putils.SendLog("VerifyCampaignProof", "", LOG_MODE)
	putils.SendLog("camId", camId, LOG_MODE)
	putils.SendLog("proofId", proofId, LOG_MODE)

	_, err := s.GetProofById(ctx, proofId)

	if err != nil {
		return false, err
	}

	// get campaign
	campaign, err := CallGetCampaignById(ctx, camId)
	if err != nil {
		return false, err
	}

	putils.SendLog("campaign.Id", campaign.Id, LOG_MODE)
	putils.SendLog("campaign.Name", campaign.Name, LOG_MODE)
	putils.SendLog("campaign.VerifierURLs", string(len(campaign.VerifierURLs)), LOG_MODE)

	// get proof
	proof, err := s.GetProofById(ctx, proofId)

	if err != nil {
		return false, err
	}

	// putils.SendLog("proof.H", proof.H)
	putils.SendLog("proof.Comm", proof.Comm, LOG_MODE)

	verificationResult, err := VerifyCommitmentSocketAsync(campaign, &proof.CustomerProof)

	return verificationResult, err
}

func VerifyCommitmentSocketAsync(campaign *putils.Campaign, proof *putils.Proof) (bool, error) {
	fmt.Printf("proof.Rs: %s\n", proof.Rs)
	numVerifiers := len(campaign.VerifierURLs)
	vcChannel := make(chan putils.VerifierCommitmentChannelResult)
	wg := sync.WaitGroup{}
	wg.Add(numVerifiers)

	// callinng verifiers to calculate proof.Comm again based on proof.Rs
	for i, verifierURL := range campaign.VerifierURLs {
		// call verifier to compute sub commitment
		go RequestVerificationAsync(campaign.Id, proof.Rs[i], verifierURL, vcChannel, &wg)
	}

	fmt.Println("Calculating total comm")
	// var subComs, randomValues []string
	var C *ristretto.Point
	for i := 0; i < numVerifiers; i++ {
		result := <-vcChannel
		fmt.Printf("URL: %s\n", result.URL)
		// fmt.Printf("Error: %v\n" + result.Error)
		fmt.Printf("Comm: %s\n", result.Comm)

		if result.Error != nil {
			return false, result.Error
		}

		Ci, err := tecc_utils.ConvertStringToPoint(result.Comm)
		if err != nil {
			return false, err
		}
		// CiBytes, _ := b64.StdEncoding.DecodeString(subProof.Comm)
		// Ci = convertBytesToPoint(CiBytes)

		if i == 0 {
			C = Ci
		} else {
			C.Add(C, Ci)

			// putils.SendLog("Current Comm", putils.ConvertPointToString(C), DEBUG_LOG)
		}

		// randomValues = append(randomValues, result.Proof.R)
		// subComs = append(subComs, result.Proof.Comm)
	}

	close(vcChannel)

	comm, err := tecc_utils.ConvertStringToPoint(proof.Comm)
	if err != nil {
		return false, err
	}

	// putils.SendLog("proof.Com", proof.Comm, LOG_MODE)
	// putils.SendLog("calculated Com", b64.StdEncoding.EncodeToString(C.Bytes()), LOG_MODE)
	if C.Equals(comm) {
		return true, nil
	} else {
		return false, nil
	}
}

func RequestVerificationAsync(camId string, r string, url string, results chan putils.VerifierCommitmentChannelResult, wg *sync.WaitGroup) {
	ci, err := RequestVerification(camId, r, url)

	wg.Done()

	fmt.Println("Done with " + url)
	fmt.Printf("ci: %s\n", *ci)

	results <- putils.VerifierCommitmentChannelResult{
		URL:   url,
		Comm:  *ci,
		Error: err,
	}
}

func RequestVerification(camId string, r string, url string) (*string, error) {
	conn, err := net.Dial("tcp", url)
	if err != nil {
		fmt.Println("Error connecting:" + err.Error())
		return nil, errors.New("ERROR:" + err.Error())
	}

	requestArgs := putils.VerificationRequest{
		CamId: camId,
		R:     r,
	}

	jsonArgs, err := json.Marshal(requestArgs)

	putils.SendRequest(conn, "verify", string(jsonArgs))
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

	var verificationResponse putils.VerificationResponse
	err = json.Unmarshal([]byte(response.Data), &verificationResponse)

	if err != nil {
		fmt.Println("error: " + err.Error())
		return nil, err
	}

	// putils.SendLog("verificationResponse.H:", verificationResponse.H, LOG_MODE)
	// putils.SendLog("verificationResponse.s:", verificationResponse.S, LOG_MODE)
	// putils.SendLog("verificationResponse.r:", verificationResponse.R, LOG_MODE)
	// putils.SendLog("verificationResponse.Comm:", verificationResponse.Comm, LOG_MODE)

	return &verificationResponse.Comm, nil
}

func (s *ProofSmartContract) FindTokenTransactionsByTimestamps(ctx contractapi.TransactionContextInterface, startTimeStr string, endTimeStr string) ([]*putils.CustomerCampaignTokenTransaction, error) {
	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		return nil, err
	}

	endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		return nil, err
	}

	queryString := fmt.Sprintf(`{"selector":{"addedTime":{"$gte": %v,"$lte": %v}}}`, startTime, endTime)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var trans []*putils.CustomerCampaignTokenTransaction
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		var tokenTransaction putils.CustomerCampaignTokenTransaction
		err = json.Unmarshal(queryResult.Value, &tokenTransaction)
		if err != nil {
			return nil, err
		}

		trans = append(trans, &tokenTransaction)
	}

	return trans, nil
}

func (s *ProofSmartContract) FindTokenTransactionsByCampaignId(ctx contractapi.TransactionContextInterface, camId string, mode string) ([]*putils.CustomerCampaignTokenTransaction, error) {
	if mode != DEVICE_VALIDATION_MODE && mode != ALL_VALIDATION_MODE {
		return nil, fmt.Errorf("mode '%s' is unsupported", mode)
	}

	campaign, err := CallGetCampaignById(ctx, camId)

	if err != nil {
		return nil, err
	}

	queryString := fmt.Sprintf(`{"selector":{"addedTime":{"$gte": %v,"$lte": %v}}}`, campaign.StartTime, campaign.EndTime)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var trans []*putils.CustomerCampaignTokenTransaction

	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		var tokenTransaction putils.CustomerCampaignTokenTransaction
		err = json.Unmarshal(queryResult.Value, &tokenTransaction)
		if err != nil {
			return nil, err
		}

		// verify
		var tpocs []putils.TPoCProof

		if mode == DEVICE_VALIDATION_MODE {
			tpocs = append(tpocs, tokenTransaction.DeviceTPoC)
		} else if mode == ALL_VALIDATION_MODE {
			tpocs = append(tpocs, tokenTransaction.DeviceTPoC)
			tpocs = append(tpocs, tokenTransaction.CustomerTPoC)
		}

		validity := false
		for _, tpoc := range tpocs {
			validity, err = putils.VerifyTPoCSocket(campaign, &tpoc)

			if err != nil {
				return nil, err
			}
		}

		if validity {
			trans = append(trans, &tokenTransaction)
		}
	}

	return trans, nil
}

func (s *ProofSmartContract) DeleteAllProofs(ctx contractapi.TransactionContextInterface) error {
	proofs, err := s.GetAllProofs(ctx)

	if err != nil {
		return err
	}

	var result bool

	for _, proof := range proofs {
		result, err = s.DeleteProofById(ctx, proof.Id)

		if err != nil {
			return err
		}

		if !result {
			return fmt.Errorf("Cannot remove proof %s", proof.Id)
		}
	}

	return nil
}

// func (t *ProofSmartContract) QueryAssets(ctx contractapi.TransactionContextInterface, queryString string) ([]*putils.CustomerCampaignTokenTransaction, error) {
// 	return getQueryResultForQueryString(ctx, queryString)
// }

// // getQueryResultForQueryString executes the passed in query string.
// func (s *ProofSmartContract) getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*putils.CustomerCampaignTokenTransaction, error) {

// 	resultsIterator, err := ctx.GetStub().GetPrivateDataQueryResult(assetCollection, queryString)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resultsIterator.Close()

// 	results := []*Asset{}

// 	for resultsIterator.HasNext() {
// 		response, err := resultsIterator.Next()
// 		if err != nil {
// 			return nil, err
// 		}
// 		var asset *Asset

// 		err = json.Unmarshal(response.Value, &asset)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
// 		}

// 		results = append(results, asset)
// 	}
// 	return results, nil
// }
