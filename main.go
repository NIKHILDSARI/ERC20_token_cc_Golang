package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Token_Contract struct {
	contractapi.Contract
}

type token_holder_info struct {
	Token_name         string `json:"token_name"`
	Token_symbol       string `json:"token_symbol"`
	Holder_name        string `json:"client_name"`
	Holder_id          string `json:"holder_id"`
	Holder_balance     int    `json:"holder_balance"`
	Holder_designation string `json:"holder_designation"`
}

const (
	ERC20_token_name   = "ERC20_token_name"
	ERC20_token_symbol = "ERC20_token_symbol"
	Token_creator_name = "Token_creator_name"
	Total_supply       = "Total_supply"
	allowance_prefix   = "allowance"
)

//************************************************************************************************************************************
// Contract functions for manuplating ledger

func (s *Token_Contract) Initledger(ctx contractapi.TransactionContextInterface, name string, symbol string, creator_name string, total_supply string) error {

	Creator_id_string, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("(Initledger(1:1))could not get client id: %v", err)
	}

	Total_sypply_int, err := strconv.Atoi(total_supply)
	if err != nil {
		return fmt.Errorf("(Initledger(1:2))could not convert total_supply: %v", err)
	}

	Creator_info := token_holder_info{
		Token_name:         name,
		Token_symbol:       symbol,
		Holder_name:        creator_name,
		Holder_id:          Creator_id_string,
		Holder_balance:     Total_sypply_int,
		Holder_designation: "Minter",
	}
	err = ctx.GetStub().PutState(ERC20_token_name, []byte(name))
	if err != nil {
		return fmt.Errorf("(Initledger(1:3)) putstate failed: %v", err)
	}
	err = ctx.GetStub().PutState(ERC20_token_symbol, []byte(symbol))
	if err != nil {
		return fmt.Errorf("(Initledger(1:4)) putstate failed: %v", err)
	}
	err = ctx.GetStub().PutState(Token_creator_name, []byte(creator_name))
	if err != nil {
		return fmt.Errorf("(Initledger(1:5)) putstate failed: %v", err)
	}
	err = ctx.GetStub().PutState(Total_supply, []byte(total_supply))
	if err != nil {
		return fmt.Errorf("(Initledger(1:6)) putstate failed: %v", err)
	}
	Creator_info_json, err := json.Marshal(Creator_info)
	if err != nil {
		return fmt.Errorf("(Initledger(1:7)) Marshal failed: %v", err)
	}
	err = ctx.GetStub().PutState(Creator_id_string, Creator_info_json)
	if err != nil {
		return fmt.Errorf("(Initledger(1:8)) putstate failed: %v", err)
	}
	return nil
}

// here spender is type string his name
func (s *Token_Contract) Approve(ctx contractapi.TransactionContextInterface, spender string, amount int) error {
	owner_id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("(Approve(2:1)) getting owner_id failed:%v", err)
	}
	compositekey, err := ctx.GetStub().CreateCompositeKey(allowance_prefix, []string{owner_id, spender})
	if err != nil {
		return fmt.Errorf("(Approve(2:2)) getting compositekey failed:%v", err)
	}
	err = ctx.GetStub().PutState(compositekey, []byte(strconv.Itoa(amount)))
	if err != nil {
		return fmt.Errorf("(Approve(2:3)) putstate failed:%v", err)
	}
	return nil
}
func (s *Token_Contract) Allowance(ctx contractapi.TransactionContextInterface, owner_id string, spender string) (int, error) {
	//here spender is of type string his name not spender id string
	compositekey, err := ctx.GetStub().CreateCompositeKey(allowance_prefix, []string{owner_id, spender})
	if err != nil {
		return 0, fmt.Errorf("(Allowance(3:1) create compositekey failed):%v", err)
	}
	allowance_byte, err := ctx.GetStub().GetState(compositekey)
	if err != nil {
		return 0, fmt.Errorf("(Allowance(3:2) getting allowance_byte failed):%v", err)
	}
	allowance_int, err := strconv.Atoi(string(allowance_byte))
	if err != nil {
		return 0, fmt.Errorf("(Allowance(3:1) converting to allowance_int failed):%v", err)
	}
	return allowance_int, nil
}
func (s *Token_Contract) CreateyourAccountID(ctx contractapi.TransactionContextInterface, your_name string) (string, error) {
	token_name_byte, err := ctx.GetStub().GetState(ERC20_token_name)
	if err != nil {
		return "", fmt.Errorf("(CreateAccountID(4:1)) getstate token_name_byte failed:%v", err)
	}
	token_symbol_byte, err := ctx.GetStub().GetState(ERC20_token_symbol)
	if err != nil {
		return "", fmt.Errorf("(CreateAccountID(4:2)) getstate token_symbol_byte failed:%v", err)
	}
	client_id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("(CreateAccountID(4:3)) getstate client_id failed:%v", err)
	}
	Client_info := token_holder_info{
		Token_name:         string(token_name_byte),
		Token_symbol:       string(token_symbol_byte),
		Holder_name:        your_name,
		Holder_id:          client_id,
		Holder_balance:     0,
		Holder_designation: "Client",
	}
	Client_info_json, err := json.Marshal(Client_info)
	if err != nil {
		return "", fmt.Errorf("(CreateAccountID(4:4)) Marshal Client_info_json failed:%v", err)
	}
	err = ctx.GetStub().PutState(your_name, Client_info_json)
	if err != nil {
		return "", fmt.Errorf("(CreateAccountID(4:5)) putstate Client_info_json failed:%v", err)
	}
	return client_id, nil
}
func (s *Token_Contract) Transfer(ctx contractapi.TransactionContextInterface, from string, to string, amount int) (string, error) {
	from_id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("(Transfer(5:1)) getstate from_id failed:%v", err)
	}
	compositekey, err := ctx.GetStub().CreateCompositeKey(allowance_prefix, []string{from_id, to})
	if err != nil {
		return "", fmt.Errorf("(Transfer(5:2)) creating compositekey failed:%v", err)
	}
	allowance_value_byte, err := ctx.GetStub().GetState(compositekey)
	if err != nil {
		return "", fmt.Errorf("(Transfer(5:3)) getstate failed:%v", err)
	}
	if allowance_value_byte == nil {
		return "in order to transfer please approve your transaction", nil
	}
	allowance_value_int, err := strconv.Atoi(string(allowance_value_byte))
	if err != nil {
		return "", fmt.Errorf("(Transfer(5:4)) convertion failed:%v", err)
	}
	if allowance_value_int <= 0 {
		return "your approved amount is 0", nil
	}

	from_info_obj := token_holder_info{}
	from_info_byte, err := ctx.GetStub().GetState(from)
	if err != nil {
		return "", fmt.Errorf("(Transfer(5:5)) getstate failed:%v", err)
	}
	err = json.Unmarshal(from_info_byte, &from_info_obj)
	if err != nil {
		return "", fmt.Errorf("(Transfer(5:6)) Unmarshal from_info_obj failed:%v", err)
	}

	to_info_obj := token_holder_info{}
	to_info_byte, err := ctx.GetStub().GetState(to)
	if err != nil {
		return "", fmt.Errorf("(Transfer(5:7)) getstate failed:%v", err)
	}
	err = json.Unmarshal(to_info_byte, &to_info_obj)
	if err != nil {
		return "", fmt.Errorf("(Transfer(5:8)) Unmarshal to_info_obj failed:%v", err)
	}

	if from_info_obj.Holder_balance >= amount {
		from_info_obj.Holder_balance = (from_info_obj.Holder_balance) - amount
		to_info_obj.Holder_balance = (to_info_obj.Holder_balance) + amount
	} else {
		return "balance not safient to make transfer", nil
	}
	from_info_obj_jon, err := json.Marshal(from_info_obj)
	if err != nil {
		return "", fmt.Errorf("(Transfer(5:9)) marshal failed:%v", err)
	}
	err = ctx.GetStub().PutState(from_info_obj.Holder_name, from_info_obj_jon)
	if err != nil {
		return "", fmt.Errorf("(Transfer(5:10)) putstate from_info_obj_json failed:%v", err)
	}

	to_info_obj_json, err := json.Marshal(to_info_obj)
	if err != nil {
		return "", fmt.Errorf("(Transfer(5:11)) marshal failed:%v", err)
	}
	err = ctx.GetStub().PutState(to_info_obj.Holder_name, to_info_obj_json)
	if err != nil {
		return "", fmt.Errorf("(Transfer(5:12)) putstate to_info_obj_json failed:%v", err)
	}
	return "transfer succefull check your account for updated balance", nil
}
func (s *Token_Contract) Balance(ctx contractapi.TransactionContextInterface, name string) (int, error) {
	client_info_byte, err := ctx.GetStub().GetState(name)
	if err != nil {
		return 0, fmt.Errorf("(Balnce(6:1) getstate client_info_byte failed:%v", err)
	}
	if client_info_byte != nil {
		client_obj := token_holder_info{}
		err := json.Unmarshal(client_info_byte, &client_obj)
		if err != nil {
			return 0, fmt.Errorf("(Balnce(6:2) Unmarshal client_obj failed:%v", err)
		}
		return client_obj.Holder_balance, nil
	}
	return 0, nil
}

// func (s *Token_Contract) Transferform() {}
// *****************************************************************************************************************************************
func main() {
	cc := new(Token_Contract)
	Chaincode, err := contractapi.NewChaincode(cc)
	if err != nil {
		log.Panicf("Error creating Token_Contract chaincode: %v", err)
	}
	if err := Chaincode.Start(); err != nil {
		log.Panicf("Error starting Token_Contract chaincode: %v", err)
	}

}
