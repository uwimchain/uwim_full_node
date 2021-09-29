package api

import (
	"errors"
	"node/apparel"
	"node/blockchain/contracts/trade_token_con"
	"node/crypt"
	"node/metrics"
	"node/storage"
)

// GetTradeToken method arguments
type GetTradeTokenArgs struct {
	TokenLabel string `json:"token_label"`
}

func (api *Api) GetTradeToken(args *GetTradeTokenArgs, result *string) error {
	args.TokenLabel = apparel.TrimToLower(args.TokenLabel)

	token := storage.GetToken(args.TokenLabel)
	if token.Id == 0 {
		return errors.New("this token does not exists")
	}

	publicKey, _ := crypt.PublicKeyFromAddress(token.Proposer)
	scAddress := crypt.AddressFromPublicKey(metrics.SmartContractPrefix, publicKey)
	tradeTokenJson, err := trade_token_con.GetToken(scAddress)
	if err != nil {
		return err
	}

	*result = string(tradeTokenJson)
	return nil
}
