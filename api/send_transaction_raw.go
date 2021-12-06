package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"node/api/api_error"
	"node/apparel"
	"node/blockchain/contracts/business_token_con"
	"node/blockchain/contracts/custom_turing_token_con"
	"node/blockchain/contracts/default_con"
	"node/blockchain/contracts/delegate_con"
	"node/blockchain/contracts/donate_token_con"
	"node/blockchain/contracts/holder_con"
	"node/blockchain/contracts/my_token_con"
	"node/blockchain/contracts/trade_token_con"
	"node/config"
	"node/crypt"
	"node/metrics"
	"node/storage"
	"node/storage/deep_actions"
	"node/websocket/sender"
	"strconv"
)

type SendTransactionRawArgs struct {
	TransactionRaw string `json:"transaction_raw"`
}

var taxNullTrueTxCommentTitles = []string{
	"custom_turing_token_add_emission_transaction",
	"custom_turing_token_de_delegate_transaction",
	"custom_turing_token_de_delegate_another_address_transaction",
	"custom_turing_token_get_reward_transaction",
	"custom_turing_token_re_delegate_transaction",
	"default_contract_create_transaction",
	"default_contract_set_price_transaction",
	"undelegate_contract_transaction",
	"my_token_contract_confirmation_transaction",
	"my_token_contract_get_percent_transaction",
	"donate_token_contract_fill_config_transaction",
	"business_token_contract_get_percent_transaction",
	"business_token_contract_fill_config_transaction",
	"trade_token_contract_add_transaction",
	"trade_token_contract_fill_config_transaction",
	"trade_token_contract_get_com_transaction",
	"trade_token_contract_get_liq_transaction",
	"holder_contract_get_transaction",
	"default_contract_fill_config_transaction",
}

var taxNullFalseTxCommentTitles = []string{
	"default_transaction",
	"custom_turing_token_delegate_transaction",
	"default_contract_buy_transaction",
	"delegate_contract_transaction",
	"business_token_contract_buy_transaction",
	"donate_token_contract_buy_transaction",
	"trade_token_contract_swap_transaction",
	"holder_contract_add_transaction",
}

var amountNullTrueTxCommentTitles = []string{
	"custom_turing_token_add_emission_transaction",
	"custom_turing_token_de_delegate_transaction",
	"custom_turing_token_de_delegate_another_address_transaction",
	"custom_turing_token_get_reward_transaction",
	"custom_turing_token_re_delegate_transaction",
	"default_contract_set_price_transaction",
	"undelegate_contract_transaction",
	"my_token_contract_confirmation_transaction",
	"business_token_contract_get_percent_transaction",
	"trade_token_contract_get_com_transaction",
	"trade_token_contract_get_liq_transaction",
	"holder_contract_get_transaction",
}

var amountNullFalseTxCommentTitles = []string{
	"default_transaction",
	"custom_turing_token_delegate_transaction",
	"default_contract_create_transaction",
	"default_contract_buy_transaction",
	"create_token_transaction",
	"delegate_contract_transaction",
	"my_token_contract_get_percent_transaction",
	"donate_token_contract_buy_transaction",
	"donate_token_contract_fill_config_transaction",
	"business_token_contract_buy_transaction",
	"business_token_contract_fill_config_transaction",
	"trade_token_contract_add_transaction",
	"trade_token_contract_fill_config_transaction",
	"trade_token_contract_swap_transaction",
	"holder_contract_add_transaction",
	"default_contract_fill_config_transaction",
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

	tx.SetHash()

	publicKey, err := crypt.PublicKeyFromAddress(tx.From)
	if err != nil {
		return api_error.NewApiError(11, "incorrect sender address").AddError()
	}

	if publicKey == nil {
		return api_error.NewApiError(12, fmt.Sprintf("signature verify error %s, incorrect public key, public key length = %d", tx.Comment.Title, len(publicKey))).AddError()
	}

	jsonStringForValidateSignature := tx.GetJsonForValidateSignature()
	if !crypt.VerifySign(publicKey, jsonStringForValidateSignature, tx.Signature) {
		return api_error.NewApiError(13, fmt.Sprintf("signature verify error %s, %v", tx.Comment.Title, tx.Signature)).AddError()
	}

	if validateTransaction := validateTransactionRaw(tx.From, tx.To, tx.TokenLabel, tx.Comment.Title,
		txRaw.Comment.Data, tx.Amount, tx.Type); validateTransaction != nil {
		return validateTransaction.AddError()
	}

	sender.SendTx(tx)

	storage.AppendTxToTransactionMemory(tx)

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
				el := apparel.ConvertInterfaceToMapStringInterface(txCommentData[i])
				if err := default_con.ValidateCreate(apparel.ConvertInterfaceToString(el["name"]), senderAddress,
					recipientAddress, tokenLabel, apparel.ConvertInterfaceToFloat64(el["price"]), amount,
					apparel.ConvertInterfaceToString(el["data"])); err != 0 {
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
			if err := default_con.ValidateSetPrice(apparel.ConvertInterfaceToInt64(txCommentData["id"]), senderAddress,
				recipientAddress, tokenLabel); err != 0 {
				return api_error.NewApiError(err, "Default contract set price error")
			}
			break
		case "default_contract_fill_config_transaction":
			if err := default_con.ValidateFillConfig(senderAddress, recipientAddress, apparel.ConvertInterfaceToFloat64(txCommentData["comission"]), amount, tokenLabel); err != 0 {
				return api_error.NewApiError(err, "Default contract fill config error")
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
		case "my_token_contract_confirmation_transaction":
			if err := validateMyConfirmation(senderAddress, recipientAddress); err != nil {
				return err
			}
			break
		case "my_token_contract_get_percent_transaction":
			if err := validateMyGetPercent(senderAddress, recipientAddress, tokenLabel); err != nil {
				return err
			}
			break
		case "business_token_contract_buy_transaction":
			if err := validateBusinessBuy(senderAddress, recipientAddress, tokenLabel, amount); err != nil {
				return err
			}
			break
		case "business_token_contract_get_percent_transaction":
			if err := validateBusinessGetPercent(senderAddress, recipientAddress,
				apparel.ConvertInterfaceToString(txCommentData["token_label"]),
				apparel.ConvertInterfaceToFloat64(txCommentData["amount"])); err != nil {
				return err
			}
			break
		case "business_token_contract_fill_config_transaction":
			if err := validateBusinessFillConfig(senderAddress, recipientAddress,
				apparel.ConvertInterfaceToFloat64(txCommentData["sales_value"]),
				apparel.ConvertInterfaceToFloat64(txCommentData["conversion"]), txCommentData["partners"],
				amount, tokenLabel); err != nil {
				return err
			}
			break
		case "donate_token_contract_buy_transaction":
			if err := validateDonateBuy(senderAddress, recipientAddress, tokenLabel, amount); err != nil {
				return err
			}
			break
		case "donate_token_contract_fill_config_transaction":
			if err := validateDonateFillConfig(senderAddress, recipientAddress,
				apparel.ConvertInterfaceToFloat64(txCommentData["conversion"]),
				apparel.ConvertInterfaceToFloat64(txCommentData["max_buy"]), amount, tokenLabel); err != nil {
				return err
			}
			break
		case "trade_token_contract_add_transaction":
			if err := validateTradeAdd(senderAddress, recipientAddress, tokenLabel, amount); err != nil {
				return err
			}
			break
		case "trade_token_contract_fill_config_transaction":
			if err := validateTradeFillConfig(senderAddress, recipientAddress,
				apparel.ConvertInterfaceToFloat64(txCommentData["commission"]), amount, tokenLabel); err != nil {
				return err
			}
			break
		case "trade_token_contract_get_com_transaction":
			if err := validateTradeGetCom(senderAddress, recipientAddress, tokenLabel); err != nil {
				return err
			}
			break
		case "trade_token_contract_get_liq_transaction":
			if err := validateTradeGetLiq(senderAddress, recipientAddress, tokenLabel); err != nil {
				return err
			}
			break
		case "trade_token_contract_swap_transaction":
			if err := validateTradeSwap(senderAddress, recipientAddress, tokenLabel,
				apparel.ConvertInterfaceToString(txCommentData["swap_token_label"]), amount); err != nil {
				return err
			}
			break
		case "holder_contract_add_transaction":
			if err := validateHolderAdd(senderAddress, recipientAddress,
				apparel.ConvertInterfaceToString(txCommentData["get_tokens_address"]),
				amount, tokenLabel, apparel.ConvertInterfaceToInt64(txCommentData["get_block_height"])); err != nil {
				return err
			}
			break
		case "holder_contract_get_transaction":
			if err := validateHolderGet(senderAddress, recipientAddress, amount, tokenLabel); err != nil {
				return err
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

	if validateBalance := validateBalanceTest(senderAddress, amount, tokenLabel, taxNull, amountNull);
		validateBalance != nil {
		return api_error.NewApiError(9, "Low balance")
	}

	return nil
}

func validateMyConfirmation(senderAddress, recipientAddress string) *api_error.ApiError {
	recipientUwAddress := crypt.AddressFromAnotherAddress(metrics.AddressPrefix, recipientAddress)
	address := deep_actions.GetAddress(recipientUwAddress)
	token := address.GetToken()

	if token.Id == 0 || token.Label == "uwm" {
		return api_error.NewApiError(101, "Token does not exist")
	}

	if !crypt.IsAddressUw(senderAddress) {
		return api_error.NewApiError(102, "Sender address is not a uwim address")
	}

	if !crypt.IsAddressSmartContract(recipientAddress) {
		return api_error.NewApiError(103, "Recipient address is not a smart-contract address")
	}

	senderAddressBalance := storage.GetBalance(senderAddress)
	if senderAddressBalance == nil {
		return api_error.NewApiError(104, fmt.Sprintf("This address \"%s\" don`t have this token \"%s\"", senderAddress, token.Label))
	}

	have := false
	for _, i := range senderAddressBalance {
		if i.TokenLabel == token.Label {
			have = true
			break
		}
	}

	if !have {
		return api_error.NewApiError(104, fmt.Sprintf("This address \"%s\" don`t have this token \"%s\"", senderAddress, token.Label))
	}

	if validateConfirmation := my_token_con.ValidateConfirmation(recipientAddress, senderAddress); validateConfirmation != 0 {
		switch validateConfirmation {
		case 013:
			return api_error.NewApiError(105, fmt.Sprintf("This address \"%s\" already in the pool", senderAddress))
		}
	}

	validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "my_token_contract_confirmation_transaction", 1)
	if validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateMyGetPercent(senderAddress, recipientAddress string, tokenLabel string) *api_error.ApiError {
	if !crypt.IsAddressUw(senderAddress) {
		return api_error.NewApiError(103, "Sender address is not a uwim address")
	}

	if !crypt.IsAddressSmartContract(recipientAddress) {
		return api_error.NewApiError(104, "Recipient address is not a smart-contract address")
	}

	recipientUwAddress := crypt.AddressFromAnotherAddress(metrics.AddressPrefix, recipientAddress)
	recipientAddressObj := deep_actions.GetAddress(recipientUwAddress)
	token := recipientAddressObj.GetToken()

	if token.Id == 0 || token.Label == "uwm" {
		return api_error.NewApiError(101, "Token does not exist")
	}

	if token.Label != tokenLabel {
		return api_error.NewApiError(102, "Invalid token label")
	}

	if validateGetPercent := my_token_con.ValidateGetPercent(recipientAddress, senderAddress); validateGetPercent != 0 {
		switch validateGetPercent {
		case 023:
			return api_error.NewApiError(105, fmt.Sprintf("This address \"%s\" not in the pool", senderAddress))
		}
	}

	validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "my_token_contract_get_percent_transaction", 2)
	if validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateBusinessBuy(senderAddress, recipientAddress, tokenLabel string, amount float64) *api_error.ApiError {
	validateBuy := business_token_con.ValidateBuy(recipientAddress, senderAddress, tokenLabel, amount)
	if validateBuy != 0 {
		switch validateBuy {
		case 411:
			return api_error.NewApiError(101, "Recipient address is not a smart-contract address")
		case 412:
			return api_error.NewApiError(102, "Sender address is not a uwim address")
		case 413:
			return api_error.NewApiError(103, "Invalid token label")
		case 414:
			return api_error.NewApiError(104, "Invalid amount")
		}
	}

	scAddressBalance := storage.GetBalance(recipientAddress)
	if scAddressBalance == nil {
		return api_error.NewApiError(105, "Smart-contract balance is null")
	}

	recipientUwAddress := crypt.AddressFromAnotherAddress(metrics.AddressPrefix, recipientAddress)
	recipientAddressObj := deep_actions.GetAddress(recipientUwAddress)
	token := recipientAddressObj.GetToken()
	if token.Id == 0 {
		return api_error.NewApiError(106, "Token does not exist")
	}

	if token.Standard != 4 {
		return api_error.NewApiError(107, "Token standard is not a business")
	}

	standardCard := token.GetStandardCard()

	txAmount := amount * apparel.ConvertInterfaceToFloat64(standardCard["conversion"])
	for _, i := range scAddressBalance {
		if i.TokenLabel == token.Label {
			if i.Amount < txAmount || i.Amount-txAmount < 1 {
				return api_error.NewApiError(105, "Smart-contract has low balance for this buy")
			}
			break
		}
	}

	if validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "business_token_contract_buy_transaction",
		1); validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateBusinessGetPercent(senderAddress, recipientAddress, tokenLabel string, amount float64) *api_error.ApiError {
	if validateGetPercent := business_token_con.ValidateGetPercent(recipientAddress, senderAddress, tokenLabel, amount); validateGetPercent != 0 {
		switch validateGetPercent {
		case 411:
			return api_error.NewApiError(101, "Recipient address is not a smart-contract address")
		case 412:
			return api_error.NewApiError(102, "Sender address is not a uwim address")
		case 414:
			return api_error.NewApiError(103, "Zero or negative amount")
		case 415:
			return api_error.NewApiError(104, "Smart-contract balance has low balance")
		case 416:
			return api_error.NewApiError(105, fmt.Sprintf("This address \"%s\" not a partner for this token", senderAddress))
		case 417:
			return api_error.NewApiError(105, fmt.Sprintf("This address \"%s\" not a partner for this token", senderAddress))
		case 4121:
			return api_error.NewApiError(105, fmt.Sprintf("This address \"%s\" not a partner for this token", senderAddress))
		case 418:
			return api_error.NewApiError(106, "Partner percent amount is null")
		case 419:
			return api_error.NewApiError(107, "Amount for get more than partner percent amount")
		case 4120:
			return api_error.NewApiError(108, fmt.Sprintf("This token \"%s\" does not exist in partners percents", tokenLabel))
		}
	}

	if validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "business_token_contract_get_percent_transaction", 1); validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateBusinessFillConfig(senderAddress, recipientAddress string, salesValue, conversion float64, partners interface{}, amount float64, tokenLabel string) *api_error.ApiError {
	if validateFillConfig := business_token_con.ValidateFillConfig(senderAddress, recipientAddress, conversion, salesValue, partners, amount, tokenLabel); validateFillConfig != 0 {
		switch validateFillConfig {
		case 431:
			return api_error.NewApiError(101, "Recipient address is not a main node address")
		case 432:
			return api_error.NewApiError(102, "Sender address is not a uw address")
		case 433:
			return api_error.NewApiError(103, "Incorrect amount")
		case 434:
			return api_error.NewApiError(104, "Incorrect token label")
		case 435:
			return api_error.NewApiError(105, "Incorrect conversion")
		case 436:
			return api_error.NewApiError(106, "Incorrect sales value")
		case 437:
			return api_error.NewApiError(107, "Incorrect partners")
		case 438:
			return api_error.NewApiError(108, "Token does not exists")
		case 439:
			return api_error.NewApiError(109, "Token standard is not a \"Business\"")
		}
	}

	if validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "business_token_contract_fill_config_transaction", 1); validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateDonateBuy(senderAddress, recipientAddress, tokenLabel string, amount float64) *api_error.ApiError {
	if validateBuy := donate_token_con.ValidateBuy(recipientAddress, senderAddress, tokenLabel, amount); validateBuy != 0 {
		switch validateBuy {
		case 111:
			return api_error.NewApiError(101, "Recipient address is not a smart-contract address")
		case 112:
			return api_error.NewApiError(102, "Sender address is not a uwim address")
		case 113:
			return api_error.NewApiError(103, "Zero or negative amount")
		case 114:
			return api_error.NewApiError(104, "Incorrect token label")
		case 116:
			return api_error.NewApiError(104, "Incorrect token label")
		case 115:
			return api_error.NewApiError(105, "Token does not exist")
		case 117:
			return api_error.NewApiError(106, "Invalid token standard card data")
		case 118:
			return api_error.NewApiError(107, "Token conversion is null")
		case 119:
			return api_error.NewApiError(109, "Amount of token more than max buy")
		case 1110:
			return api_error.NewApiError(108, "Smart-contract balance have low balance")
		case 1111:
			return api_error.NewApiError(108, "Smart-contract balance have low balance")
		}
	}

	if validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "donate_token_contract_buy_transaction", 1); validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateDonateFillConfig(senderAddress, recipientAddress string, commission, maxBuy, amount float64, tokenLabel string) *api_error.ApiError {
	if validateFillConfig := donate_token_con.ValidateFillConfig(senderAddress, recipientAddress, commission, maxBuy, amount, tokenLabel); validateFillConfig != 0 {
		switch validateFillConfig {
		case 551:
			return api_error.NewApiError(101, "Recipient address is not a main node address")
		case 552:
			return api_error.NewApiError(102, "Sender address is not a uw address")
		case 553:
			return api_error.NewApiError(103, "Incorrect amount")
		case 554:
			return api_error.NewApiError(104, "Incorrect token label")
		case 555:
			return api_error.NewApiError(105, "Incorrect conversion")
		case 556:
			return api_error.NewApiError(106, "Incorrect max buy")
		case 557:
			return api_error.NewApiError(107, "Token does not exists")
		case 558:
			return api_error.NewApiError(108, "Token standard is not a \"Donate\"")
		}
	}

	if validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "donate_token_contract_fill_config_transaction", 1); validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateTradeAdd(senderAddress, recipientAddress, tokenLabel string, amount float64) *api_error.ApiError {
	if !deep_actions.CheckToken(tokenLabel) {
		return api_error.NewApiError(101, fmt.Sprintf("Token with label \"%s\" does not exist", tokenLabel))
	}

	if validateAdd := trade_token_con.ValidateAdd(recipientAddress, senderAddress, amount, tokenLabel); validateAdd != 0 {
		switch validateAdd {
		case 511:
			return api_error.NewApiError(102, "Recipient address is not a smart-contract address")
		case 512:
			return api_error.NewApiError(103, "Sender address is not a uwim address")
		case 513:
			return api_error.NewApiError(104, "Zero or negative amount")
		case 514:
			return api_error.NewApiError(101, fmt.Sprintf("Token with label \"%s\" does not exist", tokenLabel))
		case 515:
			return api_error.NewApiError(105, fmt.Sprintf("Token \"%s\" is not a trade standard token", tokenLabel))
		case 516:
			return api_error.NewApiError(106, "Invalid token label")
		}
	}

	if validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "trade_token_contract_add_transaction", 2); validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateTradeFillConfig(senderAddress, recipientAddress string, commission, amount float64, tokenLabel string) *api_error.ApiError {
	if validateFillConfig := trade_token_con.ValidateFillConfig(senderAddress, recipientAddress, commission, amount, tokenLabel); validateFillConfig != 0 {
		switch validateFillConfig {
		case 551:
			return api_error.NewApiError(101, "Recipient address is not a main node address")
		case 552:
			return api_error.NewApiError(102, "Sender address is not a uw address")
		case 553:
			return api_error.NewApiError(103, "Incorrect amount")
		case 554:
			return api_error.NewApiError(104, "Incorrect token label")
		case 555:
			return api_error.NewApiError(105, "Incorrect commission")
		case 556:
			return api_error.NewApiError(106, "Token does not exists")
		case 557:
			return api_error.NewApiError(107, "Token standard is not a \"Trade\"")
		}
	}

	if validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "trade_token_contract_fill_config_transaction", 1); validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateTradeGetCom(senderAddress, recipientAddress, tokenLabel string) *api_error.ApiError {
	if validateGetLiq := trade_token_con.ValidateGetCom(recipientAddress, senderAddress, tokenLabel); validateGetLiq != 0 {
		switch validateGetLiq {
		case 541:
			return api_error.NewApiError(101, "Recipient address is not a smart-contract address")
		case 542:
			return api_error.NewApiError(102, "Sender address is not a uwim address")
		case 543:
			return api_error.NewApiError(103, fmt.Sprintf("Token with label \"%s\" does not exist", tokenLabel))
		case 544:
			return api_error.NewApiError(104, fmt.Sprintf("Token \"%s\" is not a trade standard token", tokenLabel))
		case 5411:
			return api_error.NewApiError(105, "Incorrect token label")
		case 5412:
			return api_error.NewApiError(105, "Incorrect token label")
		case 545:
			return api_error.NewApiError(105, "Incorrect token label")
		case 546:
			return api_error.NewApiError(106, "Invalid smart-contract data")
		case 547:
			return api_error.NewApiError(106, "Invalid smart-contract data")
		case 548:
			return api_error.NewApiError(107, "This token don`t have holders")
		case 549:
			return api_error.NewApiError(108, "Nothing to get")
		case 5410:
			return api_error.NewApiError(108, "Nothing to get")
		}
	}

	validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "trade_token_contract_get_com_transaction", 1)
	if validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateTradeGetLiq(senderAddress, recipientAddress, tokenLabel string) *api_error.ApiError {
	if validateGetLiq := trade_token_con.ValidateGetLiq(recipientAddress, senderAddress, tokenLabel); validateGetLiq != 0 {
		switch validateGetLiq {
		case 531:
			return api_error.NewApiError(101, "Recipient address is not a smart-contract address")
		case 532:
			return api_error.NewApiError(102, "Sender address is not a uwim address")
		case 533:
			return api_error.NewApiError(103, fmt.Sprintf("Token with label \"%s\" does not exist", tokenLabel))
		case 534:
			return api_error.NewApiError(104, fmt.Sprintf("Token \"%s\" is not a trade standard token", tokenLabel))
		case 5311:
			return api_error.NewApiError(105, "Incorrect token label")
		case 5312:
			return api_error.NewApiError(105, "Incorrect token label")
		case 535:
			return api_error.NewApiError(105, "Incorrect token label")
		case 536:
			return api_error.NewApiError(106, "Invalid smart-contract data")
		case 537:
			return api_error.NewApiError(106, "Invalid smart-contract data")
		case 538:
			return api_error.NewApiError(107, "This token don`t have holders")
		case 539:
			return api_error.NewApiError(108, "Nothing to get")
		case 5310:
			return api_error.NewApiError(108, "Nothing to get")
		}
	}

	validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "trade_token_contract_get_liq_transaction", 1)
	if validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateTradeSwap(senderAddress, recipientAddress, tokenLabel, swapTokenLabel string, amount float64) *api_error.ApiError {
	if !deep_actions.CheckToken(tokenLabel) {
		return api_error.NewApiError(101, fmt.Sprintf("Token with label \"%s\" does not exist", tokenLabel))
	}

	if !deep_actions.CheckToken(swapTokenLabel) {
		return api_error.NewApiError(102, fmt.Sprintf("Token with label \"%s\" does not exist", swapTokenLabel))
	}

	if validateSwap := trade_token_con.ValidateSwap(recipientAddress, senderAddress, amount, tokenLabel); validateSwap != 0 {
		switch validateSwap {
		case 521:
			return api_error.NewApiError(103, "Recipient address is not a smart-contract address")
		case 522:
			return api_error.NewApiError(104, "Sender address is not a uwim address")
		case 523:
			return api_error.NewApiError(105, "Zero or negative amount")
		case 524:
			return api_error.NewApiError(101, fmt.Sprintf("Token with label \"%s\" does not exist", tokenLabel))
		case 525:
			return api_error.NewApiError(106, fmt.Sprintf("Token \"%s\" is not a trade standard token", tokenLabel))
		case 526:
			return api_error.NewApiError(107, "Invalid token label")
		case 5210:
			return api_error.NewApiError(107, "Invalid token label")
		case 527:
			return api_error.NewApiError(108, "Invalid smart-contract address pool data")
		case 528:
			return api_error.NewApiError(109, "Invalid amount")
		case 529:
			return api_error.NewApiError(109, "Invalid amount")
		}
	}

	validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "trade_token_contract_swap_transaction", 1)
	if validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateHolderAdd(senderAddress, recipientAddress, getTokensAddress string, amount float64, tokenLabel string, getBlockHeight int64) *api_error.ApiError {
	if validateAdd := holder_con.ValidateAdd(senderAddress, recipientAddress, getTokensAddress, amount, tokenLabel, getBlockHeight); validateAdd != 0 {
		switch validateAdd {
		case 711:
			return api_error.NewApiError(101, "Incorrect sender address")
		case 712:
			return api_error.NewApiError(102, "Recipient address is not a holder smart-contract address")
		case 713:
			return api_error.NewApiError(103, "Incorrect recipient address")
		case 714:
			return api_error.NewApiError(104, "Incorrect token label")
		case 715:
			return api_error.NewApiError(105, "Incorrect amount")
		case 716:
			return api_error.NewApiError(106, "Incorrect block height")
		}
	}

	validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "holder_contract_add_transaction", 1)
	if validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateHolderGet(senderAddress, recipientAddress string, amount float64, tokenLabel string) *api_error.ApiError {
	if validateGet := holder_con.ValidateGet(senderAddress, recipientAddress, amount, tokenLabel); validateGet != 0 {
		switch validateGet {
		case 721:
			return api_error.NewApiError(101, "Recipient address is not a holder smart-contract address")
		case 722:
			return api_error.NewApiError(102, "Incorrect sender address")
		case 723:
			return api_error.NewApiError(103, "Incorrect amount")
		case 724:
			return api_error.NewApiError(104, "Incorrect token label")
		case 725:
			return api_error.NewApiError(105, "Nothing to get")
		case 726:
			return api_error.NewApiError(105, "Nothing to get")
		case 727:
			return api_error.NewApiError(106, "Holder smart-contract address has low balance")
		}
	}

	validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "holder_contract_get_transaction", 1)
	if validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}
