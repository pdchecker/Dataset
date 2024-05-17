package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

const (
	KeyAssets        = "assets"
	KeyAuctions      = "auctions"
	KeyLastAuctionID = "lastAuction"
)

func (cc *SmartContract) AddAsset(
	ctx contractapi.TransactionContextInterface, id, owner, auc_type string,
) error {
	existing, err := ctx.GetStub().GetState(cc.makeAssetKey(id))
	if err != nil {
		return fmt.Errorf("unable to interact with worldstate: %v", err)
	}

	if existing != nil {
		return fmt.Errorf("asset with ID %s already exists", id)
	}

	asset := Asset{
		ID:    id,
		Owner: owner,
	}

	err = cc.setAsset(ctx, &asset)
	if err != nil {
		return err
	}

	// Emit an event when an asset is added
	event := AssetAddingEventPayload{AssetID: id, AucType: auc_type}
	eventPayload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshalling event: %v", err)
	}

	err = ctx.GetStub().SetEvent("AddAsset", eventPayload)
	if err != nil {
		return fmt.Errorf("error setting event: %v", err)
	}
	return nil
}

func (cc *SmartContract) StartAuction(
	ctx contractapi.TransactionContextInterface, argjson string,
) error {
	var args StartAuctionArgs
	err := json.Unmarshal([]byte(argjson), &args)
	if err != nil {
		return err
	}

	asset, err := cc.GetAsset(ctx, args.AssetID)
	if err != nil {
		return err
	}
	if asset.PendingAuctionID > 0 {
		return fmt.Errorf("pending auction on asset")
	}

	lastID, err := cc.GetLastAuctionID(ctx)
	if err != nil {
		return err
	}
	auction := Auction{
		AuctionID: lastID + 1,
		AssetID:   args.AssetID,
		AucType:   args.AucType,

		EthAddr:    args.EthAddr,
		QuorumAddr: args.QuorumAddr,
		Status:     "open",
	}
	err = cc.setAuction(ctx, &auction)
	if err != nil {
		return err
	}

	err = cc.setLastAuctionID(ctx, auction.AuctionID)
	if err != nil {
		return err
	}

	asset.PendingAuctionID = auction.AuctionID
	err = cc.setAsset(ctx, asset)
	if err != nil {
		return fmt.Errorf("error setting asset: %v", err)
	}

	// Emit an event when an asset is added
	event := StartAuctionEventPayload{ID: auction.AuctionID, AucType: args.AucType, Owner: asset.Owner}
	eventPayload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshalling event: %v", err)
	}

	err = ctx.GetStub().SetEvent("StartAuction", eventPayload)
	if err != nil {
		return fmt.Errorf("error setting event: %v", err)
	}

	return nil
}

func (cc *SmartContract) RevealAuction(
	ctx contractapi.TransactionContextInterface, IDStr string,
) error {

	ID, _ := strconv.Atoi(IDStr)
	auction, err := cc.GetAuction(ctx, ID)

	if err != nil {
		return err
	}

	auction.Status = "reveal"
	err = cc.setAuction(ctx, auction)
	if err != nil {
		return fmt.Errorf("error setting auction: %v", err)
	}

	// Emit an event when an auction is revealing
	eventPayload, err := json.Marshal(auction)
	if err != nil {
		return fmt.Errorf("error marshalling event: %v", err)
	}
	err = ctx.GetStub().SetEvent("RevealAuction", []byte(eventPayload))
	if err != nil {
		return fmt.Errorf("error setting event: %v", err)
	}

	return nil
}

func (cc *SmartContract) CancelAuction(
	ctx contractapi.TransactionContextInterface, IDStr string,
) error {

	ID, _ := strconv.Atoi(IDStr)
	auction, err := cc.GetAuction(ctx, ID)
	if err != nil {
		return err
	}

	auction.Status = "closed"
	err = cc.setAuction(ctx, auction)
	if err != nil {
		return fmt.Errorf("error setting auction: %v", err)
	}

	asset, err := cc.GetAsset(ctx, auction.AssetID)
	if err != nil {
		return err
	}

	asset.PendingAuctionID = 0
	err = cc.setAsset(ctx, asset)
	if err != nil {
		return fmt.Errorf("error setting asset: %v", err)
	}

	// Emit an event when an auction is canceling
	eventPayload, err := json.Marshal(auction)
	if err != nil {
		return fmt.Errorf("error marshalling event: %v", err)
	}

	err = ctx.GetStub().SetEvent("CancelAuction", []byte(eventPayload))
	if err != nil {
		return fmt.Errorf("error setting event: %v", err)
	}

	return nil
}

func (cc *SmartContract) CloseAuction(
	ctx contractapi.TransactionContextInterface, IDStr string,
) error {

	ID, _ := strconv.Atoi(IDStr)
	auction, err := cc.GetAuction(ctx, ID)

	if err != nil {
		return err
	}

	auction.Status = "closing"
	err = cc.setAuction(ctx, auction)
	if err != nil {
		return fmt.Errorf("error setting auction: %v", err)
	}

	// Emit an event when an auction is closing
	eventPayload, err := json.Marshal(auction)
	if err != nil {
		return fmt.Errorf("error marshalling event: %v", err)
	}
	err = ctx.GetStub().SetEvent("CloseAuction", []byte(eventPayload))
	if err != nil {
		return fmt.Errorf("error setting event: %v", err)
	}
	return nil
}

func (cc *SmartContract) FinAuction(
	ctx contractapi.TransactionContextInterface, argjson string, prcdStr string,
) error {
	// only owner or admin can call this

	var args AuctionResult
	err := json.Unmarshal([]byte(argjson), &args)
	if err != nil {
		return err
	}

	prcd, err := strconv.ParseBool(prcdStr)
	if err != nil {
		return err
	}

	auction, err := cc.GetAuction(ctx, args.AuctionID)
	if err != nil {
		return err
	}

	if !cc.verifyAuctionResult(args) {
		return fmt.Errorf("invalid auction result")
	}

	auction.Status = "closed"
	err = cc.setAuction(ctx, auction)
	if err != nil {
		return err
	}

	asset, err := cc.GetAsset(ctx, auction.AssetID)
	if err != nil {
		return err
	}

	if prcd {
		asset.Owner = auction.HighestBidder
		//eventPayload = fmt.Sprintf("Auction: %d, closed with new owner: %s", auction.AuctionID, asset.Owner)
	}

	asset.PendingAuctionID = 0
	err = cc.setAsset(ctx, asset)
	if err != nil {
		return err
	}

	// Emit an event when an auction is canceling
	eventPayload, err := json.Marshal(auction)
	if err != nil {
		return fmt.Errorf("error marshalling event: %v", err)
	}

	err = ctx.GetStub().SetEvent("AuctionClosed", []byte(eventPayload))
	if err != nil {
		return fmt.Errorf("error setting event: %v", err)
	}

	return nil
}

// can add some mech to check if bidder has DID creditional
func (cc *SmartContract) verifyAuctionResult(result AuctionResult) bool {

	tmp := &AuctionResult{
		Platform:    result.Platform,
		AuctionID:   result.AuctionID,
		AuctionAddr: result.AuctionAddr,

		HighestBid:    result.HighestBid,
		HighestBidder: result.HighestBidder,
	}

	return VerifySignature(tmp.Hash(), result.Signature, result.HighestBidder)
}

func (cc *SmartContract) GetAsset(
	ctx contractapi.TransactionContextInterface, assetID string,
) (*Asset, error) {
	var asset Asset
	b, err := ctx.GetStub().GetState(cc.makeAssetKey(assetID))
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, fmt.Errorf("asset not found")
	}
	err = json.Unmarshal(b, &asset)
	return &asset, err
}

func (cc *SmartContract) GetAuction(
	ctx contractapi.TransactionContextInterface, auctionID int,
) (*Auction, error) {
	b, err := ctx.GetStub().GetState(cc.makeAuctionKey(auctionID))
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, fmt.Errorf("auction not found")
	}
	var auction Auction
	err = json.Unmarshal(b, &auction)
	return &auction, err
}

func (cc *SmartContract) GetLastAuctionID(
	ctx contractapi.TransactionContextInterface,
) (int, error) {
	b, err := ctx.GetStub().GetState(KeyLastAuctionID)
	if err != nil {
		return 0, err
	}
	var count int
	json.Unmarshal(b, &count)
	return count, nil
}

func (cc *SmartContract) setAsset(
	ctx contractapi.TransactionContextInterface, asset *Asset,
) error {
	b, _ := json.Marshal(asset)
	err := ctx.GetStub().PutState(cc.makeAssetKey(asset.ID), b)
	if err != nil {
		return fmt.Errorf("set asset error: %v", err)
	}
	return nil
}

func (cc *SmartContract) setAuction(
	ctx contractapi.TransactionContextInterface, auction *Auction,
) error {
	b, _ := json.Marshal(auction)
	err := ctx.GetStub().PutState(cc.makeAuctionKey(auction.AuctionID), b)
	if err != nil {
		return fmt.Errorf("set auction error: %v", err)
	}
	return nil
}

func (cc *SmartContract) setLastAuctionID(
	ctx contractapi.TransactionContextInterface, id int,
) error {
	b, _ := json.Marshal(id)
	return ctx.GetStub().PutState(KeyLastAuctionID, b)
}

func (cc *SmartContract) makeAssetKey(assetID string) string {
	return fmt.Sprintf("%s_%s", KeyAssets, assetID)
}

func (cc *SmartContract) makeAuctionKey(auctionID int) string {
	return fmt.Sprintf("%s_%d", KeyAuctions, auctionID)
}
