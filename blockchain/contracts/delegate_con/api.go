package delegate_con

import (
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
)

func ValidateDelegate(scAddress, uwAddress string) int {
	if scAddress != config.DelegateScAddress {
		return 1211
	}

	if !crypt.IsAddressUw(uwAddress) {
		return 1212
	}

	return 0
}

func ValidateUnDelegate(scAddress, uwAddress string, amount float64) int {
	if scAddress != config.DelegateScAddress {
		return 1221
	}

	client := GetBalance(uwAddress)
	if client.Balance <= 0 {
		return 1222
	}

	delegateContractBalance := contracts.GetBalance(config.DelegateScAddress)
	if delegateContractBalance == nil {
		return 1223
	}

	for _, coin := range delegateContractBalance {
		if coin.TokenLabel == config.DelegateToken {
			if coin.Amount < apparel.CalcTax(amount)+amount {
				return 1224
			}
			break
		}
	}

	return 0
}
