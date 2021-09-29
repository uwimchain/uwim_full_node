package api

import (
	"errors"
	"node/apparel"
	"node/crypt"
	"node/metrics"
	"strings"
)

// GetAddressFromAddress method arguments
type GetAddressFromAddressArgs struct {
	Address string `json:"address"`
	Prefix  string `json:"prefix"`
}

func (api *Api) GetAddressFromAddress(args *GetAddressFromAddressArgs, result *string) error {
	address := apparel.TrimToLower(args.Address)
	prefix := apparel.TrimToLower(args.Prefix)

	if len(prefix) != 2 {
		return errors.New("1")
	}

	if address == "" || len(address) != 61 {
		return errors.New("2")
	}

	if !strings.HasPrefix(address, metrics.AddressPrefix) && !strings.HasPrefix(address, metrics.NodePrefix) && !strings.HasPrefix(address, metrics.SmartContractPrefix) {
		return errors.New("2")
	}

	switch prefix {
	case "uw":
		if strings.HasPrefix(address, "uw") {
			*result = address
		} else {
			publicKey, err := crypt.PublicKeyFromAddress(address)
			if err != nil {
				return errors.New("3")
			}

			*result = crypt.AddressFromPublicKey(metrics.AddressPrefix, publicKey)
		}
	case "nd":
		if strings.HasPrefix(address, "nd") {
			*result = address
		} else {
			publicKey, err := crypt.PublicKeyFromAddress(address)
			if err != nil {
				return errors.New("3")
			}

			*result = crypt.AddressFromPublicKey(metrics.NodePrefix, publicKey)
		}
	case "sc":
		if strings.HasPrefix(address, "sc") {
			*result = address
		} else {
			publicKey, err := crypt.PublicKeyFromAddress(address)
			if err != nil {
				return errors.New("3")
			}

			*result = crypt.AddressFromPublicKey(metrics.SmartContractPrefix, publicKey)
		}
	}

	return nil
}
