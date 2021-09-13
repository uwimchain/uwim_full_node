package api

import (
	"bytes"
	"node/apparel"
	"node/config"
	"node/crypt"
	"node/metrics"
	"node/storage"
	"strings"
)

// Функция для валидации мнемофразы и адреса пользователя
// Возвращает:
// 0: Данные валидны
// 2: Неверная или некорректная мнемофраза
// 3: Неверный или некорректный адрес
// 4: Мнемофраза не совпадает с адресом
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

// Функция для валидации баланса
// Возвращает:
// 0: Данные валидны
// 6: Не хвататет средств для совершения операции
func validateBalance(address string, amount float64, tokenLabel string, taxNull bool) int64 {
	balance := storage.GetBalance(address)
	if balance == nil {
		return 6
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

// function for validating the number of identical transactions in memory
// return
// 0: ok
// 245: слишком много однотипных транзакций в памяти
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
