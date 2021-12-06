package donate_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
)

func GetEvents(scAddress string) (interface{}, error) {
	var scAddressEvents []contracts.Event
	scAddressEventsJson := EventDB.Get(scAddress).Value
	if scAddressEventsJson != "" {
		err := json.Unmarshal([]byte(scAddressEventsJson), &scAddressEvents)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error 1: %v", err))
		}
	}

	if scAddressEvents == nil {
		return nil, nil
	}

	result := make(map[string][]interface{})

	for _, i := range scAddressEvents {
		result["buy"] = append(result["buy"], i)
	}

	return result, nil
}

func GetConfig(scAddress string) map[string]interface{} {
	scAddressConfig := contracts.GetConfig(ConfigDB, scAddress)
	return scAddressConfig.GetData()
}

func ValidateBuy(scAddress, uwAddress, tokenLabel string, amount float64) int64 {
	if !crypt.IsAddressSmartContract(scAddress) {
		return 111
	}

	if !crypt.IsAddressUw(uwAddress) {
		return 112
	}

	if amount <= 0 {
		return 113
	}

	if tokenLabel != config.BaseToken {
		return 114
	}

	scAddressToken := contracts.GetTokenInfoForScAddress(scAddress)
	if scAddressToken.Id == 0 {
		return 115
	}

	if scAddressToken.Label == tokenLabel {
		return 116
	}

	scAddressConfig := contracts.GetConfig(ConfigDB, scAddress)
	configData := scAddressConfig.GetData()
	conversion := apparel.ConvertInterfaceToFloat64(configData["conversion"])
	maxBuy := apparel.ConvertInterfaceToFloat64(configData["max_buy"])

	if conversion <= 0 {
		return 118
	}

	scAddressBalanceForToken := contracts.GetBalanceForToken(scAddress, scAddressToken.Label)
	scAddressBalanceForTokenUwm := contracts.GetBalanceForToken(scAddress, config.BaseToken)
	txAmount := conversion * amount
	if txAmount <= 0 || txAmount > maxBuy {
		return 119
	}

	txTax := apparel.CalcTax(txAmount)
	if scAddressBalanceForToken.Amount <= txAmount || scAddressBalanceForToken.Amount-txAmount < 1 {
		return 1110
	}

	if scAddressBalanceForTokenUwm.Amount < txTax || scAddressBalanceForTokenUwm.Amount-txTax < 1 {
		return 1111
	}

	return 0
}

func ValidateFillConfig(senderAddress, recipientAddress string, conversion, maxBuy, amount float64, tokenLabel string) int {
	if recipientAddress != config.MainNodeAddress {
		return 121
	}

	if !crypt.IsAddressUw(senderAddress) {
		return 122
	}

	if amount != config.FillTokenConfigCost {
		return 123
	}

	if tokenLabel != config.BaseToken {
		return 124
	}

	if conversion <= 0 {
		return 125
	}

	if maxBuy <= 0 {
		return 126
	}

	address := contracts.GetAddress(senderAddress)
	token := address.GetToken()

	if token.Id == 0 {
		return 127
	}

	if token.Standard != 1 {
		return 128
	}

	return 0
}
