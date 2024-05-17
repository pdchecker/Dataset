package main

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// DummyContract contract for testing purposes
type DummyContract struct {
	contractapi.Contract
}

func (s *DummyContract) ReadIdentity(ctx contractapi.TransactionContextInterface) (string, error) {
	identity := ctx.GetClientIdentity()

	ID, _ := identity.GetID()
	mspID, _ := identity.GetMSPID()

	resultString := fmt.Sprintf(
		` 	- Client's identity:
				Default format: %v
				Go-syntax format: %#v
			- Client's ID: %s
			- Client's MSPID: %s
		`, identity, identity, ID, mspID)

	return resultString, nil
}

func (s *DummyContract) ReadCert(ctx contractapi.TransactionContextInterface) (string, error) {
	identity := ctx.GetClientIdentity()

	cert, _ := identity.GetX509Certificate()

	orgs := cert.Issuer.Organization
	orgUnits := cert.Issuer.OrganizationalUnit

	resultString := fmt.Sprintf(
		` 	- cert.Issuer.Organization: %#v
			- cert.Issuer.OrganizationalUnit: %#v
		`, orgs, orgUnits)

	return resultString, nil
}

func (s *DummyContract) ReadAttr(ctx contractapi.TransactionContextInterface) (string, error) {
	identity := ctx.GetClientIdentity()

	attrRole, _, _ := identity.GetAttributeValue("Role")
	attrAffiliation, _, _ := identity.GetAttributeValue("hf.Affiliation")
	attrEnrollmentID, _, _ := identity.GetAttributeValue("hf.EnrollmentID")
	attrType, _, _ := identity.GetAttributeValue("hf.Type")

	resultString := fmt.Sprintf(
		` 	- Attribute "Role": %s
			- Attribute "hf.Affiliation": %s
			- Attribute "hf.EnrollmentID": %s
			- Attribute "hf.Type": %s
		`, attrRole, attrAffiliation, attrEnrollmentID, attrType)

	return resultString, nil
}

// GetEvaluateTransactions returns functions of ProductContract not to be tagged as submit.
func (s *DummyContract) GetEvaluateTransactions() []string {
	return []string{"ReadIdentity", "ReadCert", "ReadAttr"}
}
