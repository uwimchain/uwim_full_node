package api

import (
	"encoding/json"
	"node/blockchain/contracts/delegate_con"
)

type DelegateBalanceArgs struct {
	Address string `json:"address"`
}

func (api *Api) DelegateBalance(args *DelegateBalanceArgs, result *string) error {
	balance := delegate_con.GetBalance(args.Address)
	jsonString, _ := json.Marshal(balance)
	*result = string(jsonString)
	return nil
}
