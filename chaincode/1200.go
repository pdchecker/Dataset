/****************************************************************
 * Código de smart contract correspondiente al proyecto de grado
 * Álvaro Miguel Salinas Dockar
 * Universidad Católica Boliviana "San Pablo"
 * Ingeniería Mecatrónica
 * La Paz - Bolivia, 2020
 ***************************************************************/
package main

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"

	peer "github.com/hyperledger/fabric-protos-go/peer"
)

// voterStatusInspection sirve solo para debugear que se ingrese la key correspondiente
func voterStatusInspection(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// Comprobar el número de argumentos
	if len(args) != 1 {
		return errorResponse("Es necesario proveer la clave", 400)
	}
	keyID := args[0]
	// Se espera que el key de los votantes sea de longitud 64 en forma de string
	if len(keyID) != 64 {
		return errorResponse("No se introdujo un identificador de votante", 400)
	}
	bytes, err := stub.GetState(keyID)
	if err != nil {
		return errorResponse(err.Error(), 500)
	}

	response := string(bytes)

	fmt.Println(response)

	return successResponse(response)
}
