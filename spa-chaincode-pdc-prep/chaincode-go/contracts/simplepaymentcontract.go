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
	DocType           string `json:"docType"`
	AccountID         string `json:"accountID"`
	Name              string `json:"name"`
	KYCStatus         bool   `json:"kycStatus"`
	Balance           int64  `json:"balance"`
	Bank              string `json:"bank"`
	LatestTransaction string `json:"transaction"`
}

// Transaction : Captures transaction data
type Transaction struct {
	DocType       string `json:"doctype"`
	TransactionID string `json:"txID"`
	Remitter      string `json:"remitter"`
	Beneficiary   string `json:"beneficiary"`
	Amount        int64  `json:"amount"`
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
		DocType:           "Account",
		AccountID:         id,
		Name:              name,
		KYCStatus:         false,
		Balance:           defaultBalance,
		Bank:              bank,
		LatestTransaction: ctx.GetStub().GetTxID(),
	}

	accountBytes, err = json.Marshal(account)
	if err != nil {
		return nil, err
	}

	err = ctx.GetStub().PutState(id, accountBytes)
	if err != nil {
		return nil, err
	}

	transaction := Transaction{
		DocType:       "Transaction",
		TransactionID: ctx.GetStub().GetTxID(),
		Beneficiary:   id,
		Remitter:      bank,
		Amount:        defaultBalance,
	}

	var transactionBytes []byte
	transactionBytes, err = json.Marshal(transaction)
	if err != nil {
		return nil, err
	}

	err = ctx.GetStub().PutState(ctx.GetStub().GetTxID(), transactionBytes)
	if err != nil {
		return nil, err
	}

	return &account, nil
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
	txid := ctx.GetStub().GetTxID()

	remitterAccountBytes, err := ctx.GetStub().GetState(remitter)

	if err != nil {
		return txid, fmt.Errorf("failed to read from world state for remitter: %v", err)
	}

	beneficiaryAccountBytes, err := ctx.GetStub().GetState(beneficiary)
	if err != nil {
		return txid, fmt.Errorf("failed to read from world state for beneficiary: %v", err)
	}

	var remitterAccount Account
	err = json.Unmarshal(remitterAccountBytes, &remitterAccount)
	if err != nil {
		return txid, err
	}

	var beneficiaryAccount Account
	err = json.Unmarshal(beneficiaryAccountBytes, &beneficiaryAccount)
	if err != nil {
		return txid, err
	}

	if remitterAccount.Balance < amount {
		return txid, fmt.Errorf(" Insufficient balance with the remitter ")
	}

	remitterAccount.Balance -= amount
	beneficiaryAccount.Balance += amount

	remitterAccount.LatestTransaction = txid
	beneficiaryAccount.LatestTransaction = txid

	transaction := Transaction{
		DocType:       "Transaction",
		TransactionID: txid,
		Beneficiary:   beneficiaryAccount.AccountID,
		Remitter:      remitterAccount.AccountID,
		Amount:        amount,
	}

	var transactionBytes []byte
	transactionBytes, err = json.Marshal(transaction)
	if err != nil {
		return txid, err
	}

	err = ctx.GetStub().PutState(txid, transactionBytes)
	if err != nil {
		return txid, err
	}

	remitterAccountBytes, err = json.Marshal(remitterAccount)
	if err != nil {
		return txid, err
	}

	beneficiaryAccountBytes, err = json.Marshal(beneficiaryAccount)
	if err != nil {
		return txid, err
	}

	err = ctx.GetStub().PutState(remitterAccount.AccountID, remitterAccountBytes)
	if err != nil {
		return txid, err
	}

	err = ctx.GetStub().PutState(beneficiaryAccount.AccountID, beneficiaryAccountBytes)
	if err != nil {
		return txid, err
	}

	return txid, nil
}

// DeleteUserAccount deletes an given asset from the world state.
func (spc *SimplePaymentContract) DeleteUserAccount(ctx contractapi.TransactionContextInterface, id string) error {

	_, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}

	return ctx.GetStub().DelState(id)
}

// History Queries here

// TransactionHistory structure used for returning result of history query
type TransactionHistory struct {
	Balance           int64     `json:"balance"`
	TxID              string    `json:"txId"`
	TransactionPatner string    `json:"tradePartner"`
	TransactionAmount int64     `json:"tradeAmount"`
	Timestamp         time.Time `json:"timestamp"`
	IsDelete          bool      `json:"isDelete"`
}

// GetAccountStatement returns the chain of custody for an asset since issuance.
func (spc *SimplePaymentContract) GetAccountStatement(ctx contractapi.TransactionContextInterface) ([]TransactionHistory, error) {
	id, _ := ctx.GetClientIdentity().GetID()
	log.Printf("GetAccountStatement: ID %v", id)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(id)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []TransactionHistory

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

		tradePartner, amount, err := getTransactionData(ctx, account)

		record := TransactionHistory{
			TxID:              response.TxId,
			Timestamp:         timestamp,
			Balance:           account.Balance,
			IsDelete:          response.IsDelete,
			TransactionPatner: tradePartner,
			TransactionAmount: amount,
		}
		records = append(records, record)
	}

	return records, nil
}

func getTransactionData(ctx contractapi.TransactionContextInterface, account Account) (string, int64, error) {

	txid := account.LatestTransaction
	transactionBytes, err := ctx.GetStub().GetState(txid)
	if err != nil {
		return "nil", 0, fmt.Errorf("failed to read from world state: %v", err)
	}

	var trx Transaction
	err = json.Unmarshal(transactionBytes, &trx)
	if err != nil {
		return "nil", 0, err
	}

	var tradePartner string
	if trx.Remitter == account.AccountID {
		tradePartner = trx.Beneficiary
	} else {
		tradePartner = trx.Remitter
	}

	return tradePartner, trx.Amount, nil
}

// UserBalanceReport structure used for returning result of history query
type UserBalanceReport struct {
	Balance int64  `json:"balance"`
	Name    string `json:"name"`
}

// Rich Queries Here

//GetAllUserBalanceForOrg : Gets all User Balances
func (spc *SimplePaymentContract) GetAllUserBalanceForOrg(ctx contractapi.TransactionContextInterface, bank string) ([]*UserBalanceReport, error) {

	queryString := fmt.Sprintf(`{"selector":{"docType":"Account","bank":"%s"}}`, bank)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var report []*UserBalanceReport
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var userBalance UserBalanceReport
		err = json.Unmarshal(queryResult.Value, &userBalance)
		if err != nil {
			return nil, err
		}
		report = append(report, &userBalance)
	}

	return report, nil
}
