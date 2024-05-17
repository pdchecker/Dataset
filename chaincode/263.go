package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/metadata"
)

func main() {
	Contract := new(NftChaincode)
	Contract.Info.Version = "1"
	Contract.Info.Description = "My Smart Contract"
	Contract.Info.License = new(metadata.LicenseMetadata)
	Contract.Info.License.Name = "Apache-2.0"
	Contract.Info.Contact = new(metadata.ContactMetadata)
	Contract.Info.Contact.Name = "Hjy"

	chaincode, err := contractapi.NewChaincode(Contract)
	chaincode.Info.Title = "NFT chaincode"
	chaincode.Info.Version = "1"

	if err != nil {
		panic("Could not create chaincode from NftChaincode." + err.Error())
	}

	err = chaincode.Start()

	if err != nil {
		panic("Failed to start chaincode. " + err.Error())
	}
}
