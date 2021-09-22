package api

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/storage"
	"node/storage/deep_actions"
	"node/websocket/sender"
	"strconv"
	"strings"
)

// CreateToken method arguments
type CreateTokenArgs struct {
	Mnemonic string `json:"mnemonic"`
	//Proposer string  `json:"proposer"`
	Label    string  `json:"label"`
	Type     int64   `json:"type"`
	Emission float64 `json:"emission"`
	Name     string  `json:"name"`
}

func (api *Api) CreateToken(args *CreateTokenArgs, result *string) error {
	args.Mnemonic, args.Label = apparel.TrimToLower(args.Mnemonic), apparel.TrimToLower(args.Label)

	proposer := crypt.AddressFromMnemonic(args.Mnemonic)

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))

	if check := validateToken(args.Mnemonic, proposer, args.Label, args.Name, args.Emission, args.Type); check == 0 {

		token := deep_actions.Token{
			Type:      args.Type,
			Label:     args.Label,
			Name:      args.Name,
			Proposer:  proposer,
			Signature: nil,
			Emission:  args.Emission,
			Timestamp: apparel.TimestampUnix(),
		}

		jsonString, _ := json.Marshal(token)

		token.Signature = crypt.SignMessageWithSecretKey(secretKey, jsonString)
		jsonString, _ = json.Marshal(token)

		timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)
		tokenCost := config.NewTokenCost1
		if args.Emission == 10000000 {
			tokenCost = config.NewTokenCost1
		} else if args.Emission > config.MinEmission && args.Emission < config.MaxEmission {
			tokenCost = config.NewTokenCost2
		}

		commentData := jsonString
		comment := deep_actions.Comment{
			Title: "create_token_transaction",
			Data:  commentData,
		}

		tx := deep_actions.Tx{
			Type:       3,
			Nonce:      apparel.GetNonce(timestamp),
			HashTx:     "",
			Height:     config.BlockHeight,
			From:       proposer,
			To:         config.NodeNdAddress,
			Amount:     tokenCost,
			TokenLabel: config.BaseToken,
			Timestamp:  timestamp,
			Tax:        0,
			Signature:  nil,
			Comment:    comment,
		}

		jsonString, _ = json.Marshal(deep_actions.Tx{
			Type:       tx.Type,
			Nonce:      tx.Nonce,
			From:       tx.From,
			To:         tx.To,
			Amount:     tx.Amount,
			TokenLabel: tx.TokenLabel,
			Comment:    tx.Comment,
		})
		tx.Signature = crypt.SignMessageWithSecretKey(secretKey, jsonString)

		sender.SendTx(tx)
		storage.TransactionsMemory = append(storage.TransactionsMemory, tx)
		*result = "Token created"
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
// 6: Длина Label больше 5 символов
// 7: Длина Label меньше 3 символов
// 8: Такой токен уже существует
// 9: Название токена не задано
// 10: Длина названия токена больше 80
// 11: Недопустимый тип токена
// 12: Эмиссия не задана
// 13: Эмиссия больше максимальной
// 14: Эмиссия меньше минимальной
// 18: Недостаточно средств для создания токена
// 19: У пользователя уже есть токен
/*func validateToken(args *CreateTokenArgs) int64 {
	if !memory.IsMainNode() {
		return 1
	}

	validateMnemonic := validateMnemonic(args.Mnemonic, args.Proposer)
	if validateMnemonic != 0 {
		return validateMnemonic
	}

	publicKeyFromAddress, _ := crypt.PublicKeyFromAddress(args.Proposer)
	publicKeyFromMnemonic := crypt.PublicKeyFromSecretKey(crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic)))
	if string(publicKeyFromAddress) != string(publicKeyFromMnemonic) {
		return 4
	}

	if args.Label == "" {
		return 5
	}

	if strings.Contains(args.Label, " ") || strings.Contains(args.Label, "-") || strings.Contains(args.Label, "_") {
		return 6
	}

	if int64(len(args.Label)) > config.MaxLabel {
		return 7
	}

	if int64(len(args.Label)) < config.MinLabel {
		return 8
	}

	if storage.CheckToken(args.Label) {
		return 9
	}

	if args.Name == "" {
		return 10
	}

	if int64(len(args.Name)) > config.MaxName {
		return 11
	}

	if args.Type != 0 {
		return 12
	}

	if args.Emission == 0 {
		return 13
	}

	if args.Emission > config.MaxEmission {
		return 14
	}
	if args.Emission < config.MinEmission {
		return 15
	}

	balance := storage.GetBalance(args.Proposer)
	if balance != nil {
		for _, coin := range balance {
			if coin.TokenLabel == "uwm" {
				if args.Emission == 10000000 {
					if coin.Amount < config.NewTokenCost1 {
						return 16
					}
				} else if args.Emission > 10000000 {
					if coin.Amount < config.NewTokenCost2 {
						return 17
					}
				}
			}
		}
	} else {
		return 18
	}

	token := storage.GetAddressToken(args.Proposer)
	if token.Label != "" {
		return 19
	}

	return 0
}*/
func validateToken(mnemonic, proposer, label, name string, emission float64, tokenType int64) int64 {
	if !memory.IsMainNode() {
		return 1
	}

	validateMnemonic := validateMnemonic(mnemonic, proposer)
	if validateMnemonic != 0 {
		return validateMnemonic
	}

	if label == "" {
		return 5
	}

	if strings.Contains(label, " ") || strings.Contains(label, "-") || strings.Contains(label, "_") {
		return 6
	}

	if int64(len(label)) > config.MaxLabel {
		return 7
	}

	if int64(len(label)) < config.MinLabel {
		return 8
	}

	if storage.CheckToken(label) {
		return 9
	}

	if name == "" {
		return 10
	}

	if int64(len(name)) > config.MaxName {
		return 11
	}

	if tokenType != 0 {
		return 12
	}

	if emission == 0 {
		return 13
	}

	if emission > config.MaxEmission {
		return 14
	}
	if emission < config.MinEmission {
		return 15
	}

	balance := storage.GetBalance(proposer)
	if balance != nil {
		for _, coin := range balance {
			if coin.TokenLabel == "uwm" {
				if emission == 10000000 {
					if coin.Amount < config.NewTokenCost1 {
						return 16
					}
				} else if emission > 10000000 {
					if coin.Amount < config.NewTokenCost2 {
						return 17
					}
				}
			}
		}
	} else {
		return 18
	}

	token := storage.GetAddressToken(proposer)
	if token.Id != 0 {
		return 19
	}

	return 0
}
