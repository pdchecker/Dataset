package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/ticken-ts/ticken-chaincodes/common"
	"math/big"
)

type Contract struct {
	contractapi.Contract
}

// ********** cc-event integration ********** //

const ccEventName = "cc-event"
const ccEventSellTicketFunc = "SellTicket"

// *****+************************************ //

const index = "eventID~section~ticketID"

const Name = "cc-ticket"

type Ticket struct {
	TicketID string `json:"ticket_id"`
	EventID  string `json:"event_id"`
	Section  string `json:"section"`

	// represents the public blockchain
	// token ID
	TokenID string `json:"token_id"`

	// represents the owner id
	// in the web service database
	OwnerID string `json:"owner"`
}

// Issue a new ticket for the event with ID "eventID" in the section "section"
// to the owner with ID "ownerID". This method will call the "cc-event" chaincode
// to check if the event is "on sale" or the section has remaining tickets.
//
// Params
// * - ticketID -> uuid format
// * - eventID  -> uuid format
// * - section  -> string (must be equal to the section name of the event)
// * - ownerID  -> uuid format
// * - tokenID  -> hexadecimal string representing the tokenID of the public blockchain (uint256)
//
// The return value can be:
//   - - the ticket created serialized in JSON format
//   - - error in case some conditions to issue the ticket are not fulfilled
//     such as the event is not on sale or the section has not more remaining tickets
func (c *Contract) Issue(ctx common.ITickenTxContext, ticketID, eventID, section, ownerID, tokenID string) (*Ticket, error) {
	existentTicket, err := c.GetTicket(ctx, ticketID)
	if existentTicket != nil {
		return nil, ccErr("ticket with ID %s already exists", ticketID)
	}

	ownerIDParsed, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, ccErr("error parsing owner id: %v", err)
	}
	eventIDParsed, err := uuid.Parse(eventID)
	if err != nil {
		return nil, ccErr("error parsing event id: %v", err)
	}
	ticketIDParsed, err := uuid.Parse(ticketID)
	if err != nil {
		return nil, ccErr("error parsing ticket id: %v", err)
	}
	tokenIDParsed, ok := new(big.Int).SetString(tokenID, 16)
	if !ok {
		return nil, ccErr("token ID is not a valid uint256")
	}

	ticket := Ticket{
		TicketID: ticketIDParsed.String(),
		EventID:  eventIDParsed.String(),
		Section:  section,
		TokenID:  tokenIDParsed.Text(16),
		OwnerID:  ownerIDParsed.String(),
	}

	ticketJSON, err := json.Marshal(ticket)
	if err != nil {
		return nil, ccErr("failed to serialize ticket: %v", err)
	}

	// add ticket into the chaincode cc-event
	// note: this operation is atomically handled
	// by the orderers. So, the ticket and the ticket
	// count are updated simultaneously in the same tx
	ccEventSellTicketResponse := ctx.GetStub().InvokeChaincode(
		ccEventName,
		getCCCallArgs(ccEventSellTicketFunc, eventID, section),
		ctx.GetStub().GetChannelID(),
	)

	if ccEventSellTicketResponse.Status != shim.OK {
		return nil, ccErr(ccEventSellTicketResponse.Message)
	}

	//  Create an index to enable section-based range queries, e.g. return all tickets from section V.I.P.
	//  An 'index' is a normal key-value entry in the ledger.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  This will enable very efficient state range queries based on composite keys matching indexName~section~*
	sectionIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{eventID, ticket.Section, ticket.TicketID})
	if err != nil {
		return nil, ccErr("failed to create section index key: %v", err)
	}

	if err := ctx.GetStub().PutState(sectionIndexKey, ticketJSON); err != nil {
		return nil, ccErr("failed to updated the state: %v", err)
	}

	return &ticket, nil
}

// GetTicket returns the ticket information of the event with id "ticketID".
//
// Params
// * - ticketID -> uuid format
//
// The return value can be:
// * - error in case of the ticket is not found
func (c *Contract) GetTicket(ctx common.ITickenTxContext, ticketID string) (*Ticket, error) {
	ticketJSON, err := ctx.GetStub().GetState(ticketID)
	if err != nil {
		return nil, ccErr("failed to read ticket: %v", err)
	}
	if ticketJSON == nil {
		return nil, ccErr("ticket %s does not exist", ticketID)
	}

	var ticket Ticket
	if err := json.Unmarshal(ticketJSON, &ticket); err != nil {
		return nil, ccErr("failed to deserialize ticket: %v", err)
	}

	return &ticket, err
}

// GetSectionTickets returns all the tickets of the section "section"
// from the event with ID "eventID".
//
// Params
// * - ticketID -> uuid format
// * - section  -> string (must be equal to the section name of the event)
//
// The return value can be:
// * - error in case of the event is not found or the section
//   - is not present in the event
func (c *Contract) GetSectionTickets(ctx common.ITickenTxContext, eventID, section string) ([]*Ticket, error) {
	// Execute a key range query on all keys starting with 'section'
	sectionTicketsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(index, []string{eventID, section})
	if err != nil {
		return nil, ccErr("failed to create a section ticket iterator: %v", err)
	}
	defer sectionTicketsIterator.Close()

	return constructQueryResponseFromIterator(sectionTicketsIterator)
}

func ccErr(format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("[%s] | %s", Name, msg)
}

func getCCCallArgs(opName string, args ...string) [][]byte {
	queryArgs := make([][]byte, len(args)+1)

	queryArgs[0] = []byte(opName)
	for i, arg := range args {
		queryArgs[i+1] = []byte(arg)
	}

	return queryArgs
}

// constructQueryResponseFromIterator constructs a slice of assets from the resultsIterator
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Ticket, error) {
	var assets []*Ticket

	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset Ticket

		if err := json.Unmarshal(queryResult.Value, &asset); err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
