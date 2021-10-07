package api

import (
	"encoding/json"
	"node/blockchain/contracts/default_con"
)

type DefaultContractGetTokenForParentLabelArgs struct {
	Label string
}

func (api *Api) DefaultContractGetTokenForParentLabel(args *DefaultContractGetTokenForParentLabelArgs, result *string) error {
	token := default_con.GetNftTokenElsForParentLabel(args.Label)

	jsonString, _ := json.Marshal(token)
	*result = string(jsonString)

	return nil
}
