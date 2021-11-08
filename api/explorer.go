package api

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/config"
	"node/storage"
	"node/storage/deep_actions"
	"node/storage/leveldb"
	"sort"
	"strconv"
)

type ExplorerArgs struct {
	BlocksStart int64 `json:"blocks_start"`
	BlocksLimit int64 `json:"blocks_limit"`
	BlocksLast  bool  `json:"blocks_last"`

	TokensStart int64 `json:"tokens_start"`
	TokensLimit int64 `json:"tokens_limit"`
}

type Block struct {
	Height int64               `json:"height"`
	Hash   string              `json:"hash"`
	Header deep_actions.Header `json:"header"`
	Txs    []deep_actions.Tx   `json:"txs"`
}

func (api *Api) Explorer(args *ExplorerArgs, result *string) error {
	explorer := make(map[string]interface{})
	explorer["blocks"] = explorerBlocks(args.BlocksStart, args.BlocksLimit, args.BlocksLast)
	explorer["tokens"], _ = explorerTokens(args.TokensStart, args.TokensLimit)
	transactions := storage.TransactionsMemory
	sort.Slice(transactions, func(i, j int) bool {
		timestamp1, _ := strconv.ParseInt(transactions[i].Timestamp, 10, 64)
		timestamp2, _ := strconv.ParseInt(transactions[j].Timestamp, 10, 64)
		return timestamp1 < timestamp2
	})
	explorer["transactions"] = 	transactions

	explorerJson, err := json.Marshal(explorer)
	if err != nil {
		return errors.New(fmt.Sprintf("Api explorer error 2: %v", err))
	}

	*result = string(explorerJson)
	return nil
}



func explorerBlocks(start, limit int64, last bool) interface{} {
	if last {
		limit = config.BlockHeight
		start = limit - 30
		if start <= 0 {
			start = 1
		}
	} else {
		if limit > 30 {
			limit = 30
		}

		if start >= limit {
			start = limit - 30
		}

		if start <= 0 {
			start = 1
		}
	}

	var blocks []Block
	for i := start; i <= limit; i++ {
		block := Block{}
		row := leveldb.ChainDB.Get(strconv.FormatInt(i, 10))
		err := json.Unmarshal([]byte(row.Value), &block)
		if err == nil {
			block.Height = apparel.ParseInt64(row.Key)
			blocks = append(blocks, block)
		}
	}

	if blocks != nil {
		sort.Slice(blocks, func(i, j int) bool {
			return blocks[i].Height > blocks[j].Height
		})
	}

	return blocks
}

func explorerTokens(start, limit int64) (interface{}, error) {
	tokens, err := storage.GetTokens(start, limit)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("erorr 1: %v", err))
	}

	return tokens, nil
}
