package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	perror "github.com/pkg/errors"
)

type Contract struct {
	contractapi.Contract
}

func (c *Contract) Instantiate() {
	fmt.Println("Init")
}

//--------Create a new process
func (c *Contract) StartProcess(ctx TransactionContextInterface, processLocalId string, optionName string, startTime int64, startPosition string, preKey string, spec string) (string, error) {
	ownerOrg, orgErr := ctx.GetOrg()
	if orgErr != nil {
		return "", perror.Errorf("Get org info error. %s", orgErr)
	}
	process := &Process{}
	process.ProcessLocalId = processLocalId
	process.Class = ClassName
	process.State = InProcess
	process.OwnerOrg = ownerOrg
	process.StartTime = startTime
	process.StartPosition = startPosition
	process.OptionName = optionName
	process.Spec = spec
	preKeys, parsingErr := stringToArray(preKey)
	if parsingErr != nil {
		return "", parsingErr
	}
	key := CreateCompositeKey(ownerOrg, processLocalId)
	preKeyField, err := c.createPreKeySlice(ctx, preKeys, key)
	if err != nil {
		return "", err
	} else {
		process.PreKey = preKeyField
	}

	process.Key = key
	return key, ctx.GetProcessLedger().AddProcess(key, process)
}

//--------Complete a new process
func (c *Contract) CompleteProcess(ctx TransactionContextInterface, key string, completeTime int64, completePosition string) error {
	//MSPID is needed to check auth, so cannot call QueryProcess here which may replace ownerOrg with display name
	process, err := ctx.GetProcessLedger().GetProcess(key)
	if err != nil {
		return err
	}
	if !ctx.CheckOrgValid(process.OwnerOrg) {
		return perror.New("Org check failed")
	}
	process.State = Completed
	process.CompleteTime = completeTime
	process.CompletePosition = completePosition
	return ctx.GetProcessLedger().UpdateProcess(key, process)
}

//--------Link an existing process to its previous ones, WILL OVERWRITE if preKey field already exists
func (c *Contract) LinkProcess(ctx TransactionContextInterface, key string, preKey string) error {
	//MSPID is needed to check auth, so cannot call QueryProcess here which may replace ownerOrg with display name
	process, err := ctx.GetProcessLedger().GetProcess(key)
	if err != nil {
		return err
	}
	if !ctx.CheckOrgValid(process.OwnerOrg) {
		return perror.New("Org check failed")
	}
	preKeys, parsingErr := stringToArray(preKey)
	if parsingErr != nil {
		return parsingErr
	}
	preKeyField, err := c.createPreKeySlice(ctx, preKeys, key)
	if err != nil {
		return err
	}
	process.PreKey = preKeyField
	return ctx.GetProcessLedger().UpdateProcess(key, process)
}

//--------Link an existing process to its previous ones, WILL APPEND if preKey field already exists
func (c *Contract) AddLinkedProcess(ctx TransactionContextInterface, key string, preKey string) error {
	//MSPID is needed to check auth, so cannot call QueryProcess here which may replace ownerOrg with display name
	process, err := ctx.GetProcessLedger().GetProcess(key)
	if err != nil {
		return err
	}
	if !ctx.CheckOrgValid(process.OwnerOrg) {
		return perror.New("Org check failed")
	}
	preKeys, parsingErr := stringToArray(preKey)
	if parsingErr != nil {
		return parsingErr
	}
	preKeyField, err := c.createPreKeySlice(ctx, preKeys, key)
	if err != nil {
		return err
	}
	if len(process.PreKey) == 0 {
		process.PreKey = preKeyField
	} else {
		process.PreKey = append(process.PreKey, preKeyField...)
	}

	return ctx.GetProcessLedger().UpdateProcess(key, process)
}

//-----------Query a process, the org field will be replaced with org's display name
func (c *Contract) QueryProcess(ctx TransactionContextInterface, key string) (*Process, error) {
	process, err := ctx.GetProcessLedger().GetProcess(key)
	if err != nil {
		return nil, err
	}
	name, nameErr := c.GetDisplayName(ctx, process.OwnerOrg)
	if nameErr != nil {
		return nil, nameErr
	}
	//replace the ownerOrg field with display name
	if name != "" {
		process.OwnerOrg = name
	}
	return process, nil
}

//---------Get the previous processes within depth of 1
func (c *Contract) PrevProcess(ctx TransactionContextInterface, key string) ([]*Process, error) {
	process, err := c.QueryProcess(ctx, key)
	if err != nil {
		return nil, perror.Errorf("Key not exist. %s", err)
	}
	preKey := process.PreKey
	preProcess := make([]*Process, 0, 2)
	for _, k := range preKey {
		p, err := c.QueryProcess(ctx, k)
		if err != nil {
			return nil, err
		}
		preProcess = append(preProcess, p)
	}
	return preProcess, nil
}

//---------Get the sourcing chain of on one branch within given depth (actually position 0 of preKey array
func (c *Contract) DigProcess(ctx TransactionContextInterface, key string, depth int) ([]*Process, error) {
	process, err := c.QueryProcess(ctx, key)
	if err != nil {
		return nil, perror.Errorf("Key not exist. %s", err)
	}
	preProcess := make([]*Process, 0, MaxSourcingDepth)
	for d := 0; d <= depth && d <= MaxSourcingDepth; d++ {
		preKey := process.PreKey
		if len(preKey) == 0 {
			break
		}
		pk := preKey[0]
		process, err = c.QueryProcess(ctx, pk)
		if err != nil {
			return nil, perror.Errorf("Error digging source preKey %s. %s", pk, err)
		}
		preProcess = append(preProcess, process)
	}
	return preProcess, nil
}

//----------Update org display name into the ledger
func (c *Contract) UpdateDisplayName(ctx TransactionContextInterface, displayName string) error {
	ownerOrg, orgErr := ctx.GetOrg()
	if orgErr != nil {
		return perror.Errorf("Get org info error. %s", orgErr)
	}
	if len(displayName) > 32 {
		return perror.New("Display name format error")
	}
	return ctx.GetProcessLedger().UpdateDisplayName(ownerOrg, displayName)
}

//----------Get org display name
func (c *Contract) GetDisplayName(ctx TransactionContextInterface, org string) (string, error) {
	return ctx.GetProcessLedger().GetDisplayName(org)
}

func (c *Contract) createPreKeySlice(ctx TransactionContextInterface, preKey []string, currentKey string) ([]string, error) {
	if len(preKey) != 0 {
		preKeyField := make([]string, 0, 1)
		for _, k := range preKey {
			_, err := c.QueryProcess(ctx, k)
			if err != nil {
				return nil, perror.Errorf("PreKey %s not exist. %s", k, err)
			}
			cOrg := SplitCompositeKey(currentKey)
			kOrg := SplitCompositeKey(k)
			if kOrg[0] == cOrg[0] {
				return nil, perror.Errorf("PreKey %s pointed to same org", k)
			}
			preKeyField = append(preKeyField, k)
		}
		return preKeyField, nil
	}
	return make([]string, 0), nil
}

func stringToArray(preKey string) ([]string, error) {
	preKeys := make([]string, 0)
	if preKey != "" && preKey != "[]" {
		arrayErr := json.Unmarshal([]byte(preKey), &preKeys)
		if arrayErr != nil {
			return nil, perror.New("PreKey array parsing failed")
		}
	}
	return preKeys, nil
}
