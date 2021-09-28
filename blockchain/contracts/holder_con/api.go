package holder_con

import (
	"encoding/json"
	"log"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
)

func ValidateAdd(depositorAddress, recipientAddress, tokenLabel string, amount float64, getBlockHeight int64) int64 {
	if !crypt.IsAddressUw(depositorAddress) && !crypt.IsAddressSmartContract(depositorAddress) && !crypt.IsAddressNode(depositorAddress) {
		return 711
	}

	if recipientAddress != "" {
		if !crypt.IsAddressUw(recipientAddress) && !crypt.IsAddressSmartContract(recipientAddress) && !crypt.IsAddressNode(recipientAddress) {
			return 712
		}
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

func ValidateGet(recipientAddress string) int64 {
	if !crypt.IsAddressUw(recipientAddress) && !crypt.IsAddressSmartContract(recipientAddress) && !crypt.IsAddressNode(recipientAddress) {
		return 721
	}

	var holder []Holder
	holderJson := HolderDB.Get(recipientAddress).Value
	if holderJson == "" {
		return 722
	}

	err := json.Unmarshal([]byte(holderJson), &holder)
	if err != nil {
		log.Println("validate get error 1:", err)
		return 723
	}

	if holder == nil {
		return 724
	}

	check := false
	var allTxsAmount float64 = 0
	for _, i := range holder {
		if i.RecipientAddress == recipientAddress && i.GetBlockHeight <= config.BlockHeight {
			check = true
			allTxsAmount += i.Amount
		}
	}

	if !check {
		return 725
	}

	scAddressBalance := contracts.GetBalanceForToken(config.HolderScAddress, config.BaseToken)
	if scAddressBalance.Amount < allTxsAmount {
		log.Println("GG", scAddressBalance.Amount, allTxsAmount)
		return 726
	}

	return 0
}
