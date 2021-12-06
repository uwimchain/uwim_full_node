package api

import (
	"fmt"
	"node/apparel"
	"node/blockchain/contracts/trade_token_con"
)

type TradeTokenContractSwapFixAmountArgs struct {
	Recipient  string  `json:"recipient"`
	Amount     float64 `json:"amount"`
	TokenLabel string  `json:"token_label"`
}

func (api *Api) TradeTokenContractSwapFixAmount(args *TradeTokenContractSwapFixAmountArgs, result *string) error {
	args.Recipient, args.TokenLabel = apparel.TrimToLower(args.Recipient), apparel.TrimToLower(args.TokenLabel)
	fixedAmount := trade_token_con.FixAmount(args.Recipient, args.Amount, args.TokenLabel)
	*result = fmt.Sprint(fixedAmount)
	return nil
}
