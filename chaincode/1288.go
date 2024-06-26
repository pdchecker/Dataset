/*
 *  Copyright 2017 - 2019 KB Kontrakt LLC - All Rights Reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */
package main

//go:generate mockgen -source=../vendor/github.com/hyperledger/fabric-chaincode-go/shim/interfaces.go -package=tests -destination=chaincode_mocks.go
//go:generate sed "10 i    \"github.com/hyperledger/fabric-chaincode-go/shim\"" -i chaincode_mocks.go
//go:generate sed -E -e "s/ ChaincodeStubInterface/ shim.ChaincodeStubInterface/g" -e "s/\\(StateQueryIteratorInterface/ (shim.StateQueryIteratorInterface/g" -e "s/\\(HistoryQueryIteratorInterface/(shim.HistoryQueryIteratorInterface/g" -i chaincode_mocks.go
//go:generate go fmt chaincode_mocks.go
