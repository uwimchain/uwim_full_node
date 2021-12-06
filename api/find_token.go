package api

import (
	"encoding/json"
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
	info := make(map[string]interface{})

	token := deep_actions.GetToken(args.Label)
	info["token"] = token

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

		info["token_card_history"] = tokenCardHistory
	}

	if token.Proposer != "" {
		scAddress := crypt.AddressFromAnotherAddress(metrics.SmartContractPrefix, token.Proposer)
		info["token_sc_address"] = scAddress
		scAddressObj := deep_actions.GetAddress(scAddress)
		info["token_sc_balance"] = scAddressObj.GetBalance()
		info["token_sc_transactions"] = scAddressObj.GetTxs()

		switch token.Standard {
		case 0:
			scAddressPool, _ := my_token_con.GetPool(scAddress)
			if scAddressPool != nil {
				jsonString, _ := json.Marshal(scAddressPool)
				//if err != nil {
				//	log.Println("Api find token error 2:", err)
				//} else {
				//	tokenScInfo = string(jsonString)
				//}
				info["token_sc_info"] = string(jsonString)
			}
		case 1:
			info["config"] = donate_token_con.GetConfig(scAddress)
			break
		case 4:
			info["config"] = business_token_con.GetConfig(scAddress)
			break
		case 5:
			info["config"] = trade_token_con.GetConfig(scAddress)
			tokenScInfo, _ := trade_token_con.GetToken(scAddress)
			info["token_sc_info"] = string(tokenScInfo)
			break
		}
	}

	resultJsonString, _ := json.Marshal(info)

	*result = string(resultJsonString)
	return nil
}
