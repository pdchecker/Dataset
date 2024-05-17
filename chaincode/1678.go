/////////////////////////////////////////////
//    THE BLOCKCHAIN PKI EXPERIMENT     ////
///////////////////////////////////////////
/*
	This is the fabpki, a chaincode that implements a Public Key Infrastructure (PKI)
	for measuring instruments. It runs in Hyperledger Fabric 1.4.
	He was created as part of the PKI Experiment. You can invoke its methods
	to store measuring instruments public keys in the ledger, and also to verify
	digital signatures that are supposed to come from these instruments.

	@author: Wilson S. Melo Jr.
	@date: Oct/2019
*/
package main

import (
	//the majority of the imports are trivial...
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	//these imports are for Hyperledger Fabric interface
	//"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	//sc "github.com/hyperledger/fabric/protos/peer"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

/* All the following functions are used to implement fabpki chaincode. This chaincode
basically works with 2 main features:
	1) A Register Authority RA (e.g., Inmetro) verifies a new measuring instrument (MI) and attests
	the correspondence between the MI's private key and public key. After doing this, the RA
	inserts the public key into the ledger, associating it with the respective instrument ID.

	2) Any client can ask for a digital signature ckeck. The client informs the MI ID, an
	information piece (usually a legally relevant register) and its supposed digital signature.
	The chaincode retrieves the MI public key and validates de digital signature.
*/

// SmartContract defines the chaincode base structure. All the methods are implemented to
// return a SmartContrac type.
type SmartContract struct {
}

// ECDSASignature represents the two mathematical components of an ECDSA signature once
// decomposed.
type ECDSASignature struct {
	R, S *big.Int
}

// Meter constitutes our key|value struct (digital asset) and implements a single
// record to manage the
// meter public key and measures. All blockchain transactions operates with this type.
// IMPORTANT: all the field names must start with upper case

type PrevisaoRio struct {
	Regiao       string `json:"regiao"`
	TempMinMax   string `json:"tempminmax"`
	CeuMadrugada string `json:"ceumadrugada"`
	CeuManha     string `json:"ceumanha"`
	ForecastTime string `json:"forecasttime"`
	InsertTime   string `json:"inserttime"`
}

type WeatherAPI struct {
	CityName    string `json:"cityname"`
	Situation   string `json:"situation"`
	Temperature string `json:"temperature"`
	//  Timestamp   string `json:"timestamp"`
	Date string `json:"date"`
	Hour string `json:"hour"`
}

type DadosEstacao struct {
	// ID da estação é a chave e não entra no struct
	HoraLeitura       string `json:"horaleitura"`
	TotalUltimaHora   string `json:"totalultimahora"`
	Situacao          string `json:"situacao"`
	DirecaoVentoGraus string `json:"direcaoventograus"`
	VelocidadeVento   string `json:"velocidadevento"`
	Temperatura       string `json:"temperatura"`
	Pressao           string `json:"pressao"`
	Umidade           string `json:"umidade"`
	TimestampEstacao  string `json:"timestampestacao"`
	TimestampCliente  string `json:"timestampcliente"`
}

// PublicKeyDecodePEM method decodes a PEM format public key. So the smart contract can lead
// with it, store in the blockchain, or even verify a signature.
// - pemEncodedPub - A PEM-format public key
func PublicKeyDecodePEM(pemEncodedPub string) ecdsa.PublicKey {
	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return *publicKey
}

// Init method is called when the fabpki is instantiated.
// Best practice is to have any Ledger initialization in separate function.
// Note that chaincode upgrade also calls this function to reset
// or to migrate data, so be careful to avoid a scenario where you
// inadvertently clobber your ledger's data!
func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// Invoke function is called on each transaction invoking the chaincode. It
// follows a structure of switching calls, so each valid feature need to
// have a proper entry-point.
func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	// extract the function name and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	//implements a switch for each acceptable function
	if fn == "registerWeatherFromWeb" {
		// register api weather info
		return s.registerWeatherFromWeb(stub, args)
	} else if fn == "getWeatherFromWeb" {
		// gets api weather info
		return s.getWeatherFromWeb(stub, args)
	} else if fn == "queryWebWeatherHistory" {
		// gets weather from past
		return s.queryWebWeatherHistory(stub, args)
	} else if fn == "insertWeatherForecastRJ" {
		// insert rio climate forecast
		return s.insertWeatherForecastRJ(stub, args)
	} else if fn == "getWeatherForecastRJ" {
		// get weathr forecast
		return s.getWeatherForecastRJ(stub, args)
	} else if fn == "insertStationData" {
		return s.insertStationData(stub, args)
	} else if fn == "getStationData" {
		return s.getStationData(stub, args)
	}

	//function fn not implemented, notify error
	return shim.Error("Chaincode does not support this function.")
}

func (s *SmartContract) registerWeatherFromWeb(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	//validate args vector lenght
	if len(args) != 3 {
		return shim.Error("It was expected the parameters: <\"city name\"> <situation> <temperature>")
	}

	//gets the parameters
	cityName := args[0]
	situation := args[1]
	temperature := args[2]

	// Receives the time of creation
	timestamp := time.Now() // 2009-11-10 23:00:00 +0000 UTC m=+0.000000001
	//	timestampString := timestamp.String()

	// PASSAR DATA E HORA PRO FRONT
	// USAR TEMPO UNIX

	// extract date
	year, month, day := timestamp.Date()
	dateString := fmt.Sprintf("%d-%02d-%02d", year, month, day)

	// extract hour
	hour, minute, second := timestamp.Clock()
	hourString := fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)

	//creates the meter record with the respective public key
	// var station = Station{PubKey: strpubkey, MyDate: creationDate}
	var weatherApi = WeatherAPI{CityName: cityName, Situation: situation, Temperature: temperature, Date: dateString, Hour: hourString}

	//encapsulates station data in a JSON structure
	weatherApiAsBytes, _ := json.Marshal(weatherApi)

	//loging...
	fmt.Println("Registering climate info from web...")

	//registers meter in the ledger
	stub.PutState(cityName, weatherApiAsBytes)

	//notify procedure success
	return shim.Success(nil)
}

func (s *SmartContract) getWeatherFromWeb(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	//validate args vector lenght
	if len(args) != 1 {
		return shim.Error("It was expected the parameters: <\"city name\">")
	}

	//gets the parameters
	cityName := args[0]
	fmt.Println(cityName)

	// retrieve the station data from the ledger
	weatherApiAsBytes, err := stub.GetState(cityName)
	if err != nil {
		fmt.Println(err)
		return shim.Error("Error retrieving station from the ledger")
	}

	// check if its null
	if weatherApiAsBytes == nil {
		return shim.Error("No info registered for this city")
	}

	//creates Station struct to manipulate returned bytes
	MyWeather := WeatherAPI{}

	//loging...
	fmt.Println("Retrieving station data: ", weatherApiAsBytes)

	//convert bytes into a station object
	json.Unmarshal(weatherApiAsBytes, &MyWeather)

	// log
	fmt.Println("Retrieving station data after unmarshall: ", MyWeather)

	cityName = string(MyWeather.CityName)
	situation := string(MyWeather.Situation)
	temperature := string(MyWeather.Temperature)
	date := string(MyWeather.Date)
	hour := string(MyWeather.Hour)

	var info = "Cityname: " + cityName +
		"\nSituation: " + situation +
		"\nTemperature: " + temperature +
		"\nTimestamp: " + date + " " + hour

	// returns all station info
	return shim.Success(
		[]byte(info),
	)
}

func (s *SmartContract) queryWebWeatherHistory(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("It was expected 2 parameters: <key> <year-month-day>")
	}

	historyIer, err := stub.GetHistoryForKey(args[0])
	wantedDate := args[1]

	//verifies if the history exists
	if err != nil {
		//fmt.Println(errMsg)
		return shim.Error("Fail on getting ledger history")
	}

	// flag de achou ou nao
	flag := 0
	errMsg := ""

	for historyIer.HasNext() {
		queryResponse, err := historyIer.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		/*
			- Pegar queryResponse.value da chave CityName (virá em Bytes. Separar em uma lista JSON)
			- Separar Timestamp (formatado ou separado?) e comparar com o inserido
			- Se coincidirem, retornar clima
			- Se não, ir para o próximo
			- Caso nenhum seja encontrado, retorne o aviso
		*/

		valorBytes := queryResponse.Value

		MyWeather := WeatherAPI{}

		json.Unmarshal(valorBytes, &MyWeather)
		fmt.Println("Retrieving station data after unmarshall: ", MyWeather)

		date := string(MyWeather.Date)

		if wantedDate == date {
			cityName := string(MyWeather.CityName) // talvez nem precise
			situation := string(MyWeather.Situation)
			temperature := string(MyWeather.Temperature)
			date := string(MyWeather.Date)
			hour := string(MyWeather.Hour)

			var info = "Cityname: " + cityName +
				"\nSituation: " + situation +
				"\nTemperature: " + temperature +
				"\nTimestamp: " + date + " " + hour

			flag++
			return shim.Success([]byte(info))
		}

	}
	historyIer.Close()

	//loging...
	// fmt.Printf("Consulting ledger history, found %d\n records", counter)

	if flag == 0 {
		// const shimErr = shim.Error("Não encontrado")
		errMsg = "Não encontrado"
	}

	return shim.Error(errMsg)
}

func (s *SmartContract) insertWeatherForecastRJ(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	//validate args vector lenght
	if len(args) != 6 {
		return shim.Error("It was expected the parameters: <Região> <Temperatura Min/Max> <Céu Madrugada> <Céu Manhã>  <Insert Timestamp> <Forecast Timestamp>")
	}

	regiao := args[0]
	tempMinMax := args[1]
	ceuMadrugada := args[2]
	ceuManha := args[3]
	insertTime := args[4]
	forecastTime := args[5]

	var previsaoRio = PrevisaoRio{Regiao: regiao, TempMinMax: tempMinMax, CeuMadrugada: ceuMadrugada, CeuManha: ceuManha, InsertTime: insertTime, ForecastTime: forecastTime}

	previsaoRioAsBytes, _ := json.Marshal(previsaoRio)

	//registers forecast in the ledger
	stub.PutState(regiao, previsaoRioAsBytes)

	var info = "Previsão registrada com sucesso!"

	// returns all station info
	return shim.Success(
		[]byte(info),
	)
}

func (s *SmartContract) getWeatherForecastRJ(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	//validate args vector lenght
	if len(args) != 1 {
		return shim.Error("It was expected the parameters: <\"Região\">")
	}

	//gets the parameters
	regiao := args[0]
	fmt.Println(regiao)

	// retrieve the data from the ledger
	previsaoRioAsBytes, err := stub.GetState(regiao)
	if err != nil {
		fmt.Println(err)
		return shim.Error("Error retrieving station from the ledger")
	}

	// check if its null
	if previsaoRioAsBytes == nil {
		return shim.Error("Nenhuma previsão registrada para esta região")
	}

	//creates struct to manipulate returned bytes
	MinhaPrevisao := PrevisaoRio{}

	//loging...
	fmt.Println("Retrieving station data: ", previsaoRioAsBytes)

	//convert bytes into a object
	json.Unmarshal(previsaoRioAsBytes, &MinhaPrevisao)

	// log
	fmt.Println("Retrieving station data after unmarshall: ", MinhaPrevisao)

	tempMinMax := string(MinhaPrevisao.TempMinMax)
	ceuMadrugada := string(MinhaPrevisao.CeuMadrugada)
	ceuManha := string(MinhaPrevisao.CeuManha)
	insertTime := string(MinhaPrevisao.InsertTime)
	forecastTime := string(MinhaPrevisao.ForecastTime)

	var info = "Regiao: " + regiao +
		"\nTemperatura Mínima e Máxima: " + tempMinMax +
		"\nCéu Madrugada: " + ceuMadrugada +
		"\nCéu Manhã: " + ceuManha +
		"\nInserido em (tempo unix): " + insertTime +
		"\nPrevisão atualizada em (tempo unix): " + forecastTime

	// returns all station info
	return shim.Success(
		[]byte(info),
	)

}

func (s *SmartContract) insertStationData(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	//validate args vector lenght
	if len(args) != 11 {
		return shim.Error("Os parâmetros esperados são: ...")
	}

	idEstacao := args[0]
	horaLeitura := args[1]
	totalUltimaHora := args[2]
	situacao := args[3]
	direcaoVentoGraus := args[4]
	velocidadeVento := args[5]
	temperatura := args[6]
	pressao := args[7]
	umidade := args[8]
	timestampEstacao := args[9]
	timestampCliente := args[10]

	var dadosEstacao = DadosEstacao{
		HoraLeitura:       horaLeitura,
		TotalUltimaHora:   totalUltimaHora,
		Situacao:          situacao,
		DirecaoVentoGraus: direcaoVentoGraus,
		VelocidadeVento:   velocidadeVento,
		Temperatura:       temperatura,
		Pressao:           pressao,
		Umidade:           umidade,
		TimestampEstacao:  timestampEstacao,
		TimestampCliente:  timestampCliente,
	}

	dadosEstacaoAsBytes, _ := json.Marshal(dadosEstacao)

	//registra dados no ledger com o id da estação sendo a chave
	stub.PutState(idEstacao, dadosEstacaoAsBytes)

	var info = "Dados da estação registrados com sucesso!"

	// returns all info
	return shim.Success(
		[]byte(info),
	)
}

func (s *SmartContract) getStationData(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	//validate args vector lenght
	if len(args) != 1 {
		return shim.Error("It was expected the parameters: <\"ID da estação\">")
	}

	//gets the parameters
	idEstacao := args[0]
	fmt.Println(idEstacao)

	// retrieve the station data from the ledger
	dadosEstacaoAsBytes, err := stub.GetState(idEstacao)
	if err != nil {
		fmt.Println(err)
		return shim.Error("Error retrieving station from the ledger")
	}

	// check if its null
	if dadosEstacaoAsBytes == nil {
		return shim.Error("Nenhuma dado registrado para esta estação")
	}

	//creates Station struct to manipulate returned bytes
	MinhaEstacao := DadosEstacao{}

	//loging...
	fmt.Println("Retrieving data: ", dadosEstacaoAsBytes)

	//convert bytes into a station object
	json.Unmarshal(dadosEstacaoAsBytes, &MinhaEstacao)

	// log
	fmt.Println("Retrieving station data after unmarshall: ", MinhaEstacao)

	horaLeitura 	  := string(MinhaEstacao.HoraLeitura)
	totalUltimaHora   := string(MinhaEstacao.TotalUltimaHora)
	situacao 		  := string(MinhaEstacao.Situacao)
	direcaoVentoGraus := string(MinhaEstacao.DirecaoVentoGraus)
	velocidadeVento   := string(MinhaEstacao.VelocidadeVento)
	temperatura 	  := string(MinhaEstacao.Temperatura)
	pressao 		  := string(MinhaEstacao.Pressao)
	umidade 		  := string(MinhaEstacao.Umidade)
	timestampEstacao  := string(MinhaEstacao.TimestampEstacao)
	timestampCliente  := string(MinhaEstacao.TimestampCliente)

	var info = "\nInformações (tabelas) disponiveis: " + situacao +
		"\nHorario de leitura : " + horaLeitura +
		"\nPrecipitação na última hora: " + totalUltimaHora +
		"\nDireção do vento (graus): " + direcaoVentoGraus +
		"\nVelocidade do vento: " + velocidadeVento +
		"\nTemperatura: " + temperatura +
		"\nPressão: " + pressao +
		"\nUmidade: " + umidade +
		"\nTimestamp da Estação (UNIX): " + timestampEstacao +
		"\nTimestamp do Cliente (UNIX): " + timestampCliente

	// returns all station info
	return shim.Success(
		[]byte(info),
	)

}

/*
func (s *SmartContract) testCompositeKey(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	// https://github.com/hyperledger/fabric-samples/blob/c04253d55407e5fe7217d4931738fe7273b4a8a5/token-erc-721/chaincode-go/chaincode/erc721-contract_test.go#L24
	// https://github.com/hyperledger/fabric-chaincode-go/blob/b84622ba6a7a9e543f3ca1994850c41423bc29a2/shim/stub.go

	//validate args vector lenght
	if len(args) != 6 {
		return shim.Error("It was expected the parameters: <Região> <Temperatura Min/Max> <Céu Madrugada> <Céu Manhã>  <Insert Timestamp> <Forecast Timestamp>")
	}



}

*/

/*
 * The main function starts up the chaincode in the container during instantiate
 */
func main() {

	////////////////////////////////////////////////////////
	// USE THIS BLOCK TO COMPILE THE CHAINCODE
	if err := shim.Start(new(SmartContract)); err != nil {
		fmt.Printf("Error starting SmartContract chaincode: %s\n", err)
	}
	////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////
	// USE THIS BLOCK TO PERFORM ANY TEST WITH THE CHAINCODE

	// //create pair of keys
	// privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	// if err != nil {
	// 	panic(err)
	// }

	// //marshal the keys in a buffer
	// e, err := json.Marshal(privateKey)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// _ = ioutil.WriteFile("ecdsa-keys.json", e, 0644)

	// //read the saved key
	// file, _ := ioutil.ReadFile("ecdsa-keys.json")

	// myPrivKey := ecdsa.PrivateKey{}
	// //myPubKey := ecdsa.PublicKey{}

	// _ = json.Unmarshal([]byte(file), &myPrivKey)

	// fmt.Println("Essa é minha chave privada:")
	// fmt.Println(myPrivKey)

	// myPubKey := myPrivKey.PublicKey

	// //test digital signature verifying
	// msg := "message"
	// hash := sha256.Sum256([]byte(msg))
	// fmt.Println("hash: ", hash)

	// r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("signature: (0x%x, 0x%x)\n", r, s)

	// myPubKey.Curve = elliptic.P256()

	// fmt.Println("Essa é minha chave publica:")
	// fmt.Println(myPubKey)

	// valid := ecdsa.Verify(&myPubKey, hash[:], r, s)
	// fmt.Println("signature verified:", valid)

	// otherpk := "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE6NXETwtkAKGWBcIsI6/OYE0EwsVj\n3Fc4hHTaReNfq6Hz2UEzsJKCYN0stjPCXbpdUlYtETC1a3EcS3SUVYX6qA==\n-----END PUBLIC KEY-----\n"

	// newkey := PublicKeyDecodePEM(otherpk)
	// myPubKey.Curve = elliptic.P256()

	// //valid = ecdsa.Verify(newkey, hash[:], r, s)
	// //fmt.Println("signature verified:", valid)

	// mysign := "MEYCIQCY16jbdY222oEpFiSRwXPi1kS7c4wuwxYXeWJOoAjnVgIhAJQTM+itbm1mQyd40Ug0xr2/AvjZmFSdoc/iSSHA6nRI"

	// // first decode the signature to extract the DER-encoded byte string
	// der, err := base64.StdEncoding.DecodeString(mysign)
	// if err != nil {
	// 	panic(err)
	// }

	// // unmarshal the R and S components of the ASN.1-encoded signature into our
	// // signature data structure
	// sig := &ECDSASignature{}
	// _, err = asn1.Unmarshal(der, sig)
	// if err != nil {
	// 	panic(err)
	// }

	// valid = ecdsa.Verify(&newkey, hash[:], sig.R, sig.S)
	// fmt.Println("signature verified:", valid)

	// fmt.Println("Curve: ", newkey.Curve.Params())

	////////////////////////////////////////////////////////

}
