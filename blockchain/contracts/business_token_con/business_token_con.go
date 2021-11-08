package business_token_con

import (
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/blockchain/contracts"
)

func ChangeStandard(scAddress string) error {
	partners := GetPartners(scAddress)

	if partners == nil {
		return nil
	}

	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		return errors.New("Business token contract error 2: token does not exist")
	}

	var refundTransactions []contracts.Tx
	scAddressBalance := contracts.GetBalance(scAddress)
	if scAddressBalance == nil {
		return nil
	}

	var allRefundAmount []contracts.Balance

	for _, i := range partners {
		if i.Balance == nil {
			continue
		}

		for _, j := range i.Balance {
			if j.Amount > 0 {
				refundTransactions = append(refundTransactions, contracts.Tx{
					To:         i.Address,
					Amount:     j.Amount,
					TokenLabel: j.TokenLabel,
				})

				allRefundAmount = appendToAllRefundAmount(allRefundAmount, j)
			}
		}
	}

	if refundTransactions != nil && allRefundAmount != nil {
		for _, i := range allRefundAmount {
			check := false
			for _, j := range scAddressBalance {
				if i.TokenLabel == j.TokenLabel {
					if i.Amount > j.Amount {
						return nil
					}

					check = true
					break
				}
			}

			if !check {
				return nil
			}
		}

		for _, i := range refundTransactions {
			_ = contracts.RefundTransaction(scAddress, i.To, i.Amount, i.TokenLabel)
		}

		for _, i := range scAddressBalance {
			if i.Amount > 0 {
				check := false
				for _, j := range allRefundAmount {
					if i.TokenLabel == j.TokenLabel {
						_ = contracts.RefundTransaction(scAddress, token.Proposer, i.Amount-j.Amount, i.TokenLabel)
						check = true
						break
					}
				}

				if !check {
					_ = contracts.RefundTransaction(scAddress, token.Proposer, i.Amount, i.TokenLabel)
				}
			}
		}
	}

	return nil
}

func appendToAllRefundAmount(allRefundAmount []contracts.Balance, item contracts.Balance) []contracts.Balance {
	if allRefundAmount == nil {
		allRefundAmount = append(allRefundAmount, item)
		return allRefundAmount
	}

	check := false
	for idx, el := range allRefundAmount {
		if el.TokenLabel == item.TokenLabel {
			allRefundAmount[idx].Amount += item.Amount
			check = true
			break
		}
	}

	if !check {
		allRefundAmount = append(allRefundAmount, item)
	}

	return allRefundAmount
}
