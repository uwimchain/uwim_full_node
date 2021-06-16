package api

import (
	"node/storage"
)

type GetTokensArgs struct {
}

func (api *Api) GetTokens(args *GetTokensArgs, result *string) error {

	*result = storage.GetTokens()

	return nil
}
