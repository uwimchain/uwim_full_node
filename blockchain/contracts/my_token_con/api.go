package my_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/blockchain/contracts"
	"node/crypt"
	"node/metrics"
	"node/storage"
)

func GetAddressPercent(scAddress string, uwAddress string, tokenLabel string, emission float64, amount float64) (bool, float64, error) {
	if !crypt.IsAddressSmartContract(scAddress) {
		return false, 0, errors.New("error 1: invalid insert data " + scAddress)
	}

	if !crypt.IsAddressUw(uwAddress) {
		return false, 0, errors.New("error 2: invalid insert data " + uwAddress)
	}

	var scAddressPool []Pool
	scAddressPoolJson := PoolDB.Get(scAddress).Value
	if scAddressPoolJson != "" {
		err := json.Unmarshal([]byte(scAddressPoolJson), &scAddressPool)
		if err != nil {
			return false, 0, errors.New(fmt.Sprintf("error 3: %v", err))
		}
	}

	if scAddressPool == nil {
		return false, 0, nil
	}

	for _, i := range scAddressPool {
		if i.Address == uwAddress {
			scAddressBalance := storage.GetBalanceForToken(scAddress, tokenLabel)
			uwAddressPercent := amount / (emission - scAddressBalance.Amount)
			return true, uwAddressPercent, nil
		}
	}

	return false, 0, nil
}

func ValidateConfirmation(scAddress string, uwAddress string) int64 {
	publicKey, err := crypt.PublicKeyFromAddress(uwAddress)
	if err != nil {
		return 011
	}

	if scAddress == crypt.AddressFromPublicKey(metrics.SmartContractPrefix, publicKey) {
		return 012
	}

	var scAddressPool []Pool
	scAddressPoolJson := PoolDB.Get(scAddress).Value
	if scAddressPoolJson != "" {
		err := json.Unmarshal([]byte(scAddressPoolJson), &scAddressPool)
		if err != nil {
			log.Println("my token contract confirmation validation error", err)
			return 0
		}
	}

	for _, i := range scAddressPool {
		if i.Address == uwAddress {
			return 013
		}
	}

	return 0
}

func ValidateGetPercent(scAddress string, uwAddress string) int64 {
	publicKey, err := crypt.PublicKeyFromAddress(uwAddress)
	if err != nil {
		return 021
	}

	if scAddress == crypt.AddressFromPublicKey(metrics.SmartContractPrefix, publicKey) {
		return 022
	}

	var scAddressPool []Pool
	scAddressPoolJson := PoolDB.Get(scAddress).Value
	if scAddressPoolJson != "" {
		err := json.Unmarshal([]byte(scAddressPoolJson), &scAddressPool)
		if err != nil {
			log.Println("my token contract get percent validation error", err)
			return 0
		}
	}

	check := false
	for _, i := range scAddressPool {
		if i.Address == uwAddress {
			check = true
			break
		}
	}

	if !check {
		return 023
	}

	return 0
}

func GetPool(scAddress string) ([]interface{}, error) {
	var scAddressPool []Pool
	scAddressPoolJson := PoolDB.Get(scAddress).Value
	if scAddressPoolJson != "" {
		err := json.Unmarshal([]byte(scAddressPoolJson), &scAddressPool)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("erorr 1: %v", err))
		}
	}
	scAddressToken := contracts.GetTokenInfoForScAddress(scAddress)
	if scAddressToken.Id == 0 {
		return nil, errors.New("error 2: token does not exist")
	}

	scAddressBalanceForToken := storage.GetBalanceForToken(scAddress, scAddressToken.Label)
	var result []interface{}
	for _, i := range scAddressPool {
		uwAddressBalanceForToken := contracts.GetBalanceForToken(i.Address, scAddressToken.Label)

		info := make(map[string]interface{})
		info["address"] = i.Address
		info["amount"] = uwAddressBalanceForToken.Amount
		info["token_label"] = scAddressToken.Label
		info["percent"] = uwAddressBalanceForToken.Amount / (scAddressToken.Emission - scAddressBalanceForToken.Amount)

		result = append(result, info)
	}

	return result, nil
}
