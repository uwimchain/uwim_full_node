package api

import "node/blockchain/contracts/trade_token_con"

// GetTradeTokens method arguments
type GetTradeTokensArgs struct{}

func (api *Api) GetTradeTokens(args *GetTradeTokensArgs, result *string) error {
	tradeTokens, err := trade_token_con.GetTokens()
	if err != nil {
		return err
	}

	*result = string(tradeTokens)
	return nil
}
