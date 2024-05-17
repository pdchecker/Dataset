package main

import (
    "encoding/json"
    "fmt"
    "github.com/hyperledger/fabric-contract-api-go/contractapi"
    "github.com/hyperledger/fabric/common/flogging"
    "github.com/hyperledger/fabric/core/chaincode/shim"
    "github.com/hyperledger/fabric/protos/peer"
)

var logger = flogging.MustGetLogger("blind_auction")

type BlindAuctionContract struct {
    contractapi.Contract
}

type Bid struct {
    Bidder    string `json:"bidder"`
    Amount    int    `json:"amount"`
    Revealed  bool   `json:"revealed"`
}

type Auction struct {
    ID          string `json:"id"`
    Description string `json:"description"`
    Bids        map[string]Bid
    Closed      bool   `json:"closed"`
    Winner      string `json:"winner"`
    HighestBid  int    `json:"highestBid"`
}

func (b *BlindAuctionContract) Init(ctx contractapi.TransactionContextInterface) error {
    return nil
}

func (b *BlindAuctionContract) CreateAuction(ctx contractapi.TransactionContextInterface, auctionID string, description string) error {
    exists, err := b.AuctionExists(ctx, auctionID)
    if err != nil {
        return err
    }
    if exists {
        return fmt.Errorf("Auction with ID %s already exists", auctionID)
    }

    auction := Auction{
        ID:          auctionID,
        Description: description,
        Bids:        make(map[string]Bid),
        Closed:      false,
        Winner:      "",
        HighestBid:  0,
    }

    auctionJSON, err := json.Marshal(auction)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(auctionID, auctionJSON)
}

func (b *BlindAuctionContract) PlaceBid(ctx contractapi.TransactionContextInterface, auctionID string, bidder string, amount int) error {
    auction, err := b.GetAuction(ctx, auctionID)
    if err != nil {
        return err
    }
    if auction.Closed {
        return fmt.Errorf("Auction with ID %s is closed", auctionID)
    }

    bid := Bid{
        Bidder:    bidder,
        Amount:    amount,
        Revealed:  false,
    }

    auction.Bids[bidder] = bid
    auctionJSON, err := json.Marshal(auction)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(auctionID, auctionJSON)
}

func (b *BlindAuctionContract) RevealBid(ctx contractapi.TransactionContextInterface, auctionID string, bidder string, bidAmount int) error {
    auction, err := b.GetAuction(ctx, auctionID)
    if err != nil {
        return err
    }
    if auction.Closed {
        return fmt.Errorf("Auction with ID %s is closed", auctionID)
    }

    bid, exists := auction.Bids[bidder]
    if !exists {
        return fmt.Errorf("Bidder %s has not placed a bid in the auction", bidder)
    }

    if bid.Revealed {
        return fmt.Errorf("Bidder %s has already revealed their bid", bidder)
    }

    if bid.Amount != bidAmount {
        return fmt.Errorf("Bidder %s's revealed bid amount does not match the recorded bid", bidder)
    }

    if bidAmount > auction.HighestBid {
        auction.HighestBid = bidAmount
        auction.Winner = bidder
    }

    bid.Revealed = true
    auction.Bids[bidder] = bid

    auctionJSON, err := json.Marshal(auction)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(auctionID, auctionJSON)
}

func (b *BlindAuctionContract) CloseAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {
    auction, err := b.GetAuction(ctx, auctionID)
    if err != nil {
        return err
    }
    if auction.Closed {
        return fmt.Errorf("Auction with ID %s is already closed", auctionID)
    }

    auction.Closed = true
    auctionJSON, err := json.Marshal(auction)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(auctionID, auctionJSON)
}

func (b *BlindAuctionContract) GetAuction(ctx contractapi.TransactionContextInterface, auctionID string) (*Auction, error) {
    auctionJSON, err := ctx.GetStub().GetState(auctionID)
    if err != nil {
        return nil, err
    }
    if auctionJSON == nil {
        return nil, fmt.Errorf("Auction with ID %s does not exist", auctionID)
    }

    var auction Auction
    err = json.Unmarshal(auctionJSON, &auction)
    if err != nil {
        return nil, err
    }

    return &auction, nil
}

func (b *BlindAuctionContract) AuctionExists(ctx contractapi.TransactionContextInterface, auctionID string) (bool, error) {
    auctionJSON, err := ctx.GetStub().GetState(auctionID)
    if err != nil {
        return false, err
    }
    return auctionJSON != nil, nil
}

func main() {
    chaincode, err := contractapi.NewChaincode(&BlindAuctionContract{})
    if err != nil {
        logger.Error("Error while creating chaincode: ", err)
        return
    }

    if err := chaincode.Start(); err != nil {
        logger.Error("Error while starting chaincode: ", err)
    }
}
