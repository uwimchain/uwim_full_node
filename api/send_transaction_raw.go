package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"node/api/api_error"
	"node/apparel"
	"node/blockchain/contracts/custom_turing_token_con"
	"node/blockchain/contracts/default_con"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/storage"
	"node/storage/deep_actions"
	"node/storage/validation"
	"node/websocket/sender"
	"strconv"
	"strings"
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
		"default_contract_create_transaction",
		"default_contract_set_price_transaction",
		"create_token_transaction",
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
		"default_contract_create_transaction",
		"default_contract_set_price_transaction",
		"create_token_transaction",
	}

	taxNullFalseTxCommentTitles := []string{
		"default_transaction",
		"custom_turing_token_delegate_transaction",
		"default_contract_buy_transaction",
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

	switch txType {
	case 1:
		switch txCommentTitle {
		case "default_transaction":
			break
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
		case "default_contract_create_transaction":
			var txCommentData []interface{}
			_ = json.Unmarshal([]byte(txCommentDataJson), &txCommentData)
			if txCommentData == nil {
				return api_error.NewApiError(15, "empty fields")
			}

			if len(txCommentData) > config.NftTokenElCreateLimit {
				return api_error.NewApiError(16, "limit token elements to one transaction")
			}

			for i := range txCommentData {
				// convert txCommentData element to map string interface for validating
				el := apparel.ConvertInterfaceToMapStringInterface(txCommentData[i])
				if err := default_con.ValidateCreate(apparel.ConvertInterfaceToString(el["name"]), senderAddress,
					recipientAddress, tokenLabel, apparel.ConvertInterfaceToFloat64(el["price"]), amount, apparel.ConvertInterfaceToString(el["data"])); err != 0 {
					return api_error.NewApiError(err, "Default contract create error")
				}
			}
			break
		case "default_contract_buy_transaction":
			if err := default_con.ValidateBuy(apparel.ConvertInterfaceToInt64(txCommentData["id"]), recipientAddress,
				senderAddress, tokenLabel, amount); err != 0 {
				return api_error.NewApiError(err, "Default contract create error")
			}
			break
		case "default_contract_set_price_transaction":
			if err := default_con.ValidateSetPrice(apparel.ConvertInterfaceToInt64(txCommentData["id"]), senderAddress, recipientAddress, tokenLabel); err != 0 {
				return api_error.NewApiError(err, "Default contract set price error")
			}
		default:
			return api_error.NewApiError(21, "Invalid transaction comment title")
		}
		break
	case 3:
		switch txCommentTitle {
		case "create_token_transaction":
			standard := apparel.ConvertInterfaceToInt64(txCommentData["standard"])
			tokenType := apparel.ConvertInterfaceToInt64(txCommentData["type"])
			if err := validateCreateToken(senderAddress, apparel.ConvertInterfaceToString(txCommentData["label"]),
				apparel.ConvertInterfaceToString(txCommentData["name"]),
				apparel.ConvertInterfaceToFloat64(txCommentData["emission"]),
				tokenType,
				standard); err != nil {
				return err
			}

			if standard == 7 && tokenType == 2 {
				commission := apparel.ConvertInterfaceToFloat64(txCommentData["commission"])
				if commission < config.NftTokenMinCommission || commission > config.NftTokenMaxCommission {
					return api_error.NewApiError(15, "invalid commission amount")
				}
			}
			break
		}
	}

	var amountNull bool

	amountNullTrueTxCommentTitles := []string{
		"custom_turing_token_add_emission_transaction",
		"custom_turing_token_de_delegate_transaction",
		"custom_turing_token_de_delegate_another_address_transaction",
		"custom_turing_token_get_reward_transaction",
		"custom_turing_token_re_delegate_transaction",
		"default_contract_set_price_transaction",
	}

	amountNullFalseTxCommentTitles := []string{
		"default_transaction",
		"custom_turing_token_delegate_transaction",
		"default_contract_create_transaction",
		"default_contract_buy_transaction",
		"create_token_transaction",
	}

	if apparel.ContainsStringInStringArr(amountNullTrueTxCommentTitles, txCommentTitle) {
		amountNull = true
	} else if apparel.ContainsStringInStringArr(amountNullFalseTxCommentTitles, txCommentTitle) {
		amountNull = false
	} else {
		return api_error.NewApiError(6, "Incorrect transaction comment title")
	}

	if validateBalance := validateBalanceTest(senderAddress, amount, tokenLabel, taxNull, amountNull);
		validateBalance != nil {
		return validateBalance
	}

	if validateTxInMemory := validateTxInMemoryTest(senderAddress, recipientAddress, txCommentTitle, 1);
		validateTxInMemory != nil {
		return validateTxInMemory
	}

	return nil
}

func validateCreateToken(proposer, label, name string, emission float64, tokenType, standard int64) *api_error.ApiError {
	if !memory.IsMainNode() {
		return api_error.NewApiError(1, "this node is not main")
	}

	if label == "" {
		return api_error.NewApiError(5, "empty label string")
	}

	if strings.Contains(label, " ") || strings.Contains(label, "-") || strings.Contains(label, "_") {
		return api_error.NewApiError(6, "invalid label string")
	}

	if int64(len(label)) > config.MaxLabel {
		return api_error.NewApiError(6, "invalid label string")
	}

	if int64(len(label)) < config.MinLabel {
		return api_error.NewApiError(6, "invalid label string")
	}

	if storage.CheckToken(label) {
		return api_error.NewApiError(7, "token with this label does exist")
	}

	if name == "" {
		return api_error.NewApiError(8, "empty name string")
	}

	if int64(len(name)) > config.MaxName {
		return api_error.NewApiError(8, "invalid name string")
	}

	if !validation.CheckInInt64Array(config.TokenTypes, tokenType) {
		return api_error.NewApiError(9, "invalid type")
	}

	if tokenType != 2 {
		if emission == 0 {
			return api_error.NewApiError(10, "empty emission")
		}

		if emission > config.MaxEmission {
			return api_error.NewApiError(11, "invalid emission")
		}
		if emission < config.MinEmission {
			return api_error.NewApiError(11, "invalid emission")
		}

		if standard != 0 {
			return api_error.NewApiError(14, "token type does not comply with the standard")
		}
	} else {
		if emission != 0 {
			return api_error.NewApiError(11, "invalid emission amount")
		}

		if standard != 7 {
			return api_error.NewApiError(14, "token type does not comply with the standard")
		}
	}

	balance := storage.GetBalance(proposer)
	if balance != nil {
		for _, coin := range balance {
			if coin.TokenLabel == "uwm" {
				if emission == 10000000 {
					if coin.Amount < config.NewTokenCost1 {
						return api_error.NewApiError(12, "low balance")
					}
				} else if emission > 10000000 {
					if coin.Amount < config.NewTokenCost2 {
						return api_error.NewApiError(12, "low balance")
					}
				} else {
					if coin.Amount < config.NewTokenCost1 {
						return api_error.NewApiError(12, "low balance")
					}
				}
			}
		}
	} else {
		return api_error.NewApiError(12, "low balance")
	}

	address := deep_actions.GetAddress(proposer)
	token := deep_actions.GetToken(address.TokenLabel)
	if token.Id != 0 {
		return api_error.NewApiError(13, "this address already have a token")
	}

	return nil
}
