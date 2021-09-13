package api

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
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

type FillTokenStandardCardArgs struct {
	Mnemonic             string `json:"mnemonic"`           // Мнемофраза
	Proposer             string `json:"proposer"`           // Владелец токена
	StandardCardDataJson string `json:"standard_card_data"` // Данные карточки стандарта токена
}

func (api *Api) FillTokenStandardCard(args *FillTokenStandardCardArgs, result *string) error {
	args.Mnemonic, args.Proposer, args.StandardCardDataJson =
		apparel.TrimToLower(args.Mnemonic), apparel.TrimToLower(args.Proposer), strings.TrimSpace(args.StandardCardDataJson)

	if check := validateStandardCardFields(args); check != 0 {
		return errors.New(strconv.FormatInt(check, 10))
	} else {

		signature := crypt.SignMessageWithSecretKey(
			crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic)),
			[]byte(args.Proposer),
		)

		timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

		transaction := deep_actions.NewTx(
			3,
			apparel.GetNonce(timestamp),
			"",
			config.BlockHeight,
			args.Proposer,
			config.NodeNdAddress,
			config.FillTokenCardCost,
			config.BaseToken,
			timestamp,
			0,
			signature,
			*deep_actions.NewComment(
				"fill_token_standard_card_transaction",
				[]byte(args.StandardCardDataJson),
			),
		)

		jsonString, err := json.Marshal(transaction)
		if err != nil {
			log.Println("Api fill token standard card error 1:", err)
		} else {
			sender.SendTx(jsonString)
			storage.TransactionsMemory = append(storage.TransactionsMemory, *transaction)
			*result = "Token standard card filled"
		}
	}

	return nil
}

// Функция для валидации данных карты стандарта токена
// Возвращает:
// 0: Данные валидны
// 1: Запрос отправлен не на главную ноду
// 2: Неверная или некорректная мнемофраза
// 3: Неверный или некорректный адрес
// 4: Мнемофраза не совпадает с адресом
// 5:
// 6: Не хвататет средств для совершения операции
func validateStandardCardFields(args *FillTokenStandardCardArgs) int64 {
	if !memory.IsMainNode() {
		return 1
	}

	if check := validateMnemonic(args.Mnemonic, args.Proposer); check != 0 {
		return check
	}

	if check := validateBalance(args.Proposer, config.FillTokenStandardCardCost, config.BaseToken, true); check != 0 {
		return check
	}

	if !storage.CheckAddressToken(args.Proposer) {
		return 7
	}

	token := storage.GetAddressToken(args.Proposer)
	switch token.Standard {
	case 0:
		if check := validate0standard(args.StandardCardDataJson); check != 0 {

			return check
		}
		break
	case 2:
		if check := validate2standard(args.StandardCardDataJson); check != 0 {
			return check
		}
		break
	case 3:
		if check := validate3standard(args.StandardCardDataJson); check != 0 {
			return check
		}
		break
	case 4:
		if check := validate4standard(args.StandardCardDataJson); check != 0 {
			return check
		}
		break
	}

	return 0
}

func validate0standard(data string) int64 {
	//if data == "" {
	//	return 101
	//}

	return 0
}

/*func validate1standard(data string) int64 {
	tokenStandardCard := deep_actions.ThxStandardCardData{}
	err := json.Unmarshal([]byte(data), &tokenStandardCard)
	if err != nil {
		return 111
	}

	return 0
}*/

func validate2standard(data string) int64 {
	tokenStandardCard := deep_actions.DonateStandardCardData{}
	err := json.Unmarshal([]byte(data), &tokenStandardCard)
	if err != nil {
		return 121
	}

	return 0
}

func validate3standard(data string) int64 {
	tokenStandardCard := deep_actions.StartUpStandardCardData{}
	err := json.Unmarshal([]byte(data), &tokenStandardCard)
	if err != nil {
		return 131
	}

	return 0
}

func validate4standard(data string) int64 {
	tokenStandardCard := deep_actions.BusinessStandardCardData{}
	err := json.Unmarshal([]byte(data), &tokenStandardCard)
	if err != nil {
		return 141
	}

	if tokenStandardCard.Partners != nil {
		for _, i := range tokenStandardCard.Partners {
			if len(i.Address) != 61 || !crypt.IsAddressUw(i.Address) {
				return 142
			}
		}
	}

	return 0
}
