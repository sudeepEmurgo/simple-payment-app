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

// UserPrivateDetails : Captures Account Users Personal Information
type UserPrivateDetails struct {
	DocType string `json:"doctype"`
	UserID  string `json:"userID"`
	Name    string `json:"Name"`
	Address string `json:"address"`
	Sex     string `json:"sex"`
}

var defaultBalance int64
var kycBalance int64

// InitLedger : Init the ledger
func (spc *SimplePaymentContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	defaultBalance = 0
	kycBalance = 100

	return nil
}

//ReadUserDetails : Read the user private details
func (spc *SimplePaymentContract) ReadUserDetails(ctx contractapi.TransactionContextInterface, userID string) (*UserPrivateDetails, error) {

	collectionName, _ := getCollectionName(ctx)

	// Check if user already exists
	userAsBytes, err := ctx.GetStub().GetPrivateData(collectionName, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %v", err)
	}

	if userAsBytes == nil {
		return nil, fmt.Errorf("failed to get user info %v", err)
	}

	var userPrivateDetails UserPrivateDetails
	err = json.Unmarshal(userAsBytes, &userPrivateDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return &userPrivateDetails, nil
}

//ReadUserFromCollection : Read the user private details
func (spc *SimplePaymentContract) ReadUserFromCollection(ctx contractapi.TransactionContextInterface, collection string, userID string) (*UserPrivateDetails, error) {

	// Check if user already exists
	userAsBytes, err := ctx.GetStub().GetPrivateData(collection, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %v", err)
	}

	if userAsBytes == nil {
		return nil, fmt.Errorf("failed to get user infor %v", err)
	}

	var userPrivateDetails UserPrivateDetails
	err = json.Unmarshal(userAsBytes, &userPrivateDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return &userPrivateDetails, nil
}

// RegisterUserAccount : User registers his account
func (spc *SimplePaymentContract) RegisterUserAccount(ctx contractapi.TransactionContextInterface, bank string) (*Account, error) {

	// Handling the private Data of the user
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return nil, fmt.Errorf("Error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		return nil, fmt.Errorf("User details not found in the transient map input")
	}

	type accountUserTransientInput struct {
		Name    string `json:"name"`
		Address string `json:"address"`
		Sex     string `json:"sex"`
	}

	var accountUserInput accountUserTransientInput
	err = json.Unmarshal(transientAssetJSON, &accountUserInput)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(accountUserInput.Name) == 0 {
		return nil, fmt.Errorf("Name field must be a non-empty string")
	}
	if len(accountUserInput.Address) == 0 {
		return nil, fmt.Errorf("Address field must be a non-empty string")
	}
	if len(accountUserInput.Sex) == 0 {
		return nil, fmt.Errorf("Sex field must be a non-empty string")
	}
	if len(bank) == 0 {
		return nil, fmt.Errorf("Bank field must be a non-empty string")
	}

	id := accountUserInput.Name + "@" + bank
	collectionName, _ := getCollectionName(ctx)

	// Check if user already exists
	userAsBytes, err := ctx.GetStub().GetPrivateData(collectionName, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get user info: %v", err)
	}
	if userAsBytes != nil {
		fmt.Println("User already exists ")
		return nil, fmt.Errorf("This User already exists %v", id)
	}

	userInfo := UserPrivateDetails{
		DocType: "UserPrivateDetails",
		UserID:  id,
		Name:    accountUserInput.Name,
		Address: accountUserInput.Address,
		Sex:     accountUserInput.Sex,
	}

	userInfoBytes, err := json.Marshal(userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshall user info: %v", err)
	}

	err = ctx.GetStub().PutPrivateData(collectionName, id, userInfoBytes)
	if err != nil {
		return nil, err
	}

	accountBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("Failed to read Account Info from world state: %v", err)
	}

	if accountBytes != nil {
		return nil, fmt.Errorf("The User account already exists for user")
	}

	account := Account{
		DocType:           "Account",
		AccountID:         id,
		KYCStatus:         false,
		Balance:           defaultBalance,
		Bank:              bank,
		LatestTransaction: "Account Created",
	}

	accountBytes, err = json.Marshal(account)
	if err != nil {
		return nil, err
	}

	err = ctx.GetStub().PutState(id, accountBytes)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// ApproveKYCStatus : Admin approves KYC Status
func (spc *SimplePaymentContract) ApproveKYCStatus(ctx contractapi.TransactionContextInterface, id string) (bool, error) {

	// e := ctx.GetClientIdentity().AssertAttributeValue("OU", "admin")
	// if e != nil {
	// 	return false, fmt.Errorf("User must be admin to provide kyc approva: %v", e)
	// }

	accountBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("Failed to read Account Info from world state: %v", err)
	}

	if accountBytes == nil {
		return false, fmt.Errorf("The User account does not exist")
	}

	var account Account
	err = json.Unmarshal(accountBytes, &account)
	if err != nil {
		return false, err
	}

	account.KYCStatus = true
	account.Balance = kycBalance
	account.LatestTransaction = ctx.GetStub().GetTxID()

	accountBytes, err = json.Marshal(account)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(id, accountBytes)
	if err != nil {
		return false, err
	}

	// Handle Transaction
	transaction := Transaction{
		DocType:       "Transaction",
		TransactionID: ctx.GetStub().GetTxID(),
		Beneficiary:   id,
		Remitter:      account.Bank,
		Amount:        kycBalance,
	}

	var transactionBytes []byte
	transactionBytes, err = json.Marshal(transaction)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(ctx.GetStub().GetTxID(), transactionBytes)
	if err != nil {
		return false, err
	}

	return true, nil
}

// KYCStatus : to check the senders KYC status
func (spc *SimplePaymentContract) KYCStatus(ctx contractapi.TransactionContextInterface, id string) (bool, error) {

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
func (spc *SimplePaymentContract) Balance(ctx contractapi.TransactionContextInterface, id string) (int64, error) {

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
func (spc *SimplePaymentContract) Transfer(ctx contractapi.TransactionContextInterface, remitter, beneficiary string, amount int64) (string, error) {

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

	err := ctx.GetClientIdentity().AssertAttributeValue("organizationalUnitName", "admin")
	if err != nil {
		return fmt.Errorf("User must be admin to provide kyc approva: %v", err)
	}

	_, err = ctx.GetStub().GetState(id)
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
func (spc *SimplePaymentContract) GetAccountStatement(ctx contractapi.TransactionContextInterface, id string) ([]TransactionHistory, error) {
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
	Balance   int64  `json:"balance"`
	AccountID string `json:"accountID"`
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

// getCollectionName is an internal helper function to get collection of submitting client identity.
func getCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get the MSP ID of submitting client identity
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get verified MSPID: %v", err)
	}

	// Create the collection name
	var orgCollection string

	if clientMSPID == "Org1MSP" {

		orgCollection = "DreamLandUserCollection"
	} else {
		orgCollection = "ToyLandUserCollection"
	}

	return orgCollection, nil
}
