package main

import (
	"encoding/json"
	"os"
	"testing"
	"smartcontract"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

func TestCreateAssetForFailure(t *testing.T){
	cc_instance:= new (smartcontract.SmartContract)

	ctx:= new(contractapi.TransactionContext)

	ctx.SetTransactionID("testCCxID")
	ctx.SetClientIdentity(newTestClientIdentity("Tally", "Admin"))    // TO DO: Possibility of passing different users-- this function passes the client id as a param to be set to the context

	// LET'S SAY i WANT TO GET CLIENT ID
	clientIdentity := ctx.GetClientIdentity()
	clientID := clientIdentity.GetID()
	// test when providing non strings
	err:= cc_instance.CreateAsset(ctx, "Asset1", clientID)
	assert.NoError(t, err, "Creating asset failed.")


}

 





func newTestClientIdentity(mspID string, id string) *contractapi.TestIdentity {
    return &contractapi.TestIdentity{
        MSPID: mspID,
        ID:    id,
    }
}