package my_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
)

type ConfirmationArgs struct {
	ScAddress   string `json:"sc_address"`
	UwAddress   string `json:"uw_address"`
	BlockHeight int64  `json:"block_height"`
	TxHash      string `json:"tx_hash"`
}

func NewConfirmationArgs(scAddress string, uwAddress string, blockHeight int64, txHash string) (*ConfirmationArgs, error) {
	return &ConfirmationArgs{ScAddress: scAddress, UwAddress: uwAddress, BlockHeight: blockHeight, TxHash: txHash}, nil
}

func Confirmation(args *ConfirmationArgs) error {
	err := confirmation(args.ScAddress, args.UwAddress, args.TxHash, args.BlockHeight)
	if err != nil {
		return errors.New(fmt.Sprintf("error 1: confirmation %v", err))
	}

	return nil
}

func confirmation(scAddress, uwAddress, txHash string, blockHeight int64) error {
	timestamp := apparel.TimestampUnix()

	var scAddressPool []Pool
	scAddressPoolJson := PoolDB.Get(scAddress).Value

	if scAddressPoolJson != "" {
		err := json.Unmarshal([]byte(scAddressPoolJson), &scAddressPool)
		if err != nil {
			return errors.New(fmt.Sprintf("error 1: %v", err))
		}
	}

	if scAddressPool != nil {
		for _, i := range scAddressPool {
			if i.Address == uwAddress {
				return errors.New("error 2: this address already exists of this token pool")
			}
		}
	}

	scAddressPool = append(scAddressPool, Pool{
		Address: uwAddress,
	})

	jsonString, err := json.Marshal(scAddressPool)
	if err != nil {
		return errors.New(fmt.Sprintf("error 3: %v", err))
	}

	err = contracts.AddEvent(scAddress, *contracts.NewEvent("Confirmation", timestamp, blockHeight, txHash, uwAddress, nil), EventDB, ConfigDB)
	if err != nil {
		return errors.New(fmt.Sprintf("erorr 4: add event %v", err))
	}

	PoolDB.Put(scAddress, string(jsonString))

	return nil
}
