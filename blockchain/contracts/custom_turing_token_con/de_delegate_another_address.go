package custom_turing_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
	"node/memory"
	"strconv"
)

type DeDelegateAnotherAddressArgs struct {
	Sender      string  `json:"sender"`
	Recipient   string  `json:"recipient"`
	Amount      float64 `json:"amount"`
	BlockHeight int64   `json:"block_height"`
	TxHash      string  `json:"tx_hash"`
}

func NewDeDelegateAnotherAddressArgs(sender string, recipient string, amount float64, blockHeight int64, txHash string) *DeDelegateAnotherAddressArgs {
	return &DeDelegateAnotherAddressArgs{Sender: sender, Recipient: recipient, Amount: amount, BlockHeight: blockHeight, TxHash: txHash}
}

func DeDelegateAnotherAddress(args *DeDelegateAnotherAddressArgs) error {
	if err := deDelegateAnotherAddress(args.Sender, args.Recipient, args.TxHash, args.Amount, args.BlockHeight); err != nil {
		if err := contracts.RefundTransaction(ScAddress, args.Sender, args.Amount, TokenLabel); err != nil {
			return errors.New(fmt.Sprintf("de-delegate refund error 1: %v", err))
		}

		return errors.New(fmt.Sprintf("de-delegate error 1: %v", err))
	}

	return nil
}

func deDelegateAnotherAddress(senderAddress, recipientAddress, txHash string, amount float64, blockHeight int64) error {
	senderJson := HolderDB.Get(senderAddress).Value
	recipientJson := HolderDB.Get(recipientAddress).Value

	timestamp := apparel.TimestampUnix()
	timestampD := strconv.FormatInt(timestamp, 10)

	sender := Holder{}
	recipient := Holder{}

	if senderJson == "" {
		return errors.New(fmt.Sprintf("De-delegate another address error 1: %s does not exist in smart-contract holders list", senderAddress))
	}

	_ = json.Unmarshal([]byte(senderJson), &sender)

	amount1 := amount - (amount * (ScAddressPercent / 100))
	amount2 := amount - amount1

	if recipientJson != "" {
		_ = json.Unmarshal([]byte(recipientJson), &recipient)

		recipient.Amount += amount2
		recipient.UpdateTime = timestampD
	} else {
		recipient.Address = recipientAddress
		recipient.Amount = amount2
		recipient.UpdateTime = timestampD
	}

	sender.Amount -= amount
	sender.UpdateTime = timestampD

	txCommentSign, _ := json.Marshal(contracts.NewBuyTokenSign(
		config.NodeNdAddress,
	))

	tx := contracts.NewTx(
		5,
		apparel.GetNonce(timestampD),
		"",
		blockHeight,
		ScAddress,
		UwAddress,
		amount2,
		TokenLabel,
		timestampD,
		0,
		nil,
		*contracts.NewComment("default_transaction", txCommentSign))

	jsonString, _ := json.Marshal(contracts.Tx{
		Type:       tx.Type,
		Nonce:      tx.Nonce,
		From:       tx.From,
		To:         tx.To,
		Amount:     tx.Amount,
		TokenLabel: tx.TokenLabel,
		Comment:    tx.Comment,
	})
	tx.Signature = crypt.SignMessageWithSecretKey(config.NodeSecretKey, jsonString)

	jsonString, _ = json.Marshal(tx)
	tx.HashTx = crypt.GetHash(jsonString)

	err := contracts.AddEvent(ScAddress, *contracts.NewEvent("De-delegate another address", timestamp, blockHeight, txHash, senderAddress, ""), EventDB, ConfigDB)
	if err != nil {
		return err
	}

	jsonSender, _ := json.Marshal(sender)
	HolderDB.Put(senderAddress, string(jsonSender))

	jsonRecipient, _ := json.Marshal(recipient)
	HolderDB.Put(recipientAddress, string(jsonRecipient))

	if memory.IsNodeProposer() {
		contracts.SendTx(*tx)
		*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *tx)
	}
	return nil
}