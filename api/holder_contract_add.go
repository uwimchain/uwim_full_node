package api

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
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

type HolderContractAddArgs struct {
	Mnemonic         string  `json:"mnemonic"`
	DepositorAddress string  `json:"depositor_address"`
	RecipientAddress string  `json:"recipient_address"`
	TokenLabel       string  `json:"token_label"`
	Amount           float64 `json:"amount"`
	GetBlockHeight   int64   `json:"get_block_height"`
}

func (api *Api) HolderContractAdd(args *HolderContractAddArgs, result *string) error {
	args.Mnemonic, args.DepositorAddress, args.RecipientAddress, args.TokenLabel = apparel.TrimToLower(args.Mnemonic),
		apparel.TrimToLower(args.DepositorAddress), apparel.TrimToLower(args.RecipientAddress),
		apparel.TrimToLower(args.TokenLabel)

	validateHolderAdd := validateHolderAdd(args.Mnemonic, args.DepositorAddress, args.RecipientAddress, args.TokenLabel,
		args.Amount, args.GetBlockHeight)
	if validateHolderAdd != 0 {
		return errors.New(strconv.FormatInt(validateHolderAdd, 10))
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	commentData := make(map[string]interface{})
	commentData["recipient_address"] = args.RecipientAddress
	commentData["token_label"] = args.TokenLabel
	commentData["get_block_height"] = args.GetBlockHeight

	commentDataJson, err := json.Marshal(commentData)
	if err != nil {
		log.Println("Send Transaction error:", err)
		return nil
	}

	comment := deep_actions.Comment{
		Title: "holder_contract_add_transaction",
		Data:  commentDataJson,
	}

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))

	tx := deep_actions.Tx{
		Type:       1,
		Nonce:      apparel.GetNonce(timestamp),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       args.DepositorAddress,
		To:         config.HolderScAddress,
		Amount:     args.Amount,
		TokenLabel: config.BaseToken,
		Timestamp:  timestamp,
		Tax:        config.HolderAddCost,
		Signature:  nil,
		Comment:    comment,
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

func validateHolderAdd(mnemonic, depositorAddress, recipientAddress, tokenLabel string, amount float64,
	getBlockHeight int64) int64 {
	validateMnemonic := validateMnemonic(mnemonic, depositorAddress)
	if validateMnemonic != 0 {
		return validateMnemonic
	}

	validateAdd := holder_con.ValidateAdd(depositorAddress, recipientAddress, tokenLabel, amount, getBlockHeight)
	if validateAdd != 0 {
		return validateAdd
	}

	validateBalanceForCost := validateBalance(depositorAddress, config.HolderAddCost, config.BaseToken, false)
	if validateBalanceForCost != 0 {
		return validateBalanceForCost
	}

	validateBalanceForSend := validateBalance(depositorAddress, amount, tokenLabel, false)
	if validateBalanceForSend != 0 {
		return validateBalanceForSend
	}

	validateTxInMemory := validateTxInMemory(depositorAddress, config.HolderScAddress,
		"holder_contract_add_transaction", 1)
	if validateTxInMemory != 0 {
		return validateTxInMemory
	}

	return 0
}
