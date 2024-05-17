package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"hyperlibrary/common"
	"log"
	"time"
)

type SmartContract struct {
	contractapi.Contract
	BorrowDuration time.Duration
	LateFeePerDay  float64
	Events         []common.Event
	MaxBooksOut    uint8
}

func NewSmartContract(contract contractapi.Contract) *SmartContract {
	s := &SmartContract{Contract: contract}
	s.LateFeePerDay = .50
	s.MaxBooksOut = 2
	bd, _ := time.ParseDuration(fmt.Sprintf("%dh", 14*24))
	s.BorrowDuration = bd
	return s
}

func (t *SmartContract) Init(ctx contractapi.TransactionContextInterface) error {
	log.Println("Init invoked")

	t.CreateBook(ctx, &common.Book{"book", "abcd1234", "Charles Dickens", "A Tale of Two Cities", common.FICTION, 0, 0, 0})
	t.CreateBook(ctx, &common.Book{"book", "abcd45454", "William Shakespeare", "Romeo and Juliet", common.FICTION, 0, 0, 0})
	t.CreateBook(ctx, &common.Book{"book", "abcd45455", "William Shakespeare", "Julis Casar", common.FICTION, 0, 0, 0})
	return nil
}

func (t *SmartContract) SetBorrowDuration(ctx contractapi.TransactionContextInterface, days int) error {
	if !t.HasRole(ctx, "LIBRARIAN") {
		return errors.New("Access Denied!")
	}

	log.Println(fmt.Sprintf("Setting borrow duration to %d days", days))
	bd, _ := time.ParseDuration(fmt.Sprintf("%dh", days*24))
	t.BorrowDuration = bd
	return nil
}

func (t *SmartContract) SetMaxBooksOut(ctx contractapi.TransactionContextInterface, days int) error {
	if !t.HasRole(ctx, "LIBRARIAN") {
		return errors.New("Access Denied!")
	}

	t.MaxBooksOut = uint8(days)
	return nil
}

func (t *SmartContract) SetLateFeePerDay(ctx contractapi.TransactionContextInterface, fee float64) error {
	if !t.HasRole(ctx, "LIBRARIAN") {
		return errors.New("Access Denied!")
	}
	t.LateFeePerDay = fee
	return nil
}

func (t *SmartContract) AddEvent(ctx contractapi.TransactionContextInterface, name string, payload interface{}) {
	ts, _ := ctx.GetStub().GetTxTimestamp()
	var myMap map[string]interface{}
	data, _ := json.Marshal(payload)
	json.Unmarshal(data, &myMap)
	t.Events = append(t.Events, common.Event{name, myMap, common.GetApproxTime(ts)})
}

func (t *SmartContract) SetEvents(ctx contractapi.TransactionContextInterface) {
	for i := range t.Events {
		event := t.Events[i]
		log.Println("Event", event)
	}

	if len(t.Events) > 0 {
		eventBytes, err := json.Marshal(t.Events)

		if err != nil {
			log.Fatalf(err.Error())
		}

		err = ctx.GetStub().SetEvent("Events", eventBytes)

		if err != nil {
			log.Fatal(err.Error())
		}

		t.Events = make([]common.Event, 0)
	}
}

func (t *SmartContract) Before(ctx contractapi.TransactionContextInterface, value interface{}) {
	log.Println("BEFORE", ctx.GetStub().GetTxID())
}

func (t *SmartContract) After(ctx contractapi.TransactionContextInterface, value interface{}) {
	log.Println("AFTER", ctx.GetStub().GetTxID())
}

func (t *SmartContract) CreateBook(ctx contractapi.TransactionContextInterface, book *common.Book) error {
	if !t.HasRole(ctx, "LIBRARIAN") {
		return errors.New("Access Denied!")
	}

	assetBytes, err := json.Marshal(book)
	if err != nil {
		return err
	}

	// Check for existing ISBN
	books, err := t.QueryBook(ctx, "isbn", book.Isbn)
	if len(books) > 0 {
		return errors.New(fmt.Sprintf(`A book with the "%s" ISBN already exists!`, book.Isbn))
	}

	//ctx.GetStub().SetEvent("Book.Created", assetBytes)
	t.AddEvent(ctx, "Book.Created", book)
	t.SetEvents(ctx)
	return ctx.GetStub().PutState("book."+book.Isbn, assetBytes)
}

func (t *SmartContract) PurchaseBook(ctx contractapi.TransactionContextInterface, bookId string, quantity uint16, cost float32) ([]*common.BookInstance, error) {
	if !t.HasRole(ctx, "LIBRARIAN") {
		return nil, errors.New("Access Denied!")
	}

	if quantity < 1 {
		return nil, errors.New("Quantity must be at least 1!")
	}

	assetBytes, err := ctx.GetStub().GetState("book." + bookId)

	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s: %v", bookId, err)
	}
	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", bookId)
	}

	var book common.Book
	err = json.Unmarshal(assetBytes, &book)

	log.Println(fmt.Sprintf("There are currently %d owned.", book.Owned))

	var instances []*common.BookInstance

	var i uint16
	starting_id := book.MaxId + 1
	var last_id uint16

	ts, _ := ctx.GetStub().GetTxTimestamp()
	now := common.GetApproxTime(ts)

	for i = 0; i < quantity; i++ {
		instId := fmt.Sprintf("%s-%d", book.Isbn, starting_id+i)
		last_id = starting_id + i

		inst := common.BookInstance{"bookInstance", instId, bookId, now, cost,
			common.AVAILABLE, common.NEW, time.Time{}, common.MakeEmptyUser()}
		instBytes, err := json.Marshal(inst)

		if err != nil {
			log.Println("Unable to marshal instance!")
			return nil, err
		}

		err = ctx.GetStub().PutState("bookInstance."+instId, instBytes)
		//ctx.GetStub().SetEvent("BookInstance.Created", instBytes)
		t.AddEvent(ctx, "BookInstance.Created", inst)

		if err != nil {
			log.Println("Unable to store instance state!")
			return nil, err
		}

		instances = append(instances, &inst)
	}

	book.Owned += uint(quantity)
	book.Available += uint(quantity)
	book.MaxId = last_id

	assetBytes, err = json.Marshal(book)
	if err != nil {
		log.Println("Unable to marshal book!")
		return nil, err
	}

	err = ctx.GetStub().PutState("book."+bookId, assetBytes)
	t.SetEvents(ctx)

	log.Println(fmt.Sprintf("Created %d instances", quantity))
	return instances, nil
}

func (t *SmartContract) QueryBook(ctx contractapi.TransactionContextInterface, key string, value string) ([]*common.Book, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"book","%s":"%s"}}`, key, value)
	res, err := GetQueryResultForQueryString(ctx, queryString)

	if err != nil {
		return nil, err
	}

	var books []*common.Book
	for i := range res {
		bookBytes := res[i]
		var book common.Book
		err = json.Unmarshal(bookBytes, &book)
		books = append(books, &book)
	}
	return books, nil
}

func (t *SmartContract) ListBooks(ctx contractapi.TransactionContextInterface) ([]*common.Book, error) {
	res, err := GetQueryResultForQueryString(ctx, `{"selector":{"docType":"book"}}`)

	if err != nil {
		return nil, err
	}

	var books []*common.Book
	for i := range res {
		bookBytes := res[i]
		var book common.Book
		err = json.Unmarshal(bookBytes, &book)
		books = append(books, &book)
	}
	return books, nil
}

func (t *SmartContract) ListBookInstances(ctx contractapi.TransactionContextInterface, isbn string, statuses []common.Status) ([]*common.BookInstance, error) {
	selector := map[string]interface{}{
		"docType": "bookInstance",
		"bookId":  isbn,
	}

	if len(statuses) > 0 {
		var orStatuses []map[string]common.Status
		for i := range statuses {
			orStatuses = append(orStatuses, map[string]common.Status{
				"status": statuses[i],
			})
		}
		selector["$or"] = orStatuses
	} else {

	}

	query := map[string]interface{}{
		"selector": selector,
	}

	queryString, err := json.Marshal(query)

	if err != nil {
		return nil, err
	}

	res, err := GetQueryResultForQueryString(ctx, string(queryString))

	if err != nil {
		return nil, err
	}

	var books []*common.BookInstance
	for i := range res {
		bookBytes := res[i]
		var book common.BookInstance
		err = json.Unmarshal(bookBytes, &book)

		if err != nil {
			return nil, err
		}

		books = append(books, &book)
	}
	return books, nil
}

func (t *SmartContract) GetBook(ctx contractapi.TransactionContextInterface, isbn string) (*common.Book, error) {
	bookBytes, err := ctx.GetStub().GetState(fmt.Sprintf("book.%s", isbn))

	if err != nil {
		return nil, err
	}

	var book common.Book
	err = json.Unmarshal(bookBytes, &book)

	if err != nil {
		return nil, err
	}

	return &book, nil
}

func (t *SmartContract) GetBookInstance(ctx contractapi.TransactionContextInterface, instId string) (*common.BookInstance, error) {
	bookBytes, err := ctx.GetStub().GetState(fmt.Sprintf("bookInstance.%s", instId))

	if err != nil {
		return nil, err
	}

	var bookInstance common.BookInstance
	err = json.Unmarshal(bookBytes, &bookInstance)

	if err != nil {
		return nil, err
	}

	return &bookInstance, nil
}

func (t *SmartContract) updateBook(ctx contractapi.TransactionContextInterface, book *common.Book) error {
	bookBytes, err := json.Marshal(book)

	if err != nil {
		return err
	}

	ctx.GetStub().PutState(fmt.Sprintf("book.%s", book.Isbn), bookBytes)
	return nil
}

func (t *SmartContract) updateBookInstance(ctx contractapi.TransactionContextInterface, inst *common.BookInstance) error {
	instBytes, err := json.Marshal(inst)

	if err != nil {
		return err
	}

	ctx.GetStub().PutState(fmt.Sprintf("bookInstance.%s", inst.Id), instBytes)
	return nil
}

func (t *SmartContract) GetMyBooksOut(ctx contractapi.TransactionContextInterface) ([]*common.BookInstance, error) {
	user := t.GetUserByClientId(ctx)
	queryString := fmt.Sprintf(`{"selector":{"docType":"bookInstance", "borrower.clientId": "%s"}}`, user.ClientId)
	log.Println("Querying", queryString)
	res, err := GetQueryResultForQueryString(ctx, queryString)

	if err != nil {
		return nil, err
	}

	var books []*common.BookInstance

	for _, b := range res {
		var book *common.BookInstance
		err := json.Unmarshal(b, &book)

		if err != nil {
			return nil, err
		}

		books = append(books, book)
	}

	return books, nil
}

func (t *SmartContract) MaxBooksOutCheck(ctx contractapi.TransactionContextInterface) error {
	books, err := t.GetMyBooksOut(ctx)

	if err != nil {
		return err
	}
	count := uint8(len(books))
	log.Println("Comparing", count, t.MaxBooksOut)

	if count >= t.MaxBooksOut {
		return errors.New("You have the maximum number of books out!")
	}

	return nil
}

func (t *SmartContract) BorrowBookInstance(ctx contractapi.TransactionContextInterface, instId string) (*common.BookInstance, error) {
	err := t.MaxBooksOutCheck(ctx)

	if err != nil {
		log.Println("Max books out!")
		return nil, err
	}

	inst, err := t.GetBookInstance(ctx, instId)

	if err != nil {
		return nil, err
	}

	book, err := t.GetBook(ctx, inst.BookId)

	log.Println("Checking instance status", inst)
	if inst.Status == common.AVAILABLE {
		log.Println(fmt.Sprintf("Going to borrow book \"%s\"", instId))
		inst.Borrower = t.GetUserByClientId(ctx)
		inst.Status = common.OUT

		inst.DueDate = time.Now().Add(t.BorrowDuration).Round(time.Hour)
		instBytes1, err := json.Marshal(&inst)

		if err != nil {
			return nil, err
		}

		err = ctx.GetStub().PutState(fmt.Sprintf("bookInstance.%s", instId), instBytes1)

		if err != nil {
			return nil, err
		}

		book.Available -= 1
		err = t.updateBook(ctx, book)

		if err != nil {
			return nil, err
		}

		//ctx.GetStub().SetEvent("BookInstance.Borrowed", instBytes1)
		t.AddEvent(ctx, "BookInstance.Borrowed", inst)
	} else if inst.Status == common.OUT {
		return nil, errors.New("This book is already out!")
	}

	t.SetEvents(ctx)
	return inst, nil
}

func (t *SmartContract) InspectReturnedBook(ctx contractapi.TransactionContextInterface, instId string, cond common.Condition,
	feeAmount float64, available bool) (*common.Fee, error) {
	if !t.HasRole(ctx, "LIBRARIAN") {
		return nil, errors.New("Access Denied!")
	}

	inst, err := t.GetBookInstance(ctx, instId)

	if err != nil {
		return nil, err
	}

	var fee common.Fee
	if feeAmount > 0 {
		ts, _ := ctx.GetStub().GetTxTimestamp()
		date := common.GetApproxTime(ts)

		fee = common.Fee{"fee", ctx.GetStub().GetTxID(), inst.Borrower,
			feeAmount, common.DAMAGE_FEE, date, 0, false}

		t.storeFee(ctx, &fee)
		t.AddEvent(ctx, "Fee.Created", fee)
	}

	inst.Condition = cond
	inst.Borrower = common.MakeEmptyUser()

	if available {
		inst.Status = common.AVAILABLE

		book, err := t.GetBook(ctx, inst.BookId)

		if err != nil {
			return nil, err
		}

		book.Available += 1
		err = t.updateBook(ctx, book)

		if err != nil {
			return nil, err
		}
	}

	err = t.updateBookInstance(ctx, inst)

	if err != nil {
		return nil, err
	}

	if &fee != nil {
		return &fee, nil
	}

	return nil, nil
}

func (t *SmartContract) LostMyBook(ctx contractapi.TransactionContextInterface, instId string) (*common.Fee, error) {
	inst, err := t.GetBookInstance(ctx, instId)

	if inst.Borrower.ClientId != t.GetUserByClientId(ctx).ClientId {
		return nil, errors.New("You don't currently have this book taken out!")
	}

	if inst.Status != common.OUT {
		return nil, errors.New("This book is not currently out!")
	}

	if err != nil {
		return nil, err
	}

	inst.Status = common.LOST
	err = t.updateBookInstance(ctx, inst)

	if err != nil {
		return nil, err
	}

	ts, _ := ctx.GetStub().GetTxTimestamp()
	date := common.GetApproxTime(ts)

	fee := common.Fee{"fee", ctx.GetStub().GetTxID(), t.GetUserByClientId(ctx),
		float64(inst.Cost), common.LOST_FEE, date, 0, false}

	_, err = t.storeFee(ctx, &fee)
	t.AddEvent(ctx, "Fee.Created", fee)

	if err != nil {
		return nil, err
	}

	return &fee, nil
}

func (t *SmartContract) ReturnBookInstance(ctx contractapi.TransactionContextInterface, instId string) (*common.Fee, error) {
	inst, err := t.GetBookInstance(ctx, instId)

	if err != nil {
		return nil, err
	}

	if inst.Status == common.OUT {
		lateFee, err := t.checkForLateFee(ctx, inst)

		if err != nil {
			return nil, err
		}

		inst.Status = common.RETURNED
		inst.DueDate = time.Time{}

		instBytes, err := json.Marshal(inst)

		if err != nil {
			return nil, err
		}

		ctx.GetStub().PutState(fmt.Sprintf("bookInstance.%s", instId), instBytes)
		//ctx.GetStub().SetEvent("BookInstance.Returned", instBytes)
		t.AddEvent(ctx, "BookInstance.Returned", inst)

		t.SetEvents(ctx)

		if lateFee != nil {
			return lateFee, nil
		} else {
			return nil, nil
		}

	} else {
		errors.New("Book cannot be returned if it is not out!")
	}

	return nil, err
}
