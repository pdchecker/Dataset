package main

import (
	"encoding/json"
	"reflect"

	"fmt"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/assert"
	// logger "github.com/sirupsen/logrus"
)

// callToggleLike Common test function to create like on channel and to be used for all test cases
func callToggleLike(t *testing.T, channelStub *CouchDBMockStub, likeJson string) pb.Response {
	// Create Like on Carrier channel
	toggleLikeResponse := checkInvoke(t, channelStub, "ToggleLike", []byte(likeJson))
	if toggleLikeResponse.Status != shim.OK {
		t.FailNow()
	}

	return toggleLikeResponse
}

// callListLikesByDataCall Common test function to list likes based on input criteria from specified channel
func callListLikesByDataCall(t *testing.T, channelStub *CouchDBMockStub, requestJson string) []ListLikeResponse {
	// Create Like on channel
	var input ListLikeRequest
	json.Unmarshal([]byte(requestJson), &input)
	listLikesResponse := checkInvoke(t, channelStub, "ListLikesByDataCall", []byte(requestJson))
	if listLikesResponse.Status != shim.OK {
		fmt.Println("Test_CountLikes: ListLikesByDataCall failed with message res.Message: ", string(listLikesResponse.Message))
		t.FailNow()
	}

	var output []ListLikeResponse
	json.Unmarshal([]byte(listLikesResponse.Payload), &output)
	return output
}

func Test_ToggleLike_Should_Create_New_Like_For_A_Datacall_When_Like_Does_Not_Exists(t *testing.T) {
	setupMultiChannelTest()

	// Create A DataCall on default channel
	callCreateDatacall(t, defaultStub, CREATE_DATA_CALL_FOR_UPDATE_JSON)

	//Toggle Like
	var input Like
	json.Unmarshal([]byte(LIKE_JSON_MULTI_CARRIER), &input)
	callToggleLike(t, mutlicarrierStub, LIKE_JSON_MULTI_CARRIER)

	// Get Likes based on input criteria
	output := callListLikesByDataCall(t, mutlicarrierStub, LIST_LIKE_CRITERIA_JSON)

	// The returned like to should match the created one
	assert.True(t, reflect.DeepEqual(output[0].Like, input))
}
func Test_ToggleLike_Should_Not_Create_New_Like_When_Datacall_Does_Not_Exists(t *testing.T) {
	fmt.Println("Test_ToggleLike_Should_Not_Create_New_Like_When_Datacall_Does_Not_Exists")
	// setup Multi-channel Test Envrionment
	setupMultiChannelTest()

	// Try to create a like for a datacall that doesn't exist on default channel
	var input Like
	json.Unmarshal([]byte(LIKE_JSON_MULTI_CARRIER), &input)
	toggleLikeResponse := checkInvoke_forError(t, mutlicarrierStub, "ToggleLike", []byte(LIKE_JSON_MULTI_CARRIER))

	// Validate Toggle Response, It should return an Error
	if toggleLikeResponse.Status != shim.OK {
		fmt.Println("Test_CountLikes: ListLikesByDataCall failed with message res.Message: ", string(toggleLikeResponse.Message))
	} else {
		t.FailNow()
	}

	// Fetch likes from stub
	output := callListLikesByDataCall(t, mutlicarrierStub, LIST_LIKE_CRITERIA_JSON)

	// There shouldn't be any likes returned
	assert.Equal(t, 0, len(output), "Test_ToggleLike: Function's success, status code 200.")
}

func Test_ToggleLike_Should_Update_Like_When_Like_Exists_With_Different_Status(t *testing.T) {
	fmt.Println("Test_ToggleLike_Should_Update_Like_When_Like_Exists_With_Different_Status")
	// setup Multi-channel Test Envrionment
	setupMultiChannelTest()
	// Create A DataCall on default channel
	callCreateDatacall(t, defaultStub, CREATE_DATA_CALL_FOR_UPDATE_JSON)

	//Toggle Like
	var input Like
	json.Unmarshal([]byte(LIKE_JSON_MULTI_CARRIER), &input)
	callToggleLike(t, mutlicarrierStub, LIKE_JSON_MULTI_CARRIER)

	// Get Likes based on input criteria
	output := callListLikesByDataCall(t, mutlicarrierStub, LIST_LIKE_CRITERIA_JSON)

	// The returned like to should match the created one
	assert.True(t, reflect.DeepEqual(output[0].Like, input))

	// now we will Unlike the datacall
	callToggleLike(t, mutlicarrierStub, UNLIKE_JSON_MULTI_CARRIER)
	outputs := callListLikesByDataCall(t, mutlicarrierStub, LIST_LIKE_CRITERIA_JSON)
	outputLike := outputs[0]

	// now we will check whether the len has become 0 , if the dataCall has been unliked ( as the listLikes function returns nothing for unlike)
	assert.Equal(t, false, outputLike.Like.Liked, "Test_ToggleLike: Function's success, datacall updated for Unlike.")
}

func Test_ToggleLike_Should_Not_Update_Like_When_Like_Exists_With_Same_Status(t *testing.T) {
	fmt.Println("Test_ToggleLike_Should_Not_Update_Like_When_Like_Exists_With_Same_Status")
	// setup Multi-channel Test Envrionment
	setupMultiChannelTest()

	// Create A DataCall on default channel
	callCreateDatacall(t, defaultStub, CREATE_DATA_CALL_FOR_UPDATE_JSON)

	// like the  Datacall
	callToggleLike(t, mutlicarrierStub, LIKE_JSON_MULTI_CARRIER)

	//Like  the same Datacall
	callToggleLike(t, mutlicarrierStub, LIKE_JSON_MULTI_CARRIER)

	// Fetch likes from stub
	output := callListLikesByDataCall(t, mutlicarrierStub, LIST_LIKE_CRITERIA_JSON)

	var input Like
	json.Unmarshal([]byte(LIKE_JSON_MULTI_CARRIER), &input)

	// Output should be like for the same datacall
	assert.True(t, reflect.DeepEqual(output[0].Like, input))

	// Only One like should be returned
	assert.Equal(t, 1, len(output), "Test_ToggleLike: Function's success, status code 200.")

}

func Test_CountLikes_Should_Return_Number_Of_Likes_Same_As_Set_By_Creating_LikeCountEntry(t *testing.T) {
	fmt.Println("Test_CountLikes_Should_Return_Number_Of_Likes_Same_As_Set_By_Creating_LikeCountEntry")

	setupMultiChannelTest()

	// Create Two Like Entries with delta 1 as defined in mock json
	firstLikeEntryResponse := checkInvoke(t, defaultStub, "CreateLikeCountEntry", []byte(LIKE_ENTRY_JSON))
	if firstLikeEntryResponse.Status != shim.OK {
		fmt.Println("Test_CountLikes: CountLikes failed with message res.Message: ", string(firstLikeEntryResponse.Message))
		t.FailNow()
	}

	secondLikeEntryResponse := checkInvoke(t, defaultStub, "CreateLikeCountEntry", []byte(LIKE_ENTRY_JSON))
	if secondLikeEntryResponse.Status != shim.OK {
		fmt.Println("Test_CountLikes: CountLikes failed with message res.Message: ", string(secondLikeEntryResponse.Message))
		t.FailNow()
	}

	countLikesResponse := checkInvoke(t, defaultStub, "CountLikes", []byte(COUNT_LIKES_JSON))

	if countLikesResponse.Status != shim.OK {
		fmt.Println("ToggleLike failed with message res.Message: ", string(firstLikeEntryResponse.Message))
		t.FailNow()
	}

	var output LikeCountEntry
	json.Unmarshal([]byte(countLikesResponse.Payload), &output)

	// Count should match number of likecount entries made
	assert.Equal(t, 2, output.Delta, "Test_ToggleLike: Function's success, status code 200.")
}

func Test_ListLikesByDataCall_Should_Return_List_Of_Likes_Based_On_Input_Criteria(t *testing.T) {
	fmt.Println("Test_ListLikesByDataCall_Should_Return_List_Of_Likes_Based_On_Input_Criteria")
	// setup Multi-channel Test Envrionment
	setupMultiChannelTest()

	// Create A DataCall on default channel
	callCreateDatacall(t, defaultStub, CREATE_DATA_CALL_FOR_UPDATE_JSON)

	var carrierLikeInput Like
	json.Unmarshal([]byte(LIKE_JSON_CARRIER1), &carrierLikeInput)

	var multiCarrierLike Like
	json.Unmarshal([]byte(LIKE_JSON_MULTI_CARRIER), &multiCarrierLike)

	// Create Like on aais-carrie1 channel
	callToggleLike(t, carrierStub, LIKE_JSON_CARRIER1)

	// Create Like on aais-carries channel
	callToggleLike(t, mutlicarrierStub, LIKE_JSON_MULTI_CARRIER)

	outputLikesFromMultiCarrier := callListLikesByDataCall(t, mutlicarrierStub, LIST_LIKE_CRITERIA_JSON)

	assert.True(t, reflect.DeepEqual(outputLikesFromMultiCarrier[0].Like, carrierLikeInput))
	assert.True(t, reflect.DeepEqual(outputLikesFromMultiCarrier[1].Like, multiCarrierLike))
}
func Test_GetLikeByDataCallAndOrganization_Should_Return_Like_Based_On_Input_Organization(t *testing.T) {
	fmt.Println("Test_ListLikesByDataCall_Should_Return_List_Of_Likes_Based_On_Input_Criteria")
	// setup Multi-channel Test Envrionment
	setupMultiChannelTest()

	// Create A DataCall on default channel
	callCreateDatacall(t, defaultStub, CREATE_DATA_CALL_FOR_UPDATE_JSON)

	var multiCarrierLike Like
	json.Unmarshal([]byte(LIKE_JSON_MULTI_CARRIER), &multiCarrierLike)

	// Create Like on aais-carries channel
	callToggleLike(t, mutlicarrierStub, LIKE_JSON_MULTI_CARRIER)

	getLikeResponse := checkInvoke(t, mutlicarrierStub, "GetLikeByDataCallAndOrganization", []byte(GET_LIKE_CRITERIA_JSON))
	if getLikeResponse.Status != shim.OK {
		fmt.Println("Test_CountLikes: ListLikesByDataCall failed with message res.Message: ", string(getLikeResponse.Message))
		t.FailNow()
	}
	var outputLikesFromMultiCarrier []ListLikeResponse
	json.Unmarshal([]byte(getLikeResponse.Payload), &outputLikesFromMultiCarrier)

	assert.True(t, reflect.DeepEqual(outputLikesFromMultiCarrier[0].Like, multiCarrierLike))
}