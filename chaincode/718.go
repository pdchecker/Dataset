package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Board Struct
/*
		블록내용 구성
	1. 작성자
	2. 글제목
	3. 글내용
	4. 등록일자
	5. 등록시간

*/
type Board struct {
	ID       string `json:"ID"`
	Title    string `json:"title"`
	Contents string `json:"contents"`
}

// Create Board Function
/* 첫 글 생성
Input ID(작성자) , title(글제목) , contents(글내용)
*/
func (s *SmartContract) CreateBoard(ctx contractapi.TransactionContextInterface,
	id string,
	title string,
	contents string) error {

	boardExist, err := s.BoardExists(ctx, id, title)

	if err != nil {
		return err
	}
	if boardExist {
		/* 동일한 글이 이미 존재합니다. */
		return fmt.Errorf("the Board %s already exists", id)
	}

	board := Board{
		ID:       id,
		Title:    title,
		Contents: contents,
	}

	boardJSON, err := json.Marshal(board)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, boardJSON)
}

// BoardExist Function
/*
	동일한 작성자 , 동일한 글제목 존재하는지 확인
*/
func (s *SmartContract) BoardExists(ctx contractapi.TransactionContextInterface,
	id string,
	title string) (bool, error) {
	boardJSON, err := ctx.GetStub().GetState(id)

	if err != nil {
		/* 작성자가 동일한 글제목으로 글을 작성했습니다. 생성할 수 없습니다.*/
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return boardJSON != nil, nil
}

// Board Read
/*
	등록된 게시글 블러오기
*/
func (s *SmartContract) ReadBoard(ctx contractapi.TransactionContextInterface,
	id string,
	title string) (*Board, error) {
	boardJSON, err := ctx.GetStub().GetState(id)

	/* GetState 에러확인 */
	if err != nil {
		/* WSDB 자체를 읽는데 실패했습니다. */
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if boardJSON == nil {
		/* WSDB내에 해당 ID의 정보가 존재하지 않습니다.*/
		return nil, fmt.Errorf("the board %s does not exist", id)
	}

	var outputBoard Board
	/* JSON -> Board로 변환*/
	err = json.Unmarshal(boardJSON, &outputBoard)

	/* Unmarshal 도중 에러 발생시*/
	if err != nil {
		/* board는 NULL , 에러값 리턴*/
		return nil, err
	}
	return &outputBoard, nil
}

// Update Board
/*
	글 수정하기 (수정 이력정보 등록)
*/
func (s *SmartContract) UpdateBoard(ctx contractapi.TransactionContextInterface,
	id string,
	title string,
	contents string) error {

	/* 수정하려는 글 존재확인 */
	boardExist, err := s.BoardExists(ctx, id, title)
	if err != nil {
		return err
	}
	if !boardExist {
		/* 대상 글이 존재하지 않습니다.*/
		return fmt.Errorf("the board %s does not exist", id)
	}

	/* 수정한 글의 정보를 받아 Board구조체 저장 */
	inputBoard := Board{
		ID:       id,
		Title:    title,
		Contents: contents,
	}

	inputBoardJSON, err := json.Marshal(inputBoard)
	/* Board -> JSON 변환중 에러 */
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, inputBoardJSON)
}

// Delete Board
/*
	글 삭제하기
*/
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface,
	id string,
	title string) error {
	/* 삭제하려는 글 존재확인 */
	boardExist, err := s.BoardExists(ctx, id, title)
	if err != nil {
		return err
	}
	if !boardExist {
		/* 대상 글이 존재하지 않습니다.*/
		return fmt.Errorf("the board %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}
