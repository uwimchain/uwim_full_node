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
)

// CreateToken method arguments
type ChangeTokenStandardArgs struct {
	Mnemonic string `json:"mnemonic"`
	// 0 - My
	// 1 - Donate
	// 3 - StartUp
	// 4 - Business
	// 5 - Trade
	Standard int64 `json:"standard"`
	//Proposer string `json:"proposer"`
}

func (api *Api) ChangeTokenStandard(args *ChangeTokenStandardArgs, result *string) error {

	args.Mnemonic = apparel.TrimToLower(args.Mnemonic)

	proposer := crypt.AddressFromMnemonic(args.Mnemonic)

	//if check := validateChangeTokenStandard(args); check != 0 {
	//	return errors.New(strconv.FormatInt(check, 10))
	//} else {
	if check := validateChangeTokenStandard(args.Mnemonic, proposer, args.Standard); check != 0 {
		return errors.New(strconv.FormatInt(check, 10))
	} else {
		signature := crypt.SignMessageWithSecretKey(
			crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic)),
			//[]byte(args.Proposer),
			[]byte(proposer),
		)

		//t := storage.GetAddressToken(args.Proposer)
		t := storage.GetAddressToken(proposer)

		if t.Label == "" {
			return errors.New(strconv.FormatInt(9, 10))
		} else {
			token := deep_actions.Token{
				Label:    t.Label,
				Standard: args.Standard,
			}

			jsonString, err := json.Marshal(token)
			if err != nil {
				log.Println("Api change token standard error 1:", err)
			} else {
				timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)
				transaction := deep_actions.NewTx(
					3,
					apparel.GetNonce(timestamp),
					"",
					config.BlockHeight,
					//args.Proposer,
					proposer,
					config.NodeNdAddress,
					config.ChangeTokenStandardCost,
					"uwm",
					timestamp,
					0,
					signature,
					*deep_actions.NewComment(
						"change_token_standard_transaction",
						jsonString,
					),
				)

				jsonString, err := json.Marshal(transaction)
				if err != nil {
					log.Println("Api change token standard error 2:", err)
				} else {
					sender.SendTx(jsonString)
					storage.TransactionsMemory = append(storage.TransactionsMemory, *transaction)
					*result = "Token standard changed"
				}
			}
		}

		return nil
	}
}

// function for validate change token args form
// returns:
// 0 - ok
// 1 - Запрос отправлен не на главную ноду
// 2: Неверная или некорректная мнемофраза
// 3: Неверный или некорректный адрес
// 4: Мнемофраза не совпадает с адресом
// 5: Указан неверный стандарт токена
// 6: Не хвататет средств для совершения операции
/*func validateChangeTokenStandard(args *ChangeTokenStandardArgs) int64 {
	if !memory.IsMainNode() {
		return 1
	}

	if check := validateMnemonic(args.Mnemonic, args.Proposer); check != 0 {
		return check
	}

	//if !apparel.SearchInArray([]int64{0, 1, 2, 3, 4}, args.Standard) {
	if !apparel.SearchInArray([]int64{1, 3, 4, 5}, args.Standard) {
		return 5
	}

	t := storage.GetAddressToken(args.Proposer)
	if !storage.CheckToken(t.Label) {
		return 7
	}

	if args.Standard == t.Standard {
		return 8
	}

	//if t.Standard == 0 && !apparel.SearchInArray([]int64{1, 2}, args.Standard) {
	if t.Standard == 0 && !apparel.SearchInArray([]int64{1, 3, 4, 5}, args.Standard) {
		return 9
	}

	if t.Standard == 1 && !apparel.SearchInArray([]int64{3, 4, 5}, args.Standard) {
		return 9
	}

	//if t.Standard == 3 && args.Standard != 4 {
	if t.Standard == 3 && !apparel.SearchInArray([]int64{4, 6}, args.Standard) {
		return 9
	}

	if check := validateBalance(args.Proposer, config.ChangeTokenStandardCost, config.BaseToken, false); check != 0 {
		return check
	}

	return 0
}*/

func validateChangeTokenStandard(mnemonic, proposer string, standard int64) int64 {
	if !memory.IsMainNode() {
		return 1
	}

	if check := validateMnemonic(mnemonic, proposer); check != 0 {
		return check
	}

	//if !apparel.SearchInArray([]int64{0, 1, 2, 3, 4}, args.Standard) {
	if !apparel.SearchInArray([]int64{1, 3, 4, 5}, standard) {
		return 5
	}

	t := storage.GetAddressToken(proposer)
	if !storage.CheckToken(t.Label) {
		return 7
	}

	if standard == t.Standard {
		return 8
	}

	//if t.Standard == 0 && !apparel.SearchInArray([]int64{1, 2}, args.Standard) {
	if t.Standard == 0 && !apparel.SearchInArray([]int64{1, 3, 4, 5}, standard) {
		return 9
	}

	if t.Standard == 1 && !apparel.SearchInArray([]int64{3, 4, 5}, standard) {
		return 9
	}

	//if t.Standard == 3 && args.Standard != 4 {
	if t.Standard == 3 && !apparel.SearchInArray([]int64{4, 6}, standard) {
		return 9
	}

	if check := validateBalance(proposer, config.ChangeTokenStandardCost, config.BaseToken, false); check != 0 {
		return check
	}

	return 0
}
