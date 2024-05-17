/****************************************************************
 * Código de smart contract correspondiente al proyecto de grado
 * Álvaro Miguel Salinas Dockar
 * Universidad Católica Boliviana "San Pablo"
 * Ingeniería Mecatrónica
 * La Paz - Bolivia, 2020
 ***************************************************************/
package main

import (
	"encoding/json"

	"github.com/hyperledger/fabric-chaincode-go/shim"

	peer "github.com/hyperledger/fabric-protos-go/peer"
)

func openElection(stub shim.ChaincodeStubInterface) peer.Response {
	var err error
	configJSON := Config{}

	// Obtener el estado actual de la apertura de mesa
	config, err := stub.GetState("config")
	if err != nil {
		return errorResponse("Error en el proceso interno", 500)
	} else if config == nil {
		return errorResponse("Primero se debe inicializar el proceso", 400)
	}

	err = json.Unmarshal(config, &configJSON)
	if err != nil {
		return errorResponse("Error en el proceso de datos", 500)
	}

	// La elección ya está abierta?
	if configJSON.ElectionOpen == true {
		return errorResponse("La elección ya se encuentra abierta", 400)
	}

	// Abrir la mesa
	configJSON.ElectionOpen = true

	// Actualizar la apertura de la elección

	configByte, err := json.Marshal(configJSON)
	if err != nil {
		return errorResponse("Error en el proceso interno", 500)
	}
	err = stub.PutState("config", configByte)
	if err != nil {
		return errorResponse("Error en el interno al abrir la eleccón datos, revise la configuracion", 500)
	}

	return successResponse("Elección abierta correctamente")

}
