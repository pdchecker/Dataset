package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/tidwall/gjson"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// UsageSlice describes a K8s usage block that
type UsageSlice struct {
	ID         string  `json:"id"`
	Owner      string  `json:"owner_id"`
	Consumer   string  `json:"consumer_id"`
	K8sCPUMins float64 `json:"k8s_cpu_mins"`
	K8sRAMMins float64 `json:"k8s_ram_mins"`
	Timestamp  int64   `json:"epoch"`
	Paid       bool    `json:"is_paid"`
}

// PrometheusBaseURI - local peer Prometheus path
const PrometheusBaseURI = "http://192.168.0.181:3080/api/v1/"

// ORG1K8sNamespace - k8s Namespace for ORG1
const ORG1K8sNamespace = "kube-system"

// ORG2K8sNamespace - k8s Namespace for ORG2
const ORG2K8sNamespace = "kube-system"

// TransferUsageSlice checks for new usage and transfers slice from owner to consumer
func (s *SmartContract) TransferUsageSlice(ctx contractapi.TransactionContextInterface, owner string, consumer string) error {
	//TODO: check that either owner/consumer is executing this
	id := ctx.GetStub().GetTxID()

	//get current state of unpaid UsageSlice of from state
	currentSliceIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("UsageSlice", []string{owner, consumer, "false"})
	if err != nil {
		return err
	}
	defer currentSliceIterator.Close()

	var currentSlice UsageSlice

	//if no slice with the key exists, create it else get from state
	if !currentSliceIterator.HasNext() {
		currentSlice = UsageSlice{
			ID:         id,
			Owner:      owner,
			Consumer:   consumer,
			K8sCPUMins: 0.00,
			K8sRAMMins: 0.00,
			Timestamp:  0,
			Paid:       false,
		}
	} else {
		//get slice from state
		currentSliceQueryResponse, err := currentSliceIterator.Next()
		if err != nil {
			return err
		}

		//make sure that there are no more unpaid slices, can remove this maybe
		if currentSliceIterator.HasNext() {
			return fmt.Errorf("Multiple UsageSlices with same key")
		}

		//parse reponse into object
		err = json.Unmarshal(currentSliceQueryResponse.Value, &currentSlice)
		if err != nil {
			return err
		}
	}

	// calculate timestamp of the last minute rather than latest, this is to make sure that the external request
	// is idempotent so that it does not interfere with consensus
	// TODO: Improve the logic to handle edge case
	now := time.Now()
	_, _, secs := now.Clock()
	timestamp := now.Local().Add(-(time.Second * time.Duration(secs))).Unix()

	fmt.Println("then:", timestamp)

	// Get CPU data of Pods from Prometheus
	prometheusResp, err := GetRequest(fmt.Sprintf("%squery?query=sum(container_cpu_user_seconds_total{namespace=\"%s\"})&timestmap=%d",
		PrometheusBaseURI,
		ORG1K8sNamespace,
		timestamp))
	if err != nil {
		return err
	}

	//parse JSON, TODO: add error handling
	// timestamp := gjson.Result.Float(gjson.Get(prometheusResp, "data.result.0.value.0"))
	cpuMins := gjson.Result.Float(gjson.Get(prometheusResp, "data.result.0.value.1"))

	// Get RAM Usage Data of pods from Prometheus until CPU response timestamp
	// Optimize to make this in a single call
	prometheusResp, err = GetRequest(fmt.Sprintf("%squery?query=sum(sum_over_time(container_memory_usage_bytes{namespace=\"%s\"}[35y]))&timestmap=%d",
		PrometheusBaseURI,
		ORG1K8sNamespace,
		timestamp))
	if err != nil {
		return err
	}
	//TODO: add error handling
	ramMins := gjson.Result.Float(gjson.Get(prometheusResp, "data.result.0.value.1"))

	//TODO: Figure out units
	ramMins = ramMins / 1000000

	//check if there is any usage since the last update
	if cpuMins-currentSlice.K8sCPUMins <= 0 && ramMins-currentSlice.K8sRAMMins <= 0 {
		return fmt.Errorf("there is no resource to log since last sync, usage has to be a positive number")
	}

	slice := UsageSlice{
		ID:         id,
		Owner:      owner,
		Consumer:   consumer,
		K8sCPUMins: cpuMins,
		K8sRAMMins: ramMins,
		Timestamp:  timestamp,
		Paid:       false,
	}

	err = WriteUsageSliceToState(ctx, slice)
	if err != nil {
		return err
	}
	return nil
}

//GetRequest accepts query string and returns output string
func GetRequest(query string) (string, error) {
	resp, err := http.Get(query)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(respBody), nil
}

// WriteUsageSliceToState writes a UsageSlice to the state
func WriteUsageSliceToState(ctx contractapi.TransactionContextInterface, slice UsageSlice) error {
	sliceJSON, err := json.Marshal(slice)
	if err != nil {
		return err
	}

	//Create composite ID of owner, consumer, timestamp so that it easy while getting state
	sliceCompositeKey, _ := ctx.GetStub().CreateCompositeKey("UsageSlice", []string{slice.Owner, slice.Consumer, strconv.FormatBool(slice.Paid)}) // forammted bool to string as only accept strings in arrray

	// add to state
	err = ctx.GetStub().PutState(sliceCompositeKey, sliceJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}
	return nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*UsageSlice, error) {
	// open-ended query of all UsageSlices in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("UsageSlice", []string{})
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*UsageSlice
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset UsageSlice
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
