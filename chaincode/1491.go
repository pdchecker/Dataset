package main

import (
	"encoding/json"
	"fmt"
	"crypto/sha256"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"math/big"
)

//SmartContract Define the Smart Contract structure
type SmartContract struct {
}

//File : Define the file structure, with 2 attributes.  Structure tags are used by encoding/json library
type File struct {
	Hash []byte `json:"hash"` 
	AccumulatedValues []*big.Int `json:"accumulatedValues"`
}

// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()

	switch function {
		case "queryFile":
			return s.queryFile(APIstub, args)
		case "createFile":
			return s.createFile(APIstub, args)
		case "giveAccess":
			return s.giveAccess(APIstub, args)
		case "updateFile":
			return s.updateFile(APIstub, args)
		default:
			return shim.Error("Invalid Smart Contract function name.")
	}
}

// queryFile :  Method for querying a file
func (s *SmartContract) queryFile(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	fileAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(fileAsBytes)
}

// createFile :  Method for creating a file
// args[0] : id
// args[1] : hash
// args[2] : accumulatedValues[0] (u in the accumulator)
// args[3] : accumulatedValues[1] (u^e in the accumulator where e is known only by the owner of the file)
func (s *SmartContract) createFile(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	
	value1 := new(big.Int)
	value1.SetString(args[1], 2)

	value2 := new(big.Int)
	value2.SetString(args[2], 2)

	var file = File{Hash: []byte(args[1]), AccumulatedValues: []*big.Int{value1, value2}}
	fileAsBytes, _ := json.Marshal(file)
	APIstub.PutState(args[0], fileAsBytes)

	indexName := "hash~id"
	hashNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{file.AccumulatedValues[0].String(), args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(hashNameIndexKey, value)

	return shim.Success(fileAsBytes)
}

// giveAccess :  Method for giving access to a file
// args[0] : id
// args[1:29] : values of the commitment to coin as given in the paper to verify if the user has the knowledge of (u,r) s.t. u^e = v
//args[30] : w^e (w is the value that was last added to the accumulator Thus AccumulatedValues[length -1] is the witness for the last value in the accumulator and e is known only by the owner and the user to whom the latest access was shared)
func (s *SmartContract) giveAccess(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 31 {
		return shim.Error("Incorrect number of arguments. Expecting 32")
	} 

	fileAsBytes, _ := APIstub.GetState(args[0])
	file := File{}

	json.Unmarshal(fileAsBytes, &file)
	
	//verify if the user is the owner of the file as only the user knowns the value of e s.t. AccumulatedValues[0]^e = AccumulatedValues[1]
	valueOfCommitmentToCoin := new(big.Int)
	valueOfCommitmentToCoin.SetString(args[1], 2)

	sg:= new(big.Int)
	sg.SetString(args[2], 2)

	sh:= new(big.Int)
	sh.SetString(args[3], 2)

	g_n:= new(big.Int)
	g_n.SetString(args[4], 2)

	h_n:= new(big.Int)
	h_n.SetString(args[5], 2)

	modulus:= new(big.Int)
	modulus.SetString(args[6], 2)

	C_e:= new(big.Int)
	C_e.SetString(args[7], 2)

	C_u:= new(big.Int)
	C_u.SetString(args[8], 2)

	C_r:= new(big.Int)
	C_r.SetString(args[9], 2)

	s_alpha:= new(big.Int)
	s_alpha.SetString(args[10], 2)

	s_beta:= new(big.Int)
	s_beta.SetString(args[11], 2)

	s_gamma:= new(big.Int)
	s_gamma.SetString(args[12], 2)

	s_delta:= new(big.Int)
	s_delta.SetString(args[13], 2)

	s_sigma:= new(big.Int)
	s_sigma.SetString(args[14], 2)

	s_zeta:= new(big.Int)
	s_zeta.SetString(args[15], 2)

	s_eta:= new(big.Int)
	s_eta.SetString(args[16], 2)

	s_epsilon:= new(big.Int)
	s_epsilon.SetString(args[17], 2)

	s_xi:= new(big.Int)
	s_xi.SetString(args[18], 2)

	s_phi:= new(big.Int)
	s_phi.SetString(args[19], 2)

	s_psi:= new(big.Int)
	s_psi.SetString(args[20], 2)

	st_1:= new(big.Int)
	st_1.SetString(args[21], 2)

	st_2:= new(big.Int)
	st_2.SetString(args[22], 2)

	st_3:= new(big.Int)
	st_3.SetString(args[23], 2)

	t_1:= new(big.Int)
	t_1.SetString(args[24], 2)

	t_2:= new(big.Int)
	t_2.SetString(args[25], 2)

	t_3:= new(big.Int)
	t_3.SetString(args[26], 2)

	t_4:= new(big.Int)
	t_4.SetString(args[27], 2)

	min_s_alpha:= new(big.Int)
	min_s_alpha.SetString(args[28], 2)

	max_s_alpha:= new(big.Int)
	max_s_alpha.SetString(args[29], 2)

	//If user is not the owner of the file
	if !verify(valueOfCommitmentToCoin, sg, sh, g_n, h_n, modulus, file.AccumulatedValues[1], C_e, C_u, C_r, s_alpha, s_beta, s_zeta, s_sigma, s_eta, s_epsilon, s_delta, s_xi, s_phi, s_gamma, s_psi, st_1, st_2, st_3, t_1, t_2, t_3, t_4, min_s_alpha, max_s_alpha){
		return shim.Error("User is not the owner of the file")
	}

	//if user is the owner of the file
	var value = new(big.Int)
	value.SetString(args[30], 2)
	file.AccumulatedValues = append(file.AccumulatedValues, value)
	fileAsBytes, _ = json.Marshal(file)
	APIstub.PutState(args[0], fileAsBytes)

	return shim.Success(fileAsBytes)
}

//updateFile : Allows the users to update the file
//// args[0] : id
// args[1:29] : values of the commitment to coin as given in the paper to verify if the user has the knowledge of (u,r) s.t. u^e = v
// args[30] : new hash of the file
func (s *SmartContract) updateFile(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	
	if len(args) != 31 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	} 

	fileAsBytes, _ := APIstub.GetState(args[0])
	file := File{}

	json.Unmarshal(fileAsBytes, &file)
	
	//verify if the user has access to the file
	valueOfCommitmentToCoin := new(big.Int)
	valueOfCommitmentToCoin.SetString(args[1], 2)

	sg:= new(big.Int)
	sg.SetString(args[2], 2)

	sh:= new(big.Int)
	sh.SetString(args[3], 2)

	g_n:= new(big.Int)
	g_n.SetString(args[4], 2)

	h_n:= new(big.Int)
	h_n.SetString(args[5], 2)

	modulus:= new(big.Int)
	modulus.SetString(args[6], 2)

	C_e:= new(big.Int)
	C_e.SetString(args[7], 2)

	C_u:= new(big.Int)
	C_u.SetString(args[8], 2)

	C_r:= new(big.Int)
	C_r.SetString(args[9], 2)

	s_alpha:= new(big.Int)
	s_alpha.SetString(args[10], 2)

	s_beta:= new(big.Int)
	s_beta.SetString(args[11], 2)

	s_gamma:= new(big.Int)
	s_gamma.SetString(args[12], 2)

	s_delta:= new(big.Int)
	s_delta.SetString(args[13], 2)

	s_sigma:= new(big.Int)
	s_sigma.SetString(args[14], 2)

	s_zeta:= new(big.Int)
	s_zeta.SetString(args[15], 2)

	s_eta:= new(big.Int)
	s_eta.SetString(args[16], 2)

	s_epsilon:= new(big.Int)
	s_epsilon.SetString(args[17], 2)

	s_xi:= new(big.Int)
	s_xi.SetString(args[18], 2)

	s_phi:= new(big.Int)
	s_phi.SetString(args[19], 2)

	s_psi:= new(big.Int)
	s_psi.SetString(args[20], 2)

	st_1:= new(big.Int)
	st_1.SetString(args[21], 2)

	st_2:= new(big.Int)
	st_2.SetString(args[22], 2)

	st_3:= new(big.Int)
	st_3.SetString(args[23], 2)

	t_1:= new(big.Int)
	t_1.SetString(args[24], 2)

	t_2:= new(big.Int)
	t_2.SetString(args[25], 2)

	t_3:= new(big.Int)
	t_3.SetString(args[26], 2)

	t_4:= new(big.Int)
	t_4.SetString(args[27], 2)

	min_s_alpha:= new(big.Int)
	min_s_alpha.SetString(args[28], 2)

	max_s_alpha:= new(big.Int)
	max_s_alpha.SetString(args[29], 2)

	//We iterate over all possible witnesses in this case the accumulated values to check if the user has access to the file
	has_access := false
	for i := 0; i < len(file.AccumulatedValues); i++ {
		has_access = has_access || verify(valueOfCommitmentToCoin, sg, sh, g_n, h_n, modulus, file.AccumulatedValues[i], C_e, C_u, C_r, s_alpha, s_beta, s_zeta, s_sigma, s_eta, s_epsilon, s_delta, s_xi, s_phi, s_gamma, s_psi, st_1, st_2, st_3, t_1, t_2, t_3, t_4, min_s_alpha, max_s_alpha)
	}

	if !has_access {
		return shim.Error("User does not have access to the file")
	}
	//if yes
	file.Hash = []byte(args[31])
	fileAsBytes, _ = json.Marshal(file)
	APIstub.PutState(args[0], fileAsBytes)

	return shim.Success(fileAsBytes)
}

func verify(valueOfCommitmentToCoin *big.Int, sg *big.Int, sh *big.Int, g_n *big.Int, h_n *big.Int, modulus *big.Int, accumulator *big.Int, C_e *big.Int, C_u *big.Int, C_r *big.Int, s_alpha *big.Int, s_beta *big.Int,s_zeta *big.Int, s_sigma *big.Int, s_eta *big.Int, s_epsilon *big.Int, s_delta *big.Int, s_xi *big.Int,s_phi *big.Int, s_gamma *big.Int, s_psi *big.Int,st_1 *big.Int,st_2 *big.Int, st_3 *big.Int, t_1 *big.Int,t_2 *big.Int,t_3 *big.Int, t_4 *big.Int ,min_s_alpha *big.Int, max_s_alpha *big.Int) bool {
	concatenator := append(valueOfCommitmentToCoin.Bytes(), sg.Bytes()...)
	concatenator = append(concatenator, sh.Bytes()...)
	concatenator = append(concatenator, g_n.Bytes()...)
	concatenator = append(concatenator, h_n.Bytes()...)
	concatenator = append(concatenator, modulus.Bytes()...)
	concatenator = append(concatenator, accumulator.Bytes()...)
	concatenator = append(concatenator, C_e.Bytes()...)
	concatenator = append(concatenator, C_u.Bytes()...)
	concatenator = append(concatenator, C_r.Bytes()...)
	concatenator = append(concatenator, s_alpha.Bytes()...)
	concatenator = append(concatenator, s_beta.Bytes()...)
	concatenator = append(concatenator, s_zeta.Bytes()...)
	concatenator = append(concatenator, s_sigma.Bytes()...)
	concatenator = append(concatenator, s_eta.Bytes()...)
	concatenator = append(concatenator, s_epsilon.Bytes()...)
	concatenator = append(concatenator, s_delta.Bytes()...)
	concatenator = append(concatenator, s_xi.Bytes()...)
	concatenator = append(concatenator, s_phi.Bytes()...)
	c_byte := sha256.Sum256(concatenator)
	c := new(big.Int)
	c.SetBytes(c_byte[:])

	st_1_prime := new(big.Int).Mul(new(big.Int).Exp(valueOfCommitmentToCoin,c,modulus),new(big.Int).Mul(new(big.Int).Exp(sg,s_alpha,modulus),new(big.Int).Exp(sh,s_phi,modulus)))
	st_1_prime.Mod(st_1_prime,modulus)

	st_2_prime := new(big.Int).Mul(new(big.Int).Exp(sg,c,modulus),new(big.Int).Mul(new(big.Int).Exp(new(big.Int).Mul(valueOfCommitmentToCoin,new(big.Int).ModInverse(sg,modulus)),s_gamma,modulus),new(big.Int).Exp(sh,s_psi,modulus)))
	st_2_prime.Mod(st_2_prime,modulus)

	st_3_prime := new(big.Int).Mul(new(big.Int).Exp(sg,c,modulus),new(big.Int).Mul(new(big.Int).Exp(new(big.Int).Mul(valueOfCommitmentToCoin,sg),s_sigma,modulus),new(big.Int).Exp(sh,s_xi,modulus)))
	st_3_prime.Mod(st_3_prime,modulus)

	t_1_prime := new(big.Int).Mul(new(big.Int).Exp(C_r,c,modulus),new(big.Int).Mul(new(big.Int).Exp(h_n,s_zeta,modulus),new(big.Int).Exp(g_n,s_epsilon,modulus)))
	t_1_prime.Mod(t_1_prime,modulus)

	t_2_prime := new(big.Int).Mul(new(big.Int).Exp(C_e,c,modulus),new(big.Int).Mul(new(big.Int).Exp(h_n,s_eta,modulus),new(big.Int).Exp(g_n,s_alpha,modulus)))
	t_2_prime.Mod(t_2_prime,modulus)

	t_3_prime := new(big.Int).Mul(new(big.Int).Exp(accumulator,c,modulus),new(big.Int).Mul(new(big.Int).Exp(C_u,s_alpha,modulus),new(big.Int).Exp(new(big.Int).ModInverse(h_n,modulus),s_beta,modulus)))
	t_3_prime.Mod(t_3_prime,modulus)

	t_4_prime := new(big.Int).Mul(new(big.Int).Exp(C_r,s_alpha,modulus),new(big.Int).Mul(new(big.Int).Exp(new(big.Int).ModInverse(h_n,modulus),s_delta,modulus),new(big.Int).Exp(new(big.Int).ModInverse(g_n,modulus),s_beta,modulus)))
	t_4_prime.Mod(t_4_prime,modulus)

	result_st_1 := st_1.Cmp(st_1_prime) ==0
	result_st_2 := st_2.Cmp(st_2_prime)==0
	result_st_3 := st_3.Cmp(st_3_prime)==0

	result_t_1 := t_1.Cmp(t_1_prime)==0
	result_t_2 := t_2.Cmp(t_2_prime)==0
	result_t_3 := t_3.Cmp(t_3_prime)==0
	result_t_4 := t_4.Cmp(t_4_prime)==0

	result_range:= (s_alpha.Cmp(min_s_alpha)>=0) && (s_alpha.Cmp(max_s_alpha)<=0)

	return result_st_1 && result_st_2 && result_st_3 && result_t_1 && result_t_2 && result_t_3 && result_t_4 && result_range

}
// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}