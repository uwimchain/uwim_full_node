package api

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/config"
	"node/storage"
	"node/storage/deep_actions"
	"node/storage/leveldb"
	"sort"
	"strconv"
)

// Explorer method arguments
type ExplorerArgs struct {
	Start  int64  `json:"start"`
	Limit  int64  `json:"limit"`
	Side   bool   `json:"side"`
	Option string `json:"option"`
}

type Explorer struct {
	Blocks       string `json:"blocks"`
	Transactions string `json:"transactions"`
	Tokens       string `json:"tokens"`
}

func NewExplorer(blocks string, transactions string, tokens string) *Explorer {
	return &Explorer{Blocks: blocks, Transactions: transactions, Tokens: tokens}
}

func (api *Api) Explorer(args *ExplorerArgs, result *string) error {

	if args.Start > config.BlockHeight {
		return errors.New("Starting point is greater than the current block height")
	}

	if args.Limit > config.ExplorerLimit {
		args.Limit = config.ExplorerLimit
	}

	jsonString, _ := json.Marshal(NewExplorer(explorerBlocks(args), explorerTransactions(), explorerTokens()))
	*result = string(jsonString)

	return nil
}

func explorerTransactions() string {
	transactions, _ := json.Marshal(storage.TransactionsMemory)
	return string(transactions)
}

type Block struct {
	Height int64               `json:"height"`
	Hash   string              `json:"hash"`
	Header deep_actions.Header `json:"header"`
	Txs    []deep_actions.Tx   `json:"txs"`
}

func explorerBlocks(args *ExplorerArgs) string {
	switch args.Option {
	case "last":
		start := config.BlockHeight - config.ExplorerLimit
		if start <= 0 {
			start = 1
		}
		return getBlocks(start, config.BlockHeight, true)
	default:
		if args.Start == 0 {
			args.Start = 1
		}
		return getBlocks(args.Start, args.Limit, args.Side)
	}
}

func getBlocks(start int64, limit int64, side bool) string {
	var rows []leveldb.Row

	if side {
		for i := start; i <= limit; i++ {
			if row := leveldb.ChainDB.Get(strconv.FormatInt(i, 10)); row.Value != "" {
				rows = append(rows, row)
			}
		}
	} else {
		for i := start; i >= start-limit; i-- {
			if row := leveldb.ChainDB.Get(strconv.FormatInt(i, 10)); row.Value != "" {
				rows = append(rows, row)
			}
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		return apparel.ParseInt64(rows[i].Key) > apparel.ParseInt64(rows[j].Key)
	})

	var blocks []Block
	for _, row := range rows {
		block := Block{}
		err := json.Unmarshal([]byte(row.Value), &block)
		if err == nil {
			block.Height = apparel.ParseInt64(row.Key)
			blocks = append(blocks, block)
		}
	}

	jsonString, _ := json.Marshal(blocks)
	return string(jsonString)
}

func explorerTokens() string {
	return storage.GetTokens()
}
