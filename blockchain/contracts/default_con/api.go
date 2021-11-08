package default_con

import (
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
	"node/metrics"
)

func GetToken(id int64) interface{} {
	return GetNftTokenElForId(id)
}

func GetAllTokens() interface{} {
	return GetNftAllTokensEls()
}

func ValidateCreate(name, owner, recipient, tokenLabel string, price, amount float64, data string) int {
	if name == "" {
		return 1011
	}

	if !crypt.IsAddressUw(owner) {
		return 1012
	}

	if recipient != config.MainNodeAddress {
		return 1013
	}

	if tokenLabel != config.BaseToken {
		return 1014
	}

	address := contracts.GetAddress(owner)
	if address.TokenLabel == "" {
		return 1015
	}

	token := address.GetToken()
	if token.Standard == 0 {
		return 1016
	}

	tokenElsCount := getParentTokensElsCount(address.TokenLabel)
	if tokenElsCount >= config.NftTokenElsCountMax {
		return 1017
	}

	if price < 0 {
		return 1018
	}

	if amount != config.NftCreateCost {
		return 1019
	}

	//if len(data) > config.NftTokenElMaxDataFieldLen {
	//	return 10110
	//}

	return 0
}

func ValidateBuy(tokenElId int64, recipient, buyer, tokenLabel string, amount float64) int {
	if tokenElId == 0 {
		return 1021
	}

	if !crypt.IsAddressUw(buyer) {
		return 1022
	}

	if tokenLabel != config.BaseToken {
		return 1023
	}

	tokenEl := GetNftTokenElForId(tokenElId)
	if tokenEl == nil {
		return 1024
	}

	parentToken := contracts.GetToken(tokenEl.ParentLabel)
	scAddress := crypt.AddressFromAnotherAddress(metrics.SmartContractPrefix, parentToken.Proposer)
	if recipient != scAddress {
		return 1025
	}

	if tokenEl.Owner == buyer {
		return 1026
	}

	if tokenEl.Price == 0 {
		return 1027
	}

	if tokenEl.Price != amount {
		return 1028
	}

	return 0
}

func ValidateSetPrice(tokenElId int64, sender, recipient, tokenLabel string) int {
	if tokenElId == 0 {
		return 1031
	}

	if !crypt.IsAddressUw(sender) {
		return 1032
	}

	if tokenLabel != config.BaseToken {
		return 1033
	}

	tokenEl := GetNftTokenElForId(tokenElId)
	if tokenEl == nil {
		return 1034
	}

	parentToken := contracts.GetToken(tokenEl.ParentLabel)
	scAddress := crypt.AddressFromAnotherAddress(metrics.SmartContractPrefix, parentToken.Proposer)
	if recipient != scAddress {
		return 1035
	}

	if tokenEl.Owner != sender {
		return 1036
	}

	return 0
}
