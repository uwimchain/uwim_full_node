package api

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts/trade_token_con"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/storage"
	"node/storage/deep_actions"
	"node/websocket/sender"
	"strconv"
)

type FillConfigArgs struct {
	Mnemonic   string  `json:"mnemonic"`
	Commission float64 `json:"commission"`
}

func (api *Api) TradeTokenContractFillConfig(args *FillConfigArgs, result *string) error {
	args.Mnemonic = apparel.TrimToLower(args.Mnemonic)
	uwAddress := crypt.AddressFromMnemonic(args.Mnemonic)
	scAddress := crypt.ScAddressFromMnemonic(args.Mnemonic)
	check := validateFillConfig(args.Mnemonic, scAddress, uwAddress, args.Commission)
	if check != 0 {
		return errors.New(strconv.FormatInt(check, 10))
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)
	jsonString, _ := json.Marshal(trade_token_con.TradeConfig{
		Commission: args.Commission,
	})
	comment := deep_actions.Comment{
		Title: "trade_token_contract_fill_config_transaction",
		Data:  jsonString,
	}

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))

	tx := deep_actions.Tx{
		Type:       1,
		Nonce:      apparel.GetNonce(timestamp),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       uwAddress,
		To:         scAddress,
		Amount:     1,
		TokenLabel: config.BaseToken,
		Timestamp:  timestamp,
		Tax:        0,
		Signature:  nil,
		Comment:    comment,
	}

	jsonString, _ = json.Marshal(deep_actions.Tx{
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

func validateFillConfig(mnemonic, scAddress, uwAddress string, commission float64) int64 {
	if mnemonic == "" {
		return 1
	}

	validateMnemonic := validateMnemonic(mnemonic, uwAddress)
	if validateMnemonic != 0 {
		return validateMnemonic
	}

	validateFillConfig := trade_token_con.ValidateFillConfig(trade_token_con.NewFillConfigArgs(scAddress, commission))
	if validateFillConfig != 0 {
		return validateFillConfig
	}

	validateBalance := validateBalance(uwAddress, 1, config.BaseToken, false)
	if validateBalance != 0 {
		return validateBalance
	}

	validateTxInMemory := validateTxInMemory(uwAddress, scAddress, "trade_token_contract_fill_config_transaction", 1)
	if validateTxInMemory != 0 {
		return validateTxInMemory
	}

	return 0
}
