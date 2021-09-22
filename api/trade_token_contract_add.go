package api

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts/trade_token_con"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/metrics"
	"node/storage"
	"node/storage/deep_actions"
	"node/websocket/sender"
	"strconv"
)

// SendTransactions method arguments
type TradeTokenContractAddArgs struct {
	Mnemonic      string  `json:"mnemonic"`
	Amount        float64 `json:"amount"`
	TokenLabel    string  `json:"token_label"`
	AddTokenLabel string  `json:"add_token_label"`
}

func (api *Api) TradeTokenContractAdd(args *TradeTokenContractAddArgs, result *string) error {
	args.Mnemonic, args.TokenLabel, args.AddTokenLabel = apparel.TrimToLower(args.Mnemonic), apparel.TrimToLower(args.TokenLabel), apparel.TrimToLower(args.AddTokenLabel)

	uwAddress := crypt.AddressFromMnemonic(args.Mnemonic)
	token := storage.GetToken(args.TokenLabel)
	if token.Id == 0 {
		return errors.New(strconv.Itoa(10))
	}

	publicKey, err := crypt.PublicKeyFromAddress(token.Proposer)
	if err != nil {
		return errors.New(strconv.Itoa(11))
	}

	scAddress := crypt.AddressFromPublicKey(metrics.SmartContractPrefix, publicKey)

	check := validateTradeAdd(args.Mnemonic, args.TokenLabel, args.AddTokenLabel, uwAddress, scAddress, args.Amount)
	if check != 0 {
		return errors.New(strconv.FormatInt(check, 10))
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	args.Amount, _ = apparel.Round(args.Amount)
	var tax float64 = 0.01
	if args.TokenLabel == config.BaseToken {
		tax = apparel.CalcTax(args.Amount)
	}

	comment := deep_actions.Comment{
		Title: "trade_token_contract_add_transaction",
		Data:  nil,
	}

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))

	tx := deep_actions.Tx{
		Type:       1,
		Nonce:      apparel.GetNonce(timestamp),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       uwAddress,
		To:         scAddress,
		Amount:     args.Amount,
		TokenLabel: args.AddTokenLabel,
		Timestamp:  timestamp,
		Tax:        tax,
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

func validateTradeAdd(mnemonic, tokenLabel, addTokenLabel, uwAddress, scAddress string, amount float64) int64 {
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

	validateAdd := trade_token_con.ValidateAdd(trade_token_con.NewTradeArgsForValidate(scAddress, uwAddress, amount, tokenLabel))
	if validateAdd != 0 {
		return validateAdd
	}

	validateBalance := validateBalance(uwAddress, amount, addTokenLabel, true)
	if validateBalance != 0 {
		return validateBalance
	}

	validateTxInMemory := validateTxInMemory(uwAddress, scAddress, "trade_token_contract_add_transaction", 2)
	if validateTxInMemory != 0 {
		return validateTxInMemory
	}

	return 0
}
