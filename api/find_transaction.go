package api

import (
	"encoding/json"
	"node/storage"
)

type FindTransactionArgs struct {
	Hash string `json:"hash"`
}

/*type FindTransaction struct {
	Transaction deep_actions.Tx `json:"transaction"`
	Comment     string          `json:"comment"`
}*/

func (api *Api) FindTransaction(args *FindTransactionArgs, result *string) error {
	jsonString := storage.GetTxForHash(args.Hash)

	transaction := make(map[string]interface{})
	if jsonString != "" {
		_ = json.Unmarshal([]byte(jsonString), &transaction)
		transaction["status"] = 1 // success
	} else {
		if storage.TransactionsMemory != nil {
			for idx, i := range storage.TransactionsMemory {
				if i.HashTx == args.Hash {
					tmpJsonString, _ := json.Marshal(storage.TransactionsMemory[idx])
					_ = json.Unmarshal(tmpJsonString, &transaction)
					transaction["status"] = 2 // in progress
					break
				}
			}
		}
	}

	if transaction == nil {
		*result = ""
		return nil
	} else {
		resultJsonString, _ := json.Marshal(transaction)
		*result = string(resultJsonString)
		return nil
	}
}
