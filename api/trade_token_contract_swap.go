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
type TradeTokenContractSwapArgs struct {
	Mnemonic       string  `json:"mnemonic"`
	Amount         float64 `json:"amount"`
	TokenLabel     string  `json:"token_label"`
	SwapTokenLabel string  `json:"swap_token_label"`
}

func (api *Api) TradeTokenContractSwap(args *TradeTokenContractSwapArgs, result *string) error {
	args.Mnemonic, args.TokenLabel, args.SwapTokenLabel = apparel.TrimToLower(args.Mnemonic), apparel.TrimToLower(args.TokenLabel), apparel.TrimToLower(args.SwapTokenLabel)

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

	check := validateSwap(args.Mnemonic, args.TokenLabel, args.SwapTokenLabel, uwAddress, scAddress, args.Amount)
	if check != 0 {
		return errors.New(strconv.FormatInt(check, 10))
	}

	timestampD := strconv.FormatInt(apparel.TimestampUnix(), 10)

	var tax float64 = 0.01
	if args.TokenLabel == config.BaseToken {
		tax = apparel.CalcTax(args.Amount)
	}

	comment := deep_actions.Comment{
		Title: "trade_token_contract_swap_transaction",
		Data:  nil,
	}

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))

	tx := deep_actions.Tx{
		Type:       1,
		Nonce:      apparel.GetNonce(timestampD),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       uwAddress,
		To:         scAddress,
		Amount:     args.Amount,
		TokenLabel: args.SwapTokenLabel,
		Timestamp:  timestampD,
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

func validateSwap(mnemonic, tokenLabel, swapTokenLabel, uwAddress, scAddress string, amount float64) int64 {
	if mnemonic == "" {
		return 1
	}

	if !storage.CheckToken(tokenLabel) {
		return 11
	}

	if !storage.CheckToken(swapTokenLabel) {
		return 11
	}

	validateMnemonic := validateMnemonic(mnemonic, uwAddress)
	if validateMnemonic != 0 {
		return validateMnemonic
	}

	validateSwap := trade_token_con.ValidateSwap(trade_token_con.NewTradeArgsForValidate(scAddress, uwAddress, amount, tokenLabel))
	if validateSwap != 0 {
		return validateSwap
	}

	validateBalance := validateBalance(uwAddress, amount, swapTokenLabel, true)
	if validateBalance != 0 {
		return validateBalance
	}

	validateTxInMemory := validateTxInMemory(uwAddress, scAddress, "trade_token_contract_swap_transaction", 1)
	if validateTxInMemory != 0 {
		return validateTxInMemory
	}

	return 0
}
