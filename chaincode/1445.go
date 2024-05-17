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

// candidateInspection sirve para obtener los datos del candidato y obtener el número de votos a favor
func candidateInspection(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// Comprobar el número de argumentos
	if len(args) != 1 {
		return errorResponse("Es necesario proveer la clave del candidato", 400)
	}
	println(args[0])
	keyID := args[0]
	bytes, err := stub.GetState(keyID)
	if err != nil {
		return errorResponse(err.Error(), 500)
	}

	response := string(bytes)

	fmt.Println(response)

	return successResponse(response)
}
