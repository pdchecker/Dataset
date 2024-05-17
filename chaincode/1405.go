package main

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func getDOBCert(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// [0]= name ,[1]= PAddress , [2]=CurrentAddress , [3]=Phone , [4]=Parent
	// args[5]= Doctor , args[6]=DOB
	if len(args) != 7 {
		return shim.Error("Please Provide name,Paddres,caadress,phone,parent name,Doctor name and DOB")
	}
	id := uuid.New().String()
	dob := DoB{
		DOBDetails: HealthReport{
			DocType:   DOB,
			ID:        id,
			Doctor:    args[5],
			Type:      "DOB",
			IssueTime: time.Now().Unix(),
		},
		DOB: args[6],
	}
	DByte, _ := json.Marshal(dob)
	err := stub.PutState(id, DByte)
	if err != nil {
		return shim.Error(err.Error())
	}
	IKey := uuid.New().String()
	HKey := uuid.New().String()
	PKey := uuid.New().String()
	EKey := uuid.New().String()
	RKey := uuid.New().String()
	PHKey := uuid.New().String()
	health := Health{
		DocType: HEALTH,
		ID:      HKey,
		Reports: make(map[string]string),
	}
	health.Reports["DoB_cert"] = id
	personal := Personal{
		DocType:  PERSONAL,
		ID:       PKey,
		Name:     args[0],
		PAddress: args[1],
		CAddress: args[2],
		Phone:    args[3],
		Parent:   args[4],
		Status:   "0",
		Photo:    uuid.New().String(),
		DOB:      args[6],
	}
	education := Education{
		DocType:       EDUCATION,
		ID:            EKey,
		Qualification: make(map[string]string),
	}
	Iden := Identity{
		DocType:          IDENTITY,
		ID:               IKey,
		PersonalDetails:  PKey,
		HealthDetails:    HKey,
		EducationDetails: EKey,
		Requests:         make(map[string]string),
	}
	request := Request{
		DocType:     PERSONALREQUEST,
		ID:          RKey,
		From:        IKey,
		Status:      "0",
		RequestTime: time.Now().Unix(),
	}
	Iden.Requests[PERSONALREQUEST] = RKey

	DByte, _ = json.Marshal(health)
	stub.PutState(HKey, DByte)

	DByte, _ = json.Marshal(personal)
	stub.PutState(PKey, DByte)

	DByte, _ = json.Marshal(education)
	stub.PutState(EKey, DByte)

	DByte, _ = json.Marshal(Iden)
	stub.PutState(IKey, DByte)

	DByte, _ = json.Marshal(request)
	stub.PutState(RKey, DByte)
	result := struct {
		IdentityID  string
		DoBFilename string `json:"dob_filename"`
		Request     string `json:"request_for_verification"`
		Photo       string `json:"user_photo"`
	}{
		IdentityID:  IKey,
		Request:     RKey,
		Photo:       PHKey,
		DoBFilename: id,
	}
	output, _ := json.Marshal(result)
	return shim.Success(output)
}

func addReports(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//IKey[0],Doctor[1],Type[2]
	DByte, err := getState(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	var identity Identity
	json.Unmarshal(DByte, &identity)
	DByte, _ = getState(stub, identity.HealthDetails)
	var healthDet Health
	json.Unmarshal(DByte, &healthDet)
	id := uuid.New().String()
	put := HealthReport{
		DocType:   HEALTHREPORT,
		ID:        id,
		Doctor:    args[1],
		Type:      args[2],
		IssueTime: time.Now().Unix(),
	}
	healthDet.Reports[args[2]] = id

	DByte, _ = json.Marshal(healthDet)
	stub.PutState(healthDet.ID, DByte)

	DByte, _ = json.Marshal(put)
	stub.PutState(put.ID, DByte)
	result := struct {
		Id string
	}{
		Id: id,
	}
	DByte, _ = json.Marshal(result)
	return shim.Success(DByte)
}
