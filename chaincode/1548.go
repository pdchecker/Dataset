/*
 * Copyright (C) 2020 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */
package main

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/polynetwork/fabric-contract/ccm"
	"os"
	"strconv"
)

func main() {
	tls, _ := strconv.ParseBool(os.Getenv("TLS"))
	server := &shim.ChaincodeServer{
		CCID:    os.Getenv("CHAINCODE_CCID"),
		Address: os.Getenv("CHAINCODE_ADDRESS"),
		CC:      new(ccm.CrossChainManager),
		TLSProps: shim.TLSProperties{
			Disabled: tls,
		},
	}
	err := server.Start()
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}

	//err := shim.Start(new(ccm.CrossChainManager))
	//if err != nil {
	//	panic(err)
	//}
}
