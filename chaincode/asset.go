package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)


type SmartContract struct {
	contractapi.Contract
}


type Asset struct {
	DEALERID    string  `json:"DEALERID"`
	MSISDN      string  `json:"MSISDN"`
	MPIN        string  `json:"MPIN"`
	BALANCE     float64 `json:"BALANCE"`
	STATUS      string  `json:"STATUS"`
	TRANSAMOUNT float64 `json:"TRANSAMOUNT"`
	TRANSTYPE   string  `json:"TRANSTYPE"`
	REMARKS     string  `json:"REMARKS"`
}


func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, dealerID string, msisdn string, mpin string, balance float64, status string, transAmount float64, transType string, remarks string) error {
	exists, err := s.AssetExists(ctx, dealerID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", dealerID)
	}

	asset := Asset{
		DEALERID:    dealerID,
		MSISDN:      msisdn,
		MPIN:        mpin,
		BALANCE:     balance,
		STATUS:      status,
		TRANSAMOUNT: transAmount,
		TRANSTYPE:   transType,
		REMARKS:     remarks,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(dealerID, assetJSON)
}


func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, dealerID string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(dealerID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", dealerID)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}


func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, dealerID string, msisdn string, mpin string, balance float64, status string, transAmount float64, transType string, remarks string) error {
	exists, err := s.AssetExists(ctx, dealerID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", dealerID)
	}

	
	asset := Asset{
		DEALERID:    dealerID,
		MSISDN:      msisdn,
		MPIN:        mpin,
		BALANCE:     balance,
		STATUS:      status,
		TRANSAMOUNT: transAmount,
		TRANSTYPE:   transType,
		REMARKS:     remarks,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(dealerID, assetJSON)
}


func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, dealerID string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(dealerID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}


func (s *SmartContract) GetHistoryForAsset(ctx contractapi.TransactionContextInterface, dealerID string) (string, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(dealerID)
	if err != nil {
		return "", fmt.Errorf("failed to get history for asset %s: %v", dealerID, err)
	}
	defer resultsIterator.Close()

	var history []map[string]interface{}
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return "", err
		}

		var value string
		if !response.IsDelete {
			value = string(response.Value)
		} else {
			value = "DELETED"
		}

		history = append(history, map[string]interface{}{
			"TxId":      response.TxId,
			"Value":     value,
			"Timestamp": response.Timestamp.AsTime().Format(time.RFC3339),
			"IsDelete":  response.IsDelete,
		})
	}

	historyJSON, err := json.Marshal(history)
	if err != nil {
		return "", err
	}

	return string(historyJSON), nil
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error creating asset-transfer-basic chaincode: %v\n", err)
		return
	}

	if err := assetChaincode.Start(); err != nil {
		fmt.Printf("Error starting asset-transfer-basic chaincode: %v\n", err)
	}
}