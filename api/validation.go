package api

import (
	"fmt"
	"node/api/api_error"
	"node/apparel"
	"node/config"
	"node/storage"
)

func validateBalance(address string, amount float64, tokenLabel string, taxNull, amountNull bool) *api_error.ApiError {
	if !amountNull && amount <= 0 {
		return api_error.NewApiError(6, "Zero or negative amount")
	}

	balance := storage.GetBalance(address)
	if balance == nil {
		return api_error.NewApiError(6, "Balance is null")
	}

	var tax float64 = 0.01
	if taxNull {
		tax = 0
	} else {
		if tokenLabel == config.BaseToken {
			tax = apparel.CalcTax(amount)
		}
	}

	for _, i := range balance {
		if i.TokenLabel == tokenLabel {
			if i.Amount < amount+tax {
				return api_error.NewApiError(6, fmt.Sprintf("Low balance for token \"%s\"", tokenLabel))
			}
		}

		if i.TokenLabel == config.BaseToken {
			if i.Amount < tax {
				return api_error.NewApiError(6, "Low balance for token \"uwm\"")
			}
		}
	}

	return nil
}

func validateTxInMemory(from, to, txType string, value int64) int64 {
	var txValue int64 = 0

	for _, i := range storage.TransactionsMemory {
		if i.From == from && i.To == to && i.Comment.Title == txType {
			txValue += 1
		}
		if txValue == value {
			return 245
		}
	}

	return 0
}