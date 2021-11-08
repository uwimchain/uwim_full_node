package api

import (
	"encoding/json"
	"node/storage"
	"node/storage/deep_actions"
)

type FindTransactionArgs struct {
	Hash string `json:"hash"`
}

type FindTransaction struct {
	Transaction deep_actions.Tx `json:"transaction"`
	Comment     string          `json:"comment"`
}

func (api *Api) FindTransaction(args *FindTransactionArgs, result *string) error {
	jsonString := storage.GetTxForHash(args.Hash)

	transaction := deep_actions.Tx{}
	_ = json.Unmarshal([]byte(jsonString), &transaction)

	resultJsonString, _ := json.Marshal(FindTransaction{
		Transaction: transaction,
		Comment:     string(transaction.Comment.Data),
	})

	*result = string(resultJsonString)
	return nil
}
