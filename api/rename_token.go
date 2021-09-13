package api

import (
	"bytes"
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/metrics"
	"node/storage"
	"node/storage/deep_actions"
	"node/websocket/sender"
	"strconv"
	"strings"
)

// CreateToken method arguments
type RenameTokenArgs struct {
	Mnemonic string `json:"mnemonic"`
	Proposer string `json:"proposer"`
	Label    string `json:"label"`
	NewName  string `json:"new_name"`
}

func (api *Api) RenameToken(args *RenameTokenArgs, result *string) error {
	args.Mnemonic, args.Proposer, args.Label = apparel.TrimToLower(args.Mnemonic), apparel.TrimToLower(args.Proposer), apparel.TrimToLower(args.Label)

	if check := validateRenameToken(args); check == 0 {
		signature := crypt.SignMessageWithSecretKey(
			crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic)),
			[]byte(args.Proposer),
		)

		token := deep_actions.NewToken(
			0,
			1,
			args.Label,
			args.NewName,
			"",
			nil,
			0,
			0,
		)

		jsonString, err := json.Marshal(token)
		if err != nil {
			log.Println("Api rename token error 1:", err)
		} else {
			timestamp :=strconv.FormatInt(apparel.TimestampUnix(), 10)
			transaction := *deep_actions.NewTx(
				3,
				apparel.GetNonce(timestamp),
				"",
				config.BlockHeight,
				args.Proposer,
				config.NodeNdAddress,
				config.RenameTokenCost,
				"uwm",
				timestamp,
				0,
				signature,
				*deep_actions.NewComment(
					"rename_token_transaction",
					jsonString,
				),
			)

			jsonString, err := json.Marshal(transaction)
			if err != nil {
				log.Println("Api rename token error 2:", err)
			} else {
				sender.SendTx(jsonString)
				storage.TransactionsMemory = append(storage.TransactionsMemory, transaction)
				*result = "Token renamed"
			}
		}
	} else {
		return errors.New(strconv.FormatInt(check, 10))
	}

	return nil
}

// Функция валидации данных запроса на создание токена.
// Возвращает:
// 0: Данные верны
// 1: Запрос на создание токена отправлен не на главную ноду
// 2: Неверная или некорректная мнемофраза
// 3: Неверный или некорректный адрес
// 4: Мнемофраза не совпадает с адресом
// 5: Label не задан
// 6: Такого токена не существует
// 7: Название токена не задано
// 8: Название токена должно быть меньше 80 символов
// 9: Недостаточно средств для совершения операции
// 10: У пользователя нет токена
// 11: Попытка переименовать основную валюту
func validateRenameToken(args *RenameTokenArgs) int64 {
	if !memory.IsMainNode() {
		return 1
	}

	if args.Mnemonic == "" {
		return 2
	}

	if len(bytes.Split([]byte(strings.TrimSpace(args.Mnemonic)), []byte(" "))) != 24 {
		return 2
	}

	if args.Proposer == "" {
		return 3
	}

	if !strings.HasPrefix(args.Proposer, metrics.AddressPrefix) {
		return 3
	}

	if len(args.Proposer) != 61 {
		return 3
	}

	if args.Proposer == config.GenesisAddress {
		return 3
	}

	publicKeyFromAddress, _ := crypt.PublicKeyFromAddress(args.Proposer)
	publicKeyFromMnemonic := crypt.PublicKeyFromSecretKey(crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic)))
	if string(publicKeyFromAddress) != string(publicKeyFromMnemonic) {
		return 4
	}

	if args.Label == "" {
		return 5
	}

	if !storage.CheckToken(args.Label) {
		return 6
	}

	if args.NewName == "" {
		return 7
	}

	if int64(len(args.NewName)) > config.MaxName {
		return 8
	}

	balance := storage.GetBalance(args.Proposer)
	if balance != nil {
		for _, coin := range balance {
			if coin.TokenLabel == "uwm" {
				if coin.Amount < config.RenameTokenCost {
					return 9
				}
			}
		}
	} else {
		return 9
	}

	token := storage.GetAddressToken(args.Proposer)
	if token.Label == "" {
		return 10
	}

	if args.Label == "uwm" {
		return 11
	}

	return 0
}
