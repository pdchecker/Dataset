

package main

import (
	"fmt"
	"encoding/json"
	"bytes"
	"strconv"
//	"time"

//	"github.com/go-ping/ping"
        "os/exec"
        "strings"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)


type NodesContract struct {
}


// Define Status codes for the response
const (
	OK    = 200
	ERROR = 500
)

//key=port, tipo regola = iptables , oggetto=nodo su cui fai le regole, value=opzionale 
type Node struct {
	Name		string	`json:"name"` //nome della regola, univoca altrimenti aggiorno il valore anziché crearlo
	Timestamp	string 	`json:"timestamp"`
	ip 		string 	`json:"event"`
}


// Init is called when the smart contract is instantiated
func (s *NodesContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	fmt.Println("Chaincode istanziato")
	return shim.Success(nil)
}



func (s *NodesContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "addnode" {			
		return s.addNode(APIstub, args)
	} else if function == "getnodes" {
		return s.getNodes(APIstub, args)
	} else if function == "initledger" {
		return s.initLedger(APIstub, args)
	} else if function == "pingandaddnode" {		
		return s.pingAndAddNode(APIstub, args)
	} 

	return shim.Error("Invalid Smart Contract function name on Nodes smart contract.")
}





func (s *NodesContract) initLedger (APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 0{ 
		return shim.Error("Incorrect number of arguments. Expecting 0")
	}


	var nodes [6]Node;	
	for i := 0; i< len(nodes); i++ {

		nodes[i].Name="E"+strconv.Itoa(i)
		fmt.Println("Nodes:",nodes[i])
	}


	i := 0
	for i < len(nodes) {
	//	fmt.Println("i is ", i)
		nodeAsBytes, _ := json.Marshal(nodes[i])
		APIstub.PutState(nodes[i].Name, nodeAsBytes)
		fmt.Println("Added", nodes[i])
		i = i + 1
	}
	return shim.Success(nil)
}


/*
func checkPing(ip string) <-chan int {
	packetsRecv := make(chan int)
	go func() {
		defer close(packetsRecv)

		pinger, err := ping.NewPinger(ip)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
		}

		pinger.OnRecv = func(pkt *ping.Packet) {
			fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v\n",
				pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.Ttl)
		}

		pinger.OnFinish = func(stats *ping.Statistics) {
			fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
			fmt.Printf("%d packets transmitted, %d packets received, %v%% packet loss\n",
				stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
			fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
				stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)

			packetsRecv <- stats.PacketsRecv

		}

		pinger.Count = 3
		pinger.Interval = time.Duration(500)*time.Millisecond
		pinger.Timeout = time.Duration(4)*time.Second
//		pinger.SetPrivileged(true)

		fmt.Printf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())
		pinger.Run()
	}()
	return packetsRecv
}
*/

func checkPing(ip string) <-chan int {
	packetsRecv := make(chan int)

	go func() {
		defer close(packetsRecv)
		for i := 0; i != 3; i++ {
			out, _ := exec.Command("ping", ip, "-c 1",  "-w 1").Output()
//			fmt.Println("out: ",string(out))

			if !strings.Contains(string(out), "0 received") {
			    packetsRecv <- 1
			    i=10
			} 
		}
	}()
	return packetsRecv
}





func (s *NodesContract) pingAndAddNode (APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3{ 
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	name := args[0]
	timestamp := args[1]
	ip := args[2]

	//check ping
	packetsRecv := <-checkPing(ip)
	if packetsRecv==0{
		return shim.Error( fmt.Sprintf("Could not ping a node with ip %s, 0 packets received", ip) )
	}

	//add node
	getState, err := APIstub.GetState(name)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error from getState into addNode: %s", err.Error()))
	}
	if bytes.Equal(getState,[]byte("")) {//then create new node
		node := Node{name, timestamp, ip}
		nodeAsBytes, marshalErr := json.Marshal(node)
		if marshalErr != nil {
			return shim.Error(fmt.Sprintf("Could not marshal new %s node: %s", name, marshalErr.Error()))
		}
		putErr := APIstub.PutState(name, nodeAsBytes)
		if putErr != nil {
			return shim.Error(fmt.Sprintf("Could not put new %s node in the ledger: %s", name, putErr.Error()))
		}


		//emit add node event
		eventPayload := "Node "+name+" with ip "+ip+" is added"
		payloadAsBytes := []byte(eventPayload)
		eventErr := APIstub.SetEvent("addNodeEvent",payloadAsBytes)
		if (eventErr != nil) {
		  return shim.Error(fmt.Sprintf("Failed to emit add node event"))
		}


		fmt.Println("Added new node: ", node)
		return shim.Success([]byte(fmt.Sprintf("Successfully added %s node",  name )))
	}

	return shim.Error("Error in addNode.")

}





func (s *NodesContract) addNode (APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3{ 
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	name := args[0]
	timestamp := args[1]
	ip := args[2]

	//add node
	getState, err := APIstub.GetState(name)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error from getState into addNode: %s", err.Error()))
	}
	if bytes.Equal(getState,[]byte("")) {//then create new node
		node := Node{name, timestamp, ip}
		nodeAsBytes, marshalErr := json.Marshal(node)
		if marshalErr != nil {
			return shim.Error(fmt.Sprintf("Could not marshal new %s node: %s", name, marshalErr.Error()))
		}
		putErr := APIstub.PutState(name, nodeAsBytes)
		if putErr != nil {
			return shim.Error(fmt.Sprintf("Could not put new %s node in the ledger: %s", name, putErr.Error()))
		}


		//emit add node event
		eventPayload := "Node "+name+" with ip "+ip+" is added"
		payloadAsBytes := []byte(eventPayload)
		eventErr := APIstub.SetEvent("addNodeEvent",payloadAsBytes)
		if (eventErr != nil) {
		  return shim.Error(fmt.Sprintf("Failed to emit add node event"))
		}


		fmt.Println("Added new node: ", node)
		return shim.Success([]byte(fmt.Sprintf("Successfully added %s node",  name )))
	}

	return shim.Error("Error in addNode.")

}







func (s *NodesContract) getNodes(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Check we have a valid number of args
	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments, expecting 0")
	}


	resultsIterator, err := APIstub.GetStateByRange("","")
	if err != nil {
		fmt.Println("Errore getStateByRange")
		return shim.Error(fmt.Sprintf("Errore getStateByRange -> %s",err.Error()))
	}
	defer resultsIterator.Close()

	var nodes []Node //[]byte //il risultato di tutte le macchine
	var node Node //byte //variabile machine temporanea per poi assegnarla all'array con append
 
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		json.Unmarshal(queryResponse.Value, &node)
		nodes = append(nodes, node)
	//	fmt.Println("Added into machines array:", machine.Name)
	}



	if len(nodes) == 0 {
		return shim.Error("Errore, nodes array is empty.")
	} else {
		fmt.Println("Len Nodes array: ",len(nodes))
	}


	nodesAsBytes, _ := json.Marshal(nodes)
	return shim.Success(nodesAsBytes)
}




















// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(NodesContract))
	if err != nil {
		fmt.Printf("Error creating new Nodes Smart Contract: %s", err)
	}
}



