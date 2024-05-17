package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type RoleContract struct {
	contractapi.Contract
}

// User 用户结构体
type User struct {
	UserName        string	`json:"user_name"`
	UserRole		string	`json:"user_role"`
	UserCredit      int		`json:"user_credit"`
	Power           int		`json:"power"`
}

// UserList 用户列表
type UserList struct {
	Users []string
}

// ADMIN 管理员 PowerPlant 发电方 PowerUser 用电方
const ADMIN string = "admin"
const PowerPlant string = "powerPlant"
const PowerUser string = "powerUser"

// Register 注册用户
func (r *RoleContract) Register(
	ctx contractapi.TransactionContextInterface,
	userName string,
	userRole string) (*User, error) {
	// 0.判断用户是否存在
	if r.UserExist(ctx, userName) {
		return nil, fmt.Errorf("The user is exist ! ")
	}

	// 1.用户加入用户列表
	userListAsBytes, _ := ctx.GetStub().GetState("UserList")

	// 2.检查角色是否符合标准
	if userRole != ADMIN && userRole != PowerPlant && userRole != PowerUser {
		return nil, fmt.Errorf("userRole %s is not right", userRole)
	}

	// 3.用户结构体赋值
	user := User{
		UserName: userName,
		UserRole: userRole,
		UserCredit: InitCredit,
		Power: 0,
	}

	// 4.用户加入用户列表
	userList := new(UserList)
	_ = json.Unmarshal(userListAsBytes, &userList)
	userList.Users= append(userList.Users, userName)

	userAsBytes, _ := json.Marshal(user)
	userListAsBytes, _ = json.Marshal(userList)

	// 5.用户上链
	err1 := ctx.GetStub().PutState("UserList", userListAsBytes)
	err2 := ctx.GetStub().PutState(userName, userAsBytes)

	if err1 != nil && err2 != nil {
		return nil, fmt.Errorf(err1.Error(), err2.Error())
	}

	return &user, nil
}


// QueryUser 查询用户
func (r *RoleContract) QueryUser(
	ctx contractapi.TransactionContextInterface,
	userName string) (*User, error) {
	userAsBytes, err := ctx.GetStub().GetState(userName)

	if err != nil {
		return nil, fmt.Errorf("Failed to query User Info from world state. %s ", err.Error())
	}

	if userAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", userName)
	}

	user := new(User)
	_ = json.Unmarshal(userAsBytes, user)

	return user, nil
}

// ChangeCredit 更改用户信用值
func (r *RoleContract) ChangeCredit(
	ctx contractapi.TransactionContextInterface,
	userName string,
	userCredit int) error {
	// 1.获取用户
	user, err := r.QueryUser(ctx, userName)

	if err != nil {
		return err
	}

	// 2.更改信用值
	user.UserCredit = user.UserCredit + userCredit
	userAsBytes, _ := json.Marshal(user)

	// 3.重新上链
	return ctx.GetStub().PutState(userName, userAsBytes)
}

// ChangePower 更改用户交易量
func (r *RoleContract) ChangePower(
	ctx contractapi.TransactionContextInterface,
	userName string,
	power int) error {
	// 1.获取用户
	user, err := r.QueryUser(ctx, userName)

	if err != nil {
		return err
	}

	// 2.更改能量
	user.Power = user.Power + power
	userAsBytes, _ := json.Marshal(user)

	// 3.重新上链
	return ctx.GetStub().PutState(userName, userAsBytes)
}

// QueryUserList 获取用户列表
func (r *RoleContract) QueryUserList(
	ctx contractapi.TransactionContextInterface) *UserList {
	// 1.获取用户列表
	userListAsBytes, _ := ctx.GetStub().GetState("UserList")

	// 2.赋值
	userList := new(UserList)
	_ = json.Unmarshal(userListAsBytes, &userList)

	// 3.返回用户列表
	return userList
}

//UserExist 判断用户是否存在
func(r *RoleContract) UserExist(
	ctx contractapi.TransactionContextInterface,
	userName string) bool {
	// 1.获取用户
	userAsBytes, _ := ctx.GetStub().GetState(userName)

	// 2.如果返回值为空，用户不存在，则返回false
	if userAsBytes == nil {
		return false
	}

	// 3.用户存在，返回true
	return true
}