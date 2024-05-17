package main

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"

	"strconv"
)

//"github.com/hyperledger/fabric/core/chaincode/shim"
//pb "github.com/hyperledger/fabric/protos/peer"

type MergeSortRecHandler struct {
}

func main() {

	/*var file, errOpen = os.OpenFile("mergeSortRec-fabric.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errOpen != nil {
		log.Println(errOpen)
	}
	defer file.Close()

	log.SetOutput(file)*/

	fmt.Println("Started mergeSortRec")

	var errStart = shim.Start(new(MergeSortRecHandler))
	if errStart != nil {
		fmt.Printf("Error starting chaincode: %v \n", errStart)
	}

}

func (self *MergeSortRecHandler) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (self *MergeSortRecHandler) MergeSortRec(A []int, l int, r int) {

	if l < r {
		var q = (l + r) / 2
		self.MergeSortRec(A, l, q)
		self.MergeSortRec(A, q+1, r)
		merge(A, l, q, r)
	}

}

func merge(A []int, l int, q int, r int) {
	var B = make([]int, len(A))
	for i := l; i <= q; i++ {
		B[i] = A[i]
	}
	for j := q + 1; j <= r; j++ {
		B[r+q+1-j] = A[j]
	}
	var s = l
	var t = r
	for k := l; k <= r; k++ {
		if B[s] <= B[t] {
			A[k] = B[s]
			s++
		} else {
			A[k] = B[t]
			t--
		}
	}
}

func (self *MergeSortRecHandler) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var function, args = stub.GetFunctionAndParameters()

	if function == "Sort" {

		if len(args) < 4 {
			return shim.Error("Call to Sort must have at least 4 parameters")
		}

		var A = make([]int, len(args)-3)
		for i := range A {
			A[i], _ = strconv.Atoi(args[i])
		}

		var l, errL = strconv.Atoi(args[len(A)])
		if errL != nil {
			return shim.Error("Error while setting L in Sort")
		}
		var r, errR = strconv.Atoi(args[len(A)+1])
		if errR != nil {
			return shim.Error("Error while setting R in Sort")
		}

		var signature = args[len(A)+2]

		self.MergeSortRec(A, l, r)

		var errEvent = stub.SetEvent("sort/mergeSortRec", []byte(signature))
		if errEvent != nil {
			return shim.Error("Error setting event")
		}

		return shim.Success(nil)
	}

	return shim.Error("Not yet implemented function called")
}
