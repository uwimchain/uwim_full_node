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

type FillTokenCardArgs struct {
	Mnemonic string `json:"mnemonic"` // Мнемофраза
	Proposer string `json:"proposer"` // Владелец токена

	FullName   string                   `json:"full_name"`  // Фамилия Имя Отчество
	BirthDay   string                   `json:"birthday"`   // Дата рождения
	Gender     string                   `json:"gender"`     // Пол
	Country    string                   `json:"country"`    // Страна
	Region     string                   `json:"region"`     // Область
	City       string                   `json:"city"`       // Город
	Social     *deep_actions.Social     `json:"social"`     // Социальные сети
	Messengers *deep_actions.Messengers `json:"messengers"` // Месседжеры
	Email      string                   `json:"email"`      // Эллектронная почта
	Site       string                   `json:"site"`       // Сайт
	Hashtags   string                   `json:"hashtags"`   // Интересы (хэштэги)
	WorkPlace  string                   `json:"work_place"` // Место работы
	School     string                   `json:"school"`     // Школа
	Education  string                   `json:"education"`  // Образование (Колледж/Университет)
	Comment    string                   `json:"comment"`    // Комментарий
}

func (api *Api) FillTokenCard(args *FillTokenCardArgs, result *string) error {
	args.Mnemonic, args.Proposer, args.FullName, args.BirthDay, args.Gender, args.Country, args.Region, args.City,
		args.Email, args.Site, args.Hashtags, args.WorkPlace, args.School, args.Education, args.Comment =
		apparel.TrimToLower(args.Mnemonic), apparel.TrimToLower(args.Proposer), strings.TrimSpace(args.FullName),
		strings.TrimSpace(args.BirthDay), strings.TrimSpace(args.Gender), strings.TrimSpace(args.Country),
		strings.TrimSpace(args.Region), strings.TrimSpace(args.City), strings.TrimSpace(args.Email),
		strings.TrimSpace(args.Site), strings.TrimSpace(args.Hashtags), strings.TrimSpace(args.WorkPlace),
		strings.TrimSpace(args.School), strings.TrimSpace(args.Education), strings.TrimSpace(args.Comment)

	if check := validateCardFields(args); check != 0 {
		return errors.New(strconv.FormatInt(check, 10))
	} else {

		signature := crypt.SignMessageWithSecretKey(
			crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic)),
			[]byte(args.Proposer),
		)

		tokenCard := deep_actions.PersonalTokenCard{
			FullName:   args.FullName,
			BirthDay:   args.BirthDay,
			Gender:     args.Gender,
			Country:    args.Country,
			Region:     args.Region,
			City:       args.City,
			Social:     args.Social,
			Messengers: args.Messengers,
			//Photos:     args.Photos,
			Email:     args.Email,
			Site:      args.Site,
			Hashtags:  args.Hashtags,
			WorkPlace: args.WorkPlace,
			School:    args.School,
			Education: args.Education,
			Comment:   args.Comment,
		}

		jsonString, err := json.Marshal(tokenCard)
		if err != nil {
			log.Println("Api fill token card error 1:", err)
		} else {
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
					"fill_token_card_transaction",
					jsonString,
				),
			)

			jsonString, err := json.Marshal(transaction)
			if err != nil {
				log.Println("Api fill token card error 2:", err)
			} else {
				sender.SendTx(jsonString)
				storage.TransactionsMemory = append(storage.TransactionsMemory, *transaction)
				*result = "Token card filled"
			}
		}
	}

	return nil
}

// Функция для валидации данных карты токена
// Возвращает:
// 0: Данные валидны
// 1: Запрос отправлен не на главную ноду
// 2: Неверная или некорректная мнемофраза
// 3: Неверный или некорректный адрес
// 4: Мнемофраза не совпадает с адресом
// 5:
// 6: Не хвататет средств для совершения операции
func validateCardFields(args *FillTokenCardArgs) int64 {
	if !memory.IsMainNode() {
		return 1
	}

	if check := validateMnemonic(args.Mnemonic, args.Proposer); check != 0 {
		return check
	}

	if check := validateBalance(args.Proposer, config.FillTokenCardCost, config.BaseToken, true); check != 0 {
		return check
	}

	if args.Hashtags != "" {
		if check := validateHashtags(args.Hashtags); check != 0 {
			return check
		}
	}

	return 0
}

func validateHashtags(hashtagsString string) int64 {
	if hashtagsString == "" {
		return 21
	}

	hashtags := strings.Split(strings.TrimSpace(hashtagsString), "#")
	if len(hashtags)-1 < 3 {
		return 22
	}

	if len(hashtags)-1 > 10 {
		return 23
	}

	return 0
}
