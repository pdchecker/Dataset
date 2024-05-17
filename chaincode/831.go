package main

// 
// Copyright ConvergenciaX Spa.
// Autor: Joel Cotrado Vargas <joel.cotrado@gmail.com>
// Date: 01-04-2023
// Implementación de smartcontract foodcontrol con control de PKI OU
//


import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)
type SmartContract struct {
	contractapi.Contract
}

// COodigos de error retornados por fallas con IOU states

var (
	errMissingOU = errors.New("La identidad no trae el OU requerido para ejeuttar la transacción")

)

// Food provee funciones basicas para el control de comida
type Food struct {
	Agricultor string `json:"agricultor"`
	Organizacion string `json:"organizacion"`
	Variedad string `json:"variedad"`
	
} 

// Se definie que esta funcion ahora es parte del smartcontract
//func (s *SmartContract) Set(ctx contractapi.TransactionContextInterface, foodId string, agricultor string,  variedad string)  error {
func (s *SmartContract) Set(ctx contractapi.TransactionContextInterface, foodId string, variedad string)  error {

	//Se obtiene el agricultor desde contexto con paquete cid y valida el OU del certificado del cliente
	//hasOU, err := cid.hasOUValue(ctx.GetStub(), "department2") 


	hasOUctx, err := cid.HasOUValue(ctx.GetStub(),"department2")

	if err != nil {
		return nil
	}
	
 
    fmt.Printf("hasOUctx : %s", hasOUctx)
 
	if err != nil{
		return err
	}

	if !hasOUctx {
		return errMissingOU
	}


	identidad := ctx.GetClientIdentity()
	
	agricultor, err := identidad.GetID()
	if err != nil {
		return err
	}

	org, err := identidad.GetID()
	if err != nil {
		return err
	}


	//Validaciones de sintaxis.

	//Validaciones de Negocios

	//food, err := s.Query(ctx,foodId)
	//if food != nil {
	//		fmt.Printf("Set-> Error - foodid: %s ya existe", err.Error())
	//		return err
	//}

	//Otras Validaciones
	//joel.cotrado++ convierte la estructura en bytes para almacenarla en la DLT
	food := Food{
		Agricultor:  agricultor,
		Organizacion: org,
		Variedad: variedad,
	}


	food2Bytes, err := json.Marshal(food)

	if err != nil {										
		fmt.Printf("Marshall error : %s", err.Error())
		//return nil, err
		return err
	}

  // con el stub se puede acceder al ledger distribuido
 return ctx.GetStub().PutState(foodId, food2Bytes)  //Crea el timestamp de la operación y 
}


func (s *SmartContract) Query(ctx contractapi.TransactionContextInterface, foodId string) (*Food, error){

   	food2Bytes, err :=	ctx.GetStub().GetState(foodId)  // sE asigna el estado a variable food2bytes
	if err != nil {
	 	return nil, fmt.Errorf("Falla en la lecutra del estado mundo %s", err.Error())
	}

	if food2Bytes == nil {
		return nil, fmt.Errorf("Identificador :[%s] no existe en la blockchain", foodId )
	}
	food := new(Food)

	err = json.Unmarshal(food2Bytes, food)
	if err != nil {
		return nil, fmt.Errorf ("Query-> Error Unmarshal. %s", err.Error())
    }

	return food, nil
}

func main(){
	
	// inicializa un chaincode
	chaincode, err := contractapi.NewChaincode(new(SmartContract) )

	if err != nil {
		fmt.Printf("Error creando foodcontrol chaincode %s", err.Error())
		return
	}

	if err := chaincode.Start();  err != nil {
		fmt.Printf("Error iniciando foodcontrol chaincode %s", err.Error())
	}
}