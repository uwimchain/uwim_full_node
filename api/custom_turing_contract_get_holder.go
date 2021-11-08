package api

import (
	"encoding/json"
	"node/apparel"
	"node/blockchain/contracts/custom_turing_token_con"
)

type CustomTuringContractGetHolderArgs struct {
	Address string `json:"address"`
}

func (api *Api) CustomTuringContractGetHolder(args *CustomTuringContractGetHolderArgs, result *string) error {
	args.Address = apparel.TrimToLower(args.Address)

	holder := custom_turing_token_con.GetHolder(args.Address)

	jsonString, _ := json.Marshal(holder)

	*result = string(jsonString)

	return nil
}
