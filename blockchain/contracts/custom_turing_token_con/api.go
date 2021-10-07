package custom_turing_token_con

import (
	"encoding/json"
	"node/blockchain/contracts"
	"node/config"
)

func GetHolder(address string) interface{} {
	holderJson := HolderDB.Get(address).Value

	holder := Holder{}
	_ = json.Unmarshal([]byte(holderJson), &holder)

	return holder
}

func GetHolders() interface{} {
	holdersJson := HolderDB.GetAll("")

	var holders []Holder

	if holdersJson == nil {
		return nil
	}

	for _, i := range holdersJson {
		holder := Holder{}
		_ = json.Unmarshal([]byte(i.Value), &holder)

		holders = append(holders, holder)
	}

	return holders
}

func ValidateAddEmission(uwAddress, scAddress string, addEmissionAmount float64) int {
	if uwAddress != UwAddress {
		return 911
	}

	if scAddress != ScAddress {
		return 912
	}

	if addEmissionAmount <= 0 {
		return 913
	}

	token := contracts.GetToken(TokenLabel)
	if token.Emission+addEmissionAmount > config.MaxEmission {
		return 914
	}

	return 0
}

func ValidateDeDelegate(uwAddress, scAddress string, amount float64) int {
	if scAddress != ScAddress {
		return 921
	}

	if amount <= 0 {
		return 922
	}

	holderJson := HolderDB.Get(uwAddress).Value
	holder := Holder{}

	if holderJson == "" {
		return 923
	}

	_ = json.Unmarshal([]byte(holderJson), &holder)

	if holder.Amount < amount {
		return 924
	}

	return 0
}

func ValidateDeDelegateAnotherAddress(senderAddress, scAddress string, amount float64) int {
	if scAddress != ScAddress {
		return 931
	}

	if amount <= 0 {
		return 932
	}

	holderJson := HolderDB.Get(senderAddress).Value
	holder := Holder{}

	if holderJson == "" {
		return 933
	}

	_ = json.Unmarshal([]byte(holderJson), &holder)

	if holder.Amount < amount {
		return 934
	}

	return 0
}

func ValidateDelegate(scAddress string, tokenLabel string) int {
	if scAddress != ScAddress {
		return 941
	}

	if tokenLabel != TokenLabel {
		return 942
	}

	return 0
}

func ValidateGetReward(uwAddress, scAddress string) int {
	if uwAddress != UwAddress {
		return 951
	}

	if scAddress != ScAddress {
		return 952
	}

	return 0
}

func ValidateReDelegate(senderAddress, recipientAddress, scAddress string, amount float64) int {
	if scAddress != ScAddress {
		return 961
	}

	if senderAddress == recipientAddress {
		return 962
	}

	if amount <= 0 {
		return 963
	}

	senderJson := HolderDB.Get(senderAddress).Value
	sender := Holder{}

	_ = json.Unmarshal([]byte(senderJson), &sender)
	if sender.Amount < amount {
		return 964
	}

	return 0
}
