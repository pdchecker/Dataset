package main

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"

	"strconv"
)

//"github.com/hyperledger/fabric/core/chaincode/shim"
//pb "github.com/hyperledger/fabric/protos/peer"

type QuickSortRecHandler struct {
}

func main() {

	/*var file, errOpen = os.OpenFile("quickSortRec-fabric.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errOpen != nil {
		log.Println(errOpen)
	}
	defer file.Close()

	log.SetOutput(file)*/

	fmt.Println("Started quickSortRec")

	var errStart = shim.Start(new(QuickSortRecHandler))
	if errStart != nil {
		fmt.Printf("Error starting chaincode: %v \n", errStart)
	}

}

func (self *QuickSortRecHandler) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (self *QuickSortRecHandler) QuickSortRec(A []int, l int, r int) {

	if l < r {
		var q = partition(A, l, r)
		self.QuickSortRec(A, l, q)
		self.QuickSortRec(A, q+1, r)
	}

}

func partition(A []int, l int, r int) int {
	var x = A[((l + r) / 2)]
	var i = l - 1
	var j = r + 1
	return partition1(A, x, i, j)
}

func partition1(A []int, x int, i int, j int) int {
	j--
	if A[j] > x {
		return partition1(A, x, i, j)
	} else {
		return partition2(A, x, i, j)
	}
}

func partition2(A []int, x int, i int, j int) int {
	i++
	if A[i] < x {
		return partition2(A, x, i, j)
	} else {
		if i < j {
			exchange(A, i, j)
			return partition1(A, x, i, j)
		} else {
			return j
		}
	}
}

func exchange(A []int, q int, i int) {
	var tmp = A[q]
	A[q] = A[i]
	A[i] = tmp
}

func (self *QuickSortRecHandler) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
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

		self.QuickSortRec(A, l, r)

		var errEvent = stub.SetEvent("sort/quickSortRec", []byte(signature))
		if errEvent != nil {
			return shim.Error("Error setting event")
		}

		return shim.Success(nil)
	}

	return shim.Error("Not yet implemented function called")
}
