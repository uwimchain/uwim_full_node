package api

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/storage"
)

type GetTokensArgs struct {
	Start int64 `json:"start"`
	//TokensCount int64 `json:"tokens_count"`
	Limit int64 `json:"limit"`
}

func (api *Api) GetTokens(args *GetTokensArgs, result *string) error {

	//tokens, err := storage.GetTokens(args.TokensCount)
	tokens, err := storage.GetTokens(args.Start, args.Limit)
	if err != nil {
		return errors.New(fmt.Sprintf("Api get tokens error 1: %v", err))
	}

	tokensJson, err := json.Marshal(tokens)
	if err != nil {
		return errors.New(fmt.Sprintf("Api get tokens erorr 2: %v", err))
	}
	*result = string(tokensJson)
	return nil
}
