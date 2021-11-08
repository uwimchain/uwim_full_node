package business_token_con

import (
	"encoding/json"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
)

func GetPartner(scAddress string, address string) Partner {
	partners := GetPartners(scAddress)

	if partners != nil {
		for i := range partners {
			if partners[i].Address == address {
				return partners[i]
			}
		}
	}

	return Partner{}
}

func GetConfig(scAddress string) map[string]interface{} {
	scAddressConfig := contracts.GetConfig(ConfigDB, scAddress)
	configData := scAddressConfig.GetData()

	configData["partners"] = GetPartners(scAddress)

	return configData
}

func ValidateBuy(scAddress, uwAddress, tokenLabel string, amount float64) int64 {
	if !crypt.IsAddressSmartContract(scAddress) {
		return 411
	}

	if !crypt.IsAddressUw(uwAddress) {
		return 412
	}

	if tokenLabel != config.BaseToken {
		return 413
	}

	if amount <= 0 {
		return 414
	}

	scAddressConfig := contracts.GetConfig(ConfigDB, scAddress)
	configData := scAddressConfig.GetData()
	conversion := apparel.ConvertInterfaceToFloat64(configData["conversion"])
	salesValue := apparel.ConvertInterfaceToFloat64(configData["sales_value"])

	if conversion <= 0 {
		return 415
	}

	if salesValue <= 0 {
		return 416
	}

	if (amount * conversion) < salesValue {
		return 417
	}

	return 0
}

func ValidateGetPercent(scAddress, uwAddress, tokenLabel string, amount float64) int64 {
	if !crypt.IsAddressSmartContract(scAddress) {
		return 411
	}

	if !crypt.IsAddressUw(uwAddress) {
		return 412
	}

	if amount <= 0 {
		return 414
	}

	scAddressBalanceForToken := contracts.GetBalanceForToken(scAddress, tokenLabel)
	if scAddressBalanceForToken.Amount < amount {
		return 415
	}

	partners := GetPartners(scAddress)

	if partners == nil {
		return 417
	}

	partnerCheck := false
	for _, i := range partners {
		if i.Address == uwAddress {
			if i.Balance == nil {
				return 418
			}

			tokenCheck := false
			for _, j := range i.Balance {
				if j.TokenLabel == tokenLabel {
					if j.Amount < amount {
						return 419
					}
					tokenCheck = true
					break
				}
			}
			if !tokenCheck {
				return 4120
			}

			partnerCheck = true
			break
		}
	}
	if !partnerCheck {
		return 4121
	}

	return 0
}

func ValidateFillConfig(senderAddress, recipientAddress string, conversion, salesValue float64, partners interface{}, amount float64, tokenLabel string) int {
	if recipientAddress != config.MainNodeAddress {
		return 431
	}

	if !crypt.IsAddressUw(senderAddress) {
		return 432
	}

	if amount != config.FillTokenConfigCost {
		return 433
	}

	if tokenLabel != config.BaseToken {
		return 434
	}

	if conversion <= 0 {
		return 435
	}

	if salesValue <= 0 {
		return 436
	}

	if partners != nil {
		jsonString, _ := json.Marshal(partners)
		partnersArr := Partners{}
		_ = json.Unmarshal(jsonString, &partnersArr)

		if partnersArr != nil {
			for _, i := range partnersArr {
				if i.Address == "" || i.Balance != nil || i.Percent < 0 {
					return 437
				}
			}
		}
	}

	address := contracts.GetAddress(senderAddress)
	token := address.GetToken()
	if token.Id == 0 {
		return 438
	}

	if token.Standard != 4 {
		return 439
	}

	return 0
}
