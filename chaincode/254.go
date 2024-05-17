package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"sort"
)

type ElectionContract struct {
	contractapi.Contract
}

// ElectionProposal 选举提案
type ElectionProposal struct {
	ElectionProposalName 	string						`json:"election_proposal_name"`
	ProposerName 			string						`json:"proposer_name"`
	CandidateMap 			map[string]Candidate		`json:"candidate_map"`
	VoterMap 				map[string]Voter			`json:"voter_map"`
	State 					string						`json:"state"`
	StartTime 				string						`json:"start_time"`
	EndTime 				string						`json:"end_time"`
}

// Candidate 候选人
type Candidate struct {
	CandidateName 	string  `json:"candidate_name"`
	UserCredit 		int		`json:"user_credit"`
	Power           int		`json:"power"`
	Votes 			int		`json:"votes"`
}

// Voter 投票人
type Voter struct {
	VoterName 		string	`json:"voter_name"`
	UserCredit 		int		`json:"user_credit"`
	Power           int		`json:"power"`
	Voted 			bool	`json:"voted"`
}

// Committee 委员会
type Committee struct {
	Users []string
}

// CreateElectionProposal 创建选举提案
func (e *ElectionContract) CreateElectionProposal(
	ctx contractapi.TransactionContextInterface,
	electionProposalName string,
	proposerName string,
	startTime string,
	endTime string) (*ElectionProposal,error) {
	// 1.判断选举提案是否存在
	if e.ElectionProposalExist(ctx, electionProposalName) {
		return nil, fmt.Errorf("Election proposal existed ! ")
	}

	// 2.判断时间是否符合规范
	var t TimeContract
	if !t.CompareTime(startTime, endTime) {
		return nil, fmt.Errorf("End time earlier than start time ! ")
	}

	// 3.查看proposer是否存在
	var r RoleContract
	proposer, err := r.QueryUser(ctx, proposerName)

	if err != nil {
		return nil, fmt.Errorf("Query proposer false, %s ", err.Error())
	}

	// 4.查看powerUser信用值， 若小于某个额度，则拒绝发起提案
	if proposer.UserCredit - CreditBorder < 0 {
		return nil, fmt.Errorf("proposer credit less than %d ", CreditBorder)
	}

	// 5.获取候选人与投票人
	voterMap := make(map[string]Voter)
	candidateMap := make(map[string]Candidate)
	userList := r.QueryUserList(ctx)

	for _, userName := range userList.Users {
		//获取用户
		user, _ := r.QueryUser(ctx, userName)

		voterMap[user.UserName] = Voter{
			VoterName: user.UserName,
			UserCredit: user.UserCredit,
			Power: user.Power,
			Voted: false,
		}

		if user.UserCredit - 90 > 0  {
			candidateMap[user.UserName] = Candidate{
				CandidateName: user.UserName,
				UserCredit: user.UserCredit,
				Power: user.Power,
				Votes: 0,
			}
		}
	}

	// 6.赋值
	electionProposal := ElectionProposal{
		ElectionProposalName: electionProposalName,
		ProposerName: proposerName,
		CandidateMap: candidateMap,
		VoterMap: voterMap,
		State: "Voting",
		StartTime: startTime,
		EndTime: endTime,
	}

	electionProposalAsBytes, _ := json.Marshal(electionProposal)
	// 7.上链
	err = ctx.GetStub().PutState(electionProposalName, electionProposalAsBytes)

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return &electionProposal, nil
}

func (e *ElectionContract) VoteElectionProposal(
	ctx contractapi.TransactionContextInterface,
	electionProposalName string,
	voterName string,
	candidateName string) (*ElectionProposal, error) {
	var r RoleContract
	// 1.判断选举提案是否存在
	if !e.ElectionProposalExist(ctx, electionProposalName) {
		return nil, fmt.Errorf("Election proposal is not existed ! ")
	}

	// 2.获取选举提案
	electionProposal,err := e.QueryElectionProposal(ctx, electionProposalName)

	if err != nil {
		return nil, err
	}

	// 3.判断是否到达选举时间
	//var t TimeContract
	//if !t.CompareWithNow(electionProposal.StartTime) || t.CompareWithNow(electionProposal.EndTime) {
	//	return nil, fmt.Errorf("The proposal is not voting ! ")
	//}

	// 4.判断投票人是否存在
	if !r.UserExist(ctx, voterName) {
		return nil, fmt.Errorf("%s is not exist ! ", voterName)
	}

	// 5.获取投票人信息
	voter := electionProposal.VoterMap[voterName]

	// 6.判断投票人是否投过票，如果投过票返回
	if voter.Voted == true {
		return nil, fmt.Errorf("voter had voted")
	}

	// 7.修改投票信息
	electionProposal.VoterMap[voterName] = Voter{
		voterName,
		voter.UserCredit,
		voter.Power,
		true}

	// 8.判断候选人是否存在
	if !r.UserExist(ctx, candidateName) {
		return nil, fmt.Errorf("%s is not exist ! ", candidateName)
	}

	// 9.获取候选人的信息
	candidate := electionProposal.CandidateMap[candidateName]

	// 10.增加候选人票数
	electionProposal.CandidateMap[candidateName] = Candidate{
		candidateName,
		candidate.UserCredit,
		candidate.Power,
		candidate.Votes + voter.UserCredit}

	// 11.选举提案上链
	electionProposalAsBytes, _ := json.Marshal(electionProposal)

	err = ctx.GetStub().PutState(electionProposalName, electionProposalAsBytes)

	if err != nil {
		return nil, err
	}

	// 12.更新信用值
	_ = r.ChangeCredit(ctx, voterName, BallotAwardCredit)

	return electionProposal, nil
}

// CheckElectionProposal 检查选举提案结果
func (e *ElectionContract) CheckElectionProposal(
	ctx contractapi.TransactionContextInterface,
	electionProposalName string) (*Committee, error) {
	// 1.判断选举提案是否存在
	if !e.ElectionProposalExist(ctx, electionProposalName) {
		return nil, fmt.Errorf("Election proposal is not existed ! ")
	}

	// 2.获取选举提案
	var r RoleContract
	electionProposal, err := e.QueryElectionProposal(ctx, electionProposalName)

	if err != nil {
		return nil, fmt.Errorf("Failed to query election proposal Info from world state. %s ", err.Error())
	}

	// 3.判断是否到达选举时间
	//var t TimeContract
	//if !t.CompareWithNow(electionProposal.EndTime) {
	//	return nil, fmt.Errorf("The proposal is voting ! ")
	//}

	// 4.更改选举提案的状态
	electionProposal.State = "Done"

	// 5.新建委员会
	committee := new(Committee)
	candidates := []Candidate{}

	// 6.候选人成员放入数组中
	for _, val  := range electionProposal.CandidateMap {
		candidates = append(candidates, val)
	}

	// 7.候选人数组按照票数多少排序
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Votes > candidates[j].Votes
	})

	// 8.选出委员会成员
	k := 0
	for _, v := range candidates {
		if k == CommitteeMemberNumber {
			break
		}

		candidateName := v.CandidateName
		user, _ := r.QueryUser(ctx, candidateName)
		committee.Users = append(committee.Users, user.UserName)
		k++
	}

	// 9.委员会成员上链
	committeeListAsBytes, _ := json.Marshal(committee)
	err1 := ctx.GetStub().PutState("COMMITTEE", committeeListAsBytes)

	if err1 != nil {
		return nil, err1
	}

	return committee, nil
}

// QueryElectionProposal 获取选举提案信息
func (e *ElectionContract) QueryElectionProposal(
	ctx contractapi.TransactionContextInterface,
	electionProposalName string) (*ElectionProposal, error) {
	// 1.获取选举提案信息
	electionProposalAsBytes, err := ctx.GetStub().GetState(electionProposalName)

	if err != nil {
		return nil, fmt.Errorf("Failed to query User Info from world state. %s ", err.Error())
	}

	if electionProposalAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", electionProposalName)
	}
	// 2.赋值
	electionProposal := new(ElectionProposal)
	_ = json.Unmarshal(electionProposalAsBytes, electionProposal)

	return electionProposal, nil
}

// ElectionProposalExist 判断选举提案是否存在
func (e *ElectionContract) ElectionProposalExist(
	ctx contractapi.TransactionContextInterface,
	electionProposalName string) bool {
	// 1.获取选举提案
	electionProposalAsBytes, err := ctx.GetStub().GetState(electionProposalName)

	// 2.如果选举提案不存在，或者获取选举提案失败，返回false
	if err != nil {
		return false
	}

	if electionProposalAsBytes == nil {
		return false
	}
	// 3.如果选举提案存在，返回true
	return true
}

// QueryCommittee 获取委员会成员
func (e *ElectionContract) QueryCommittee(
	ctx contractapi.TransactionContextInterface) *Committee {
	// 1.获取选举提案信息
	committeeAsBytes, err := ctx.GetStub().GetState("COMMITTEE")

	if err != nil {
		return nil
	}

	if committeeAsBytes == nil {
		return nil
	}
	// 2.赋值
	committee := new(Committee)
	_ = json.Unmarshal(committeeAsBytes, committee)

	return committee
}