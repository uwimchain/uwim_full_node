package api

import (
	"encoding/json"
	"node/storage"
	"node/storage/deep_actions"
)

// FindToken method arguments
type FindTokenArgs struct {
	Label string `json:"label"`
}

func (api *Api) FindToken(args *FindTokenArgs, result *string) error {
	jsonString := storage.GetToken(args.Label)

	token := deep_actions.Token{}
	_ = json.Unmarshal([]byte(jsonString), &token)

	type TokenCardHistory struct {
		TxHash        string `json:"tx_hash"`
		Timestamp     string `json:"timestamp"`
		TxCommentData string `json:"tx_comment_data"`
	}

	var tokenCardHistory []TokenCardHistory

	if token.CardHistory != nil {

		for _, item := range token.CardHistory {
			jsonString := storage.GetTxForHash(item.TxHash)
			transaction := deep_actions.Tx{}
			_ = json.Unmarshal([]byte(jsonString), &transaction)

			if transaction.Comment.Title == "fill_token_card_transaction" {
				tokenCardHistory = append(tokenCardHistory, TokenCardHistory{transaction.HashTx, transaction.Timestamp, string(transaction.Comment.Data)})
			}
		}
	}

	type FindToken struct {
		Token            deep_actions.Token `json:"token"`
		TokenCardHistory []TokenCardHistory `json:"token_card_history"`
	}

	resultJsonString, _ := json.Marshal(FindToken{
		Token:            token,
		TokenCardHistory: tokenCardHistory,
	})

	*result = string(resultJsonString)
	return nil
}
