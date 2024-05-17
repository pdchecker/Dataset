package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
	"log"
	"math"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type BawangContract struct {
	contractapi.Contract
}

type UserContract struct {
	contractapi.Contract
}

type User struct {
	ID   string `json:"id"`
	NoHP string `json:"noHP"`
	Nama string `json:"nama"`

	Username	string `json:"username"`
	OrgName 	string `json:"orgName"`

	TTL  	string	`json:"ttl"`
	NoKK   	int64 	`json:"noKK"`
	NoNPWP 	int64 	`json:"noNPWP"`
	NIK 	int64 	`json:"nik"`

	LuasLahanHa 	float64 `json:"luasLahanHa"`
	AlamatToko 		string 	`json:"alamatToko"`
	AlamatLahan 	string 	`json:"alamatLahan"`
	KelompokTani 	string 	`json:"kelompokTani"`

	CreatedAt int64 `json:"createdAt"`
}

// Bawang struct
type Bawang struct {
	ID 				string `json:"id"` // for query
	BenihAsetID 	string `json:"benihAsetID"`
	BawangAsetID 	string `json:"bawangAsetID"`

	UsernamePengirim string `json:"usernamePengirim"`
	UsernamePenerima string `json:"usernamePenerima"`
	AlamatPengirim   string `json:"alamatPengirim"`
	AlamatPenerima   string `json:"alamatPenerima"`

	KuantitasBenihKg 	float64 `json:"kuantitasBenihKg"`
	HargaBenihPerKg  	float64 `json:"hargaBenihPerKg"`
	HargaBenihTotal		float64	`json:"hargaBenihTotal"`

	KuantitasBawangKg 	float64 `json:"kuantitasBawangKg"`
	HargaBawangPerKg  	float64 `json:"hargaBawangPerKg"`
	HargaBawangTotal	float64	`json:"hargaBawangTotal"`

	CreatedAt 	int64	`json:"createdAt"`

	// Unique Value
	// From Penangkar
	UmurBenih       string `json:"umurBenih"`
	UmurPanen       string `json:"umurPanen"`
	LamaPenyimpanan string `json:"lamaPenyimpanan"`
	Varietas        string `json:"varietas"`

	// From Petani
	UkuranUmbi    	string 	`json:"ukuranUmbi"`
	KadarAirPersen 	float64 `json:"kadarAirPersen"`
	Pupuk         	string 	`json:"pupuk"`
	Pestisida     	string 	`json:"pestisida"`
	Perlakuan     	string 	`json:"perlakuan"`
	Produktivitas 	string 	`json:"produktivitas"`
	TanggalTanam 	int64 	`json:"tanggalTanam"`
	TanggalPanen 	int64 	`json:"tanggalPanen"`

	// From Pengumpul
	TanggalMasuk     int64 `json:"tanggalMasuk"`
	TeknikSorting    string `json:"teknikSorting"`
	MetodePengemasan string `json:"metodePengemasan"`

	TxID1 string `json:"txID1"` // penangkar - petani
	TxID2 string `json:"txID2"` // petani - pengumpul
	TxID3 string `json:"txID3"` // pengumpul - pedagang

	IsAsset 	bool `json:"isAsset"`
	IsConfirmed bool `json:"isConfirmed"`
	IsEmpty		bool `json:"isEmpty"`
	IsRejected 	bool `json:"isRejected"`

	RejectReason	string	`json:"rejectReason"`
}

// Creating benih asset and insert it on blockchain
// Crating new benih aset
func (s *BawangContract) CreateBenih(ctx contractapi.TransactionContextInterface, bawangData string) (string, error) {
	if len(bawangData) == 0 {
		return "", fmt.Errorf("Please pass the correct benih data")
	}

	bawang := new(Bawang)

	bawang.CreatedAt = time.Now().Unix()
	// Set ID (key)
	bawang.ID = ctx.GetStub().GetTxID()

	// Set BenihAsetID
	bawang.BenihAsetID = ctx.GetStub().GetTxID()

	bawang.IsAsset = true

	bawang.IsConfirmed = false

	bawang.IsRejected = false

	// insert UmurBenih, UmurPanen, LamaPenyimpanan, Varietas, KuantitasBenihKg
	err := json.Unmarshal([]byte(bawangData), &bawang)

	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling benih. %s", err.Error())
	}

	bawangAsBytes, err := json.Marshal(bawang)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling benih. %s", err.Error())
	}

	ctx.GetStub().SetEvent("CreateAsset", bawangAsBytes)

	// Put state using key and data
	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(bawang.ID, bawangAsBytes)
}

// Creating benih trx from penangkar to petani, add txid1
// Create tx1 with the unique from the benih
func (s *BawangContract) CreateTrxBawangByPenangkar(ctx contractapi.TransactionContextInterface, bawangData, prevID string) (string, error) {

	if len(bawangData) == 0 {
		return "", fmt.Errorf("Please pass the correct benih data")
	}

	if len(prevID) == 0 {
		return "", fmt.Errorf("Please pass the correct benih aset id")
	}

	bawangPrev, err := s.GetBawangByID(ctx, prevID) 

	if err != nil {
		return "", fmt.Errorf("Failed while getting benih aset. %s", err.Error())
	}

	bawangNew := new(Bawang)

	bawangNew.CreatedAt = time.Now().Unix()

	// Set ID (key)
	bawangNew.ID = ctx.GetStub().GetTxID()
	// Set TxID
	bawangNew.TxID1 = ctx.GetStub().GetTxID()

	// Set asetIDs
	bawangNew.BenihAsetID = bawangPrev.BenihAsetID

	bawangNew.IsAsset = false
	bawangNew.IsConfirmed = false

	// Get penangkar unique value from benih aset
	bawangNew.UmurBenih = bawangPrev.UmurBenih
	bawangNew.UmurPanen = bawangPrev.UmurPanen
	bawangNew.LamaPenyimpanan = bawangPrev.LamaPenyimpanan
	bawangNew.Varietas = bawangPrev.Varietas
	bawangNew.KuantitasBenihKg = bawangPrev.KuantitasBenihKg
	bawangNew.HargaBenihPerKg = bawangPrev.HargaBenihPerKg

	// insert UsernamePengirim, UsernamePenerima, AlamatPengirim, AlamatPenerima, KuantitasBenihKg, HargaBenihPerKg
	err = json.Unmarshal([]byte(bawangData), &bawangNew)

	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling benih. %s", err.Error())
	}

	// // check identity
	// alamatPengirim, err := s.checkUserbyUsername(ctx, bawangNew.UsernamePengirim)

	// if err != nil {
	// 	return "", fmt.Errorf("Failed while checking username pengirim. %s", err.Error())
	// }

	// bawangNew.AlamatPengirim = alamatPengirim

	// alamatPenerima, err := s.checkUserbyUsername(ctx, bawangNew.UsernamePenerima)

	// if err != nil {
	// 	return "", fmt.Errorf("Failed while checking username penerima. %s", err.Error())
	// }

	// bawangNew.AlamatPenerima = alamatPenerima

	// calculate benih total price
	bawangNew.HargaBenihTotal = bawangNew.KuantitasBenihKg * bawangNew.HargaBenihPerKg

	// check quantity
	if bawangPrev.KuantitasBenihKg - bawangNew.KuantitasBenihKg >= 0 {

		_, err = s.updateKuantitasBenihByID(ctx, prevID, bawangNew.KuantitasBenihKg)

		if err != nil {
			return "", fmt.Errorf("Failed while updating benih aset kuantitas. %s", err.Error())
		}

		bawangAsBytes, err := json.Marshal(bawangNew)

		if err != nil {
			return "", fmt.Errorf("Failed while marshling benih. %s", err.Error())
		}
	
		ctx.GetStub().SetEvent("CreateAsset", bawangAsBytes)
	
		// Put state using key and data
		return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(bawangNew.ID, bawangAsBytes)
	} else {
		return "", fmt.Errorf("Benih quantity is not sufficient. %s", err.Error())
	}
}

// Planting benih by petani, inserting pupuk and timestamp for tanggalTanam
// Creating new soon to be bawang asset
func (s *BawangContract) PlantBenih(ctx contractapi.TransactionContextInterface, bawangData, prevID string) (string, error) {
	
	if len(bawangData) == 0 {
		return "", fmt.Errorf("Please pass the correct bawang data")
	}

	if len(prevID) == 0 {
		return "", fmt.Errorf("Please pass the correct bawang transaction id")
	}

	bawangPrev, err := s.GetBawangByID(ctx, prevID) 

	if err != nil {
		return "", fmt.Errorf("Failed while getting bawang init. %s", err.Error())
	}

	// set the benih quantity that petani got from penangkar to zero
	// petani plant all of its benih
	// make easier to query which benih that has been planted
	_, err = s.updateKuantitasBenihByID(ctx, prevID, bawangPrev.KuantitasBenihKg)

	bawangNew := new(Bawang)

	bawangNew.CreatedAt = time.Now().Unix()
	
	// Set ID (key)
	bawangNew.ID = ctx.GetStub().GetTxID()

	// Set asetID
	bawangNew.BawangAsetID = ctx.GetStub().GetTxID()
	bawangNew.BenihAsetID = bawangPrev.BenihAsetID

	bawangNew.TxID1 = bawangPrev.TxID1

	bawangNew.IsAsset = true
	bawangNew.IsConfirmed = false

	// Get penangkar unique field from bawangPrev
	bawangNew.UmurBenih = bawangPrev.UmurBenih
	bawangNew.UmurPanen = bawangPrev.UmurPanen
	bawangNew.LamaPenyimpanan = bawangPrev.LamaPenyimpanan
	bawangNew.Varietas = bawangPrev.Varietas
	bawangNew.KuantitasBenihKg = bawangPrev.KuantitasBenihKg
	bawangNew.HargaBenihPerKg = bawangPrev.HargaBenihPerKg
	bawangNew.HargaBenihTotal = bawangPrev.HargaBenihTotal

	// Insert Tanggal Tanam and Pupuk
	bawangNew.TanggalTanam = time.Now().Unix()

	// insert pupuk and username pengirim
	err = json.Unmarshal([]byte(bawangData), &bawangNew)

	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling bawang. %s", err.Error())
	}

	bawangAsBytes, err := json.Marshal(bawangNew)
	
	if err != nil {
		return "", fmt.Errorf("Failed while marshling bawang. %s", err.Error())
	}

	ctx.GetStub().SetEvent("CreateAsset", bawangAsBytes)

	// Put state using key and data
	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(bawangNew.ID, bawangAsBytes)
}

// Convert benih that has been planted into bawang asset 
// Add other bawang field
func (s *BawangContract) HarvestBawang(ctx contractapi.TransactionContextInterface, bawangData, bawangID string) (string, error) {
	
	if len(bawangData) == 0 {
		return "", fmt.Errorf("Please pass the correct bawang data")
	}

	if len(bawangID) == 0 {
		return "", fmt.Errorf("Please pass the correct bawang id")
	}

	bawangAsBytes, err := ctx.GetStub().GetState(bawangID)

	if err != nil {
		return "", fmt.Errorf("Failed to get bawang data. %s", err.Error())
	}

	if bawangAsBytes == nil {
		return "", fmt.Errorf("%s does not exist", bawangID)
	}

	// Create new bawang object
	bawang := new(Bawang)
	_ = json.Unmarshal(bawangAsBytes, bawang)

	// insert UkuranUmbi, KadarAirPersen, Pupuk, Pestisida, Perlakuan, Produktivitas, KuantitasBawangKg, TanggalPanen
	bawang.TanggalPanen = time.Now().Unix()
	err = json.Unmarshal([]byte(bawangData), &bawang)

	if err != nil {
		return "", fmt.Errorf("Failed while marshling bawang. %s", err.Error())
	}

	bawangAsBytes, err = json.Marshal(bawang)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling bawang. %s", err.Error())
	}

	// return the previous id and update the state
	return bawangID, ctx.GetStub().PutState(bawang.ID, bawangAsBytes)
}

// Updating bawang trx by petani, sending bawang to pengumpul, adding unique value from petani and txid2
func (s *BawangContract) UpdateBawangTrxByPetani(ctx contractapi.TransactionContextInterface, bawangData, prevID string) (string, error) {

	if len(bawangData) == 0 {
		return "", fmt.Errorf("Please pass the correct bawang data")
	}

	if len(prevID) == 0 {
		return "", fmt.Errorf("Please pass the correct bawang aset id")
	}

	bawangPrev, err := s.GetBawangByID(ctx, prevID) 

	if err != nil {
		return "", fmt.Errorf("Failed while getting bawang aset. %s", err.Error())
	}

	bawangNew := new(Bawang)

	bawangNew.CreatedAt = time.Now().Unix()

	// Set ID (key)
	bawangNew.ID = ctx.GetStub().GetTxID()
	// Set TxID
	bawangNew.TxID1 = bawangPrev.TxID1
	bawangNew.TxID2 = ctx.GetStub().GetTxID()

	// Set asetID
	bawangNew.BawangAsetID = bawangPrev.BawangAsetID
	bawangNew.BenihAsetID = bawangPrev.BenihAsetID

	bawangNew.IsAsset = false
	bawangNew.IsConfirmed = false

	// Get penangkar unique field from bawang aset
	bawangNew.UmurBenih = bawangPrev.UmurBenih
	bawangNew.UmurPanen = bawangPrev.UmurPanen
	bawangNew.LamaPenyimpanan = bawangPrev.LamaPenyimpanan
	bawangNew.Varietas = bawangPrev.Varietas
	bawangNew.KuantitasBenihKg = bawangPrev.KuantitasBenihKg
	bawangNew.HargaBenihPerKg = bawangPrev.HargaBenihPerKg
	bawangNew.HargaBenihTotal = bawangPrev.HargaBenihTotal

	// Get petani unique field from bawang aset
	bawangNew.UkuranUmbi = bawangPrev.UkuranUmbi
	bawangNew.KadarAirPersen = bawangPrev.KadarAirPersen
	bawangNew.Pupuk = bawangPrev.Pupuk
	bawangNew.Pestisida = bawangPrev.Pestisida
	bawangNew.Perlakuan = bawangPrev.Perlakuan
	bawangNew.Produktivitas = bawangPrev.Produktivitas
	bawangNew.TanggalTanam = bawangPrev.TanggalTanam
	bawangNew.TanggalPanen = bawangPrev.TanggalPanen

	// insert UsernamePengirim, UsernamePenerima, AlamatPengirim, AlamatPenerima, KuantitasBawangKg, HargaBawangPerKg
	err = json.Unmarshal([]byte(bawangData), &bawangNew)

	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling bawang. %s", err.Error())
	}

	// calculate bawang total price
	bawangNew.HargaBawangTotal = bawangNew.KuantitasBawangKg * bawangNew.HargaBawangPerKg

	// check quantity
	if bawangPrev.KuantitasBawangKg - bawangNew.KuantitasBawangKg >= 0 {

		_, err = s.updateKuantitasBawangByID(ctx, prevID, bawangNew.KuantitasBawangKg)

		if err != nil {
			return "", fmt.Errorf("Failed while updating bawang aset kuantitas. %s", err.Error())
		}

		bawangAsBytes, err := json.Marshal(bawangNew)

		if err != nil {
			return "", fmt.Errorf("Failed while marshling bawang. %s", err.Error())
		}

		ctx.GetStub().SetEvent("CreateAsset", bawangAsBytes)
	
		// Put state using key and data
		return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(bawangNew.ID, bawangAsBytes)
	} else {
		return "", fmt.Errorf("Bawang quantity is not sufficient. %s", err.Error())
	}
}

// Updating bawang trx by pengumpul, sending bawang to pedagang, adding unique value from pengumpul and txid3
func (s *BawangContract) UpdateBawangTrxByPengumpul(ctx contractapi.TransactionContextInterface, bawangData, prevID string) (string, error) {

	if len(bawangData) == 0 {
		return "", fmt.Errorf("Please pass the correct bawang data")
	}

	if len(prevID) == 0 {
		return "", fmt.Errorf("Please pass the correct bawang aset id")
	}

	bawangPrev, err := s.GetBawangByID(ctx, prevID) 

	if err != nil {
		return "", fmt.Errorf("Failed while getting bawang aset. %s", err.Error())
	}

	bawangNew := new(Bawang)

	bawangNew.CreatedAt = time.Now().Unix()

	// Set id (key)
	bawangNew.ID = ctx.GetStub().GetTxID()
	// Set txid1
	bawangNew.TxID1 = bawangPrev.TxID1
	// Set txid2
	bawangNew.TxID2 = bawangPrev.TxID2
	// Set txid3
	bawangNew.TxID3 = ctx.GetStub().GetTxID()

	// Set asetID
	bawangNew.BawangAsetID = bawangPrev.BawangAsetID
	bawangNew.BenihAsetID = bawangPrev.BenihAsetID

	bawangNew.IsAsset = false
	bawangNew.IsConfirmed = false

	// Add unique field from penangkar
	bawangNew.UmurBenih = bawangPrev.UmurBenih
	bawangNew.UmurPanen = bawangPrev.UmurPanen
	bawangNew.LamaPenyimpanan = bawangPrev.LamaPenyimpanan
	bawangNew.Varietas = bawangPrev.Varietas
	bawangNew.KuantitasBenihKg = bawangPrev.KuantitasBenihKg
	bawangNew.HargaBenihPerKg = bawangPrev.HargaBenihPerKg
	bawangNew.HargaBenihTotal = bawangPrev.HargaBenihTotal

	// Add unique field from petani
	bawangNew.UkuranUmbi = bawangPrev.UkuranUmbi
	bawangNew.KadarAirPersen = bawangPrev.KadarAirPersen
	bawangNew.Pupuk = bawangPrev.Pupuk
	bawangNew.Pestisida = bawangPrev.Pestisida
	bawangNew.Perlakuan = bawangPrev.Perlakuan
	bawangNew.Produktivitas = bawangPrev.Produktivitas
	bawangNew.TanggalTanam = bawangPrev.TanggalTanam
	bawangNew.TanggalPanen = bawangPrev.TanggalPanen

	// insert TanggalMasuk, TeknikSorting, MetodePengemasan, UsernamePengirim, UsernamePenerima, KuantitasBawangKg, HargaBawangPerKg
	err = json.Unmarshal([]byte(bawangData), &bawangNew)

	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling benih. %s", err.Error())
	}

	// calculate bawang total price
	bawangNew.HargaBawangTotal = bawangNew.KuantitasBawangKg * bawangNew.HargaBawangPerKg

	// check quantity
	if bawangPrev.KuantitasBawangKg - bawangNew.KuantitasBawangKg >= 0 {

		_, err = s.updateKuantitasBawangByID(ctx, prevID, bawangNew.KuantitasBawangKg)

		if err != nil {
			return "", fmt.Errorf("Failed while updating bawang aset kuantitas. %s", err.Error())
		}

		bawangAsBytes, err := json.Marshal(bawangNew)

		if err != nil {
			return "", fmt.Errorf("Failed while marshling bawang. %s", err.Error())
		}
	
		// Insert into blockchain
		ctx.GetStub().SetEvent("CreateAsset", bawangAsBytes)
	
		// Put state using key and data
		return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(bawangNew.ID, bawangAsBytes)
	} else {
		return "", fmt.Errorf("Bawang quantity is not sufficient. %s", err.Error())
	}
}

// Add benih asset quantity by ID
func (s *BawangContract) AddBenihKuantitasByID(ctx contractapi.TransactionContextInterface, quantity float64, bawangID string) (*Bawang, error) {
	if len(bawangID) == 0 {
		return nil, fmt.Errorf("Please pass the correct bawang id")
	}

	bawangAsBytes, err := ctx.GetStub().GetState(bawangID)

	if err != nil {
		return nil, fmt.Errorf("Failed to get bawang data. %s", err.Error())
	}

	if bawangAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", bawangID)
	}

	// Create new bawang object
	bawang := new(Bawang)
	_ = json.Unmarshal(bawangAsBytes, bawang)

	// Add benih quantity
	bawang.KuantitasBenihKg += quantity
	bawang.KuantitasBenihKg = math.Round(bawang.KuantitasBenihKg*100)/100

	if bawang.KuantitasBenihKg > 0 {
		bawang.IsEmpty = false
	}

	bawangAsBytes, err = json.Marshal(bawang)

	return bawang, ctx.GetStub().PutState(bawang.ID, bawangAsBytes)
}

// Add bawang asset quantity by ID
func (s *BawangContract) AddBawangKuantitasByID(ctx contractapi.TransactionContextInterface, quantity float64, bawangID string) (*Bawang, error) {
	if len(bawangID) == 0 {
		return nil, fmt.Errorf("Please pass the correct bawang id")
	}

	bawangAsBytes, err := ctx.GetStub().GetState(bawangID)

	if err != nil {
		return nil, fmt.Errorf("Failed to get bawang data. %s", err.Error())
	}

	if bawangAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", bawangID)
	}

	// Create new bawang object
	bawang := new(Bawang)
	_ = json.Unmarshal(bawangAsBytes, bawang)

	// Add bawang quantity
	bawang.KuantitasBawangKg += quantity

	if bawang.KuantitasBawangKg > 0 {
		bawang.IsEmpty = false
	}

	bawangAsBytes, err = json.Marshal(bawang)

	return bawang, ctx.GetStub().PutState(bawang.ID, bawangAsBytes)
}

// Check user identity and retrieve its address
func (s *BawangContract) checkUserbyUsername(ctx contractapi.TransactionContextInterface, username string) (string, error) {
	if len(username) == 0 {
		return "", fmt.Errorf("Please insert the correct username")
	}

	queryString := `{"selector":{"username":"`+ username +`"}}`

	res, err := s.GetBawangForQuery(ctx, queryString)

	log.Printf(res)

	type KeyRecord struct {
		Key string `json:"Key"`
		Record User `json:"Record"`
	}

	type Result struct {
		Result []KeyRecord `json:"result"`
		Error error `json:"error"`
		ErrorData error `json:"errorData"`
	}

	if err != nil {
		return "", fmt.Errorf("User not found. %s", err.Error())
	}

	result := new(Result)
	json.Unmarshal([]byte(res), &result)

	orgName := result.Result[0].Record.OrgName

	if orgName == "Penangkar" || orgName == "Petani" {
		return result.Result[0].Record.AlamatLahan, nil
	} else {
		return result.Result[0].Record.AlamatToko, nil
	}

	return "", nil
}

// Updating (subtracting) benih quantity by it's ID
func (s *BawangContract) updateKuantitasBenihByID(ctx contractapi.TransactionContextInterface, bawangID string, kuantitasBenihNext float64) (string, error) {
	if len(bawangID) == 0 {
		return "", fmt.Errorf("Please insert the correct benih aset id")
	}

	bawangAsBytes, err := ctx.GetStub().GetState(bawangID)

	if err != nil {
		return "", fmt.Errorf("Failed to get benih data. %s", err.Error())
	}

	if bawangAsBytes == nil {
		return "", fmt.Errorf("Benih aset with %s id does not exist", bawangID)
	}

	// Create new bawang object
	bawang := new(Bawang)
	_ = json.Unmarshal(bawangAsBytes, bawang)

	bawang.KuantitasBenihKg -= kuantitasBenihNext
	bawang.KuantitasBenihKg = math.Round(bawang.KuantitasBenihKg*100)/100


	if bawang.KuantitasBenihKg == 0 {
		bawang.IsEmpty = true
	}

	bawangAsBytes, err = json.Marshal(bawang)

	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(bawang.ID, bawangAsBytes)
}

// Updating (subtracting) bawang quantity by it's ID
func (s *BawangContract) updateKuantitasBawangByID(ctx contractapi.TransactionContextInterface, bawangID string, kuantitasBawangNext float64) (string, error) {
	if len(bawangID) == 0 {
		return "", fmt.Errorf("Please insert the correct bawang aset id")
	}

	bawangAsBytes, err := ctx.GetStub().GetState(bawangID)

	if err != nil {
		return "", fmt.Errorf("Failed to get bawang data. %s", err.Error())
	}

	if bawangAsBytes == nil {
		return "", fmt.Errorf("Bawang aset with %s id does not exist", bawangID)
	}

	// Create new bawang object
	bawang := new(Bawang)
	_ = json.Unmarshal(bawangAsBytes, bawang)

	bawang.KuantitasBawangKg -= kuantitasBawangNext
	bawang.KuantitasBawangKg = math.Round(bawang.KuantitasBawangKg*100)/100


	if bawang.KuantitasBawangKg == 0 {
		bawang.IsEmpty = true
	}

	bawangAsBytes, err = json.Marshal(bawang)

	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(bawang.ID, bawangAsBytes)
}

// Set isConfirm to true
func (s *BawangContract) ConfirmTrxByID(ctx contractapi.TransactionContextInterface, bawangID string) (string, error) {

	if len(bawangID) == 0 {
		return "", fmt.Errorf("Please pass the correct bawang id")
	}

	bawangAsBytes, err := ctx.GetStub().GetState(bawangID)

	if err != nil {
		return "", fmt.Errorf("Failed to get bawang data. %s", err.Error())
	}

	if bawangAsBytes == nil {
		return "", fmt.Errorf("%s does not exist", bawangID)
	}

	// Create new bawang object
	bawang := new(Bawang)
	_ = json.Unmarshal(bawangAsBytes, bawang)

	// Change the confirmed status
	bawang.IsConfirmed = true

	// Change createdAt
	bawang.CreatedAt = time.Now().Unix()

	// Set tanggal masuk when pengumpul confirmed the bawang trx
	if len(bawang.UkuranUmbi) != 0 || len(bawang.Pupuk) != 0 || 
	bawang.TanggalTanam != 0 || bawang.TanggalPanen != 0 {
		bawang.TanggalMasuk = time.Now().Unix()
	}

	bawangAsBytes, err = json.Marshal(bawang)

	if err != nil {
		return "", fmt.Errorf("Failed while marshling bawang. %s", err.Error())
	}

	// Update the state
	return bawangID, ctx.GetStub().PutState(bawang.ID, bawangAsBytes)
}

// Set isRejected to true and retrieve the asset quantity
func (s *BawangContract) RejectTrxByID(ctx contractapi.TransactionContextInterface, bawangIDPrev, bawangIDReject string, kuantitasPrev float64, rejectReason string) (string, error) {
	
	// Create new bawang object for previous bawang
	if len(bawangIDPrev) == 0 {
		return "", fmt.Errorf("Please pass the correct previous bawang id")
	}

	bawangPrevAsBytes, err := ctx.GetStub().GetState(bawangIDPrev)

	if err != nil {
		return "", fmt.Errorf("Failed to get bawang data. %s", err.Error())
	}

	if bawangPrevAsBytes == nil {
		return "", fmt.Errorf("%s does not exist", bawangIDPrev)
	}
	
	bawangPrev := new(Bawang)
	_ = json.Unmarshal(bawangPrevAsBytes, bawangPrev)

	if len(bawangIDReject) == 0 {
		return "", fmt.Errorf("Please pass the correct rejected bawang id")
	}

	bawangRejectAsBytes, err := ctx.GetStub().GetState(bawangIDReject)

	if err != nil {
		return "", fmt.Errorf("Failed to get bawang data. %s", err.Error())
	}

	if bawangRejectAsBytes == nil {
		return "", fmt.Errorf("%s does not exist", bawangIDReject)
	}

	bawangReject := new(Bawang)
	_ = json.Unmarshal(bawangRejectAsBytes, bawangReject)

	// Change the rejected status
	bawangReject.IsRejected = true

	if bawangPrev.BenihAsetID != "" && bawangPrev.BawangAsetID == "" {
		bawangPrev.KuantitasBenihKg += kuantitasPrev
	} else {
		bawangPrev.KuantitasBawangKg += kuantitasPrev
	}

	bawangReject.RejectReason = rejectReason

	bawangPrevAsBytes, err = json.Marshal(bawangPrev)
	bawangRejectAsBytes, err = json.Marshal(bawangReject)

	if err != nil {
		return "", fmt.Errorf("Failed while marshling bawang. %s", err.Error())
	}

	// Update the state
	ctx.GetStub().PutState(bawangPrev.ID, bawangPrevAsBytes)

	return bawangIDReject, ctx.GetStub().PutState(bawangReject.ID, bawangRejectAsBytes)
}

// Get bawang object by ID
func (s *BawangContract) GetBawangByID(ctx contractapi.TransactionContextInterface, bawangID string) (*Bawang, error) {
	if len(bawangID) == 0 {
		return nil, fmt.Errorf("Please provide correct contract ID")
		// return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// fetch bawang by its ID
	bawangAsBytes, err := ctx.GetStub().GetState(bawangID)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if bawangAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", bawangID)
	}

	// create new Bawang object and unmarshal bawangAsBytes into bawang
	bawang := new(Bawang)
	_ = json.Unmarshal(bawangAsBytes, bawang)

	return bawang, nil
}

// CreateUser function to create user and insert it on blockchain
func (s *UserContract) CreateUser(ctx contractapi.TransactionContextInterface, userData string) (string, error) {

	if len(userData) == 0 {
		return "", fmt.Errorf("Please pass the correct bawang data")
	}

	var user User

	user.CreatedAt = time.Now().Unix()

	//create user ID
	user.ID = ctx.GetStub().GetTxID()

	err := json.Unmarshal([]byte(userData), &user)
	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling user. %s", err.Error())
	}

	userAsBytes, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling bawang. %s", err.Error())
	}

	// Insert into blockchain
	ctx.GetStub().SetEvent("CreateAsset", userAsBytes)

	// Put state using key and data
	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(user.ID, userAsBytes)
}

// GetUserByID get bawang object by ID
func (s *UserContract) GetUserByID(ctx contractapi.TransactionContextInterface, userID string) (*User, error) {
	if len(userID) == 0 {
		return nil, fmt.Errorf("Please provide correct contract ID")
		// return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// fetch bawang by its ID
	userAsBytes, err := ctx.GetStub().GetState(userID)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if userAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", userID)
	}

	// create new Bawang object and unmarshal bawangAsBytes into bawang
	user := new(User)
	_ = json.Unmarshal(userAsBytes, user)

	return user, nil
}

// it should be GetBawangForQuery
func (s *BawangContract) GetBawangForQuery(ctx contractapi.TransactionContextInterface, queryString string) (string, error) {

	queryResults, err := s.getQueryResultForQueryString(ctx, queryString)

	if err != nil {
		return "", fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	return queryResults, nil

}

func (s *BawangContract) getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) (string, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)
    resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
    defer resultsIterator.Close()
    if err != nil {
        return "", err
    }
    // buffer is a JSON array containing QueryRecords
    var buffer bytes.Buffer
    buffer.WriteString("[")
    bArrayMemberAlreadyWritten := false
    for resultsIterator.HasNext() {
        queryResponse,
        err := resultsIterator.Next()
        if err != nil {
            return "", err
        }
        // Add a comma before array members, suppress it for the first array member
        if bArrayMemberAlreadyWritten == true {
            buffer.WriteString(",")
        }
        buffer.WriteString("{\"Key\":")
        buffer.WriteString("\"")
        buffer.WriteString(queryResponse.Key)
        buffer.WriteString("\"")
        buffer.WriteString(", \"Record\":")
        // Record is a JSON object, so we write as-is
        buffer.WriteString(string(queryResponse.Value))
        buffer.WriteString("}")
        bArrayMemberAlreadyWritten = true
    }
    buffer.WriteString("]")
    fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())
    return string(buffer.Bytes()), nil
}

// GetHistoryForAssetByID to get history for asset by ID
func (s *BawangContract) GetHistoryForAssetByID(ctx contractapi.TransactionContextInterface, bawangID string) (string, error) {

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(bawangID)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return "", fmt.Errorf(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		// buffer.WriteString(", \"Timestamp\":")
		// buffer.WriteString("\"")
		// buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		// buffer.WriteString("\"")

		// buffer.WriteString(", \"IsDelete\":")
		// buffer.WriteString("\"")
		// buffer.WriteString(strconv.FormatBool(response.IsDelete))
		// buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return string(buffer.Bytes()), nil
}

// InitLedger
func (s *BawangContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// for i, bawang := range bawangs {
	// 	bawangAsBytes, _ := json.Marshal(bawang)
	// 	err := ctx.GetStub().PutState("Bawang"+strconv.Itoa(i), bawangAsBytes)

	// 	if err != nil {
	// 		return fmt.Errorf("Failed to put to world state. %s", err.Error())
	// 	}
	// }

	return nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(BawangContract), new(UserContract))
	if err != nil {
		fmt.Printf("Error create bawang merah chaincode: %s", err.Error())
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincodes: %s", err.Error())
	}

}
