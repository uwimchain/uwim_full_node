package api_apparel

import (
	"encoding/json"
	"node/crypt"
	"node/storage/deep_actions"
)

func SignTransaction(secretKey []byte, from, to, tokenLabel string, amount, tax float64, comment deep_actions.Comment) []byte {
	tx := deep_actions.Tx{
		From:       from,
		To:         to,
		Amount:     amount,
		TokenLabel: tokenLabel,
		Tax:        tax,
		Comment:    comment,
	}

	jsonString, _ := json.Marshal(tx)

	return crypt.SignMessageWithSecretKey(secretKey, jsonString)
}
