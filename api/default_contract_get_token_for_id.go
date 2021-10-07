package api

import (
	"encoding/json"
	"node/blockchain/contracts/default_con"
)

type DefaultContractGetTokenForIdArgs struct {
	Id int64
}

func (api *Api) DefaultContractGetTokenForId(args *DefaultContractGetTokenForIdArgs, result *string) error {
	token := default_con.GetNftTokenElForId(args.Id)

	jsonString, _ := json.Marshal(token)
	*result = string(jsonString)

	return nil
}
