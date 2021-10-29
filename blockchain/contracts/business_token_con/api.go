package business_token_con

import (
	"encoding/json"
	"log"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
)

func GetPartners(scAddress string) []Partner {
	var partners []Partner
	jsonString := ContractsDB.Get(scAddress).Value
	_ = json.Unmarshal([]byte(jsonString), &partners)

	return partners
}

func GetPartner(scAddress string, address string) Partner {
	var partners []Partner
	jsonString := ContractsDB.Get(scAddress).Value
	_ = json.Unmarshal([]byte(jsonString), &partners)

	if partners != nil {
		for i, _ := range partners {
			if partners[i].Address == address {
				return partners[i]
			}
		}
	}

	return Partner{}
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

	var scAddressPartners []Partner
	scAddressPartnersJson := ContractsDB.Get(scAddress).Value
	if scAddressPartnersJson != "" {
		err := json.Unmarshal([]byte(scAddressPartnersJson), &scAddressPartners)
		if err != nil {
			log.Println("validate get percent error:", err)
			return 416
		}
	}

	if scAddressPartners == nil {
		return 417
	}

	partnerCheck := false
	for _, i := range scAddressPartners {
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
