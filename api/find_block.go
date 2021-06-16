package api

import (
	"node/storage"
)

// FindBlock method arguments
type FindBlockArgs struct {
	Height int64 `json:"height"`
}

func (api *Api) FindBlock(args *FindBlockArgs, result *string) error {
	jsonString := storage.GetBLockForHeight(args.Height)
	*result = jsonString
	return nil
}
