package main

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Post struct {
	Hash     string   `json:"hash"`
	Creator  string   `json:"creator"`
	CID      string   `json:"cid"`
	ReplyTo  string   `json:"replyTo"`
	BelongTo string   `json:"belongTo"`
	Assets   []string `json:"assets,omitempty"`

	Deleted bool `json:"deleted"`

	Upvotes   []string            `json:"upvotes,omitempty"`
	Downvotes []string            `json:"downvotes,omitempty"`
	Emojis    map[string][]string `json:"emojis,omitempty"`
}

type Upvote struct {
	Hash    string `json:"hash"`
	Creator string `json:"creator"`
}

type Downvote struct {
	Hash    string `json:"hash"`
	Creator string `json:"creator"`
}

type Emoji struct {
	Hash    string `json:"hash"`
	Creator string `json:"creator"`
	Code    string `json:"code"`
}

type Delete struct {
	Hash    string `json:"hash"`
	Creator string `json:"creator"`
}

// CreatePost creates a post.
func (s *SmartContract) CreatePost(ctx contractapi.TransactionContextInterface, payload string) error {

	post := Post{}
	err := json.Unmarshal([]byte(payload), &post)

	if err != nil {
		return err
	}

	exists, err := s.PostExists(ctx, post.Hash)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the post %s already exists", post.Hash)
	}

	err = ctx.GetStub().PutState(post.Hash, []byte(payload))
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("CreatePost", []byte(payload))
}

func (s *SmartContract) DeletePost(ctx contractapi.TransactionContextInterface, payload string) error {
	delete := Delete{}

	err := json.Unmarshal([]byte(payload), &delete)
	if err != nil {
		return err
	}

	exists, err := s.PostExists(ctx, delete.Hash)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("the post %s does not exist", delete.Hash)
	}

	post, _ := s.ReadPost(ctx, delete.Hash)
	if post.Creator != delete.Creator {
		return fmt.Errorf("the post %s is not created by %s", delete.Hash, delete.Creator)
	}

	post.Deleted = true
	postJSON, _ := json.Marshal(post)
	err = ctx.GetStub().PutState(delete.Hash, postJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("DeletePost", []byte(payload))
}

// PostExists returns true when post with given ID exists in world state
func (s *SmartContract) PostExists(ctx contractapi.TransactionContextInterface, postId string) (bool, error) {
	postJSON, err := ctx.GetStub().GetState(postId)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return postJSON != nil, nil
}

// ReadPost returns the post stored in the world state with given id.
func (s *SmartContract) ReadPost(ctx contractapi.TransactionContextInterface, postId string) (*Post, error) {
	postJSON, err := ctx.GetStub().GetState(postId)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if postJSON == nil {
		return nil, fmt.Errorf("the post %s does not exist", postId)
	}

	var post Post
	json.Unmarshal(postJSON, &post)

	return &post, nil
}

// UpdatePost updates an existing post in the world state with provided parameters.
func (s *SmartContract) UpdatePost(ctx contractapi.TransactionContextInterface, payload string) error {
	next := Post{}
	err := json.Unmarshal([]byte(payload), &next)

	if err != nil {
		return err
	}

	exists, err := s.PostExists(ctx, next.Hash)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the post %s does not exist", next.Hash)
	}

	prev, _ := s.ReadPost(ctx, next.Hash)

	x := reflect.ValueOf(&next).Elem()
	y := reflect.ValueOf(prev).Elem()

	// use reflection package to dynamically update non-zero value
	for i := 0; i < x.NumField(); i++ {
		name := x.Type().Field(i).Name
		yf := y.FieldByName(name)
		xf := x.FieldByName(name)
		if name != "Hash" && yf.CanSet() && !xf.IsZero() {
			yf.Set(xf)
		}
	}

	// overwriting original post with new post
	yJSON, _ := json.Marshal(prev)
	err = ctx.GetStub().PutState(prev.Hash, yJSON)

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("UpdatePost", []byte(payload))
}

func (s *SmartContract) UpvotePost(ctx contractapi.TransactionContextInterface, payload string) error {
	upvote := Upvote{}
	err := json.Unmarshal([]byte(payload), &upvote)
	if err != nil {
		return err
	}

	exists, err := s.PostExists(ctx, upvote.Hash)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the post %s does not exist", upvote.Hash)
	}

	post, _ := s.ReadPost(ctx, upvote.Hash)
	if post.Upvotes == nil {
		post.Upvotes = make([]string, 0)
	}
	flag := false
	for _, v := range post.Upvotes {
		if v == upvote.Creator {
			flag = true
			break
		}
	}
	if !flag {
		post.Upvotes = append(post.Upvotes, upvote.Creator)
		if post.Downvotes != nil {
			for i, v := range post.Downvotes {
				if v == upvote.Creator {
					post.Downvotes = append(post.Downvotes[:i], post.Downvotes[i+1:]...)
					break
				}
			}
		}
	}

	PostJSON, _ := json.Marshal(post)
	err = ctx.GetStub().PutState(upvote.Hash, PostJSON)

	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("UpvotePost", []byte(payload))
}

func (s *SmartContract) DownvotePost(ctx contractapi.TransactionContextInterface, payload string) error {
	downvote := Downvote{}
	err := json.Unmarshal([]byte(payload), &downvote)
	if err != nil {
		return err
	}

	exists, err := s.PostExists(ctx, downvote.Hash)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the post %s does not exist", downvote.Hash)
	}

	post, _ := s.ReadPost(ctx, downvote.Hash)
	if post.Downvotes == nil {
		post.Downvotes = make([]string, 0)
	}
	flag := false
	for _, v := range post.Downvotes {
		if v == downvote.Creator {
			flag = true
			break
		}
	}
	if !flag {
		post.Downvotes = append(post.Downvotes, downvote.Creator)
		if post.Upvotes != nil {
			for i, v := range post.Upvotes {
				if v == downvote.Creator {
					post.Upvotes = append(post.Upvotes[:i], post.Upvotes[i+1:]...)
					break
				}
			}
		}
	}

	PostJSON, _ := json.Marshal(post)

	err = ctx.GetStub().PutState(downvote.Hash, PostJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("DownvotePost", []byte(payload))
}

func (s *SmartContract) AddEmojiPost(ctx contractapi.TransactionContextInterface, payload string) error {
	emoji := Emoji{}
	err := json.Unmarshal([]byte(payload), &emoji)
	if err != nil {
		return err
	}

	exists, err := s.PostExists(ctx, emoji.Hash)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the post %s does not exist", emoji.Hash)
	}

	Post, _ := s.ReadPost(ctx, emoji.Hash)
	if Post.Emojis == nil {
		Post.Emojis = make(map[string][]string)
	}
	Post.Emojis[emoji.Code] = append(Post.Emojis[emoji.Code], emoji.Creator)
	PostJSON, _ := json.Marshal(Post)

	err = ctx.GetStub().PutState(emoji.Hash, PostJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("AddEmojiPost", []byte(payload))
}

func (s *SmartContract) RemoveEmojiPost(ctx contractapi.TransactionContextInterface, payload string) error {
	emoji := Emoji{}
	err := json.Unmarshal([]byte(payload), &emoji)
	if err != nil {
		return err
	}

	exists, err := s.PostExists(ctx, emoji.Hash)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the post %s does not exist", emoji.Hash)
	}

	Post, _ := s.ReadPost(ctx, emoji.Hash)
	if Post.Emojis[emoji.Code] != nil {
		for i, v := range Post.Emojis[emoji.Code] {
			if v == emoji.Creator {
				Post.Emojis[emoji.Code] = append(Post.Emojis[emoji.Code][:i], Post.Emojis[emoji.Code][i+1:]...)
				break
			}
		}
		if len(Post.Emojis[emoji.Code]) == 0 {
			delete(Post.Emojis, emoji.Code)
		}
	}

	PostJSON, _ := json.Marshal(Post)

	err = ctx.GetStub().PutState(emoji.Hash, PostJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return ctx.GetStub().SetEvent("RemoveEmojiPost", []byte(payload))
}

// GetAllPosts returns all posts found in world state
func (s *SmartContract) GetAllPosts(ctx contractapi.TransactionContextInterface) ([]*Post, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var posts []*Post
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var post Post
		json.Unmarshal(queryResponse.Value, &post)
		posts = append(posts, &post)
	}

	return posts, nil
}

func (s *SmartContract) QueryPostsByCreator(ctx contractapi.TransactionContextInterface, creator string) ([]*Post, error) {
	queryString := fmt.Sprintf(`{"selector":{"creator":"%s"}}`, creator)
	return getQueryResultForQueryString(ctx, queryString)
}

func (s *SmartContract) QueryPostsByBelongTo(ctx contractapi.TransactionContextInterface, belongTo string) ([]*Post, error) {
	queryString := fmt.Sprintf(`{"selector":{"belongTo":"%s"}}`, belongTo)
	return getQueryResultForQueryString(ctx, queryString)
}

func (s *SmartContract) QueryPostsByReplyTo(ctx contractapi.TransactionContextInterface, replyTo string) ([]*Post, error) {
	queryString := fmt.Sprintf(`{"selector":{"replyTo":"%s"}}`, replyTo)
	return getQueryResultForQueryString(ctx, queryString)
}

// getQueryResultForQueryString executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Post, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// constructQueryResponseFromIterator constructs a slice of posts from the resultsIterator
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Post, error) {
	var posts []*Post
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var post Post
		json.Unmarshal(queryResult.Value, &post)
		posts = append(posts, &post)
	}

	return posts, nil
}
