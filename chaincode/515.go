package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SimpleContract para manejar escritura y lectura desde el world-state
type SimpleContract struct {
	contractapi.Contract
}

// Crear agrega una nueva llave con valor al world-state
func (sc *SimpleContract) Crear(ctx contractapi.TransactionContextInterface, key string, value string) error {
	activoActual, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("No se puede interactuar con el world state")
	}

	if activoActual != nil {
		return fmt.Errorf("No se puede crear un par en el world state con la llave %s. Ya existe", key)
	}

	err = ctx.GetStub().PutState(key, []byte(value))

	if err != nil {
		return errors.New("No se puede interactuar con el world state")
	}

	return nil
}

// Leer devuelve el valor en clave en el world-state
func (sc *SimpleContract) Leer(ctx contractapi.TransactionContextInterface, key string) (string, error) {
	activoActual, err := ctx.GetStub().GetState(key)

	if err != nil {
		return "", errors.New("No se puede interactuar con el world state")
	}

	if activoActual == nil {
		return "", fmt.Errorf("No se puede leer un par en el world state con la llave %s. Ya existe", key)
	}

	return string(activoActual), nil
}

// Actualizar cambia el valor con llave en el world state
func (sc *SimpleContract) Actualizar(ctx contractapi.TransactionContextInterface, key string, value string) error {
	activoActual, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("No se puede interactuar con el world state")
	}

	if activoActual == nil {
		return fmt.Errorf("No se puede actualizar un par en el world state con la llave %s. Ya existe", key)
	}

	err = ctx.GetStub().PutState(key, []byte(value))

	if err != nil {
		return errors.New("No se puede interactuar con el world state")
	}

	return nil
}

func (sc *SimpleContract) MostrarInfoStub(ctx contractapi.TransactionContextInterface) error {
	log.Println("**********************")
	log.Println("")
	log.Println("[Channel ID] ", ctx.GetStub().GetChannelID())
	log.Println("")
	log.Println("**********************")
	log.Println("")
	log.Println("[Transaction ID] ", ctx.GetStub().GetTxID())
	log.Println("")
	log.Println("**********************")

	// obteniendo el ID del cliente que invoka la TX
	identityID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return errors.New("Obteniendo el ID del cliente")
	}

	// obteniendo el MSP-ID del cliente que invoka la TX
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return errors.New("Obteniendo el MSP-ID del cliente")
	}

	log.Println("[Client identity] ", identityID)
	log.Println("")
	log.Println("**********************")
	log.Println("")
	log.Println("[Client MSP-ID] ", mspID)
	log.Println("")
	log.Println("**********************")
	log.Println("")
	return nil
}
