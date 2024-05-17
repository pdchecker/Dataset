package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"jarvispowered.com/color_rpg/chain/chainTypes"
)

//go:generate counterfeiter -o mocks/transaction.go -fake-name TransactionContext . transactionContext
type transactionContext interface {
	contractapi.TransactionContextInterface
}

//go:generate counterfeiter -o mocks/chaincodestub.go -fake-name ChaincodeStub . chaincodeStub
type chaincodeStub interface {
	shim.ChaincodeStubInterface
}

//go:generate counterfeiter -o mocks/clientidentity.go -fake-name ClientIdentity . clientIdentity
type clientIdentity interface {
	cid.ClientIdentity
}

//go:generate counterfeiter -o mocks/statequeryiterator.go -fake-name StateQueryIterator . stateQueryIterator
type stateQueryIterator interface {
	shim.StateQueryIteratorInterface
}

func postLocal(endpoint string, payload string) (string, error) {
	var jsonData = []byte(payload)

	request, error := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("accept", "application/json")
	request.Header.Set("Request-Timeout", "120s")

	client := &http.Client{}
	response, error := client.Do(request)

	if error != nil {
		fmt.Println("response Status:", response.Status)
		fmt.Println("response Headers:", response.Header)

		return "", error
	}

	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)

	return string(body), nil
}

func testGetCharacter(t *testing.T, characterName string) chainTypes.Character {
	// Now request the character again to make sure it exists
	url := "http://127.0.0.1:5000/api/v1/namespaces/default/apis/crpg/query/GetCharacterByName"

	interpretedBody := fmt.Sprintf(`{
		"input": {
		  "name": "%s"
		}
	  }`, characterName)

	result, err := postLocal(url, interpretedBody)

	require.NoError(t, err)

	returnCharacter := &chainTypes.Character{}
	err = json.Unmarshal([]byte(result), returnCharacter)

	require.NoError(t, err)

	require.Equal(t, returnCharacter.Name, characterName)

	return *returnCharacter
}

func testGetProfile(t *testing.T, walletAddress string, name string) string {
	// Now request the character again to make sure it exists
	url := "http://127.0.0.1:5000/api/v1/namespaces/default/apis/crpg/query/GetPlayerProfile"

	rawBody := `{
	"input": {
	  "externalAddress": "%s"
	}
  }`

	interpretedBody := fmt.Sprintf(rawBody, walletAddress)
	result, err := postLocal(url, interpretedBody)

	require.Contains(t, result, name)
	require.NoError(t, err)

	return result
}

func testGetVault(t *testing.T, walletAddress string) string {
	// Now request the character again to make sure it exists
	url := "http://127.0.0.1:5000/api/v1/namespaces/default/apis/crpg/query/GetPlayerVault"

	rawBody := `{
	"input": {
	  "externalAddress": "%s"
	}
  }`

	interpretedBody := fmt.Sprintf(rawBody, walletAddress)
	result, err := postLocal(url, interpretedBody)

	require.NoError(t, err)

	return result
}

func testCreateProfile(t *testing.T, walletAddress string, name string) {
	url := "http://127.0.0.1:5000/api/v1/namespaces/default/apis/crpg/invoke/NewPlayerProfile?fly-sync=true"

	rawBody := `{
		"input": {
		"externalAddress": "%s",
		"name": "%s",
		"signature": "string"
		}
	}`

	interpretedBody := fmt.Sprintf(rawBody, walletAddress, name)
	result, err := postLocal(url, interpretedBody)

	_ = result
	// fmt.Println(result)
	fmt.Println("Sent request to make player")
	require.NoError(t, err)
}

func TestDefineInterface(t *testing.T) {
	url := "http://127.0.0.1:5108/api/contracts/interface"

	jsonData, err := os.ReadFile("../ffi.json")

	body := fmt.Sprintf(`{
		"format": "ffi",
		"name": "",
		"version": "",
		"schema": %s
	}`, string(jsonData))

	bodyBytes := []byte(body)

	require.NoError(t, err, "Failed to load FFI")

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)

	request.Header.Set("Accept", "*/*")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://127.0.0.1:5108")
	request.Header.Set("Pragma", "no-cache")
	request.Header.Set("Referer", "http://127.0.0.1:5108/home?action=contracts.interface")
	request.Header.Set("Sec-Fetch-Dest", "empty")
	request.Header.Set("Sec-Fetch-Mode", "cors")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	request.Header.Set("Sec-GPC", "1")
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.114 Safari/537.36")

	request.Header.Set("Request-Timeout", "120s")

	client := &http.Client{}
	response, err := client.Do(request)

	_ = response

	require.NoError(t, err)

	time.Sleep(7 * time.Second)

}

func TestRegisterApi(t *testing.T) {
	url := "http://127.0.0.1:5108/api/contracts/apifabric"

	body := `{"name":"crpg","interfaceName":"crpg","interfaceVersion":"1.0","chaincode":"crpg","channel":"firefly","address":""}`

	bodyBytes := []byte(body)

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)

	request.Header.Set("Accept", "*/*")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://127.0.0.1:5108")
	request.Header.Set("Pragma", "no-cache")
	request.Header.Set("Referer", "http://127.0.0.1:5108/home?action=contracts.api")
	request.Header.Set("Sec-Fetch-Dest", "empty")
	request.Header.Set("Sec-Fetch-Mode", "cors")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	request.Header.Set("Sec-GPC", "1")
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.114 Safari/537.36")

	request.Header.Set("Request-Timeout", "120s")

	client := &http.Client{}
	response, err := client.Do(request)

	_ = response

	require.NoError(t, err)
	time.Sleep(7 * time.Second)

}

func TestCreatePlayer1(t *testing.T) {
	testCreateProfile(t, "0x123", "Doc")
	time.Sleep(5 * time.Second)
	testGetProfile(t, "0x123", "Doc")
}

func TestCreatePlayer2(t *testing.T) {
	testCreateProfile(t, "0x456", "Dungeon")
	time.Sleep(5 * time.Second)
	testGetProfile(t, "0x123", "Doc")
}

func TestCreateCharacter(t *testing.T) {
	externalAddress := "0x123"
	profileName := "Doc"
	characterName := "Saintly"

	url := "http://127.0.0.1:5000/api/v1/namespaces/default/apis/crpg/invoke/NewCharacter?fly-sync=true"

	rawBody := `{
		"input": {
		  "externalAddress": "%s",
		  "name": "%s",
		  "signature": "string"
		}
	  }`

	interpretedBody := fmt.Sprintf(rawBody, externalAddress, characterName)
	result, err := postLocal(url, interpretedBody)

	_ = result
	// fmt.Println(result)
	require.NoError(t, err)

	time.Sleep(5 * time.Second)

	// Get the profile again and ensure it contains the character
	profileString := testGetProfile(t, externalAddress, profileName)

	require.Contains(t, profileString, profileName, "Should contain profile name")
	require.Contains(t, profileString, characterName, "Should contain character name")
}

func TestBuyPacks(t *testing.T) {
	owner := "0x456"
	ownerName := "Dungeon"
	quantity := 6

	url := "http://127.0.0.1:5000/api/v1/namespaces/default/apis/crpg/invoke/BuyPacks?fly-sync=true"

	rawBody := `{
		"input": {
		  "owner": "%s",
		  "quantity": %d,
		  "signature": "string"
		}
	  }`

	interpretedBody := fmt.Sprintf(rawBody, owner, quantity)

	result, err := postLocal(url, interpretedBody)

	_ = result
	// fmt.Println(result)
	require.NoError(t, err)

	time.Sleep(5 * time.Second)

	// Now get the vault and make sure it shows packs
	vaultBody := testGetVault(t, owner)

	var vault = &chainTypes.PlayerVault{}
	err = json.Unmarshal([]byte(vaultBody), vault)

	require.NoError(t, err, "Failed to deserialize vault from chain")

	assert.True(t, len(vault.Packs) == quantity, "Vault should have the correct number of packs purchased")

	// Now get the profile and make sure the gold is gone
	profileBody := testGetProfile(t, owner, ownerName)
	var profile = &chainTypes.PlayerProfile{}
	err = json.Unmarshal([]byte(profileBody), profile)

	require.NoError(t, err, "Failed to deserialize profile from chain")

	require.True(t, profile.Gold == chainTypes.PLAYER_DEFAULT_GOLD-quantity*100, "Gold should be reduced")

}

func TestOpenAllPacks(t *testing.T) {
	owner := "0x456"

	// Get Vault
	vaultBody := testGetVault(t, owner)

	var vault = &chainTypes.PlayerVault{}
	err := json.Unmarshal([]byte(vaultBody), vault)

	require.NoError(t, err, "Failed to deserialize vault from chain")

	packCount := len(vault.Packs)

	packsBytes, err := json.Marshal(vault.Packs)

	require.NoError(t, err, "Failed to serialize packs")

	packsString := string(packsBytes)

	url := "http://127.0.0.1:5000/api/v1/namespaces/default/apis/crpg/invoke/OpenPacks?fly-sync=true"

	interpretedBody := fmt.Sprintf(`{
		"input": {
		  "owner": "0x456",
		  "packIds": %s,
		  "signature": "string"
		}
	  }`, packsString)

	result, err := postLocal(url, interpretedBody)

	_ = result
	// fmt.Println(result)
	require.NoError(t, err)

	time.Sleep(5 * time.Second)

	// Now get the vault and make sure it shows tiles
	vaultBody = testGetVault(t, owner)

	err = json.Unmarshal([]byte(vaultBody), vault)

	require.NoError(t, err, "Failed to deserialize vault from chain")

	require.Equal(t, len(vault.Tiles), packCount*chainTypes.PACK_SIZE)

}

func TestMakeDungeon(t *testing.T) {
	url := "http://127.0.0.1:5000/api/v1/namespaces/default/apis/crpg/invoke/MakeDungeon?fly-sync=true"
	owner := "0x456"
	ownerName := "Dungeon"
	height := 5
	width := 10
	dungeonName := "Test"

	// Get Vault
	vaultBody := testGetVault(t, owner)

	var vault = &chainTypes.PlayerVault{}
	err := json.Unmarshal([]byte(vaultBody), vault)

	require.NoError(t, err, "Failed to deserialize vault from chain")

	tileCount := len(vault.Tiles)

	dungeonTiles := make([]chainTypes.DungeonTile, 0, tileCount)

	currentX, currentY := 0, 0

	for _, t := range vault.Tiles {
		newTile := chainTypes.DungeonTile{X: currentX, Y: currentY, TileId: t}
		dungeonTiles = append(dungeonTiles, newTile)

		currentX++

		if currentX >= width {
			currentX = 0
			currentY++
		}

		// Wrap back around
		if currentY >= height {
			currentY = 0
		}
	}

	dungeonTilesBytes, err := json.Marshal(dungeonTiles)

	require.NoError(t, err, "Failed to serialize dungeon tiles")

	dungeonTileString := string(dungeonTilesBytes)

	interpretedBody := fmt.Sprintf(`{
		"input": {
		  "dungeonTiles": %s,
		  "height": %d,
		  "width": %d,
		  "name": "%s",
		  "owner": "%s",
		  "signature": "string"
		}
	  }`, dungeonTileString, height, width, dungeonName, owner)

	result, err := postLocal(url, interpretedBody)

	_ = result
	// fmt.Println(result)
	require.NoError(t, err)

	time.Sleep(7 * time.Second)

	// Now get the profile and make sure it shows the dungeon
	profileBody := testGetProfile(t, owner, ownerName)

	var profile = &chainTypes.PlayerProfile{}
	err = json.Unmarshal([]byte(profileBody), profile)

	require.NoError(t, err, "Failed to deserialize profile from chain")

	require.Contains(t, profile.Dungeons, dungeonName)

}

func TestListDungeon(t *testing.T) {
	dungeonName := "Test"
	dungeonOwner := "0x456"

	url := "http://127.0.0.1:5000/api/v1/namespaces/default/apis/crpg/invoke/ListDungeon?fly-sync=true"

	interpretedBody := fmt.Sprintf(`{
		"input": {
		  "dungeonName": "%s",
		  "dungeonOwner": "%s",
		  "signature": "string"
		}
	  }`, dungeonName, dungeonOwner)

	result, err := postLocal(url, interpretedBody)

	_ = result
	// fmt.Println(result)
	require.NoError(t, err)

	time.Sleep(5 * time.Second)

}

func TestStartDungeon(t *testing.T) {
	characterName := "Saintly"
	dungeonName := "Test"
	dungeonOwner := "0x456"
	player := "0x123"

	url := "http://127.0.0.1:5000/api/v1/namespaces/default/apis/crpg/invoke/StartDungeon?fly-sync=true"

	interpretedBody := fmt.Sprintf(`{
		"input": {
		  "characterName": "%s",
		  "dungeonName": "%s",
		  "dungeonOwner": "%s",
		  "player": "%s",
		  "signature": "string"
		}
	  }`, characterName, dungeonName, dungeonOwner, player)

	result, err := postLocal(url, interpretedBody)

	_ = result
	// fmt.Println(result)
	require.NoError(t, err)

	time.Sleep(5 * time.Second)

}

func TestScoreDungeon(t *testing.T) {
	dungeonName := "Test"
	dungeonOwner := "0x456"
	player := "0x123"
	characterName := "Saintly"

	url := "http://127.0.0.1:5000/api/v1/namespaces/default/apis/crpg/invoke/ScoreDungeon?fly-sync=true"

	interpretedBody := fmt.Sprintf(`{
		"input": {
		  "dungeonName": "%s",
		  "dungeonOwner": "%s",
		  "moves": [
			0,1,2,3,4,5,6
		  ],
		  "signature": "%s",
		  "startTime": 0
		}
	  }`, dungeonName, dungeonOwner, player)

	result, err := postLocal(url, interpretedBody)

	_ = result
	// fmt.Println(result)
	require.NoError(t, err)

	time.Sleep(5 * time.Second)

	// Get the dungeon and confirm the power has gone down

	// Get the player and confirm the power has gone down
	character := testGetCharacter(t, characterName)

	require.True(t, character.Power < chainTypes.CHARACTER_START_POWER)

}
