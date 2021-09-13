package api

import (
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/blockchain/contracts/trade_token_con"
)

type GetTradeTokensForCrontabArgs struct {
	Key string `json:"key"`
}

func (api *Api) GetTradeTokensForCrontab(args *GetTradeTokensForCrontabArgs, result *string) error {
	if args.Key != "915b2032912dd27a550d84a691f3a973" {
		return errors.New("Invalid api key")
	}
	
	jsonString, err := trade_token_con.GetTokensForCrontab()
	if err != nil {
		return err
	}

	*result = string(jsonString)
	return nil
}
