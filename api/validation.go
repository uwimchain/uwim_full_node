package api

import (
	"bytes"
	"fmt"
	"node/api/api_error"
	"node/apparel"
	"node/config"
	"node/crypt"
	"node/metrics"
	"node/storage"
	"strings"
)

func validateMnemonic(mnemonic string, address string) int64 {
	if mnemonic == "" {
		return 2
	}

	if len(bytes.Split([]byte(strings.TrimSpace(mnemonic)), []byte(" "))) != 24 {
		return 2
	}

	if !strings.HasPrefix(address, metrics.NodePrefix) &&
		!strings.HasPrefix(address, metrics.SmartContractPrefix) &&
		!strings.HasPrefix(address, metrics.AddressPrefix) {
		return 3
	}

	if len(address) != 61 {
		return 3
	}

	if address == config.GenesisAddress {
		return 3
	}

	publicKeyFromAddress, _ := crypt.PublicKeyFromAddress(address)
	publicKeyFromMnemonic := crypt.PublicKeyFromSecretKey(crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(mnemonic)))
	if string(publicKeyFromAddress) != string(publicKeyFromMnemonic) {
		return 4
	}

	return 0
}

func validateBalance(address string, amount float64, tokenLabel string, taxNull bool) int64 {
	balance := storage.GetBalance(address)
	if balance == nil {
		return 6
	}

	var tax float64 = 0.01
	if taxNull {
		tax = 0
	} else if tokenLabel == config.BaseToken {
		tax = apparel.CalcTax(amount)
	}

	for _, i := range balance {
		if i.TokenLabel == tokenLabel {
			if i.Amount < amount {
				return 6
			}
		}

		if i.TokenLabel == config.BaseToken {
			if i.Amount < tax {
				return 6
			}
		}
	}

	return 0
}

func validateBalanceTest(address string, amount float64, tokenLabel string, taxNull, amountNull bool) *api_error.ApiError {
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

func validateTxInMemoryTest(from, to, txCommentTitle string, value int64) *api_error.ApiError {
	var txValue int64 = 0

	for _, i := range storage.TransactionsMemory {
		if i.From == from && i.To == to && i.Comment.Title == txCommentTitle {
			txValue += 1
		}
		if txValue == value {
			return api_error.NewApiError(245, "Limit exceeded transaction with this type")
		}
	}

	return nil
}
