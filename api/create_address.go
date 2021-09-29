package api

import (
	"bytes"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/crypt"
	"strconv"
	"strings"
)

// Balance method arguments
type CreateAddressArgs struct {
	Mnemonic string `json:"mnemonic"`
}

func (api *Api) CreateAddress(args *CreateAddressArgs, result *string) error {
	if len(bytes.Split([]byte(strings.TrimSpace(args.Mnemonic)), []byte(" "))) != 24 {
		return errors.New(strconv.Itoa(5))
	}

	*result = crypt.AddressFromMnemonic(args.Mnemonic)
	return nil
}
