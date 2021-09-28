package donate_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
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

	scAddressTokenStandardCardData := contracts.DonateStandardCardData
	err := json.Unmarshal([]byte(scAddressToken.StandardCard), &scAddressTokenStandardCardData)
	if err != nil {
		log.Println("donate token contract validate buy error:", err)
		return 117
	}

	if scAddressTokenStandardCardData.Conversion <= 0 {
		return 118
	}

	scAddressBalanceForToken := contracts.GetBalanceForToken(scAddress, scAddressToken.Label)
	scAddressBalanceForTokenUwm := contracts.GetBalanceForToken(scAddress, config.BaseToken)
	txAmount := scAddressTokenStandardCardData.Conversion * amount
	if txAmount <= 0 {
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
