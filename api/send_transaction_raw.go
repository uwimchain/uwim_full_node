package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"node/api/api_error"
	"node/apparel"
	"node/blockchain/contracts/delegate_con"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/storage"
	"node/storage/deep_actions"
	"node/websocket/sender"
	"strconv"
)

type SendTransactionRawArgs struct {
	TransactionRaw string `json:"transaction_raw"`
}

var taxNullTrueTxCommentTitles = []string{
	"undelegate_contract_transaction",
}

var taxNullFalseTxCommentTitles = []string{
	"default_transaction",
	"delegate_contract_transaction",
}

var amountNullTrueTxCommentTitles = []string{
	"undelegate_contract_transaction",
}

var amountNullFalseTxCommentTitles = []string{
	"default_transaction",
	"delegate_contract_transaction",
}

func (api *Api) SendTransactionRaw(args *SendTransactionRawArgs, result *string) error {
	txRaw, err := crypt.DecodeTransactionRaw(args.TransactionRaw)
	if err != nil {
		return err
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	tax := 0.01

	if apparel.InArray(taxNullTrueTxCommentTitles, txRaw.Comment.Title) {
		tax = 0
	} else if txRaw.TokenLabel == config.BaseToken {
		tax = apparel.CalcTax(txRaw.Amount)
	}

	comment := deep_actions.Comment{
		Title: txRaw.Comment.Title,
		Data:  []byte(txRaw.Comment.Data),
	}

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

	jsonStringForValidateSignature, _ := json.Marshal(deep_actions.Tx{
		Type:       tx.Type,
		Nonce:      tx.Nonce,
		From:       tx.From,
		To:         tx.To,
		Amount:     tx.Amount,
		TokenLabel: tx.TokenLabel,
		Comment:    tx.Comment,
	})

	publicKey, err := crypt.PublicKeyFromAddress(tx.From)
	if err != nil {
		return api_error.AddError(api_error.NewApiError(11, "incorrect sender address"))
	}

	if publicKey == nil {
		return api_error.AddError(api_error.NewApiError(12, fmt.Sprintf("signature verify error %s, incorrect public key, public key length = %d", tx.Comment.Title, len(publicKey))))
	}

	if !crypt.VerifySign(publicKey, jsonStringForValidateSignature, tx.Signature) {
		return api_error.AddError(api_error.NewApiError(13, fmt.Sprintf("signature verify error %s, %v", tx.Comment.Title, tx.Signature)))
	}

	if validateTransaction := validateTransactionRaw(tx.From, tx.To, tx.TokenLabel, tx.Comment.Title,
		txRaw.Comment.Data, tx.Amount, tx.Type); validateTransaction != nil {
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

func validateTransactionRaw(senderAddress, recipientAddress, tokenLabel, txCommentTitle, txCommentDataJson string,
	amount float64, txType int64) *api_error.ApiError {

	if senderAddress == "" {
		return api_error.NewApiError(1, "Sender address is null")
	}

	if recipientAddress == "" {
		return api_error.NewApiError(2, "Recipient address is null")
	}

	if senderAddress == recipientAddress {
		return api_error.NewApiError(3, "Recipient address is a sender address")
	}

	if crypt.IsAddressSmartContract(senderAddress) {
		return api_error.NewApiError(14, "Sender address is a smart-contract address")
	}

	if amount <= 0 && !apparel.InArray(amountNullTrueTxCommentTitles, txCommentTitle) {
		return api_error.NewApiError(4, "Zero or negative amount")
	}

	if tokenLabel == "" {
		return api_error.NewApiError(5, "Empty token label")
	}

	if !deep_actions.CheckToken(tokenLabel) {
		return api_error.NewApiError(6, fmt.Sprintf("Token with label \"%s\" does not exist", tokenLabel))
	}

	var taxNull bool

	if apparel.InArray(taxNullTrueTxCommentTitles, txCommentTitle) {
		taxNull = true
	} else if apparel.InArray(taxNullFalseTxCommentTitles, txCommentTitle) {
		taxNull = false
	} else {
		return api_error.NewApiError(7, "Incorrect transaction comment title")
	}

	txCommentData := make(map[string]interface{})
	if txCommentDataJson != "" {
		_ = json.Unmarshal([]byte(txCommentDataJson), &txCommentData)
	}

	switch txType {
	case 1:
		switch txCommentTitle {
		case "default_transaction":
			validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "default_transaction", 1)
			if validateTxInMemory != 0 {
				return api_error.NewApiError(10, "You have already sent a transaction of this type")
			}

		case "delegate_contract_transaction":
			if err := delegate_con.ValidateDelegate(recipientAddress, senderAddress); err != 0 {
				return api_error.NewApiError(err, "Delegate contract delegate error")
			}
			break
		case "undelegate_contract_transaction":
			if err := delegate_con.ValidateUnDelegate(recipientAddress, senderAddress,
				apparel.ConvertInterfaceToFloat64(txCommentData["undelegate_amount"])); err != 0 {
				return api_error.NewApiError(err, "Delegate contract undelegate error")
			}
			break
		default:
			return api_error.NewApiError(7, "Incorrect transaction comment title")
		}
		break
	default:
		return api_error.NewApiError(8, "Incorrect transaction type")
	}

	var amountNull bool

	if apparel.InArray(amountNullTrueTxCommentTitles, txCommentTitle) {
		amountNull = true
	} else if apparel.InArray(amountNullFalseTxCommentTitles, txCommentTitle) {
		amountNull = false
	} else {
		return api_error.NewApiError(7, "Incorrect transaction comment title")
	}

	if validateBalance := validateBalance(senderAddress, amount, tokenLabel, taxNull, amountNull);
		validateBalance != nil {
		return api_error.NewApiError(9, "Low balance")
	}

	return nil
}