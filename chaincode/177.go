package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Cliente struct {
	Id             string
	NombreCompleto string
	Saldo          float64
	ValorPromedio  float64
	Moneda         string
}

// Lista de clientes
var clientesBanco1 = []Cliente{
	{
		Id:             "1020498574",
		NombreCompleto: "Diego Salazar Rojas",
		Saldo:          1000.0,
		ValorPromedio:  200.0,
		Moneda:         "EUR",
	},
	{
		Id:             "1030599585",
		NombreCompleto: "Ana García López",
		Saldo:          500.0,
		ValorPromedio:  150.0,
		Moneda:         "USD",
	},
	{
		Id:             "1040600596",
		NombreCompleto: "Juan Pérez Martínez",
		Saldo:          2000000.0,
		ValorPromedio:  300000.0,
		Moneda:         "COP",
	},
	{
		Id:             "1050701607",
		NombreCompleto: "María González Pérez",
		Saldo:          1500.0,
		ValorPromedio:  250.0,
		Moneda:         "EUR",
	},
	{
		Id:             "1060802618",
		NombreCompleto: "Pedro Rodríguez García",
		Saldo:          1200.0,
		ValorPromedio:  200.0,
		Moneda:         "USD",
	},
}

var clientesBanco2 = []Cliente{
	{
		Id:             "2020498574",
		NombreCompleto: "Camila Arango Uribe",
		Saldo:          12000000.0,
		ValorPromedio:  2500000.0,
		Moneda:         "COP",
	},
	{
		Id:             "2030599585",
		NombreCompleto: "Santiago López Restrepo",
		Saldo:          800.0,
		ValorPromedio:  200.0,
		Moneda:         "USD",
	},
	{
		Id:             "2040600596",
		NombreCompleto: "Mariana Gómez Vélez",
		Saldo:          1500.0,
		ValorPromedio:  300.0,
		Moneda:         "EUR",
	},
	{
		Id:             "2050701607",
		NombreCompleto: "Andrés Martínez Pérez",
		Saldo:          10000000.0,
		ValorPromedio:  220000.0,
		Moneda:         "COP",
	},
	{
		Id:             "2060802618",
		NombreCompleto: "Laura García Díaz",
		Saldo:          1250.0,
		ValorPromedio:  250.0,
		Moneda:         "USD",
	},
}

// Estructura de la transacción
type Transaccion struct {
	IdCliente     string    `json:"idCliente"`
	Monto         float64   `json:"monto"`
	Destino       string    `json:"destino"`
	MonedaDestino string    `json:"monedaDestino"`
	IdTransaccion string    `json:"idTransaccion"`
	Hash          string    `json:"hash"`
	FirstTime     string    `json:"firstTime"`
	Timestamp     time.Time `json:"timestamp"`
}

type EntidadSancionada struct {
	Id     string
	Nombre string
}

// Lista de entidades sancionadas
var entidadesSancionadas = []EntidadSancionada{
	{
		Id:     "1",
		Nombre: "entidadSancionada1",
	},
	{
		Id:     "3",
		Nombre: "entidadSancionada3",
	},
	{
		Id:     "5",
		Nombre: "entidadSancionada5",
	},
}

func (s *SmartContract) PoblarBD(ctx contractapi.TransactionContextInterface) error {
	var mensajeError string

	// Agregar los clientes del banco 1 al ledger
	for _, cliente := range clientesBanco1 {
		clienteAsBytes, _ := json.Marshal(cliente)
		err := ctx.GetStub().PutState(cliente.Id, clienteAsBytes)
		if err != nil {
			mensajeError = fmt.Sprintf("error al crear cliente: %s", err.Error())
			return fmt.Errorf(mensajeError)
		}
	}

	// Agregar los clientes del banco 2 al ledger
	for _, cliente := range clientesBanco2 {
		clienteAsBytes, _ := json.Marshal(cliente)
		err := ctx.GetStub().PutState(cliente.Id, clienteAsBytes)
		if err != nil {
			mensajeError = fmt.Sprintf("error al crear cliente: %s", err.Error())
			return fmt.Errorf(mensajeError)
		}
	}

	// Agregar las entidades sancionadas
	for _, nombre := range entidadesSancionadas {
		nombreString := nombre.Nombre
		idString := nombre.Id

		// Crear una nueva entidad sancionada
		entidadSancionada := EntidadSancionada{
			Id:     idString,
			Nombre: nombreString,
		}

		// Convertir la entidad a JSON
		entidadAsBytes, _ := json.Marshal(entidadSancionada)

		// Almacenar la entidad en el ledger
		err := ctx.GetStub().PutState(entidadSancionada.Id, entidadAsBytes)
		if err != nil {
			mensajeError = fmt.Sprintf("error al agregar entidad sancionada: %s", err.Error())
			return fmt.Errorf(mensajeError)
		}
	}

	if mensajeError == "" {
		fmt.Println("Aprovisionamiento completado con éxito")
		return nil
	} else {
		fmt.Println("Error en aprovisionamiento:", mensajeError)
		return fmt.Errorf(mensajeError)
	}
}

func (s *SmartContract) CrearTransaccion(ctx contractapi.TransactionContextInterface, idCliente string, monto float64, monedaDestino string, destino string, idTransaccion string) error {
	// Validaciones

	// Validar si el ID de la transacción ya existe
	transactionAsBytesQuery, err := ctx.GetStub().GetState(idTransaccion)
	if err != nil {
		return fmt.Errorf("failed to read from world state. %s", err.Error())
	}
	if transactionAsBytesQuery != nil {
		return fmt.Errorf("%s already exist", idTransaccion)
	}

	// Obtener el cliente 1 del ledger
	cliente, err := BuscarClientePorIDBanco1(ctx, idCliente)
	if err != nil {
		return fmt.Errorf("error al obtener el cliente del banco 1: %s", err)

	}
	var currentClient Cliente = cliente

	// Obtener el saldo del cliente 1 del ledger
	saldoCliente1, err := GetSaldo(ctx, idCliente)
	if err != nil {
		return fmt.Errorf("error al obtener el saldo del cliente: %s", err.Error())
	}
	// Validar fondos
	if monto > saldoCliente1 {
		return fmt.Errorf("fondos insuficientes")
	}

	// Obtener la moneda del cliente
	monedaOrigen := currentClient.Moneda

	// Validar entidades sancionadas
	if EstaSancionado(ctx, destino) {
		return fmt.Errorf("entidad '%s' sancionada", destino)
	}

	// Validar transacción sospechosa
	if monto > currentClient.ValorPromedio*1.5 {
		return fmt.Errorf("transacción sospechosa")
	}

	// Convertir moneda (si es necesario)
	if monedaOrigen != monedaDestino {
		monto = ConvertirMoneda(monto, monedaOrigen, monedaDestino)
	}

	// **Obtener el equivalente del monto en la moneda local del cliente 1**
	montoLocalCliente1 := ConvertirMoneda(monto, monedaDestino, monedaOrigen)

	// Restar el monto en la moneda local del cliente 1
	cliente.Saldo -= montoLocalCliente1
	cliente1AsBytes, _ := json.Marshal(cliente)
	err = ctx.GetStub().PutState(cliente.Id, cliente1AsBytes)
	if err != nil {
		return err
	}

	// Obtener el cliente 2 del ledger
	cliente2, err := BuscarClientePorIDBanco2(ctx, destino)
	if err != nil {
		return fmt.Errorf("error al obtener el cliente 2: %s", err)
	}

	// Depositar monto al cliente 2
	cliente2.Saldo += monto
	cliente2AsBytes, _ := json.Marshal(cliente2)
	err = ctx.GetStub().PutState(cliente2.Id, cliente2AsBytes)
	if err != nil {
		return err
	}

	// Crear transacción
	transaccion := Transaccion{
		IdCliente:     idCliente,
		Monto:         monto,
		MonedaDestino: monedaDestino,
		Destino:       destino,
		IdTransaccion: idTransaccion,
		Timestamp:     time.Now(),
	}

	// Calcular el hash de la transacción
	hash := CalcularHash(transaccion)
	transaccion.Hash = hash

	transactionAsBytes, _ := json.Marshal(transaccion)

	//save in ledger
	return ctx.GetStub().PutState(idTransaccion, transactionAsBytes)
}

func (s *SmartContract) ConsultarTransaccion(ctx contractapi.TransactionContextInterface, idTransaccion string) (*Transaccion, error) {
	transactionAsBytes, err := ctx.GetStub().GetState(idTransaccion)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	// Validar que el hash de la transacción recibida exista en el ledger
	if transactionAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist in ledger", idTransaccion)
	}

	// Desempaquetar la transacción
	transaccion := new(Transaccion)
	err = json.Unmarshal(transactionAsBytes, transaccion)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error. %s", err.Error())
	}

	return transaccion, nil
}

func (s *SmartContract) MostrarClientes(ctx contractapi.TransactionContextInterface) error {
	// Validar si el banco es el banco 1
	stub := ctx.GetStub()
	creator, err := stub.GetCreator()
	if err != nil {
		return fmt.Errorf("error al obtener el creador: %s", err.Error())
	}

	// Obtener todos los clientes del ledger
	clientesIterator, err := stub.GetStateByPartialCompositeKey("cliente", []string{})
	if err != nil {
		return fmt.Errorf("error al obtener clientes del ledger: %s", err.Error())
	}

	// Desempaquetar e imprimir los clientes
	defer clientesIterator.Close()
	for clientesIterator.HasNext() {
		response, err := clientesIterator.Next()
		if err != nil {
			return fmt.Errorf("error al obtener siguiente cliente: %s", err.Error())
		}

		cliente := Cliente{}
		err = json.Unmarshal(response.Value, &cliente)
		if err != nil {
			return fmt.Errorf("error al desempaquetar cliente: %s", err.Error())
		}

		fmt.Printf("Cliente: %+v\n", cliente)
		fmt.Printf("creator es : %+v\n", string(creator))
	}

	return nil
}

func (s *SmartContract) MostrarEntidadesSancionadas(ctx contractapi.TransactionContextInterface) error {
	// Obtener todas las entidades sancionadas del ledger
	entidadesIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("entidadSancionada", []string{})
	if err != nil {
		return fmt.Errorf("error al obtener entidades sancionadas del ledger: %s", err.Error())
	}

	// Desempaquetar e imprimir las entidades
	defer entidadesIterator.Close()
	for entidadesIterator.HasNext() {
		response, err := entidadesIterator.Next()
		if err != nil {
			return fmt.Errorf("error al obtener siguiente entidad sancionada: %s", err.Error())
		}

		entidadAsBytes := response.Value

		// Desempaquetar la entidad
		entidad := EntidadSancionada{}
		err = json.Unmarshal(entidadAsBytes, &entidad)
		if err != nil {
			return fmt.Errorf("error al desempaquetar entidad sancionada: %s", err.Error())
		}

		// Imprimir la entidad sancionada
		fmt.Println("Entidad sancionada:", entidad)
	}

	return nil
}

func GetEntidadSancionadaByName(ctx contractapi.TransactionContextInterface, nombre string) (EntidadSancionada, bool) {
	for _, entidad := range entidadesSancionadas {
		if strings.EqualFold(entidad.Nombre, nombre) {
			return entidad, true
		}
	}

	return EntidadSancionada{}, false
}

func CalcularHash(transaccion Transaccion) string {
	hashBytes, _ := json.Marshal(transaccion)
	return fmt.Sprintf("%x", sha256.Sum256(hashBytes))
}

func BuscarClientePorIDBanco1(ctx contractapi.TransactionContextInterface, id string) (Cliente, error) {

	// Obtener el valor del estado del cliente
	clienteAsBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return Cliente{}, fmt.Errorf("error al obtener el cliente del banco 1: %s", err.Error())
	}

	// Si el cliente no existe, retornar un error
	if clienteAsBytes == nil {
		return Cliente{}, fmt.Errorf("Cliente del banco 1 no encontrado")
	}

	// Desempaquetar el valor del estado en una variable de tipo `Cliente`
	cliente := Cliente{}
	err = json.Unmarshal(clienteAsBytes, &cliente)
	if err != nil {
		return Cliente{}, fmt.Errorf("error al desempaquetar el cliente del banco 1: %s", err.Error())
	}

	// Retornar el cliente
	return cliente, nil

}

func BuscarClientePorIDBanco2(ctx contractapi.TransactionContextInterface, id string) (Cliente, error) {
	// Obtener el valor del estado del cliente
	clienteAsBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return Cliente{}, fmt.Errorf("error al obtener el cliente del banco 1: %s", err.Error())
	}

	// Si el cliente no existe, retornar un error
	if clienteAsBytes == nil {
		return Cliente{}, fmt.Errorf("Cliente del banco 1 no encontrado")
	}

	// Desempaquetar el valor del estado en una variable de tipo `Cliente`
	cliente := Cliente{}
	err = json.Unmarshal(clienteAsBytes, &cliente)
	if err != nil {
		return Cliente{}, fmt.Errorf("error al desempaquetar el cliente del banco 1: %s", err.Error())
	}

	// Retornar el cliente
	return cliente, nil
}

// Función para convertir moneda
func ConvertirMoneda(monto float64, monedaOrigen string, monedaDestino string) float64 {
	var tasaCambio float64

	switch {
	case monedaOrigen == "COP" && monedaDestino == "USD":
		tasaCambio = 1.0 / 4000.0
	case monedaOrigen == "USD" && monedaDestino == "COP":
		tasaCambio = 4000.0
	case monedaOrigen == "USD" && monedaDestino == "EUR":
		tasaCambio = 0.93
	case monedaOrigen == "EUR" && monedaDestino == "USD":
		tasaCambio = 1.0 / 0.93
	case monedaOrigen == "COP" && monedaDestino == "EUR":
		tasaCambio = 1.0 / 4250.0
	case monedaOrigen == "EUR" && monedaDestino == "COP":
		tasaCambio = 4250.0
	default:
		fmt.Println("Moneda no compatible:", monedaOrigen, monedaDestino)
		return monto
	}

	return monto * tasaCambio
}

// Obtener el tipo de moneda del cliente
func GetMonedaCliente(idCliente string) string {
	for _, cliente := range clientesBanco1 {
		if cliente.Id == idCliente {
			return cliente.Moneda
		}
	}
	return "" // Moneda por defecto si no se encuentra el cliente
}

// Función para verificar si una entidad está sancionada
func EstaSancionado(ctx contractapi.TransactionContextInterface, nombre string) bool {
	for _, entidad := range entidadesSancionadas {
		if strings.EqualFold(entidad.Nombre, nombre) {
			return true
		}
	}

	_, ok := GetEntidadSancionadaByName(ctx, nombre)
	return ok
}

// Función para obtener el saldo del cliente
func GetSaldo(ctx contractapi.TransactionContextInterface, idCliente string) (float64, error) {

	// Obtener el valor del estado del cliente
	clienteAsBytes, err := ctx.GetStub().GetState(idCliente)
	if err != nil {
		return 0.0, fmt.Errorf("error al obtener el saldo del cliente: %s", err.Error())
	}

	// Si el cliente no existe, retornar 0
	if clienteAsBytes == nil {
		return 0.0, nil
	}

	// Desempaquetar el valor del estado en una variable de tipo `Cliente`
	cliente := Cliente{}
	err = json.Unmarshal(clienteAsBytes, &cliente)
	if err != nil {
		return 0.0, fmt.Errorf("error al desempaquetar el cliente: %s", err.Error())
	}

	// Retornar el valor de la propiedad `Saldo`
	return cliente.Saldo, nil

}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("error create cross-border-contract chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("error create cross-border-contract chaincode: %s", err.Error())
	}
}
