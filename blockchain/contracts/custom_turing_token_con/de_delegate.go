package custom_turing_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"strconv"
)

type DeDelegateArgs struct {
	UwAddress   string  `json:"uw_address"`
	Amount      float64 `json:"amount"`
	BlockHeight int64   `json:"block_height"`
	TxHash      string  `json:"tx_hash"`
}

func NewDeDelegateArgs(uwAddress string, amount float64, blockHeight int64, txHash string) *DeDelegateArgs {
	return &DeDelegateArgs{UwAddress: uwAddress, Amount: amount, BlockHeight: blockHeight, TxHash: txHash}
}

func DeDelegate(args *DeDelegateArgs) error {
	if err := deDelegate(args.UwAddress, args.TxHash, args.Amount, args.BlockHeight); err != nil {
		if err := contracts.RefundTransaction(ScAddress, args.UwAddress, args.Amount, TokenLabel); err != nil {
			return errors.New(fmt.Sprintf("de-delegate refund error 1: %v", err))
		}

		return errors.New(fmt.Sprintf("de-delegate error 1: %v", err))
	}

	return nil
}

func deDelegate(uwAddress, txHash string, amount float64, blockHeight int64) error {
	holderJson := HolderDB.Get(uwAddress).Value

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	holder := Holder{}

	if holderJson == "" {
		return errors.New(fmt.Sprintf("De-delegate error 1: %s does not exist in smart-contract holders list", uwAddress))
	}

	_ = json.Unmarshal([]byte(holderJson), &holder)

	holder.Amount -= amount
	holder.UpdateTime = timestamp

	var txs []contracts.Tx

	if uwAddress == UwAddress {
		txs = append(txs, contracts.Tx{
			To:         uwAddress,
			Amount:     amount,
			TokenLabel: TokenLabel,
		})
	} else {
		amount1 := amount - (amount * (ScAddressPercent / 100))
		amount2 := amount - amount1

		txs = append(txs, contracts.Tx{
			To:         uwAddress,
			Amount:     amount1,
			TokenLabel: TokenLabel,
		}, contracts.Tx{
			To:         UwAddress,
			Amount:     amount2,
			TokenLabel: TokenLabel,
		})
	}

	if txs == nil {
		return errors.New("De-delegate error 2: empty transactions list")
	}

	err := contracts.AddEvent(ScAddress, *contracts.NewEvent("De-delegate", timestamp, blockHeight, txHash, uwAddress, ""), EventDB, ConfigDB)
	if err != nil {
		return err
	}

	jsonHolder, err := json.Marshal(holder)
	if err != nil {
		return err
	}
	HolderDB.Put(uwAddress, string(jsonHolder))

	for _, i := range txs {
		contracts.SendNewScTx(config.DelegateScAddress, i.To, i.Amount, i.TokenLabel, "default_transaction")
	}

	return nil
}
