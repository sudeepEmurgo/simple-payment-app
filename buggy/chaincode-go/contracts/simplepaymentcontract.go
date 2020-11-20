package contracts

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SimplePaymentContract contract for handling writing and reading from the world state
type SimplePaymentContract struct {
	contractapi.Contract
}

// Account : The asset being tracked on the chain
type Account struct {
	AccountID string `json:"accountID"`
	Name      string `json:"name"`
	KYCStatus bool   `json:"kycStatus"`
	Balance   int64  `json:"balance"`
	Bank      string `json:"bank"`
}

var defaultBalance int64

// var accountMapping map[string]Account

// InitLedger : Init the ledger
func (spc *SimplePaymentContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// accountMapping = make(map[string]Account)

	defaultBalance = 100

	return nil
}

// RegisterUserAccount : User registers his account
func (spc *SimplePaymentContract) RegisterUserAccount(ctx contractapi.TransactionContextInterface, name string, bank string) (*Account, error) {

	id, _ := ctx.GetClientIdentity().GetID()
	accountBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}

	if accountBytes != nil {
		return nil, fmt.Errorf("the account already exists for user %s", name)
	}

	account := Account{
		AccountID: id,
		Name:      name,
		KYCStatus: false,
		Balance:   defaultBalance,
		Bank:      bank,
	}
	accountBytes, err = json.Marshal(account)
	if err != nil {
		return nil, err
	}

	return &account, ctx.GetStub().PutState(id, accountBytes)
}

// KYCStatus : to check the senders KYC status
func (spc *SimplePaymentContract) KYCStatus(ctx contractapi.TransactionContextInterface) (bool, error) {

	id, _ := ctx.GetClientIdentity().GetID()
	accountBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	var account Account
	err = json.Unmarshal(accountBytes, &account)
	if err != nil {
		return false, err
	}

	return account.KYCStatus, nil
}

// Balance : to check the senders balance
func (spc *SimplePaymentContract) Balance(ctx contractapi.TransactionContextInterface) (int64, error) {

	id, _ := ctx.GetClientIdentity().GetID()
	accountBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}

	var account Account
	err = json.Unmarshal(accountBytes, &account)
	if err != nil {
		return 0, err
	}

	return account.Balance, nil
}

// Transfer : to transfer amount and update balances
func (spc *SimplePaymentContract) Transfer(ctx contractapi.TransactionContextInterface, beneficiary string, amount int64) (string, error) {

	remitter, _ := ctx.GetClientIdentity().GetID()
	remitterAccountBytes, err := ctx.GetStub().GetState(remitter)
	if err != nil {
		return "check", fmt.Errorf("failed to read from world state for remitter: %v", err)
	}

	beneficiaryAccountBytes, err := ctx.GetStub().GetState(beneficiary)
	if err != nil {
		return "check", fmt.Errorf("failed to read from world state for beneficiary: %v", err)
	}

	var remitterAccount Account
	err = json.Unmarshal(remitterAccountBytes, &remitterAccount)
	if err != nil {
		return "check", err
	}

	var beneficiaryAccount Account
	err = json.Unmarshal(beneficiaryAccountBytes, &beneficiaryAccount)
	if err != nil {
		return "check", err
	}

	if remitterAccount.Balance < amount {
		return "check", fmt.Errorf(" Insufficient balance with the remitter ")
	}

	remitterAccount.Balance -= amount
	beneficiaryAccount.Balance += amount

	remitterAccountBytes, err = json.Marshal(remitterAccount)
	if err != nil {
		return "check", err
	}

	beneficiaryAccountBytes, err = json.Marshal(beneficiaryAccount)
	if err != nil {
		return "check", err
	}

	err = ctx.GetStub().PutState(remitterAccount.AccountID, remitterAccountBytes)
	if err != nil {
		return "check", err
	}

	err = ctx.GetStub().PutState(beneficiaryAccount.AccountID, beneficiaryAccountBytes)
	if err != nil {
		return "check", err
	}

	return "check", nil
}

// DeleteUserAccount deletes an given asset from the world state.
func (s *SimplePaymentContract) DeleteUserAccount(ctx contractapi.TransactionContextInterface, id string) error {

	_, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}

	return ctx.GetStub().DelState(id)
}

// AccountStatement structure used for returning result of history query
type AccountStatement struct {
	Balance   int64     `json:"balance"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

// GetAccountStatement returns the chain of custody for an asset since issuance.
func (t *SimplePaymentContract) GetAccountStatement(ctx contractapi.TransactionContextInterface) ([]AccountStatement, error) {
	id, _ := ctx.GetClientIdentity().GetID()
	log.Printf("GetAccountStatement: ID %v", id)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(id)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []AccountStatement

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var account Account
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &account)
			if err != nil {
				return nil, err
			}
		} else {
			account = Account{
				AccountID: id,
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := AccountStatement{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Balance:   account.Balance,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}
