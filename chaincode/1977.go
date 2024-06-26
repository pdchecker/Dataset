package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dovetail-lab/fabric-chaincode/common"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/pkg/errors"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
)

// Create a new logger
var logger = log.ChildLogger(log.RootLogger(), "activity-fabric-putall")

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

func init() {
	_ = activity.Register(&Activity{}, New)
}

// Activity is a stub for executing Hyperledger Fabric put operations
type Activity struct {
}

// New creates a new Activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &Activity{}, nil
}

// Metadata implements activity.Activity.Metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	// check input args
	input := &Input{}
	if err = ctx.GetInputObject(input); err != nil {
		return false, err
	}

	if input.StateData == nil {
		logger.Errorf("input data is nil\n")
		output := &Output{Code: 400, Message: "input data is nil"}
		ctx.SetOutputObject(output)
		return false, errors.New(output.Message)
	}
	logger.Debugf("input value type %T: %+v\n", input.StateData, input.StateData)

	// get chaincode stub
	stub, err := common.GetChaincodeStub(ctx)
	if err != nil || stub == nil {
		logger.Errorf("failed to retrieve fabric stub: %+v\n", err)
		output := &Output{Code: 500, Message: err.Error()}
		ctx.SetOutputObject(output)
		return false, err
	}

	var successCount, errorCount int
	var errorKeys []string
	var resultValue []interface{}
	if input.PrivateCollection != "" {
		// store data on a private collection
		for _, v := range input.StateData {
			vmap := v.(map[string]interface{})
			vkey := vmap[common.KeyField].(string)
			if err := storePrivateData(stub, input.PrivateCollection, input.CompositeKeys, vkey, vmap[common.ValueField]); err != nil {
				errorCount++
				errorKeys = append(errorKeys, vkey)
			} else {
				successCount++
				resultValue = append(resultValue, vmap)
			}
		}
	} else {
		// store data on the ledger
		for _, v := range input.StateData {
			vmap := v.(map[string]interface{})
			vkey := vmap[common.KeyField].(string)
			if err := storeData(stub, input.CompositeKeys, vkey, vmap[common.ValueField]); err != nil {
				errorCount++
				errorKeys = append(errorKeys, vkey)
			} else {
				successCount++
				resultValue = append(resultValue, vmap)
			}
		}
	}

	if errorCount > 0 {
		output := &Output{
			Code:    500,
			Message: fmt.Sprintf("failed to store keys: %s", strings.Join(errorKeys, ",")),
			Count:   successCount,
			Errors:  errorCount,
			Result:  resultValue,
		}
		if successCount > 0 {
			// return 300 if partial successs
			output.Code = 300
			ctx.SetOutputObject(output)
			return true, nil
		}
		// return 500 if all failures
		ctx.SetOutputObject(output)
		return false, errors.New(output.Message)
	}
	// return 200 if no errors
	logger.Debugf("set activity output result: %+v\n", resultValue)
	output := &Output{
		Code:    200,
		Message: fmt.Sprintf("stored data on ledger: %+v", resultValue),
		Count:   successCount,
		Errors:  errorCount,
		Result:  resultValue,
	}
	ctx.SetOutputObject(output)
	return true, nil
}

func storePrivateData(ccshim shim.ChaincodeStubInterface, collection string, compositeKeyDefs string, key string, value interface{}) error {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		logger.Errorf("failed to marshal value '%+v', error: %+v\n", value, err)
		return errors.Wrapf(err, "failed to marshal value: %+v", value)
	}

	// store data on a private collection
	if err := ccshim.PutPrivateData(collection, key, jsonBytes); err != nil {
		logger.Errorf("failed to store data in private collection %s: %+v\n", collection, err)
		return errors.Wrapf(err, "failed to store data in private collection %s", collection)
	}
	logger.Debugf("stored in private collection %s, data: %s\n", collection, string(jsonBytes))

	// store composite keys if required
	if len(compositeKeyDefs) == 0 {
		return nil
	}
	compositeKeys := common.ExtractCompositeKeys(ccshim, compositeKeyDefs, key, value)
	if compositeKeys != nil && len(compositeKeys) > 0 {
		for _, k := range compositeKeys {
			cv := []byte{0x00}
			if err := ccshim.PutPrivateData(collection, k, cv); err != nil {
				logger.Errorf("failed to store composite key %s on collection %s: %+v\n", k, collection, err)
			} else {
				logger.Debugf("stored composite key %s on collection %s\n", k, collection)
			}
		}
	}
	return nil
}

func storeData(ccshim shim.ChaincodeStubInterface, compositeKeyDefs string, key string, value interface{}) error {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		logger.Errorf("failed to marshal value '%+v', error: %+v\n", value, err)
		return errors.Wrapf(err, "failed to marshal value: %+v", value)
	}
	// store data on the ledger
	if err := ccshim.PutState(key, jsonBytes); err != nil {
		logger.Errorf("failed to store data on ledger: %+v\n", err)
		return errors.Errorf("failed to store data on ledger: %+v", err)
	}
	logger.Debugf("stored data on ledger: %s\n", string(jsonBytes))

	// store composite keys if required
	if len(compositeKeyDefs) == 0 {
		return nil
	}
	compositeKeys := common.ExtractCompositeKeys(ccshim, compositeKeyDefs, key, value)
	if compositeKeys != nil && len(compositeKeys) > 0 {
		for _, k := range compositeKeys {
			cv := []byte{0x00}
			if err := ccshim.PutState(k, cv); err != nil {
				logger.Errorf("failed to store composite key %s: %+v\n", k, err)
			} else {
				logger.Debugf("stored composite key %s\n", k)
			}
		}
	}
	return nil
}
