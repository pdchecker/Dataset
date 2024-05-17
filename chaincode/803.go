package main

import "github.com/hyperledger/fabric-contract-api-go/contractapi"

func (s *SmartContract) IsProvider(ctx contractapi.TransactionContextInterface) error {
	return ctx.GetClientIdentity().AssertAttributeValue("abac.role", "provider")
}

func (s *SmartContract) IsIntegrator(ctx contractapi.TransactionContextInterface) error {
	return ctx.GetClientIdentity().AssertAttributeValue("abac.role", "integrator")
}
