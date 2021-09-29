package api

import (
	"node/storage"
	"strconv"
)

// GetTokensCount method arguments
type GetTokensCountArgs struct{}

func (api *Api) GetTokensCount(args *GetTokensCountArgs, result *string) error {

	*result = strconv.FormatInt(storage.GetTokensCount(), 10)
	return nil
}
