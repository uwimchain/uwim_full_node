package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"node/api/api_error"
	"node/apparel"
	"node/blockchain/contracts/custom_turing_token_con"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/storage"
	"node/storage/deep_actions"
	"node/websocket/sender"
	"strconv"
)

// SendTransactionRaw method arguments
type SendTransactionRawArgs struct {
	TransactionRaw string `json:"transaction_raw"`
}

func (api *Api) SendTransactionRaw(args *SendTransactionRawArgs, result *string) error {
	txRaw, err := crypt.DecodeTransactionRaw(args.TransactionRaw)
	if err != nil {
		return err
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	tax := 0.01

	taxNullTrueTxCommentTitles := []string{
		"custom_turing_token_add_emission_transaction",
		"custom_turing_token_de_delegate_transaction",
		"custom_turing_token_de_delegate_another_address_transaction",
		"custom_turing_token_get_reward_transaction",
		"custom_turing_token_re_delegate_transaction",
	}

	if apparel.ContainsStringInStringArr(taxNullTrueTxCommentTitles, txRaw.Comment.Title) {
		tax = 0
	} else if txRaw.TokenLabel == config.BaseToken {
		tax = apparel.CalcTax(txRaw.Amount)
	}

	comment := deep_actions.Comment{
		Title: txRaw.Comment.Title,
		Data:  []byte(txRaw.Comment.Data),
	}

	log.Println(txRaw.Signature)
	signature, err := base64.StdEncoding.DecodeString(txRaw.Signature)
	if err != nil {
		return err
	}

	tx := deep_actions.Tx{
		Type:       txRaw.Type,
		Nonce:      txRaw.Nonce,
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       txRaw.From,
		To:         txRaw.To,
		Amount:     txRaw.Amount,
		TokenLabel: txRaw.TokenLabel,
		Timestamp:  timestamp,
		Tax:        tax,
		Signature:  signature,
		Comment:    comment,
	}

	if validateTransaction := validateTransactionRaw(tx.From, tx.To, tx.TokenLabel, tx.Comment.Title, txRaw.Comment.Data, tx.Amount); validateTransaction != nil {
		return api_error.AddError(validateTransaction)
	}

	jsonString, _ := json.Marshal(tx)
	tx.HashTx = crypt.GetHash(jsonString)

	sender.SendTx(tx)

	if memory.IsValidator() {
		storage.TransactionsMemory = append(storage.TransactionsMemory, tx)
	}

	*result = "Transaction send"
	return nil
}

func validateTransactionRaw(senderAddress, recipientAddress, tokenLabel, txCommentTitle, txCommentDataJson string, amount float64) *api_error.ApiError {
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

	if !storage.CheckToken(tokenLabel) {
		return api_error.NewApiError(5, fmt.Sprintf("Token with label \"%s\" does not exist", tokenLabel))
	}

	var taxNull bool

	taxNullTrueTxCommentTitles := []string{
		"custom_turing_token_add_emission_transaction",
		"custom_turing_token_de_delegate_transaction",
		"custom_turing_token_de_delegate_another_address_transaction",
		"custom_turing_token_get_reward_transaction",
		"custom_turing_token_re_delegate_transaction",
	}

	taxNullFalseTxCommentTitles := []string{
		"default_transaction",
		"custom_turing_token_delegate_transaction",
	}

	if apparel.ContainsStringInStringArr(taxNullTrueTxCommentTitles, txCommentTitle) {
		taxNull = true
	} else if apparel.ContainsStringInStringArr(taxNullFalseTxCommentTitles, txCommentTitle) {
		taxNull = false
	} else {
		return api_error.NewApiError(6, "Incorrect transaction comment title")
	}

	txCommentData := make(map[string]interface{})
	if txCommentDataJson != "" {
		_ = json.Unmarshal([]byte(txCommentDataJson), &txCommentData)
	}

	switch txCommentTitle {
	case "custom_turing_token_add_emission_transaction":
		if err := custom_turing_token_con.ValidateAddEmission(senderAddress, recipientAddress,
			apparel.ConvertInterfaceToFloat64(txCommentData["add_emission_amount"])); err != 0 {
			return api_error.NewApiError(err, "Custom turing token contract error")
		}
		break
	case "custom_turing_token_de_delegate_transaction":
		if err := custom_turing_token_con.ValidateDeDelegate(senderAddress, recipientAddress,
			apparel.ConvertInterfaceToFloat64(txCommentData["de_delegate_amount"])); err != 0 {
			return api_error.NewApiError(err, "Custom turing token contract error")
		}
		break
	case "custom_turing_token_de_delegate_another_address_transaction":
		if err := custom_turing_token_con.ValidateDeDelegateAnotherAddress(senderAddress, recipientAddress,
			apparel.ConvertInterfaceToFloat64(txCommentData["de_delegate_amount"])); err != 0 {
			return api_error.NewApiError(err, "Custom turing token contract error")
		}
		break
	case "custom_turing_token_get_reward_transaction":
		if err := custom_turing_token_con.ValidateGetReward(senderAddress, recipientAddress); err != 0 {
			return api_error.NewApiError(err, "Custom turing token contract error")
		}
		break
	case "custom_turing_token_re_delegate_transaction":
		if err := custom_turing_token_con.ValidateReDelegate(senderAddress,
			apparel.ConvertInterfaceToString(txCommentData["re_delegate_recipient_address"]), recipientAddress,
			apparel.ConvertInterfaceToFloat64(txCommentData["re_delegate_amount"])); err != 0 {
			return api_error.NewApiError(err, "Custom turing token contract error")
		}
		break
	case "custom_turing_token_delegate_transaction":
		if err := custom_turing_token_con.ValidateDelegate(recipientAddress, tokenLabel); err != 0 {
			return api_error.NewApiError(err, "Custom turing token contract error")
		}
		break
	}

	var amountNull bool

	amountNullTrueTxCommentTitles := []string{
		"custom_turing_token_add_emission_transaction",
		"custom_turing_token_de_delegate_transaction",
		"custom_turing_token_de_delegate_another_address_transaction",
		"custom_turing_token_get_reward_transaction",
		"custom_turing_token_re_delegate_transaction",
	}

	amountNullFalseTxCommentTitles := []string{
		"default_transaction",
		"custom_turing_token_delegate_transaction",
	}

	if apparel.ContainsStringInStringArr(amountNullTrueTxCommentTitles, txCommentTitle) {
		amountNull = true
	} else if apparel.ContainsStringInStringArr(amountNullFalseTxCommentTitles, txCommentTitle) {
		amountNull = false
	} else {
		return api_error.NewApiError(6, "Incorrect transaction comment title")
	}

	if validateBalance := validateBalanceTest(senderAddress, amount, tokenLabel, taxNull, amountNull); validateBalance != nil {
		return validateBalance
	}

	if validateTxInMemory := validateTxInMemoryTest(senderAddress, recipientAddress, txCommentTitle, 1); validateTxInMemory != nil {
		return validateTxInMemory
	}

	return nil
}
