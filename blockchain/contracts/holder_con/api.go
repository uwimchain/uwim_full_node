package holder_con

import (
	"encoding/json"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
)

func GetHolder(address string) Holders {
	jsonString := HolderDB.Get(address).Value

	holders := Holders{}
	if jsonString != "" {
		_ = json.Unmarshal([]byte(jsonString), &holders)
	}

	return holders
}

func ValidateAdd(senderAddress, recipientAddress, getTokensAddress string, amount float64, tokenLabel string, getBlockHeight int64) int64 {
	if !crypt.IsAddressUw(senderAddress) && !crypt.IsAddressSmartContract(senderAddress) && !crypt.IsAddressNode(senderAddress) {
		return 711
	}

	if recipientAddress != config.HolderScAddress {
		return 711
	}

	if !crypt.IsAddressUw(getTokensAddress) && !crypt.IsAddressSmartContract(getTokensAddress) && !crypt.IsAddressNode(getTokensAddress) {
		return 711
	}

	if tokenLabel != config.BaseToken {
		return 713
	}

	if amount <= 0 {
		return 714
	}

	if getBlockHeight <= 0 {
		return 715
	}

	return 0
}

func ValidateGet(senderAddress, recipientAddress string, amount float64, tokenLabel string) int64 {
	if recipientAddress != config.HolderScAddress {
		return 721
	}

	if !crypt.IsAddressUw(senderAddress) && !crypt.IsAddressSmartContract(senderAddress) && !crypt.IsAddressNode(senderAddress) {
		return 722
	}

	if amount != 0 {
		return 723
	}

	if tokenLabel != config.BaseToken {
		return 724
	}

	holders := GetHolder(senderAddress)
	if holders == nil {
		return 725
	}

	check := false
	var allTxsAmount float64 = 0
	for _, i := range holders {
		if i.RecipientAddress == senderAddress && i.GetBlockHeight <= config.BlockHeight {
			check = true
			allTxsAmount += i.Amount
		}
	}

	if !check {
		return 726
	}

	scAddressBalance := contracts.GetBalanceForToken(config.HolderScAddress, config.BaseToken)
	if scAddressBalance.Amount < allTxsAmount {
		return 727
	}

	return 0
}
