package api

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts/my_token_con"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/metrics"
	"node/storage"
	"node/storage/deep_actions"
	"node/websocket/sender"
	"strconv"
)

type MyTokenContractGetPercentArgs struct {
	Mnemonic   string `json:"mnemonic"`
	TokenLabel string `json:"token_label"`
}

func (api *Api) MyTokenContractGetPercent(args *MyTokenContractGetPercentArgs, result *string) error {
	args.Mnemonic, args.TokenLabel = apparel.TrimToLower(args.Mnemonic), apparel.TrimToLower(args.TokenLabel)
	uwAddress := crypt.AddressFromMnemonic(args.Mnemonic)
	scAddressToken := storage.GetToken(args.TokenLabel)
	scAddressPublicKey, err := crypt.PublicKeyFromAddress(scAddressToken.Proposer)
	if err != nil {
		return err
	}

	scAddress := crypt.AddressFromPublicKey(metrics.SmartContractPrefix, scAddressPublicKey)

	check := validateGetPercent(args.Mnemonic, args.TokenLabel, uwAddress, scAddress)
	if check != 0 {
		return errors.New(strconv.FormatInt(check, 10))
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))

	comment := deep_actions.Comment{
		Title: "my_token_contract_get_percent_transaction",
		Data:  nil,
	}

	tx := deep_actions.Tx{
		Type:       1,
		Nonce:      apparel.GetNonce(timestamp),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       uwAddress,
		To:         scAddress,
		Amount:     storage.GetBalanceForToken(uwAddress, scAddressToken.Label).Amount,
		TokenLabel: args.TokenLabel,
		Timestamp:  timestamp,
		Tax:        0,
		Signature:  secretKey,
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

func validateGetPercent(mnemonic, tokenLabel, uwAddress, scAddress string) int64 {
	if mnemonic == "" {
		return 1
	}

	if !storage.CheckToken(tokenLabel) {
		return 11
	}

	validateMnemonic := validateMnemonic(mnemonic, uwAddress)
	if validateMnemonic != 0 {
		return validateMnemonic
	}

	validateGetPercent := my_token_con.ValidateGetPercent(scAddress, uwAddress)
	if validateGetPercent != 0 {
		return validateGetPercent
	}

	validateTxInMemory := validateTxInMemory(uwAddress, scAddress, "my_token_contract_confirmation_", 2)
	if validateTxInMemory != 0 {
		return validateTxInMemory
	}

	return 0
}
