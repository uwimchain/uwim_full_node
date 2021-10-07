package custom_turing_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"strconv"
)

type DelegateArgs struct {
	UwAddress   string  `json:"uw_address"`
	Amount      float64 `json:"amount"`
	BlockHeight int64   `json:"block_height"`
	TxHash      string  `json:"tx_hash"`
}

func NewDelegateArgs(uwAddress string, amount float64, blockHeight int64, txHash string) *DelegateArgs {
	return &DelegateArgs{UwAddress: uwAddress, Amount: amount, BlockHeight: blockHeight, TxHash: txHash}
}

func Delegate(args *DelegateArgs) error {
	if err := delegate(args.UwAddress, args.TxHash, args.Amount, args.BlockHeight); err != nil {
		if err := contracts.RefundTransaction(ScAddress, args.UwAddress, args.Amount, TokenLabel); err != nil {
			return errors.New(fmt.Sprintf("delegate refund error 1: %v", err))
		}

		return errors.New(fmt.Sprintf("delegate error 1: %v", err))
	}

	return nil
}

func delegate(uwAddress, txHash string, amount float64, blockHeight int64) error {
	holderJson := HolderDB.Get(uwAddress).Value

	timestamp := apparel.TimestampUnix()
	timestampD := strconv.FormatInt(timestamp, 10)

	holder := Holder{}

	if holderJson == "" {
		holder.Address = uwAddress
		holder.Amount = amount
		holder.UpdateTime = timestampD
	} else {
		_ = json.Unmarshal([]byte(holderJson), &holder)

		holder.Amount += amount
		holder.UpdateTime = timestampD
	}

	err := contracts.AddEvent(ScAddress, *contracts.NewEvent("Delegate", timestamp, blockHeight, txHash, uwAddress, ""), EventDB, ConfigDB)
	if err != nil {
		return err
	}

	jsonHolder, err := json.Marshal(holder)
	if err != nil {
		return err
	}
	HolderDB.Put(uwAddress, string(jsonHolder))

	return nil
}
