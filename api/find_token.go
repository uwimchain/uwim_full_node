package api

import (
	"encoding/json"
	"log"
	"node/blockchain/contracts/business_token_con"
	"node/blockchain/contracts/donate_token_con"
	"node/blockchain/contracts/my_token_con"
	"node/blockchain/contracts/trade_token_con"
	"node/crypt"
	"node/metrics"
	"node/storage"
	"node/storage/deep_actions"
)

type FindTokenArgs struct {
	Label string `json:"label"`
}

type TokenCardHistory struct {
	TxHash        string `json:"tx_hash"`
	Timestamp     string `json:"timestamp"`
	TxCommentData string `json:"tx_comment_data"`
}

func (api *Api) FindToken(args *FindTokenArgs, result *string) error {
	token := deep_actions.GetToken(args.Label)
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

	tokenScAddress := ""
	var tokenScBalance []deep_actions.Balance
	var tokenScTransactions []deep_actions.Tx
	tokenScInfo := ""
	if token.Proposer != "" {
		publicKey, err := crypt.PublicKeyFromAddress(token.Proposer)
		if err != nil {
			log.Println("Api find token error 1:", err)
		} else {
			tokenScAddress = crypt.AddressFromPublicKey(metrics.SmartContractPrefix, publicKey)
		}

		scAddress := deep_actions.GetAddress(tokenScAddress)
		tokenScBalance = storage.GetBalance(tokenScAddress)
		tokenScTransactions = scAddress.GetTxs()

		switch token.Standard {
		case 0:
			scAddressPool, _ := my_token_con.GetPool(tokenScAddress)
			if scAddressPool != nil {
				jsonString, err := json.Marshal(scAddressPool)
				if err != nil {
					log.Println("Api find token error 2:", err)
				} else {
					tokenScInfo = string(jsonString)
				}
			}
		}
	}

	info := make(map[string]interface{})

	switch token.Standard {
	case 1:
		info["config"] = donate_token_con.GetConfig(tokenScAddress)
		break
	case 4:
		info["config"] = business_token_con.GetConfig(tokenScAddress)
		break
	case 5:
		info["config"] = trade_token_con.GetConfig(tokenScAddress)
		break
	}

	info["token"] = token
	info["token_card_history"] = token
	info["token_sc_address"] = tokenScAddress
	info["token_sc_balance"] = tokenScBalance
	info["token_sc_transactions"] = tokenScTransactions
	info["token_sc_info"] = tokenScInfo

	resultJsonString, _ := json.Marshal(info)

	*result = string(resultJsonString)
	return nil
}
