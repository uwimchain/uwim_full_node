package validation

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts/custom_turing_token_con"
	"node/config"
	"node/storage"
	"node/storage/deep_actions"
	"strings"
)

func validateTransactionType3(t deep_actions.Tx) error {

	switch t.Comment.Title {
	case "create_token_transaction":
		token := deep_actions.Token{}
		err := json.Unmarshal(t.Comment.Data, &token)
		if err != nil {
			return errors.New("create token data error")
		}

		if token.Label == "" {
			return errors.New("token label is empty")
		}

		if int64(len(token.Label)) > config.MaxLabel {
			return errors.New("token label is greater than maximum")
		}

		if int64(len(token.Label)) < config.MinLabel {
			return errors.New("token label is less than the minimum")
		}

		if token.Name == "" {
			return errors.New("token name is empty")
		}

		if int64(len(token.Name)) > config.MaxName {
			return errors.New("token name is greater than maximum")
		}

		if !CheckInInt64Array(config.TokenTypes, token.Type) {
			return errors.New("this type of token does not exist")
		}

		if token.Label != custom_turing_token_con.TokenLabel {
			if token.Type == 2 {
				if token.Standard != 7 {
					return errors.New("invalid token standard")
				}

				if token.Emission != 0 {
					return errors.New("incorrect emission amount")
				}
			} else {
				if token.Standard == 7 {
					return errors.New("invalid token standard")
				}
				if token.Emission == 0 {
					return errors.New("token emission is empty")
				}

				if token.Emission > config.MaxEmission {
					return errors.New("token emission is greater than maximum")
				}
				if token.Emission < config.MinEmission {
					return errors.New("token emission is less than the minimum")
				}
			}
		}

		if deep_actions.CheckToken(token.Label) {
			return errors.New("token already exists")
		}

		address := deep_actions.GetAddress(t.From)

		if address.CheckAddressToken() {
			return errors.New("this user have token")
		}
		break
	case "rename_token_transaction":
		token := deep_actions.Token{}
		err := json.Unmarshal(t.Comment.Data, &token)
		if err != nil {
			return errors.New("rename token data error")
		}

		if token.Label == "" {
			return errors.New("token label is empty")
		}

		if token.Label == config.BaseToken {
			return errors.New("token label is \"uwm\"")
		}

		if token.Name == "" {
			return errors.New("token name is empty")
		}

		if int64(len(token.Name)) > config.MaxName {
			return errors.New("token name is greater than maximum")
		}

		if !deep_actions.CheckToken(token.Label) {
			return errors.New("this token does not exist`s")
		}

		address := deep_actions.GetAddress(t.From)

		if !address.CheckAddressToken() {
			return errors.New("this user haven`t token")
		}

		break
	case "change_token_standard_transaction":
		token := deep_actions.Token{}
		err := json.Unmarshal(t.Comment.Data, &token)
		if err != nil {
			return errors.New("change token standard data error")
		}

		if !apparel.SearchInArray([]int64{0, 1, 3, 4, 5}, token.Standard) {
			return errors.New("invalid token standard 1")
		}

		row := storage.GetTokenJson(token.Label)
		if row == "" {
			return errors.New("invalid token standard 2")
		}

		err = json.Unmarshal([]byte(row), &token)
		if err != nil {
			return errors.New("invalid token standard 3")
		}

		if !storage.CheckToken(token.Label) {
			return errors.New("invalid token standard 4")
		}

		if token.Standard == token.Standard {
			return errors.New("invalid token standard 5")
		}

		if token.Standard == 0 && !apparel.SearchInArray([]int64{1, 3, 4, 5}, token.Standard) {
			return errors.New("invalid token standard 6")
		}

		if token.Standard == 1 && !apparel.SearchInArray([]int64{3, 4, 5}, token.Standard) {
			return errors.New("invalid token standard 7")
		}

		if token.Standard == 3 && !apparel.SearchInArray([]int64{4, 5}, token.Standard) {
			return errors.New("invalid token standard 8")
		}

		if token.Standard == 4 {
			return errors.New("invalid token standard 9")
		}

		break
	case "fill_token_card_transaction":
		tokenCard := deep_actions.PersonalTokenCard{}
		err := json.Unmarshal(t.Comment.Data, &tokenCard)
		if err != nil {
			return errors.New("fill token card data error")
		}

		if tokenCard.Hashtags != "" {
			hashtags := strings.Split(strings.TrimSpace(tokenCard.Hashtags), "#")
			if hashtags != nil && len(hashtags)-1 < 3 {
				return errors.New("invalid token card data 16")
			}

			if hashtags != nil && len(hashtags)-1 > 10 {
				return errors.New("invalid token card data 17")
			}
		}

		break
	case "fill_token_standard_card_transaction":
		address := deep_actions.GetAddress(t.From)
		token := deep_actions.GetToken(address.TokenLabel)
		switch token.Standard {
		case 2:
			if check := validate2standard(string(t.Comment.Data)); check != nil {
				return check
			}
			break
		case 3:
			if check := validate3standard(string(t.Comment.Data)); check != nil {
				return check
			}
			break
		case 4:
			if check := validate4standard(string(t.Comment.Data)); check != nil {
				return check
			}
			break
		case 7:
			if check := validate7standard(string(t.Comment.Data)); check != nil {
				return check
			}
			break
		}
		break
	default:
		return errors.New("transaction type does not match the comment title 3")
	}

	return nil
}

func validate2standard(data string) error {
	tokenStandardCard := deep_actions.DonateStandardCardData{}
	err := json.Unmarshal([]byte(data), &tokenStandardCard)
	if err != nil {
		return errors.New("invalid token standard card data 121")
	}

	return nil
}

func validate3standard(data string) error {
	tokenStandardCard := deep_actions.StartUpStandardCardData{}
	err := json.Unmarshal([]byte(data), &tokenStandardCard)
	if err != nil {
		return errors.New("invalid token standard card data 131")
	}

	return nil
}

func validate4standard(data string) error {
	tokenStandardCard := deep_actions.BusinessStandardCardData{}
	err := json.Unmarshal([]byte(data), &tokenStandardCard)
	if err != nil {
		return errors.New("invalid token standard card data 141")
	}

	return nil
}

func validate7standard(data string) error {
	tokenStandardCard := deep_actions.NftStandardCardData{}
	err := json.Unmarshal([]byte(data), &tokenStandardCard)
	if err != nil {
		return errors.New("invalid token standard card data 171")
	}

	return nil
}
