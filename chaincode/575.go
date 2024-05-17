package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	/*
	"strconv"
	"strings"
	*/
//	"github.com/hyperledger/fabric/common/util"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
//Insert struct field in alphabetic order => to achieve determinism accross languages
// golang keeps the order when marshal to json but doesn't order automatically
type Asset struct {
	ID             string  `json:"ID"`
	Model          string  `json:"Model"`
	Price		int    `json:"Price"`
	Color          string  `json:"Color"`
	Fuel           string  `json:"Fuel"`
}


type Wallet struct{
	Owner	string			`json:"Owner"`
	NFT	map[string]Asset	`json:"NFT"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("Errore nella lettura dell'identita'")
	}
	
	if clientOrgID != "Org1MSP" {
		return fmt.Errorf("L'unico a poter chiamare questa funzione e' Org1MSP")
	}
	assets := []Asset{
		{ID: "asset1", Color: "Blue", Fuel: "Gasoline", Price: 8000, Model: "Panda"},
		{ID: "asset2", Color: "Red", Fuel: "Diesel", Price: 10000, Model: "Punto"},
		{ID: "asset3", Color: "Black", Fuel: "Gasoline", Price: 13000, Model: "Tipo"},
		
	}
	var walletOrg1 Wallet
	walletOrg1.Owner = "Org1MSP"
	
	walletOrg1.NFT = make(map[string]Asset)
	walletOrg1.NFT[assets[0].ID] = assets[0]
	walletOrg1.NFT[assets[1].ID] = assets[1]
	walletOrg1.NFT[assets[2].ID] = assets[2]
	
	walletJSON, err := json.Marshal(walletOrg1)
	
	if err != nil {
		return err
	}
	
	err = ctx.GetStub().PutState("Org1MSP", walletJSON)
	if err != nil {
		return fmt.Errorf("Errore nell'aggiunta di Org1 allo stato globale. %v", err)
	}
	
	return nil
}

// CreateAsset puo' essere richiamato solo da Org1 e crea un token, aggiungengolo al wallet di Org1 cioè al deposito della fabbrica
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, color string, fuel string, price int, model string) error {
	
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("Errore nella lettura dell'identita'")
	}
	
	if clientOrgID != "Org1MSP" {
		return fmt.Errorf("L'unico a poter chiamare questa funzione e' Org1MSP")
	}

	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("L'asset con ID: %s esiste gia' in fabbrica", id)
	}

	asset := Asset{
		ID:             id,
		Color:          color,
		Fuel:           fuel,
		Price:		price,
		Model:		model,
	}
	
	walletJSON, err := ctx.GetStub().GetState("Org1MSP")
	
	if err != nil {
		return fmt.Errorf("Errore nella lettura dello stato globale: %v", err)
	}
	
	var wallet Wallet
	err = json.Unmarshal(walletJSON, &wallet)
	if err != nil {
		return fmt.Errorf("errore unmarshal: ",err)
	}
	wallet.NFT[id] = asset
	
	walletJSON, err = json.Marshal(wallet)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState("Org1MSP", walletJSON)
	
	if err != nil {
		return fmt.Errorf("Errore nell'aggiunta di Org1 allo stato globale. %v", err)
	}
	
	return nil
}


// ReadAsset ritorna l'asset con l'ID specificato se contenuto all'interno del proprio wallet
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (Asset, error) {
	var asset Asset
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if(clientOrgID != "Org1MSP"){
		return asset,fmt.Errorf("Errore, non sei Org1")
	}
	walletJSON, err := ctx.GetStub().GetState("Org1MSP")
	
	if walletJSON == nil {
		return asset, fmt.Errorf("the wallet %s does not exist", id)
	}

	var wallet Wallet
	
	err = json.Unmarshal(walletJSON, &wallet)
	if err != nil {
		return asset, err
	}
	
	_, presente := wallet.NFT[id]
	
	if presente {
		return wallet.NFT[id], nil
	}

	return asset, fmt.Errorf("ID non trovato")
}


// UpdateAsset aggiorna le informazioni di un token contenuto nel propio wallet
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, color string, model string, price int,fuel string) error {
	
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	
	walletJSON, err := ctx.GetStub().GetState(clientOrgID)
	
	var wallet Wallet
	err = json.Unmarshal(walletJSON, &wallet)
	
	if err != nil {
		return fmt.Errorf("Errore nella lettura del wallet")
	}
	
	_ , presente := wallet.NFT[id]
	
	if ! presente {
		return fmt.Errorf("ID non trovato")
	}

	// Creo nuovo token
	asset := Asset{
		ID:             id,
		Color:          color,
		Model:          model,
		Price:		price,
		Fuel:		fuel,
	}
	
	wallet.NFT[id] = asset
	walletJSON, err = json.Marshal(wallet)
	
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(clientOrgID, walletJSON)
}

// DeleteAsset cancella un token dal proprio wallet
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) (Asset,error) {
	var empty Asset
	
	
	walletJSON, err := ctx.GetStub().GetState("Org1MSP")
	
	if err != nil {
		return empty,fmt.Errorf("Errore nella lettura dello stato globale: %v", err)
	}
	
	var wallet Wallet
	err = json.Unmarshal(walletJSON, &wallet)
	
	ass, presente := wallet.NFT[id]
	
	if ! presente {
		return empty,fmt.Errorf("L'ID non esiste");
	}
	
	delete(wallet.NFT, id)
	
	walletJSON, err = json.Marshal(wallet)
	
	if err != nil {
		return empty,fmt.Errorf("Errore nel marshalling")
	}
	ctx.GetStub().PutState("Org1MSP", walletJSON) 

	return ass,nil
}

// AssetExists controlla l'esistenza dell'ID specificato all'interno di tutti i wallet cioè se la vettura è disponibile dal concessionario o dal produttore
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	
/*	org := "Org2MSP"
	walletJSON, err := ctx.GetStub().GetState(org)
	
	if err != nil {
		return false, fmt.Errorf("Errore nella lettura dello stato globale: %v", err)
	}
	
	var wallet Wallet
	err = json.Unmarshal(walletJSON, &wallet)
	
	_ , presente := wallet.NFT[id]
	
	if presente {
		return true, nil;
	}
	*/
	org := "Org1MSP"
	walletJSON, err := ctx.GetStub().GetState(org)
	var wallet Wallet
	if err != nil {
		return false, fmt.Errorf("Errore nella lettura dello stato globale: %v", err)
	}
	
	err = json.Unmarshal(walletJSON, &wallet)
	
	if err != nil {
		return false, fmt.Errorf("Errore nell'unmarshalling: %v", err)
	}
	
	_ , presente := wallet.NFT[id]
	
	return presente, nil
}

// AssetExistsProducer controlla l'esistenza dell'ID specificato all'interno del wallet del produttore cioè se la vettura è disponibile nel deposito del produttore
func (s *SmartContract) AssetExistsProducer(ctx contractapi.TransactionContextInterface, model string) (bool, error) {
	
	org := "Org1MSP"
	walletJSON, err := ctx.GetStub().GetState(org)
	
	if err != nil {
		return false, fmt.Errorf("Errore nella lettura dello stato globale: %v", err)
	}
	
	var wallet Wallet
	err = json.Unmarshal(walletJSON, &wallet)
	if err != nil {
		return false, fmt.Errorf("Errore nell'unmarshalling: %v", err)
	}
	for key := range wallet.NFT {
		if wallet.NFT[key].Model == model {
			return true, nil
		}
	}
		return false,nil
	}


/*
// TransferAsset trasferisce un token da un wallet ad un altro
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
	
	
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	
	if err != nil {
		return err
	}
	
	if (newOwner == "Org1MSP" || newOwner == "Org2MSP") && newOwner != clientOrgID {
	 
		var wallet1, wallet2 Wallet
		
		walletJSON1, err := ctx.GetStub().GetState("Org1MSP")
		if err != nil {
				return fmt.Errorf("Errore get state wallet1: " , err)
			}
		if len(walletJSON1)==0{
			return fmt.Errorf("lunghezza 0")
		}
		
		
			
			
		walletJSON2, err := ctx.GetStub().GetState("Org2MSP")
		if err != nil {
				return fmt.Errorf("Errore get state wallet2: " , err)
			}
		if walletJSON2==nil {
			return fmt.Errorf("lunghezza 00")
		}
		
			
			
		err = json.Unmarshal(walletJSON1, &wallet1)
		if err != nil {
			return fmt.Errorf("---walletjson1:--- " + string(walletJSON1))
		}
		
		err = json.Unmarshal(walletJSON2, &wallet2)
		if err != nil {
			return fmt.Errorf("---walletjson2:--- " + string(walletJSON2))
		}
		
		if clientOrgID == "Org1MSP" {
			_, presente := wallet1.NFT[id]
			
			if ! presente {
				return fmt.Errorf("ID non trovato")
			}
			
			wallet2.NFT[id] = wallet1.NFT[id]
			delete(wallet1.NFT, id)
			
			
			walletJSON1, err = json.Marshal(wallet1)
			if err != nil {
				fmt.Errorf("Errore nel marshalling")
			}
			
			walletJSON2, err = json.Marshal(wallet2)
			if err != nil {
				fmt.Errorf("Errore nel marshalling")
			}
			
			err = ctx.GetStub().PutState("Org1MSP", walletJSON1)
			if err != nil {
				fmt.Errorf("Errore nella scrittura di wallet1 nello stato globale")
			}
			
			err = ctx.GetStub().PutState("Org2MSP", walletJSON2)
			if err != nil {
				fmt.Errorf("Errore nella scrittura di wallet2 nello stato globale")
			}
			
			return nil
			
		} else if clientOrgID == "Org2MSP" {
			
			_, presente := wallet2.NFT[id]
			
			if ! presente {
				return fmt.Errorf("ID non trovato")
			}
			
			wallet1.NFT[id] = wallet2.NFT[id]
			delete(wallet2.NFT, id)
			
			
			walletJSON1, err = json.Marshal(wallet1)
			if err != nil {
				fmt.Errorf("Errore nel marshalling")
			}
			
			walletJSON2, err = json.Marshal(wallet2)
			if err != nil {
				fmt.Errorf("Errore nel marshalling")
			}
			
			err = ctx.GetStub().PutState("Org1MSP", walletJSON1)
			if err != nil {
				fmt.Errorf("Errore nella scrittura di wallet1 nello stato globale")
			}
			
			err = ctx.GetStub().PutState("Org2MSP", walletJSON2)
			if err != nil {
				fmt.Errorf("Errore nella scrittura di wallet2 nello stato globale")
			}
			
			return nil
		} else {
			return fmt.Errorf("Attenzione, controlla chi riceve la vettura.")
		}	
	} else {
		return fmt.Errorf("Attenzione, controlla chi riceve la vettura")
	}
}

// TransferAsset trasferisce un token da un wallet ad un altro
func (s *SmartContract) AutomaticTransferAsset(ctx contractapi.TransactionContextInterface, id string) error {
	
	
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	
	if err != nil {
		return err
	}
	
	if (clientOrgID == "Org2MSP") {
	 
		var wallet1, wallet2 Wallet
		
		walletJSON1, err := ctx.GetStub().GetState("Org1MSP")
		walletJSON2, err := ctx.GetStub().GetState("Org2MSP")
		
		err = json.Unmarshal(walletJSON1, &wallet1)
		if err != nil {
			return fmt.Errorf("Errore nell'unmarshalling: %v", err)
		}
		
		err = json.Unmarshal(walletJSON2, &wallet2)
		if err != nil {
			return fmt.Errorf("Errore nell'unmarshalling: %v", err)
		}
		
			_, presente := wallet1.NFT[id]
			
			if ! presente {
				return fmt.Errorf("ID non trovato in fabbrica")
			}
			
			wallet2.NFT[id] = wallet1.NFT[id]
			delete(wallet1.NFT, id)
			
			
			walletJSON1, err = json.Marshal(wallet1)
			if err != nil {
				fmt.Errorf("Errore nel marshalling")
			}
			
			walletJSON2, err = json.Marshal(wallet2)
			if err != nil {
				fmt.Errorf("Errore nel marshalling")
			}
			
			err = ctx.GetStub().PutState("Org1MSP", walletJSON1)
			if err != nil {
				fmt.Errorf("Errore nella scrittura di wallet1 nello stato globale")
			}
			
			err = ctx.GetStub().PutState("Org2MSP", walletJSON2)
			if err != nil {
				fmt.Errorf("Errore nella scrittura di wallet2 nello stato globale")
			}
			
			return nil
	} else {
		return fmt.Errorf("Attenzione, controlla chi riceve la vettura")
	}
}
*/

// GetAllWallets ritorna tutti i wallet nello stato globale
func (s *SmartContract) GetAllWallets(ctx contractapi.TransactionContextInterface) ([]*Wallet, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var wallets []*Wallet
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var wallet Wallet
		err = json.Unmarshal(queryResponse.Value, &wallet)
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, &wallet)
	}

	return wallets, nil
}

// GetWallet returns wallet of org1
func (s *SmartContract) GetWallet(ctx contractapi.TransactionContextInterface) (map[string]Asset, error) {
	var wallet Wallet
	walletJSON, err := ctx.GetStub().GetState("Org1MSP")
	
	if err != nil {
		return wallet.NFT,fmt.Errorf("Errore nella lettura dello stato globale: %v", err)
	}
	
	err = json.Unmarshal(walletJSON, &wallet)
	if err != nil {
		return wallet.NFT,fmt.Errorf("Errore nell'unmarshalling: %v", err)
	}
	return wallet.NFT,nil
	}







