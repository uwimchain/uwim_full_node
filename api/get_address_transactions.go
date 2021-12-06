package api

import (
	"encoding/json"
	"node/storage"
	"node/storage/deep_actions"
	"sort"
	"strconv"
)

type GetAddressTransactionsArgs struct {
	Address           string `json:"address"`
	Start             int    `json:"start"`
	Count             int    `json:"count"`
	TransactionsCount int    `json:"transactions_count"`
}

func (api *Api) GetAddressTransactions(args *GetAddressTransactionsArgs, result *string) error {
	address := deep_actions.GetAddress(args.Address)

	resultTxs := deep_actions.Txs{}
	count := 0
	if args.Start == 0 && args.Count == 0 {
		resultTxs = address.GetTxs()
	} else {
		txs := address.GetAllTxs()

		if txs != nil {
			count = len(txs)

			sort.Slice(txs, func(i, j int) bool {
				timestamp1, _ := strconv.ParseInt(txs[i].Timestamp, 10, 64)
				timestamp2, _ := strconv.ParseInt(txs[j].Timestamp, 10, 64)
				return timestamp1 > timestamp2
			})

			if args.Start+args.Count >= len(txs) {
				resultTxs = txs[args.Start:]
			} else {
				resultTxs = txs[args.Start : args.Start+args.Count]
			}
		}
	}

	memoryTransactions := deep_actions.Txs{}
	if args.Start == 0 {
		if storage.TransactionsMemory != nil {
			for _, i := range storage.TransactionsMemory {
				if i.From == args.Address || i.To == args.Address {
					memoryTransactions = append(memoryTransactions, i)
				}
			}
		}
	}

	info := make(map[string]interface{})
	info["txs"] = resultTxs
	info["count"] = count
	info["memory_transactions"] = memoryTransactions

	jsonString, _ := json.Marshal(info)
	*result = string(jsonString)

	return nil
}
