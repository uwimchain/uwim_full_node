package custom_turing_token_con

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"strconv"
)

type AddEmissionArgs struct {
	AddEmissionAmount float64 `json:"add_emission_amount"`
	TxHash            string  `json:"tx_hash"`
	BlockHeight       int64   `json:"block_height"`
}

func NewAddEmissionArgs(addEmissionAmount float64, blockHeight int64, txHash string) *AddEmissionArgs {
	return &AddEmissionArgs{AddEmissionAmount: addEmissionAmount, TxHash: txHash, BlockHeight: blockHeight}
}

func AddEmission(args *AddEmissionArgs) error {
	if err := addEmission(args.AddEmissionAmount, args.TxHash, args.BlockHeight); err != nil {
		return errors.New(fmt.Sprintf("Add emission error 1: %v", err))
	}

	return nil
}

func addEmission(addEmissionAmount float64, txHash string, blockHeight int64) error {
	if err := contracts.AddTokenEmission(TokenLabel, addEmissionAmount); err != nil {
		return err
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	if err := contracts.AddEvent(ScAddress, *contracts.NewEvent("De-delegate another address", timestamp, blockHeight, txHash, UwAddress, ""), EventDB, ConfigDB); err != nil {
		return err
	}

	address := contracts.GetAddress(ScAddress)
	address.UpdateBalance(ScAddress, addEmissionAmount, TokenLabel, timestamp, true)
	return nil
}
