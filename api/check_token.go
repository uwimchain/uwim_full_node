package api

import (
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/storage"
	"strconv"
)

type CheckTokenArgs struct {
	Label string `json:"label"`
}

func (api *Api) CheckToken(args *CheckTokenArgs, result *string) error {
	args.Label = apparel.TrimToLower(args.Label)
	labelLen := len(args.Label)
	if labelLen < 3 || labelLen > 5 {
		return errors.New(strconv.Itoa(0))
	}

	if storage.CheckToken(args.Label) {
		*result = strconv.Itoa(1)
		return nil
	}

	return errors.New(strconv.Itoa(0))
}
