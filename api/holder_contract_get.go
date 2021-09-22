package api

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts/holder_con"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/storage"
	"node/storage/deep_actions"
	"node/websocket/sender"
	"strconv"
)

type HolderContractGetArgs struct {
	Mnemonic         string `json:"mnemonic"`
	RecipientAddress string `json:"recipient_address"`
}

func (api *Api) HolderContractGet(args *HolderContractGetArgs, result *string) error {
	args.Mnemonic, args.RecipientAddress = apparel.TrimToLower(args.Mnemonic),
		apparel.TrimToLower(args.RecipientAddress)

	validateHolderGet := validateHolderGet(args.Mnemonic, args.RecipientAddress)
	if validateHolderGet != 0 {
		return errors.New(strconv.FormatInt(validateHolderGet, 10))
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))

	comment := deep_actions.Comment{
		Title: "holder_contract_get_transaction",
		Data:  nil,
	}

	tx := deep_actions.Tx{
		Type:       1,
		Nonce:      apparel.GetNonce(timestamp),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       args.RecipientAddress,
		To:         config.HolderScAddress,
		Amount:     0,
		TokenLabel: config.BaseToken,
		Timestamp:  timestamp,
		Tax:        config.HolderGetCost,
		Signature:  nil,
		Comment: comment,
	}

	jsonString, _ := json.Marshal(deep_actions.Tx{
		Type:       tx.Type,
		Nonce:      tx.Nonce,
		From:       tx.From,
		To:         tx.To,
		Amount:     tx.Amount,
		TokenLabel: tx.TokenLabel,
		Comment:    tx.Comment,
	})
	tx.Signature = crypt.SignMessageWithSecretKey(secretKey, jsonString)
	sender.SendTx(tx)

	if memory.IsValidator() {
		storage.TransactionsMemory = append(storage.TransactionsMemory, tx)
	}

	*result = "Transaction send"
	return nil
}

func validateHolderGet(mnemonic, recipientAddress string) int64 {
	validateMnemonic := validateMnemonic(mnemonic, recipientAddress)
	if validateMnemonic != 0 {
		return validateMnemonic
	}

	validateGet := holder_con.ValidateGet(recipientAddress)
	if validateGet != 0 {
		return validateGet
	}

	validateBalance := validateBalance(recipientAddress, config.HolderGetCost, config.BaseToken, false)
	if validateBalance != 0 {
		return validateBalance
	}

	validateTxInMemory := validateTxInMemory(recipientAddress, config.HolderScAddress,
		"holder_contract_get_transaction", 1)
	if validateTxInMemory != 0 {
		return validateTxInMemory
	}

	return 0
}
