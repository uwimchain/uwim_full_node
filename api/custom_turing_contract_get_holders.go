package api

import (
	"encoding/json"
	"node/blockchain/contracts/custom_turing_token_con"
)

type CustomTuringContractGetHoldersArgs struct{}

func (api *Api) CustomTuringContractGetHolders(args *CustomTuringContractGetHoldersArgs, result *string) error {

	holders := custom_turing_token_con.GetHolders()

	jsonString, _ := json.Marshal(holders)

	*result = string(jsonString)

	return nil
}
