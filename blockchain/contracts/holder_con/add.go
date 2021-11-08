package holder_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
)

type AddArgs struct {
	DepositorAddress string  `json:"depositor_address"`
	RecipientAddress string  `json:"recipient_address"`
	Amount           float64 `json:"amount"`
	TokenLabel       string  `json:"token_label"`
	GetBlockHeight   int64   `json:"get_block_height"`
	TxHash           string  `json:"tx_hash"`
	BlockHeight      int64   `json:"block_height"`
}

func NewAddArgs(depositorAddress string, recipientAddress string, amount float64, tokenLabel string, getBlockHeight int64, txHash string, blockHeight int64) (*AddArgs, error) {
	//amount, _ = apparel.Round(amount)
	amount = apparel.Round(amount)
	return &AddArgs{DepositorAddress: depositorAddress, RecipientAddress: recipientAddress, Amount: amount, TokenLabel: tokenLabel, GetBlockHeight: getBlockHeight, TxHash: txHash, BlockHeight: blockHeight}, nil
}

func (args *AddArgs)Add() error {
	err := add(args.DepositorAddress, args.RecipientAddress, args.TokenLabel, args.TxHash, args.Amount, args.GetBlockHeight, args.BlockHeight)
	if err != nil {
		refundError := contracts.RefundTransaction(config.HolderScAddress, args.DepositorAddress, args.Amount, args.TokenLabel)
		if refundError != nil {
			log.Println(fmt.Sprintf("Refund transaction %v", refundError))
		}
		return errors.New(fmt.Sprintf("error 1: addHolder %v", err))
	}

	return nil
}

func add(depositorAddress, recipientAddress, tokenLabel, txHash string, amount float64, getBlockHeight, blockHeight int64) error {
	if !crypt.IsAddressUw(depositorAddress) && !crypt.IsAddressSmartContract(depositorAddress) && !crypt.IsAddressNode(depositorAddress) {
		return errors.New("error 1: incorrect depositor address")
	}

	if recipientAddress == "" {
		recipientAddress = depositorAddress
	} else if !crypt.IsAddressUw(recipientAddress) && !crypt.IsAddressSmartContract(recipientAddress) && !crypt.IsAddressNode(recipientAddress) {
		return errors.New("error 2: incorrect recipient address")
	}

	if tokenLabel != config.BaseToken {
		return errors.New("error 3: token label is not a \"uwm\"")
	}

	if amount <= 0 {
		return errors.New("error 4: zero or negative amount")
	}

	if getBlockHeight <= 0 {
		return errors.New("error 5: incorrect get block height")
	}

	var holder []Holder
	holderJson := HolderDB.Get(recipientAddress).Value
	if holderJson != "" {
		err := json.Unmarshal([]byte(holderJson), &holder)
		if err != nil {
			return errors.New(fmt.Sprintf("erorr 6: %v", err))
		}
	}

	holder = append(holder, Holder{
		DepositorAddress: depositorAddress,
		RecipientAddress: recipientAddress,
		Amount:           amount,
		TokenLabel:       tokenLabel,
		GetBlockHeight:   getBlockHeight,
	})

	jsonHolder, err := json.Marshal(holder)
	if err != nil {
		return errors.New(fmt.Sprintf("error 7: %v", err))
	}

	timestamp := apparel.TimestampUnix()
	err = contracts.AddEvent(config.HolderScAddress, *contracts.NewEvent("Add", timestamp, config.BlockHeight, txHash, depositorAddress, newEventAddTypeData(depositorAddress, recipientAddress, tokenLabel, amount)), EventDB, ConfigDB)
	if err != nil {
		return errors.New(fmt.Sprintf("error 8: %v", err))
	}

	HolderDB.Put(recipientAddress, string(jsonHolder))

	return nil
}
