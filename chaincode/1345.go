package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/nova38/saacs/pkg/chaincode/common"
	"github.com/nova38/saacs/pkg/chaincode/contracts/roles"
)

func main() {

	fmt.Println("Starting BioChain")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	config := common.ServerConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	sm := roles.BuildContract()

	if name := os.Getenv("CHAINCODE_NAME"); name != "" {
		sm.Info.Contact.Name = name
	}

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      sm,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	fmt.Print("Starting Roles Auth Chaincode Server")
	fmt.Printf("config: %+v\n", config)
	fmt.Printf("server: %+v\n", server)

	if err := server.Start(); err != nil {
		slog.Error("Failed to start", err)
	}
}
