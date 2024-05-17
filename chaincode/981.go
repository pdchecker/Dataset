/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const privateCollectionOrg1 = "Org1PrivateCollection"
const privateCollectionOrg1nOrg2 = "Org1AndOrg2PrivateCollection"

// SmartContract of this fabric sample
type SmartContract struct {
	contractapi.Contract
}

// AssetPublicDetails describes main asset details that are visible to all organizations
type AssetPublicDetails struct {
	ID     string `json:"assetID"`
	INECLV string `json:"ineClv"`
	Email  string `json:"email"`
	Owner  string `json:"owner"`
}

// AssetPrivateDetails describes details that are private to owners
type AssetPrivateDetails struct {
	ID                string `json:"assetID"`
	Telefono          string `json:"telefono"`
	Email             string `json:"email"`
	Pswrd             string `json:"pswrd"`
	Nombre            string `json:"nombre"`
	ApellidoPaterno   string `json:"apellidoPaterno"`
	ApellidoMaterno   string `json:"apellidoMaterno"`
	CURP              string `json:"curp"`
	INECLV            string `json:"ineClv"`
	AnioDeRegistro    string `json:"anioDeRegistro"`
	AnioDeEmision     string `json:"anioDeEmision"`
	Vigencia          string `json:"vigencia"`
	Calle             string `json:"calle"`
	Numero            int    `json:"numero"`
	Colonia           string `json:"colonia"`
	Localidad         string `json:"localidad"`
	Seccion           string `json:"seccion"`
	Municipio         string `json:"municipio"`
	Estado            string `json:"estado"`
	CodigoPostal      string `json:"codigoPostal"`
	OCR               string `json:"ocr"`
	IDCiudadano       string `json:"idCiudadano"`
	FechaDeNacimiento string `json:"fechaDeNacimiento"`
	Nacionalidad      string `json:"nacionalidad"`
	PaisDeResidencia  string `json:"paisDeResidencia"`
	TipoDeActividad   string `json:"tipoDeActividad"`
	NivelParecido     string `json:"nivelParecido"`
	Domicilio         string `json:"domicilio"`
	TipoDocumento     string `json:"tipoDocumento"`
	FechaDeProceso    string `json:"fechaDeProceso"`
	Genero            string `json:"genero"`
	MayorDeEdad       string `json:"mayorDeEdad"`
	CopiaBN           string `json:"copiaBN"`
	PruebaDeVida      string `json:"pruebaDeVida"`
	ImgRostro         string `json:"imgRostro"`
	ImgRostroID       string `json:"imgRostroID"`
	ImgIDFrontal      string `json:"imgIDFrontal"`
	ImgIDTrasera      string `json:"imgIDTrasera"`
	Owner             string `json:"owner"`
}

// verifyClientOrgMatchesPeerOrg is an internal function used verify client org id and matches peer org id.
func verifyClientOrgMatchesPeerOrg(ctx contractapi.TransactionContextInterface) error {
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the client's MSPID: %v", err)
	}
	peerMSPID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the peer's MSPID: %v", err)
	}

	if clientMSPID != peerMSPID {
		return fmt.Errorf("client from org %v is not authorized to read or write private data from an org %v peer", clientMSPID, peerMSPID)
	}

	return nil
}

// ====================== Onboarding Functions

// CreateAssetPrivateCollectionOrg1 creates a new asset by placing the main asset details in the assetCollection
// that can be read by both organizations. The appraisal value is stored in the owners org specific collection.
func (s *SmartContract) CreateAssetPrivateCollectionOrg1(ctx contractapi.TransactionContextInterface) error {

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		//log error to stdout
		return fmt.Errorf("asset not found in the transient map input")
	}

	type assetTransientInput struct {
		ID                string `json:"assetID"`
		Telefono          string `json:"telefono"`
		Email             string `json:"email"`
		Pswrd             string `json:"pswrd"`
		Nombre            string `json:"nombre"`
		ApellidoPaterno   string `json:"apellidoPaterno"`
		ApellidoMaterno   string `json:"apellidoMaterno"`
		CURP              string `json:"curp"`
		INECLV            string `json:"ineClv"`
		AnioDeRegistro    string `json:"anioDeRegistro"`
		AnioDeEmision     string `json:"anioDeEmision"`
		Vigencia          string `json:"vigencia"`
		Calle             string `json:"calle"`
		Numero            int    `json:"numero"`
		Colonia           string `json:"colonia"`
		Localidad         string `json:"localidad"`
		Seccion           string `json:"seccion"`
		Municipio         string `json:"municipio"`
		Estado            string `json:"estado"`
		CodigoPostal      string `json:"codigoPostal"`
		OCR               string `json:"ocr"`
		IDCiudadano       string `json:"idCiudadano"`
		FechaDeNacimiento string `json:"fechaDeNacimiento"`
		Nacionalidad      string `json:"nacionalidad"`
		PaisDeResidencia  string `json:"paisDeResidencia"`
		TipoDeActividad   string `json:"tipoDeActividad"`
		NivelParecido     string `json:"nivelParecido"`
		Domicilio         string `json:"domicilio"`
		TipoDocumento     string `json:"tipoDocumento"`
		FechaDeProceso    string `json:"fechaDeProceso"`
		Genero            string `json:"genero"`
		MayorDeEdad       string `json:"mayorDeEdad"`
		CopiaBN           string `json:"copiaBN"`
		PruebaDeVida      string `json:"pruebaDeVida"`
		ImgRostro         string `json:"imgRostro"`
		ImgRostroID       string `json:"imgRostroID"`
		ImgIDFrontal      string `json:"imgIDFrontal"`
		ImgIDTrasera      string `json:"imgIDTrasera"`
	}

	var assetInput assetTransientInput
	err = json.Unmarshal(transientAssetJSON, &assetInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetInput.ID) == 0 {
		return fmt.Errorf("assetID field must be a non-empty string")
	}
	if len(assetInput.Telefono) == 0 {
		return fmt.Errorf("telefono field must be a non-empty string")
	}
	if len(assetInput.Email) == 0 {
		return fmt.Errorf("email field must be a non-empty string")
	}
	if len(assetInput.Pswrd) == 0 {
		return fmt.Errorf("pswrd field must be a non-empty string")
	}
	if len(assetInput.Nombre) == 0 {
		return fmt.Errorf("nombre field must be a non-empty string")
	}
	if len(assetInput.ApellidoPaterno) == 0 {
		return fmt.Errorf("apellidoPaterno field must be a non-empty string")
	}
	if len(assetInput.ApellidoMaterno) == 0 {
		return fmt.Errorf("apellidoMaterno field must be a non-empty string")
	}
	if len(assetInput.CURP) == 0 {
		return fmt.Errorf("curp field must be a non-empty string")
	}
	if len(assetInput.INECLV) == 0 {
		return fmt.Errorf("ineClv field must be a non-empty string")
	}
	if len(assetInput.AnioDeRegistro) == 0 {
		return fmt.Errorf("anioDeRegistro field must be a non-empty string")
	}
	if len(assetInput.AnioDeEmision) == 0 {
		return fmt.Errorf("anioDeEmision field must be a non-empty string")
	}
	if len(assetInput.Vigencia) == 0 {
		return fmt.Errorf("vigencia field must be a non-empty string")
	}
	if len(assetInput.Calle) == 0 {
		return fmt.Errorf("calle field must be a non-empty string")
	}
	if assetInput.Numero <= 0 {
		return fmt.Errorf("numero field must be a positive integer")
	}
	if len(assetInput.Colonia) == 0 {
		return fmt.Errorf("colonia field must be a positive integer")
	}
	if len(assetInput.Localidad) == 0 {
		return fmt.Errorf("localidad field must be a positive integer")
	}
	if len(assetInput.Seccion) == 0 {
		return fmt.Errorf("seccion field must be a positive integer")
	}
	if len(assetInput.Municipio) == 0 {
		return fmt.Errorf("municipio field must be a positive integer")
	}
	if len(assetInput.Estado) == 0 {
		return fmt.Errorf("estado field must be a positive integer")
	}
	if len(assetInput.CodigoPostal) == 0 {
		return fmt.Errorf("codigoPostal field must be a positive integer")
	}
	if len(assetInput.OCR) == 0 {
		return fmt.Errorf("ocr field must be a positive integer")
	}
	if len(assetInput.IDCiudadano) == 0 {
		return fmt.Errorf("idCiudadano field must be a positive integer")
	}
	if len(assetInput.FechaDeNacimiento) == 0 {
		return fmt.Errorf("fechaDeNacimiento field must be a non-empty string")
	}
	if len(assetInput.Nacionalidad) == 0 {
		return fmt.Errorf("nacionalidad field must be a non-empty string")
	}
	if len(assetInput.PaisDeResidencia) == 0 {
		return fmt.Errorf("paisDeResidencia field must be a non-empty string")
	}
	if len(assetInput.TipoDeActividad) == 0 {
		return fmt.Errorf("tipoDeActividad field must be a non-empty string")
	}
	if len(assetInput.NivelParecido) == 0 {
		return fmt.Errorf("nivelDeParecido field must be a non-empty string")
	}
	if len(assetInput.Domicilio) == 0 {
		return fmt.Errorf("domicilio field must be a non-empty string")
	}
	if len(assetInput.TipoDocumento) == 0 {
		return fmt.Errorf("tipoDocumento field must be a non-empty string")
	}
	if len(assetInput.FechaDeProceso) == 0 {
		return fmt.Errorf("fechaDeProceso field must be a non-empty string")
	}
	if len(assetInput.Genero) == 0 {
		return fmt.Errorf("genero field must be a non-empty string")
	}
	if len(assetInput.MayorDeEdad) == 0 {
		return fmt.Errorf("mayorDeEdad field must be a non-empty string")
	}
	if len(assetInput.CopiaBN) == 0 {
		return fmt.Errorf("copiaBN field must be a non-empty string")
	}
	if len(assetInput.ImgRostro) == 0 {
		return fmt.Errorf("imgRostro field must be a non-empty string")
	}
	if len(assetInput.ImgRostroID) == 0 {
		return fmt.Errorf("imgRostroID field must be a non-empty string")
	}
	if len(assetInput.ImgIDFrontal) == 0 {
		return fmt.Errorf("imgIDFrontal field must be a non-empty string")
	}
	if len(assetInput.ImgIDTrasera) == 0 {
		return fmt.Errorf("imgIDTrasera field must be a non-empty string")
	}

	// Check if asset already exists
	assetAsBytes, err := ctx.GetStub().GetPrivateData(privateCollectionOrg1, assetInput.ID)
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	} else if assetAsBytes != nil {
		fmt.Println("Asset already exists: " + assetInput.ID)
		return fmt.Errorf("this asset already exists: " + assetInput.ID)
	}

	// Get ID of submitting client identity
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get verified OrgID: %v", err)
	}

	// Verify that the client is submitting request to peer in their organization
	// This is to ensure that a client from another org doesn't attempt to read or
	// write private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateAsset cannot be performed: Error %v", err)
	}

	// Public state
	// Make submitting client the owner
	assetPublic := AssetPublicDetails{
		ID:     assetInput.ID,
		INECLV: assetInput.INECLV,
		Email:  assetInput.Email,
		Owner:  clientID,
	}
	assetPublicJSONasBytes, err := json.Marshal(assetPublic)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}

	// escibimos en el public state
	err = ctx.GetStub().PutState(assetInput.ID, assetPublicJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put to public state. %v", err)
	}

	// Make submitting client the owner
	asset := AssetPrivateDetails{
		ID:                assetInput.ID,
		Telefono:          assetInput.Telefono,
		Email:             assetInput.Email,
		Pswrd:             assetInput.Pswrd,
		Nombre:            assetInput.Nombre,
		ApellidoPaterno:   assetInput.ApellidoPaterno,
		ApellidoMaterno:   assetInput.ApellidoMaterno,
		CURP:              assetInput.CURP,
		INECLV:            assetInput.INECLV,
		AnioDeRegistro:    assetInput.AnioDeRegistro,
		AnioDeEmision:     assetInput.AnioDeEmision,
		Vigencia:          assetInput.Vigencia,
		Calle:             assetInput.Calle,
		Numero:            assetInput.Numero,
		Colonia:           assetInput.Colonia,
		Localidad:         assetInput.Localidad,
		Seccion:           assetInput.Seccion,
		Municipio:         assetInput.Municipio,
		Estado:            assetInput.Estado,
		CodigoPostal:      assetInput.CodigoPostal,
		OCR:               assetInput.OCR,
		IDCiudadano:       assetInput.IDCiudadano,
		FechaDeNacimiento: assetInput.FechaDeNacimiento,
		Nacionalidad:      assetInput.Nacionalidad,
		PaisDeResidencia:  assetInput.PaisDeResidencia,
		TipoDeActividad:   assetInput.TipoDeActividad,
		NivelParecido:     assetInput.NivelParecido,
		Domicilio:         assetInput.Domicilio,
		TipoDocumento:     assetInput.TipoDocumento,
		FechaDeProceso:    assetInput.FechaDeProceso,
		Genero:            assetInput.Genero,
		MayorDeEdad:       assetInput.MayorDeEdad,
		CopiaBN:           assetInput.CopiaBN,
		PruebaDeVida:      assetInput.PruebaDeVida,
		ImgRostro:         assetInput.ImgRostro,
		ImgRostroID:       assetInput.ImgRostroID,
		ImgIDFrontal:      assetInput.ImgIDFrontal,
		ImgIDTrasera:      assetInput.ImgIDTrasera,
		Owner:             clientID,
	}
	assetJSONasBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}

	// Save asset to private data collection
	// Typical logger, logs to stdout/file in the fabric managed docker container, running this chaincode
	// Look for container name like dev-peer0.org1.example.com-{chaincodename_version}-xyz
	log.Printf("CreateAsset Put: collection %v, ID %v", privateCollectionOrg1, assetInput.ID)
	err = ctx.GetStub().PutPrivateData(privateCollectionOrg1, assetInput.ID, assetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset into private data collecton: %v", err)
	}
	return nil
}

// TransferAssetToPrivateCollection transfers an asset by setting a new owner name on the asset
// that can be read by both organizations. The appraisal value is stored in the owners org specific collection.
func (s *SmartContract) TransferAssetToPrivateCollection(ctx contractapi.TransactionContextInterface) error {

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field
	transientTransferJSON, ok := transientMap["asset_properties"]
	if !ok {
		return fmt.Errorf("asset owner not found in the transient map")
	}

	type assetTransferTransientInput struct {
		ID   string `json:"assetID"`
		Org1 string `json:"org1"`
		Org2 string `json:"org2"`
	}

	var assetTransferInput assetTransferTransientInput
	err = json.Unmarshal(transientTransferJSON, &assetTransferInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetTransferInput.ID) == 0 {
		return fmt.Errorf("assetID field must be a non-empty string")
	}
	if len(assetTransferInput.Org1) == 0 {
		return fmt.Errorf("ownerMSP field must be a non-empty string")
	}
	if len(assetTransferInput.Org2) == 0 {
		return fmt.Errorf("buyerMSP field must be a non-empty string")
	}
	log.Printf("TransferAsset: verify asset exists ID %v", assetTransferInput.ID)
	// Read asset from the private data collection
	asset, err := s.ReadAssetPrivateDetails(ctx, privateCollectionOrg1, assetTransferInput.ID)
	if err != nil {
		return fmt.Errorf("error reading asset: %v", err)
	}
	if asset == nil {
		return fmt.Errorf("%v does not exist", assetTransferInput.ID)
	}
	// Verify that the client is submitting request to peer in their organization
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("TransferAsset cannot be performed: Error %v", err)
	}

	assetJSONasBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed marshalling asset %v: %v", assetTransferInput.ID, err)
	}

	log.Printf("TransferAsset Put: collection %v, ID %v", privateCollectionOrg1, assetTransferInput.ID)
	err = ctx.GetStub().PutPrivateData(privateCollectionOrg1nOrg2, assetTransferInput.ID, assetJSONasBytes) //rewrite the asset
	if err != nil {
		return err
	}

	return nil
}

// UpdateAssetPrivateCollection updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAssetPrivateCollection(ctx contractapi.TransactionContextInterface) error {

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field
	transientUpdateJSON, ok := transientMap["asset_properties"]
	if !ok {
		return fmt.Errorf("asset owner not found in the transient map")
	}

	var assetUpdateInput AssetPrivateDetails
	err = json.Unmarshal(transientUpdateJSON, &assetUpdateInput)

	// Read asset from the private data collection
	asset, err := s.ReadAssetPrivateDetails(ctx, "Org1AndOrg2PrivateCollection", assetUpdateInput.ID)
	if err != nil {
		return fmt.Errorf("error reading asset: %v", err)
	}
	if asset == nil {
		return fmt.Errorf("%v does not exist", assetUpdateInput.ID)
	}

	asset.Telefono = assetUpdateInput.Telefono
	asset.Email = assetUpdateInput.Email
	asset.Nombre = assetUpdateInput.Nombre
	asset.ApellidoPaterno = assetUpdateInput.ApellidoPaterno
	asset.ApellidoMaterno = assetUpdateInput.ApellidoMaterno
	asset.CURP = assetUpdateInput.CURP
	asset.INECLV = assetUpdateInput.INECLV
	asset.AnioDeRegistro = assetUpdateInput.AnioDeRegistro
	asset.AnioDeEmision = assetUpdateInput.AnioDeEmision
	asset.Vigencia = assetUpdateInput.Vigencia
	asset.Calle = assetUpdateInput.Calle
	asset.Numero = assetUpdateInput.Numero
	asset.Colonia = assetUpdateInput.Colonia
	asset.Localidad = assetUpdateInput.Localidad
	asset.Seccion = assetUpdateInput.Seccion
	asset.Municipio = assetUpdateInput.Municipio
	asset.Estado = assetUpdateInput.Estado
	asset.CodigoPostal = assetUpdateInput.CodigoPostal
	asset.OCR = assetUpdateInput.OCR
	asset.IDCiudadano = assetUpdateInput.IDCiudadano
	asset.FechaDeNacimiento = assetUpdateInput.FechaDeNacimiento
	asset.Nacionalidad = assetUpdateInput.Nacionalidad
	asset.PaisDeResidencia = assetUpdateInput.PaisDeResidencia
	asset.TipoDeActividad = assetUpdateInput.TipoDeActividad
	asset.Domicilio = assetUpdateInput.Domicilio

	assetJSONasBytes, err := json.Marshal(asset)

	log.Printf("Actualizando Asset: %s", assetUpdateInput.ID)

	// Put asset appraised value into owners org specific private data collection
	log.Printf("Put: collection %v, ID %v", "Org1AndOrg2PrivateCollection", assetUpdateInput.ID)
	err = ctx.GetStub().PutPrivateData("Org1AndOrg2PrivateCollection", assetUpdateInput.ID, assetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}
	return nil
}

// UpdateAssetPrivateCollectionOrg1 updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAssetPrivateCollectionOrg1(ctx contractapi.TransactionContextInterface) error {

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field
	transientUpdateJSON, ok := transientMap["asset_properties"]
	if !ok {
		return fmt.Errorf("asset owner not found in the transient map")
	}

	var assetUpdateInput AssetPrivateDetails
	err = json.Unmarshal(transientUpdateJSON, &assetUpdateInput)

	// Read asset from the private data collection
	asset, err := s.ReadAssetPrivateDetails(ctx, "Org1PrivateCollection", assetUpdateInput.ID)
	if err != nil {
		return fmt.Errorf("error reading asset: %v", err)
	}
	if asset == nil {
		return fmt.Errorf("%v does not exist", assetUpdateInput.ID)
	}

	log.Printf("Actualizando Asset: %s", assetUpdateInput.ID)

	asset.Telefono = assetUpdateInput.Telefono
	asset.Email = assetUpdateInput.Email
	asset.Nombre = assetUpdateInput.Nombre
	asset.ApellidoPaterno = assetUpdateInput.ApellidoPaterno
	asset.ApellidoMaterno = assetUpdateInput.ApellidoMaterno
	asset.CURP = assetUpdateInput.CURP
	asset.INECLV = assetUpdateInput.INECLV
	asset.AnioDeRegistro = assetUpdateInput.AnioDeRegistro
	asset.AnioDeEmision = assetUpdateInput.AnioDeEmision
	asset.Vigencia = assetUpdateInput.Vigencia
	asset.Calle = assetUpdateInput.Calle
	asset.Numero = assetUpdateInput.Numero
	asset.Colonia = assetUpdateInput.Colonia
	asset.Localidad = assetUpdateInput.Localidad
	asset.Seccion = assetUpdateInput.Seccion
	asset.Municipio = assetUpdateInput.Municipio
	asset.Estado = assetUpdateInput.Estado
	asset.CodigoPostal = assetUpdateInput.CodigoPostal
	asset.OCR = assetUpdateInput.OCR
	asset.IDCiudadano = assetUpdateInput.IDCiudadano
	asset.FechaDeNacimiento = assetUpdateInput.FechaDeNacimiento
	asset.Nacionalidad = assetUpdateInput.Nacionalidad
	asset.PaisDeResidencia = assetUpdateInput.PaisDeResidencia
	asset.TipoDeActividad = assetUpdateInput.TipoDeActividad
	asset.Domicilio = assetUpdateInput.Domicilio

	assetJSONasBytes, err := json.Marshal(asset)

	// Put asset appraised value into owners org specific private data collection
	log.Printf("Put: collection %v, ID %v", "Org1PrivateCollection", assetUpdateInput.ID)
	err = ctx.GetStub().PutPrivateData("Org1PrivateCollection", assetUpdateInput.ID, assetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}
	return nil
}
