package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Project struct {
	ObjectType       string                  `json:"docType"`
	ProjectName      string                  `json:"projectName"`
	ProjectType      string                  `json:"projectType"`
	Place            string                  `json:"place"`
	Phases           []Phase                 `json:"phases"`
	CreationDate     int                     `json:"creationDate"`
	ActualStartDate  int                     `json:"actualStartDate"`
	TotalProjectCost float64                 `json:"totalProjectCost"`
	ProjectState     string                  `json:"projectState"`  //Created, Open For Funding, PartlyFunded, FullyFunded, Seeking Validation, Completed
	ApprovalState    string                  `json:"approvalState"` //Approved, Abandoned, UnApproved
	NGO              string                  `json:"ngo"`
	Contributors     map[string]string       `json:"contributors"`
	Contributions    map[string]Contribution `json:"contributions"`
	TotalReceived    float64                 `json:"totalReceived"`
	TotalRedeemed    float64                 `json:"totalRedeemed"`
	Comments         string                  `json:"comments"`
	Balance          float64                 `json:"balance"`
}

type Phase struct {
	Qty                float64                `json:"qty"`
	OutstandingQty     float64                `json:"outstandingQty"`
	PhaseState         string                 `json:"phaseState"` //Created, Open For Funding, PartlyFunded, FullyFunded, Seeking Validation, Validated, Complete
	StartDate          int                    `json:"startDate"`
	EndDate            int                    `json:"endDate"`
	ValidationCriteria map[string][]Criterion `json:"validationCriteria"`
	CAValidation       Validation             `json:"caValidation"`
}

//for CA validation
type Validation struct {
	IsValid  bool   `json:"isValid"`
	Comments string `json:"comments"`
}

type Contribution struct {
	Contributor     string  `json:"donatorAddress"`
	ContributionQty float64 `json:"contributionQty"`
}

type Criterion struct {
	Desc    string `json:"desc"`
	DocName string `json:"docName"`
	DocHash string `json:"docHash"`
}

// Actual Project Start date
func (s *SmartContract) UpdateActualStartDate(ctx contractapi.TransactionContextInterface, arg string) (bool, error) {
	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return false, fmt.Errorf("Error getting transaction creator: " + err.Error())
	}
	mspId, commonName, _ := getTxCreatorInfo(ctx, creator)
	fmt.Println("client is: " + commonName)

	if mspId != NgoMSP {
		return false, fmt.Errorf("only ca can initiate Project")
	}

	var args []string

	err = json.Unmarshal([]byte(arg), &args)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}
	if len(args) != 3 {
		return false, fmt.Errorf("Incorrect number of arguments. Expecting 3")
	} else if len(args[0]) <= 0 {
		return false, fmt.Errorf("pId must be a non-empty string")
	} else if len(args[1]) <= 0 {
		return false, fmt.Errorf("Date must be a non-empty string")
	} else if len(args[2]) <= 0 {
		return false, fmt.Errorf("tx Id must be a non-empty string")
	}

	projectId := strings.ToLower(args[0])
	date, err := strconv.Atoi(args[1])
	if err != nil {
		return false, fmt.Errorf("date should be numeric.")
	}
	txId := args[2]

	projectState := Project{}

	//check if the project exists
	projectAsBytes, err := ctx.GetStub().GetState(projectId)
	if err != nil {
		return false, fmt.Errorf("Error getting project")
	}
	if projectAsBytes == nil {
		return false, fmt.Errorf("project is not present")
	}
	json.Unmarshal(projectAsBytes, &projectState)
	if date <= 0 {
		return false, fmt.Errorf("Date should be greater than zero")
	}
	projectState.ActualStartDate = date

	projectAsBytes, _ = json.Marshal(projectState)
	ctx.GetStub().PutState(projectId, projectAsBytes)

	//create a transaction
	err = createTransaction(ctx, commonName, projectState.NGO, 0.0, date, "UpdateActualStartDate", projectId, txId, -1)
	if err != nil {
		return false, fmt.Errorf("Failed to add a Tx: " + err.Error())
	}

	return true, nil
}

//create a new project
func (s *SmartContract) CreateProject(ctx contractapi.TransactionContextInterface, arg string) (bool, error) {

	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return false, fmt.Errorf("Error getting transaction creator: " + err.Error())
	}
	mspId, commonName, _ := getTxCreatorInfo(ctx, creator)
	fmt.Println("client is: " + commonName)

	if mspId != NgoMSP {
		return false, fmt.Errorf("only ngo can initiate createProject")
	}

	var args []string

	err = json.Unmarshal([]byte(arg), &args)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	if len(args) != 3 {
		return false, fmt.Errorf("Incorrect number of arguments. Expecting 3")
	} else if len(args[0]) <= 0 {
		return false, fmt.Errorf("project details must be a non-empty json string")
	} else if len(args[1]) <= 0 {
		return false, fmt.Errorf("pId must be a non-empty string")
	} else if len(args[2]) <= 0 {
		return false, fmt.Errorf("tx Id must be a non-empty string")
	}

	pId := args[1]
	txId := args[2]

	projectObj := Project{}
	err = json.Unmarshal([]byte(args[0]), &projectObj)

	if err != nil {
		return false, fmt.Errorf("error in unmarshalling: " + err.Error())
	} else if len(projectObj.ProjectName) <= 0 {
		return false, fmt.Errorf("Project name is mandatory!")
	} else if len(projectObj.ProjectType) <= 0 {
		return false, fmt.Errorf("Project type is mandatory!")
	} else if len(projectObj.Phases) < 1 {
		return false, fmt.Errorf("please specify atleast one phase!")
	} else if projectObj.CreationDate <= 0 {
		return false, fmt.Errorf("Creation Date is mandatory!")
	} else if projectObj.Contributors != nil {
		return false, fmt.Errorf("Contributors should be none!")
	}

	projInBytes, _ := ctx.GetStub().GetState(pId)
	if projInBytes != nil {
		return false, fmt.Errorf("Project with this pId already exists")
	}

	//check if project with same name already exists.
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"Project\", \"projectName\":\"%s\"}}", projectObj.ProjectName)

	queryResults, err := GetQueryResultForQueryString(ctx, queryString)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	} else if len(queryResults) > 2 {
		return false, fmt.Errorf("A project with the same name already exists!")
	}
	fmt.Println("project -------------------")
	if projectObj.Contributions != nil {
		return false, fmt.Errorf("No contributions expected!")
	}
	projectObj.Contributions = make(map[string]Contribution)
	//set extra attributes
	projectObj.NGO = commonName
	projectObj.ObjectType = "Project"
	projectObj.ProjectState = "Created"
	projectObj.ApprovalState = "UnApproved"
	projectObj.Place = strings.ToLower(projectObj.Place)
	projectObj.Contributors = make(map[string]string)
	projectObj.Balance = 0.0

	//TODO: move it to UI
	allPhaseCosts := 0.0
	for i := 0; i < len(projectObj.Phases); i++ {
		projectObj.Phases[i].PhaseState = "Created"
		projectObj.Phases[i].OutstandingQty = projectObj.Phases[i].Qty
		allPhaseCosts = math.Round((allPhaseCosts+projectObj.Phases[i].Qty)*100) / 100

		if projectObj.Phases[i].StartDate >= projectObj.Phases[i].EndDate {
			return false, fmt.Errorf("end date must be ahead of start date!")
		}

		if projectObj.Phases[i].ValidationCriteria == nil {
			return false, fmt.Errorf("Please provide atleast one validation criteria!")
		}
	}

	projectObj.TotalProjectCost = allPhaseCosts

	newProjAsBytes, err := json.Marshal(projectObj)
	if err != nil {
		return false, fmt.Errorf("Json convert error" + err.Error())
	}

	err = ctx.GetStub().PutState(pId, newProjAsBytes)
	if err != nil {
		return false, fmt.Errorf("error saving project" + err.Error())
	}

	err = createTransaction(ctx, commonName, "All", projectObj.TotalProjectCost, projectObj.CreationDate, "ProjectCreate", pId, txId, -1)
	if err != nil {
		return false, fmt.Errorf("Failed to add a Tx: " + err.Error())
	}

	eventPayload := "A new project '" + projectObj.ProjectName + "' is waiting for your approval."

	notification := &Notification{TxId: txId, Description: eventPayload, Users: []string{ca + "." + creditsauthority + "." + domain}}
	notificationtAsBytes, err := json.Marshal(notification)
	ctx.GetStub().SetEvent("Notification", notificationtAsBytes)

	return true, nil
}

//create a new project
func (s *SmartContract) ApproveProject(ctx contractapi.TransactionContextInterface, arg string) (bool, error) {

	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return false, fmt.Errorf("Error getting transaction creator: " + err.Error())
	}
	mspId, commonName, _ := getTxCreatorInfo(ctx, creator)
	fmt.Println("client is: " + commonName)

	if mspId != CreditsAuthorityMSP {
		return false, fmt.Errorf("only ca can initiate ApproveProject")
	}

	var args []string

	err = json.Unmarshal([]byte(arg), &args)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	if len(args) != 3 {
		return false, fmt.Errorf("Incorrect number of arguments. Expecting 3")
	} else if len(args[0]) <= 0 {
		return false, fmt.Errorf("project details must be a non-empty json string")
	} else if len(args[1]) <= 0 {
		return false, fmt.Errorf("pId must be a non-empty string")
	} else if len(args[2]) <= 0 {
		return false, fmt.Errorf("tx Id must be a non-empty string")
	}

	pId := args[1]
	txId := args[2]

	newProjectObj := Project{}
	err = json.Unmarshal([]byte(args[0]), &newProjectObj)

	if err != nil {
		return false, fmt.Errorf("error in unmarshalling: " + err.Error())
	} else if len(newProjectObj.ProjectName) <= 0 {
		return false, fmt.Errorf("Project name is mandatory!")
	} else if len(newProjectObj.ProjectType) <= 0 {
		return false, fmt.Errorf("Project type is mandatory!")
	} else if len(newProjectObj.Phases) < 1 {
		return false, fmt.Errorf("please specify atleast one phase!")
	} else if newProjectObj.Contributors != nil {
		return false, fmt.Errorf("Contributors should be none!")
	}

	projInBytes, _ := ctx.GetStub().GetState(pId)
	if projInBytes == nil {
		return false, fmt.Errorf("Project with this pId doesn't exist!")
	}

	projectState := Project{}
	err = json.Unmarshal(projInBytes, &projectState)
	if err != nil {
		return false, fmt.Errorf("error in unmarshalling: " + err.Error())
	}

	if newProjectObj.ProjectName != projectState.ProjectName {
		//check if project with same name already exists.
		queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"Project\", \"projectName\":\"%s\"}}", newProjectObj.ProjectName)

		queryResults, err := GetQueryResultForQueryString(ctx, queryString)
		if err != nil {
			return false, fmt.Errorf(err.Error())
		} else if len(queryResults) < 2 {
			return false, fmt.Errorf("A project with the same name already exists!")
		}
	}

	fmt.Println("project -------------------")

	if newProjectObj.Contributions != nil {
		return false, fmt.Errorf("No contributions expected!")
	}

	//set extra attributes
	newProjectObj.NGO = projectState.NGO
	newProjectObj.ObjectType = "Project"
	newProjectObj.Place = strings.ToLower(newProjectObj.Place)
	newProjectObj.Contributors = make(map[string]string)
	newProjectObj.Contributions = make(map[string]Contribution)
	newProjectObj.CreationDate = projectState.CreationDate
	newProjectObj.ApprovalState = "Approved"
	newProjectObj.ProjectState = "Open For Funding"
	newProjectObj.Balance = 0.0

	//TODO: move it to UI
	allPhaseCosts := 0.0
	for i := 0; i < len(newProjectObj.Phases); i++ {
		newProjectObj.Phases[i].PhaseState = "Created"
		newProjectObj.Phases[i].OutstandingQty = newProjectObj.Phases[i].Qty
		allPhaseCosts = math.Round((allPhaseCosts+newProjectObj.Phases[i].Qty)*100) / 100

		if newProjectObj.Phases[i].StartDate >= newProjectObj.Phases[i].EndDate {
			return false, fmt.Errorf("end date must be ahead of start date, for each phase!")
		}

		if newProjectObj.Phases[i].ValidationCriteria == nil || len(newProjectObj.Phases[i].ValidationCriteria) <= 0 {
			return false, fmt.Errorf("Please provide atleast one validation criteria!")
		}
	}

	newProjectObj.Phases[0].PhaseState = "Open For Funding"
	newProjectObj.TotalProjectCost = allPhaseCosts

	newProjAsBytes, err := json.Marshal(newProjectObj)
	if err != nil {
		return false, fmt.Errorf("Json convert error" + err.Error())
	}

	err = ctx.GetStub().PutState(pId, newProjAsBytes)
	if err != nil {
		return false, fmt.Errorf("error saving project" + err.Error())
	}

	err = createTransaction(ctx, commonName, "All", newProjectObj.TotalProjectCost, newProjectObj.CreationDate, "ProjectApprove", pId, txId, -1)
	if err != nil {
		return false, fmt.Errorf("Failed to add a Tx: " + err.Error())
	}

	eventPayload := newProjectObj.ProjectName + " project has been approved by Rainforest Foundation US. "

	notification := &Notification{TxId: txId, Description: eventPayload, Users: []string{newProjectObj.NGO}}
	notificationtAsBytes, _ := json.Marshal(notification)
	ctx.GetStub().SetEvent("Notification", notificationtAsBytes)

	return true, nil
}

//validate/reject a phase
func (s *SmartContract) ValidatePhase(ctx contractapi.TransactionContextInterface, arg string) (bool, error) {

	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return false, fmt.Errorf("Error getting transaction creator: " + err.Error())
	}
	mspId, commonName, _ := getTxCreatorInfo(ctx, creator)
	if mspId != CreditsAuthorityMSP {
		return false, fmt.Errorf("only creditsauthority can initiate ValidatePhase")
	}

	var args []string

	err = json.Unmarshal([]byte(arg), &args)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	if len(args) != 6 {
		return false, fmt.Errorf("Incorrect number of arguments. Expecting 6")
	} else if len(args[0]) <= 0 {
		return false, fmt.Errorf("project id must be a non-empty json string")
	} else if len(args[1]) <= 0 {
		return false, fmt.Errorf("phase No. must be a non-empty string")
	} else if len(args[2]) <= 0 {
		return false, fmt.Errorf("validation must be a non-empty string")
	} else if len(args[4]) <= 0 {
		return false, fmt.Errorf("date must be a non-empty string")
	} else if len(args[5]) <= 0 {
		return false, fmt.Errorf("tx Id must be a non-empty string")
	}

	projectId := args[0]
	phaseNumber, err := strconv.Atoi(args[1])
	if err != nil || phaseNumber < 0.0 {
		return false, fmt.Errorf("Invalid phase number!")
	}
	validated, err := strconv.ParseBool(args[2])
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}
	comments := args[3]
	date, err := strconv.Atoi(args[4])
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}
	txId := args[5]

	projectInBytes, _ := ctx.GetStub().GetState(projectId)
	if projectInBytes == nil {
		return false, fmt.Errorf("Project doesn't exist")
	}

	projectObj := Project{}
	err = json.Unmarshal(projectInBytes, &projectObj)
	if err != nil {
		return false, fmt.Errorf("error in unmarshalling: " + err.Error())
	}

	if !(phaseNumber >= 0 && phaseNumber < len(projectObj.Phases)) {
		return false, fmt.Errorf("Invalid phase number!")
	} else if projectObj.Phases[phaseNumber].PhaseState != "Seeking Validation" {
		return false, fmt.Errorf("The phase must be in Seeking Validation state!")
	} else if !validated && len(comments) == 0 {
		return false, fmt.Errorf("comments are mandatory!")
	}

	//update the phase with validation details & Validated/Rejected phase state
	validationObj := Validation{
		IsValid:  validated,
		Comments: comments,
	}
	projectObj.Phases[phaseNumber].CAValidation = validationObj

	if projectObj.TotalReceived == 0.0 {
		projectObj.ProjectState = "Open For Funding"
	} else if projectObj.TotalReceived < projectObj.TotalProjectCost {
		projectObj.ProjectState = "Partially Funded"
	} else {
		projectObj.ProjectState = "Fully Funded"
	}

	if validated {
		projectObj.Phases[phaseNumber].PhaseState = "Validated"
		if phaseNumber == len(projectObj.Phases)-1 {
			projectObj.ProjectState = "Validated"
		} else {
			//change the state of next phase accordingly
			if projectObj.Balance == 0.0 {
				projectObj.Phases[phaseNumber+1].PhaseState = "Open For Funding"
			} else if projectObj.Phases[phaseNumber+1].OutstandingQty > projectObj.Balance {
				projectObj.Phases[phaseNumber+1].OutstandingQty = math.Round((projectObj.Phases[phaseNumber+1].OutstandingQty-projectObj.Balance)*100) / 100
				projectObj.Balance = 0.0
				projectObj.Phases[phaseNumber+1].PhaseState = "Partially Funded"
			} else if projectObj.Phases[phaseNumber+1].OutstandingQty == projectObj.Balance {
				projectObj.Phases[phaseNumber+1].OutstandingQty = math.Round((projectObj.Phases[phaseNumber+1].OutstandingQty-projectObj.Balance)*100) / 100
				projectObj.Balance = 0.0
				projectObj.Phases[phaseNumber+1].PhaseState = "Fully Funded"
			} else {
				projectObj.Balance = math.Round((projectObj.Balance-projectObj.Phases[phaseNumber+1].OutstandingQty)*100) / 100
				projectObj.Phases[phaseNumber+1].OutstandingQty = 0.0
				projectObj.Phases[phaseNumber+1].PhaseState = "Fully Funded"
			}
		}
	} else {
		if projectObj.Phases[phaseNumber].OutstandingQty == projectObj.Phases[phaseNumber].Qty {
			projectObj.Phases[phaseNumber].PhaseState = "Open For Funding"
		} else if projectObj.Phases[phaseNumber].OutstandingQty <= 0.0 {
			projectObj.Phases[phaseNumber].PhaseState = "Fully Funded"
		} else {
			projectObj.Phases[phaseNumber].PhaseState = "Partially Funded"
		}
	}

	projectInBytes, err = json.Marshal(projectObj)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().PutState(projectId, projectInBytes)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	err = createTransaction(ctx, commonName, projectObj.NGO, 0.0, date, "Project_Phase_Validation", projectId, txId, phaseNumber)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	eventPayload := "Phase " + strconv.Itoa(phaseNumber+1) + " of project '" + projectObj.ProjectName + "' has been "
	if validated {
		eventPayload += "validated by "
	} else {
		eventPayload += "rejected by "
	}
	eventPayload += "Rainforest Foundation US."

	notification := &Notification{TxId: txId, Description: eventPayload, Users: []string{projectObj.NGO}}
	notificationtAsBytes, err := json.Marshal(notification)
	ctx.GetStub().SetEvent("Notification", notificationtAsBytes)

	return true, nil
}

//update the validation criteria to add uploaded document hashes
func (s *SmartContract) AddDocumentHash(ctx contractapi.TransactionContextInterface, arg string) (bool, error) {

	//getusercontext to populate the required data
	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return false, fmt.Errorf("Error getting transaction creator: " + err.Error())
	}
	mspId, commonName, _ := getTxCreatorInfo(ctx, creator)
	if mspId != NgoMSP {
		return false, fmt.Errorf("only ngo can initiate addDocumentHash")
	}

	var args []string

	err = json.Unmarshal([]byte(arg), &args)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	if len(args) != 7 {
		return false, fmt.Errorf("Incorrect number of arguments. Expecting 7")
	} else if len(args[0]) <= 0 {
		return false, fmt.Errorf("project id must be a non-empty json string")
	} else if len(args[1]) <= 0 {
		return false, fmt.Errorf("phase No. must be a non-empty string")
	} else if len(args[2]) <= 0 {
		return false, fmt.Errorf("criterion must be a non-empty string")
	} else if len(args[3]) <= 0 {
		return false, fmt.Errorf("doc hash must be a non-empty string")
	} else if len(args[4]) <= 0 {
		return false, fmt.Errorf("doc name must be a non-empty string")
	} else if len(args[5]) <= 0 {
		return false, fmt.Errorf("date must be a non-empty string")
	} else if len(args[6]) <= 0 {
		return false, fmt.Errorf("tx Id must be a non-empty string")
	}

	projectId := args[0]
	phaseNumber, err := strconv.Atoi(args[1])
	if err != nil || phaseNumber < 0.0 {
		return false, fmt.Errorf("Invalid phase number!")
	}
	criterion := args[2]
	docHash := args[3]
	docName := args[4]
	date, err := strconv.Atoi(args[5])
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}
	txId := args[6]

	projectInBytes, _ := ctx.GetStub().GetState(projectId)
	if projectInBytes == nil {
		return false, fmt.Errorf("Project doesn't exist")
	}

	projectObj := Project{}
	err = json.Unmarshal(projectInBytes, &projectObj)
	if err != nil {
		return false, fmt.Errorf("error in unmarshalling: " + err.Error())
	}

	if projectObj.NGO != commonName {
		return false, fmt.Errorf("Invalid project owner")
	}

	//save the docHash
	if projectObj.Phases[phaseNumber].PhaseState == "Validated" {
		return false, fmt.Errorf("Documents cant be uploaded to a validated phase!")
	} else if projectObj.Phases[phaseNumber].ValidationCriteria[criterion] == nil {
		return false, fmt.Errorf("No such criteria exists!")
	}

	projectObj.Phases[phaseNumber].ValidationCriteria[criterion] = append(projectObj.Phases[phaseNumber].ValidationCriteria[criterion], Criterion{"desc", docName, docHash})

	newProjAsBytes, err := json.Marshal(projectObj)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().PutState(projectId, newProjAsBytes)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	err = createTransaction(ctx, commonName, projectObj.NGO, 0.0, date, "UploadDocument", projectId, txId, phaseNumber)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	tmpList := make([]string, 0, len(projectObj.Contributions))
	for k := range projectObj.Contributions {
		tmpList = append(tmpList, k)
	}

	splitName := strings.SplitN(commonName, ".", -1)
	eventPayload := splitName[0] + " has uploaded a document to the phase " + strconv.Itoa(phaseNumber+1) + " of a project."
	notification := &Notification{TxId: txId, Description: eventPayload, Users: tmpList}
	notificationtAsBytes, err := json.Marshal(notification)
	ctx.GetStub().SetEvent("Notification", notificationtAsBytes)

	return true, nil
}

//update the project/phase state
func (s *SmartContract) UpdateProject(ctx contractapi.TransactionContextInterface, arg string) (bool, error) {

	//getusercontext to populate the required data
	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return false, fmt.Errorf("Error getting transaction creator: " + err.Error())
	}
	mspId, commonName, _ := getTxCreatorInfo(ctx, creator)
	if mspId != NgoMSP {
		return false, fmt.Errorf("only ngo can initiate UpdateProject")
	}

	var args []string

	err = json.Unmarshal([]byte(arg), &args)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	projectId := strings.ToLower(args[0])
	phaseNumber, err := strconv.Atoi(args[1])
	if err != nil || phaseNumber < 0.0 {
		return false, fmt.Errorf("Invalid phase number!")
	}
	state := args[2]
	date, err := strconv.Atoi(args[3])
	if err != nil {
		return false, fmt.Errorf("date should be numeric.")
	}
	txId := args[4]

	projectState := Project{}

	//check if the project exists
	projectAsBytes, err := ctx.GetStub().GetState(projectId)
	if err != nil {
		return false, fmt.Errorf("Error getting project")
	}
	if projectAsBytes == nil {
		return false, fmt.Errorf("project is not present")
	}
	json.Unmarshal(projectAsBytes, &projectState)

	if projectState.NGO != commonName {
		return false, fmt.Errorf("Invalid project owner")
	}

	//check for the validity of the phase number
	if phaseNumber >= len(projectState.Phases) {
		return false, fmt.Errorf("invalid phase number")
	}

	currentPhaseState := projectState.Phases[phaseNumber].PhaseState
	if state == "Open For Funding" {
		if currentPhaseState == "Created" {
			projectState.Phases[phaseNumber].PhaseState = "Open For Funding"
		} else {
			return false, fmt.Errorf("Only created state can be opened for funding")
		}
		if phaseNumber > 0 {
			if projectState.Phases[phaseNumber-1].PhaseState != "Complete" {
				return false, fmt.Errorf("previous phase is not Complete")
			}
		}
	} else if state == "Seeking Validation" {
		//TODO: check if documents are uploaded for each validation criteria
		if currentPhaseState == "Open For Funding" || currentPhaseState == "Partially Funded" || currentPhaseState == "Fully Funded" {
			projectState.Phases[phaseNumber].PhaseState = "Seeking Validation"
			projectState.ProjectState = "Seeking Validation"
		} else {
			return false, fmt.Errorf("current phase is in an invalid state to seek validation")
		}
	} else if state == "Complete" {
		if currentPhaseState == "Validated" {
			projectState.Phases[phaseNumber].PhaseState = "Complete"
		} else {
			return false, fmt.Errorf("current phase is not yet validated to be marked complete")
		}
		if phaseNumber == len(projectState.Phases)-1 {
			projectState.ProjectState = "Completed"
		}
	} else {
		return false, fmt.Errorf("state can be Open For Funding or Seeking Validation or Complete")
	}

	projectAsBytes, _ = json.Marshal(projectState)
	ctx.GetStub().PutState(projectId, projectAsBytes)

	//create a transaction
	err = createTransaction(ctx, commonName, "All", 0.0, date, "UpdateProject", projectId, txId, phaseNumber)
	if err != nil {
		return false, fmt.Errorf("Failed to add a Tx: " + err.Error())
	}

	//list of users who will receive the notification.
	tmpList := make([]string, 0, len(projectState.Contributions))
	for k := range projectState.Contributions {
		tmpList = append(tmpList, k)
	}

	splitName := strings.SplitN(commonName, ".", -1)
	eventPayload := splitName[0] + " has updated the phase " + strconv.Itoa(phaseNumber+1) + " of project " + projectState.ProjectName + "."

	notification := &Notification{TxId: txId, Description: eventPayload, Users: tmpList}
	if state == "Seeking Validation" {
		notification.Users = []string{ca + "." + creditsauthority + "." + domain}
		eventPayload = "Your validation is requested for the project '" + projectState.ProjectName + "'"
		notification.Description = eventPayload
	}
	notificationtAsBytes, err := json.Marshal(notification)
	ctx.GetStub().SetEvent("Notification", notificationtAsBytes)
	return true, nil
}

//delete the project
func (s *SmartContract) DeleteProject(ctx contractapi.TransactionContextInterface, arg string) (bool, error) {

	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return false, fmt.Errorf("Error getting transaction creator: " + err.Error())
	}
	mspId, commonName, _ := getTxCreatorInfo(ctx, creator)
	if mspId != CreditsAuthorityMSP {

		return false, fmt.Errorf("only Regulator can initiate DeleteProject")
	}

	var args []string

	err = json.Unmarshal([]byte(arg), &args)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	if len(args) != 4 {
		return false, fmt.Errorf("Incorrect number of arguments. Expecting 6")
	} else if len(args[0]) <= 0 {
		return false, fmt.Errorf("project id must be a non-empty json string")
	} else if len(args[1]) <= 0 {
		return false, fmt.Errorf("comments must be a non-empty string")
	} else if len(args[2]) <= 0 {
		return false, fmt.Errorf("date must be a non-empty string")
	} else if len(args[3]) <= 0 {
		return false, fmt.Errorf("tx Id must be a non-empty string")
	}

	projectId := strings.ToLower(args[0])
	comments := args[1]

	date, err := strconv.Atoi(args[2])
	if err != nil {
		return false, fmt.Errorf("date should be numeric.")
	}
	txId := args[3]
	projectState := Project{}
	//check if the project exists
	projectAsBytes, err := ctx.GetStub().GetState(projectId)
	if err != nil {
		return false, fmt.Errorf("Error getting project")
	}
	if projectAsBytes == nil {
		return false, fmt.Errorf("project is not present")
	}

	json.Unmarshal(projectAsBytes, &projectState)

	if projectState.ApprovalState != "UnApproved" {
		return false, fmt.Errorf("Only UnApproved project can be deleted!")
	}

	er := ctx.GetStub().DelState(projectId)
	fmt.Println("Error in delstate: ", er)
	if er != nil {
		return false, fmt.Errorf("Error deleting project")
	}

	//create a transaction
	err = createTransaction(ctx, commonName, projectState.NGO, 0.0, date, "DeleteProject", projectId, txId, -1)
	if err != nil {
		return false, fmt.Errorf("Failed to add a Tx: " + err.Error())
	}

	eventPayload := projectState.ProjectName + " project has been deleted by Rainforest Foundation US. " + "Comments: " + comments

	notification := &Notification{TxId: txId, Description: eventPayload, Users: []string{projectState.NGO}}
	notificationtAsBytes, err := json.Marshal(notification)
	ctx.GetStub().SetEvent("Notification", notificationtAsBytes)
	return true, nil
}

//abandon the project
func (s *SmartContract) AbandonProject(ctx contractapi.TransactionContextInterface, arg string) (bool, error) {

	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return false, fmt.Errorf("Error getting transaction creator: " + err.Error())
	}
	mspId, commonName, _ := getTxCreatorInfo(ctx, creator)
	if mspId != CreditsAuthorityMSP {

		return false, fmt.Errorf("only Regulator can initiate AbandonProject")
	}

	var args []string

	err = json.Unmarshal([]byte(arg), &args)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	if len(args) != 4 {
		return false, fmt.Errorf("Incorrect number of arguments. Expecting 6")
	} else if len(args[0]) <= 0 {
		return false, fmt.Errorf("project id must be a non-empty json string")
	} else if len(args[1]) <= 0 {
		return false, fmt.Errorf("comments must be a non-empty string")
	} else if len(args[2]) <= 0 {
		return false, fmt.Errorf("date must be a non-empty string")
	} else if len(args[3]) <= 0 {
		return false, fmt.Errorf("tx Id must be a non-empty string")
	}

	projectId := strings.ToLower(args[0])
	comments := args[1]

	date, err := strconv.Atoi(args[2])
	if err != nil {
		return false, fmt.Errorf("date should be numeric.")
	}
	txId := args[3]
	projectState := Project{}
	//check if the project exists
	projectAsBytes, err := ctx.GetStub().GetState(projectId)
	if err != nil {
		return false, fmt.Errorf("Error getting project")
	}
	if projectAsBytes == nil {
		return false, fmt.Errorf("project is not present")
	}

	json.Unmarshal(projectAsBytes, &projectState)

	if projectState.ApprovalState != "Approved" {
		return false, fmt.Errorf("Only Approved project can be abandoned!")
	}
	projectState.ApprovalState = "Abandoned"
	projectState.Comments = comments

	projectAsBytes, _ = json.Marshal(projectState)
	ctx.GetStub().PutState(projectId, projectAsBytes)

	//create a transaction
	err = createTransaction(ctx, commonName, projectState.NGO, 0.0, date, "AbandonProject", projectId, txId, -1)
	if err != nil {
		return false, fmt.Errorf("Failed to add a Tx: " + err.Error())
	}

	eventPayload := projectState.ProjectName + " project has been abandoned by Rainforest Foundation US. " + "Comments: " + comments

	notification := &Notification{TxId: txId, Description: eventPayload, Users: []string{projectState.NGO}}
	notificationtAsBytes, err := json.Marshal(notification)
	ctx.GetStub().SetEvent("Notification", notificationtAsBytes)

	return true, nil
}

func (s *SmartContract) EditProject(ctx contractapi.TransactionContextInterface, arg string) (bool, error) {

	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return false, fmt.Errorf("Error getting transaction creator: " + err.Error())
	}
	mspId, commonName, _ := getTxCreatorInfo(ctx, creator)
	if mspId != CreditsAuthorityMSP {

		return false, fmt.Errorf("only Regulator can initiate EditProject")
	}

	var args []string

	err = json.Unmarshal([]byte(arg), &args)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	if len(args) != 4 {
		return false, fmt.Errorf("Incorrect number of arguments. Expecting 4")
	} else if len(args[0]) <= 0 {
		return false, fmt.Errorf("project id must be a non-empty json string")
	} else if len(args[1]) <= 0 {
		return false, fmt.Errorf("project must be a non-empty json string")
	} else if len(args[2]) <= 0 {
		return false, fmt.Errorf("date must be a non-empty string")
	} else if len(args[3]) <= 0 {
		return false, fmt.Errorf("tx Id must be a non-empty string")
	}

	projectId := strings.ToLower(args[0])
	projtoEdit := args[1]

	date, err := strconv.Atoi(args[2])
	if err != nil {
		return false, fmt.Errorf("date should be numeric.")
	}
	txId := args[3]
	projectEdit := Project{}
	json.Unmarshal([]byte(projtoEdit), &projectEdit)

	projGonnaEdit := Project{}
	//check if the project exists
	projectAsBytes, err := ctx.GetStub().GetState(projectId)
	if err != nil {
		return false, fmt.Errorf("Error getting project")
	}
	if projectAsBytes == nil {
		return false, fmt.Errorf("project is not present")
	}

	json.Unmarshal(projectAsBytes, &projGonnaEdit)

	var currentPhaseNum = -1
	if projGonnaEdit.ApprovalState == "Approved" {
		for i := 0; i < len(projGonnaEdit.Phases); i++ {
			if projGonnaEdit.Phases[i].PhaseState != "Validated" {
				currentPhaseNum = i
				break
			}
		}

		if currentPhaseNum != -1 {
			for j := 0; j < len(projectEdit.Phases); j++ {
				if currentPhaseNum == j {
					projGonnaEdit.Phases[j].StartDate = projectEdit.Phases[j].StartDate
					projGonnaEdit.Phases[j].EndDate = projectEdit.Phases[j].EndDate
				} else if j > currentPhaseNum {

					projGonnaEdit.Phases[j].StartDate = projectEdit.Phases[j].StartDate
					projGonnaEdit.Phases[j].EndDate = projectEdit.Phases[j].EndDate

					for k, v := range projectEdit.Phases[j].ValidationCriteria {
						if _, ok := projGonnaEdit.Phases[j].ValidationCriteria[k]; !ok {
							projGonnaEdit.Phases[j].ValidationCriteria[k] = v
						}

					}

				}

			}
		}

	} else {
		return false, fmt.Errorf("project is not approved")
	}

	projectAsBytes, _ = json.Marshal(projGonnaEdit)
	ctx.GetStub().PutState(projectId, projectAsBytes)

	//create a transaction
	err = createTransaction(ctx, commonName, projGonnaEdit.NGO, 0.0, date, "EditProject", projectId, txId, -1)
	if err != nil {
		return false, fmt.Errorf("Failed to add a Tx: " + err.Error())
	}

	eventPayload := projGonnaEdit.ProjectName + " project has been edited by Rainforest Foundation US. "

	notification := &Notification{TxId: txId, Description: eventPayload, Users: []string{projGonnaEdit.NGO}}
	notificationtAsBytes, err := json.Marshal(notification)
	ctx.GetStub().SetEvent("Notification", notificationtAsBytes)

	return true, nil
}
