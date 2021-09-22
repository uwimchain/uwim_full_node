package api

import (
	"node/api/api_error"
	"node/apparel"
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
type SendTransactionTestArgs struct {
	TransactionRaw string `json:"transaction_raw"`
}

func (api *Api) SendTransactionRaw(args *SendTransactionTestArgs, result *string) error {
	commentTitle, senderAddress, recipientAddress, amount, tokenScAddress, signature := crypt.DecodeTransactionRaw(args.TransactionRaw)

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	tax := 0.01
	tokenLabel := storage.GetAddressToken(crypt.AddressFromAnotherAddress(metrics.AddressPrefix, tokenScAddress)).Label
	if tokenLabel != config.BaseToken {
		tax = apparel.CalcTax(amount)
	}

	comment := deep_actions.Comment{
		Title: commentTitle,
		Data:  nil,
	}

	tx := deep_actions.Tx{
		Type:       1,
		Nonce:      apparel.GetNonce(timestamp),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       senderAddress,
		To:         recipientAddress,
		Amount:     amount,
		TokenLabel: tokenLabel,
		Timestamp:  timestamp,
		Tax:        tax,
		Signature:  signature,
		Comment:    comment,
	}

	if validateTransaction := validateTransactionTest(senderAddress, recipientAddress, tokenLabel, amount); validateTransaction != nil {
		return api_error.AddError(validateTransaction)
	}

	sender.SendTx(tx)

	if memory.IsValidator() {
		storage.TransactionsMemory = append(storage.TransactionsMemory, tx)
	}

	*result = "Transaction send"
	return nil
}

func validateTransactionTest(senderAddress, recipientAddress, tokenLabel string, amount float64) *api_error.ApiError {
	if senderAddress == "" {
		return api_error.NewApiError(1, "Sender address is null")
	}

	if recipientAddress == "" {
		return api_error.NewApiError(1, "Recipient address is null")
	}

	if senderAddress == recipientAddress {
		return api_error.NewApiError(2, "Recipient address is a sender address")
	}

	if amount < 0 {
		return api_error.NewApiError(3, "Zero amount")
	}

	if tokenLabel == "" {
		return api_error.NewApiError(4, "Token label is null or this token does not exists")
	}

	if validateBalance := validateBalanceTest(senderAddress, amount, tokenLabel, false); validateBalance != nil {
		return validateBalance
	}

	if validateTxInMemory := validateTxInMemoryTest(senderAddress, recipientAddress, "default_transaction", 1); validateTxInMemory != nil {
		return validateTxInMemory
	}

	return nil
}
