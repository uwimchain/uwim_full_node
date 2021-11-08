package api

import (
	"node/storage/deep_actions"
)

type FindBlockArgs struct {
	Height int64 `json:"height"`
}

func (api *Api) FindBlock(args *FindBlockArgs, result *string) error {
	chainJson := deep_actions.GetChainJson(args.Height)
	*result = chainJson
	return nil
}
