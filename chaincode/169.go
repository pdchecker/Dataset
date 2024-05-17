package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a car
type SmartContract struct {
	contractapi.Contract
}

// NextShowID and NextTicketId
const (
	NEXTSHOWID   = "NEXT_SHOW_ID"
	NEXTTICKETID = "NEXT_TICKET_ID"
)

//doctypes
const (
	SHOW   = "SHOW"
	TICKET = "TICKET"
	WINDOW = "WINDOW"
	SODA   = "SODA"
)

// Theatre Model
type Theatre struct {
	TheatreNo      int    `json:"theatreNo"`
	TheatreName    string `json:"theatreName"`
	Windows        int    `json:"windows,omitempty"`
	TicketsPerShow int    `json:"ticketsPerShow,omitempty"`
	ShowsDaily     int    `json:"showsDaily,omitempty"`
	SodaStock      int    `json:"sodaStock,omitempty"`
	Halls          int    `json:"halls,omitempty"`
	DocType        string `json:"docType"`
}

// Window model
type Window struct {
	WindowNo    int    `json:"windowNo"`
	TicketsSold int    `json:"ticketsSold"`
	DocType     string `json:"docType"`
}

// Ticket model
type Ticket struct {
	TicketNo        int     `json:"ticketNo"`
	Show            Show    `json:"show"`
	Window          Window  `json:"window"`
	Quantity        int     `json:"quantity,number"`
	Amount          float64 `json:"amount,string"`
	CouponNumber    string  `json:"couponNumber"`
	CouponAvailed   bool    `json:"couponAvailed"`
	ExchangeAvailed bool    `json:"exchangeAvailed"`
	DocType         string  `json:"docType"`
}

// Show model
type Show struct {
	ShowID    int    `json:"showID"`
	Movie     string `json:"movie"`
	ShowSlot  string `json:"showSlot"`
	Quantity  int    `json:"quantity,number"`
	HallNo    int    `json:"hallNo"`
	TheatreNo int    `json:"theatreNo"`
	DocType   string `json:"docType"`
}

// Soda model
type Soda struct {
	Stock        int    `json:"stock"`
	TicketNo     int    `json:"ticketNo"`
	CouponNumber string `json:"couponNumber"`
	DocType      string `json:"docType"`
}

// Property model
type Property struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CreateShows model
type CreateShows struct {
	TheatreNo int    `json:"theatreNo"`
	Shows     []Show `json:"shows"`
}

// Init initializes chaincode
// ===========================
func (s *SmartContract) Init(ctx contractapi.TransactionContextInterface) error {
	_, err := set(ctx, NEXTSHOWID, "0")
	_, err = set(ctx, NEXTTICKETID, "0")

	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}

// register Theatre in word state
func (s *SmartContract) registerTheatre(ctx contractapi.TransactionContextInterface, args []string) (string, error) {
	fmt.Println("************ Register Theatre Start *****************")
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect number of arguments. Expecting 1")
	}
	var theatre Theatre
	err := json.Unmarshal([]byte(args[0]), &theatre)
	fmt.Println("************ args *****************", err)

	if err != nil {
		fmt.Println("Cannot unmarshal theatre Object", err)
		return "", fmt.Errorf(err.Error())
	}
	txnID := ctx.GetStub().GetTxID()

	var number int
	for _, c := range txnID {
		number = number + int(c)
	}

	theatre.TheatreNo = number
	theatre.DocType = "THEATRE"
	theatreAsBytes, _ := json.Marshal(theatre)
	putTheatreData := ctx.GetStub().PutState("THEATRE"+strconv.Itoa(theatre.TheatreNo), theatreAsBytes)
	if putTheatreData != nil {
		return "", fmt.Errorf(err.Error())
	}

	// create windows for the theatre
	for i := 1; i <= theatre.Windows; i++ {
		var window Window
		window.WindowNo = i
		window.TicketsSold = 0
		window.DocType = WINDOW
		windowAsBytes, _ := json.Marshal(window)
		err := ctx.GetStub().PutState("WINDOW"+strconv.Itoa(i), windowAsBytes)
		if err != nil {
			return "", fmt.Errorf(err.Error())
		}
	}

	fmt.Println("************ Register Theatre End *****************")

	return string("MovieTheatre Number:" + strconv.Itoa(theatre.TheatreNo)), nil
}

// create show for theatre
func (s *SmartContract) createShow(ctx contractapi.TransactionContextInterface, args []string) (string, error) {
	fmt.Println("************ Show Created Start *****************")
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect number of arguments. Expecting 1")
	}
	var createShows CreateShows
	err := json.Unmarshal([]byte(args[0]), &createShows)
	if err != nil {
		fmt.Errorf("Cannot unmarshal createShows Object %s", err)
		return "", fmt.Errorf("Cannot unmarshal createShows Object %s", err.Error())
	}
	showSeq, err := get(ctx, NEXTSHOWID)
	fmt.Println("Generating show for showSeq", showSeq)

	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	shows := createShows.Shows
	var theatre Theatre
	theatreBytes, err := ctx.GetStub().GetState("THEATRE" + strconv.Itoa(createShows.TheatreNo))

	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	json.Unmarshal(theatreBytes, &theatre)

	if len(shows) > theatre.Halls {
		return "", fmt.Errorf("Number of Movies cannot exceed" + strconv.Itoa(theatre.Halls))
	}

	for _, show := range shows {
		for i := 1; i <= theatre.ShowsDaily; i++ {
			showSeq = showSeq + 1
			show.ShowID = +showSeq
			show.ShowSlot = strconv.Itoa(i)
			show.Quantity = theatre.TicketsPerShow
			show.TheatreNo = theatre.TheatreNo
			show.DocType = SHOW
			showAsBytes, _ := json.Marshal(show)
			err = ctx.GetStub().PutState("SHOW"+strconv.Itoa(show.ShowID), showAsBytes)
			if err != nil {
				return "", fmt.Errorf(err.Error())
			}
		}
	}
	// fmt.Println("saving current showSeq %s ", showSeq)
	_, err = set(ctx, NEXTSHOWID, strconv.Itoa(showSeq))
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	fmt.Println("************ Show Created End *****************")

	return string("Show Created Successfully"), nil
}

// purchaseTicket from window
func (s *SmartContract) purchaseTicket(ctx contractapi.TransactionContextInterface, args []string) (string, error) {
	fmt.Println("************ Purchase Ticket Start *****************")
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect number of arguments. Expecting 1")
	}

	var ticket Ticket

	err := json.Unmarshal([]byte(args[0]), &ticket)
	if err != nil {
		fmt.Println("Cannot unmarshal ticket Object", err)
		return "", fmt.Errorf(err.Error())
	}

	ticketSeq, err := get(ctx, NEXTTICKETID)
	fmt.Println("Generating Ticket for ticketSeq", ticketSeq)

	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	showBytes, err := ctx.GetStub().GetState("SHOW" + strconv.Itoa(ticket.Show.ShowID))
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	var show Show
	json.Unmarshal(showBytes, &show)

	windowBytes, err := ctx.GetStub().GetState("WINDOW" + strconv.Itoa(ticket.Window.WindowNo))
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	var window Window
	json.Unmarshal(windowBytes, &window)
	// check the show for number of seats remaining
	if show.Quantity < 0 || show.Quantity-ticket.Quantity < 0 {
		return "", fmt.Errorf("Seats Full for the requested show or Not enough seats as requested. Available:" + strconv.Itoa(show.Quantity))
	}

	show.Quantity = show.Quantity - ticket.Quantity
	window.TicketsSold = window.TicketsSold + ticket.Quantity
	fmt.Println(window.TicketsSold)
	fmt.Println(ticket.Quantity)
	ticketSeq = ticketSeq + 1
	ticket.TicketNo = ticketSeq
	ticket.Show = show
	ticket.Window = window
	ticket.DocType = TICKET

	showAsBytes, _ := json.Marshal(show)
	err = ctx.GetStub().PutState("SHOW"+strconv.Itoa(show.ShowID), showAsBytes)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	windowAsBytes, _ := json.Marshal(window)
	err = ctx.GetStub().PutState("WINDOW"+strconv.Itoa(window.WindowNo), windowAsBytes)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	fmt.Println("saving current ticketSeq", ticketSeq)
	_, err = set(ctx, NEXTTICKETID, strconv.Itoa(ticketSeq))
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	ticketAsBytes, _ := json.Marshal(ticket)
	err = ctx.GetStub().PutState("TICKET"+strconv.Itoa(ticketSeq), ticketAsBytes)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	fmt.Println("************ Purchase Ticket Start *****************")
	return string(ctx.GetStub().GetTxID()), nil
}

// issueCoupon for collect the water bottle and popcorn
func (s *SmartContract) issueCoupon(ctx contractapi.TransactionContextInterface, args []string) (string, error) {
	fmt.Println("************ Issue Coupon Start *****************")
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect number of arguments. Expecting 1")
	}

	var ticket Ticket

	err := json.Unmarshal([]byte(args[0]), &ticket)
	if err != nil {
		fmt.Println("Cannot unmarshal ticket Object", err)
		return "", fmt.Errorf(err.Error())
	}
	ticketBytes, err := ctx.GetStub().GetState("TICKET" + strconv.Itoa(ticket.TicketNo))
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	json.Unmarshal(ticketBytes, &ticket)

	if ticket.CouponAvailed {
		fmt.Println("Coupon Availed Already")
		return "", fmt.Errorf("Coupon Availed Already")
	}

	txnID := ctx.GetStub().GetTxID()
	var number int
	for _, c := range txnID {
		number = number + int(c)
	}
	ticket.CouponNumber = strconv.Itoa(number)
	ticket.CouponAvailed = true
	ticket.ExchangeAvailed = false
	ticketAsBytes, _ := json.Marshal(ticket)
	err = ctx.GetStub().PutState("TICKET"+strconv.Itoa(ticket.TicketNo), ticketAsBytes)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	fmt.Println("************ Issue Coupon End *****************")
	return string("Coupon Number:" + ticket.CouponNumber), nil
}

// change the coupon to water bottle to soda
func (s *SmartContract) couponExchanged(ctx contractapi.TransactionContextInterface, args []string) (string, error) {
	fmt.Println("************ Coupon Exchanged Start *****************")
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect number of arguments. Expecting 1")
	}

	var ticket Ticket
	err := json.Unmarshal([]byte(args[0]), &ticket)
	if err != nil {
		fmt.Println("Cannot unmarshal ticket Object", err)
		return "", fmt.Errorf(err.Error())
	}

	ticketBytes, err := ctx.GetStub().GetState("TICKET" + strconv.Itoa(ticket.TicketNo))
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	json.Unmarshal(ticketBytes, &ticket)

	if ticket.ExchangeAvailed {
		fmt.Println("Exchange Availed Already")
		return "", fmt.Errorf("Exchange Availed Already")
	}

	var theatre Theatre
	theatreBytes, err := ctx.GetStub().GetState("THEATRE" + strconv.Itoa(ticket.Show.TheatreNo))
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	json.Unmarshal(theatreBytes, &theatre)

	// check if even number for the eligible soda exchange
	// Need to check first 200 user can exchanges
	// In window 1 user can exchanges
	couponNo, err := strconv.Atoi(ticket.CouponNumber)
	if err != nil {
		fmt.Println("Ticket Not eligible for exchange")
		return "", fmt.Errorf("Ticket Not eligible for exchange")
	}
	if couponNo%2 != 0 {
		fmt.Println("Ticket Not eligible for exchange")
		return "", fmt.Errorf("Ticket Not eligible for exchange")
	}

	if theatre.SodaStock-ticket.Quantity < 0 {
		fmt.Println("Soda Stock Over")
		return "", fmt.Errorf("Soda Stock Over")
	}

	ticket.ExchangeAvailed = true
	ticketAsBytes, _ := json.Marshal(ticket)
	err = ctx.GetStub().PutState("TICKET"+strconv.Itoa(ticket.TicketNo), ticketAsBytes)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	theatre.SodaStock = theatre.SodaStock - ticket.Quantity

	theatreAsBytes, _ := json.Marshal(theatre)
	err = ctx.GetStub().PutState("THEATRE"+strconv.Itoa(theatre.TheatreNo), theatreAsBytes)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	fmt.Println("************ Coupon Exchanged End *****************")
	return string(ctx.GetStub().GetTxID()), nil
}

// get function
func get(ctx contractapi.TransactionContextInterface, key string) (int, error) {
	if key == "" {
		return 0, fmt.Errorf("Incorrect arguments. Expecting a key")
	}
	value, err := ctx.GetStub().GetState(key)
	if err != nil {
		return 0, fmt.Errorf("Failed to get asset: %s with error: %s", key, err)
	}
	if value == nil {
		return 0, fmt.Errorf("Asset not found: %s", key)
	}
	var property Property
	json.Unmarshal(value, &property)
	fmt.Println(property)
	fmt.Println(property.Value)
	i, err := strconv.Atoi(property.Value)
	if err != nil {
		return 0, fmt.Errorf("Failed to get next sequence number %s", err)
	}
	// fmt.Println("Got the the value for %s : value : %s", key, i)
	return i, nil
}

// set function
func set(ctx contractapi.TransactionContextInterface, key string, value string) (string, error) {
	fmt.Println("setting value", key, value)

	var property Property
	property.Key = key
	property.Value = value

	propertyAsBytes, _ := json.Marshal(property)
	err := ctx.GetStub().PutState(key, propertyAsBytes)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	return value, nil
}

// The main function.
func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}
