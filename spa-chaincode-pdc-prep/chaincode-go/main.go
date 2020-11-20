package main

import (
	"simple-payment-application-chaincode/contracts"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	simplePaymentContract := new(contracts.SimplePaymentContract)

	cc, err := contractapi.NewChaincode(simplePaymentContract)

	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
