package delegate_validation

import (
	"node/apparel"
	"node/blockchain/contracts"
	"node/blockchain/contracts/delegate_con"
	"node/config"
)

func UnDelegateValidate(address string, amount float64) int64 {
	client := delegate_con.GetBalance(address)
	if client.Balance <= 0 {
		return 1
	}

	delegateContractBalance := contracts.GetDelegateScBalance
	if delegateContractBalance == nil {
		return 2
	}

	for _, coin := range delegateContractBalance {
		if coin.TokenLabel == config.DelegateToken {
			if coin.Amount < apparel.CalcTax(amount)+amount {
				return 2
			}
		}
	}

	return 0
}
