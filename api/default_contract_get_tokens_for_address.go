package api

import (
	"encoding/json"
	"node/blockchain/contracts/default_con"
)

type DefaultContractGetTokenForAddressArgs struct {
	Address string
}

func (api *Api) DefaultContractGetTokenForAddress(args *DefaultContractGetTokenForAddressArgs, result *string) error {
	token := default_con.GetNftTokenElsForAddress(args.Address)

	jsonString, _ := json.Marshal(token)
	*result = string(jsonString)

	return nil
}
