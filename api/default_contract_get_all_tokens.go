package api

import (
	"encoding/json"
	"node/blockchain/contracts/default_con"
)

type DefaultContractGetAllTokensArgs struct {}

func (api *Api) DefaultContractGetAllTokens(args *DefaultContractGetAllTokensArgs, result *string) error {

	token := default_con.GetAllTokens()

	jsonString, _ := json.Marshal(token)
	*result = string(jsonString)

	return nil
}
