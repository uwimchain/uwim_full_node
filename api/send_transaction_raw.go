package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"node/api/api_error"
	"node/apparel"
	"node/blockchain/contracts/business_token_con"
	"node/blockchain/contracts/custom_turing_token_con"
	"node/blockchain/contracts/default_con"
	"node/blockchain/contracts/delegate_con"
	"node/blockchain/contracts/donate_token_con"
	"node/blockchain/contracts/my_token_con"
	"node/blockchain/contracts/trade_token_con"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/metrics"
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

var taxNullTrueTxCommentTitles = []string{
	"custom_turing_token_add_emission_transaction",
	"custom_turing_token_de_delegate_transaction",
	"custom_turing_token_de_delegate_another_address_transaction",
	"custom_turing_token_get_reward_transaction",
	"custom_turing_token_re_delegate_transaction",
	"default_contract_create_transaction",
	"default_contract_set_price_transaction",
	"create_token_transaction",
	"undelegate_contract_transaction",
	"change_token_standard_transaction",
	"fill_token_card_transaction",
	"fill_token_standard_card_transaction",
	"rename_token_transaction",
	"my_token_contract_confirmation_transaction",
	"my_token_contract_get_percent_transaction",
	"business_token_contract_get_percent_transaction",
	"trade_token_contract_add_transaction",
	"trade_token_contract_fill_config_transaction",
	"trade_token_contract_get_com_transaction",
	"trade_token_contract_get_liq_transaction",
}

var taxNullFalseTxCommentTitles = []string{
	"default_transaction",
	"custom_turing_token_delegate_transaction",
	"default_contract_buy_transaction",
	"delegate_contract_transaction",
	"business_token_contract_buy_transaction",
	"donate_token_contract_buy_transaction",
	"trade_token_contract_swap_transaction",
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
}

var amountNullFalseTxCommentTitles = []string{
	"default_transaction",
	"custom_turing_token_delegate_transaction",
	"default_contract_create_transaction",
	"default_contract_buy_transaction",
	"create_token_transaction",
	"delegate_contract_transaction",
	"change_token_standard_transaction",
	"fill_token_card_transaction",
	"fill_token_standard_card_transaction",
	"rename_token_transaction",
	"my_token_contract_get_percent_transaction",
	"business_token_contract_buy_transaction",
	"donate_token_contract_buy_transaction",
	"trade_token_contract_add_transaction",
	"trade_token_contract_fill_config_transaction",
	"trade_token_contract_swap_transaction",
}

func (api *Api) SendTransactionRaw(args *SendTransactionRawArgs, result *string) error {
	txRaw, err := crypt.DecodeTransactionRaw(args.TransactionRaw)
	if err != nil {
		return err
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	tax := 0.01

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
		return api_error.NewApiError(2, "Recipient address is null")
	}

	if senderAddress == recipientAddress {
		return api_error.NewApiError(3, "Recipient address is a sender address")
	}

	if amount <= 0 {
		return api_error.NewApiError(4, "Zero or negative amount")
	}

	if tokenLabel == "" {
		return api_error.NewApiError(5, "Empty token label")
	}

	if !deep_actions.CheckToken(tokenLabel) {
		return api_error.NewApiError(6, fmt.Sprintf("Token with label \"%s\" does not exist", tokenLabel))
	}

	var taxNull bool

	if apparel.ContainsStringInStringArr(taxNullTrueTxCommentTitles, txCommentTitle) {
		taxNull = true
	} else if apparel.ContainsStringInStringArr(taxNullFalseTxCommentTitles, txCommentTitle) {
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
			if err := validateBusinessGetPercent(senderAddress, recipientAddress, apparel.ConvertInterfaceToString(txCommentData["token_label"]), apparel.ConvertInterfaceToFloat64(txCommentData["amount"])); err != nil {
				return err
			}
			break
		case "donate_token_contract_buy_transaction":
			if err := validateDonateBuy(senderAddress, recipientAddress, tokenLabel, amount); err != nil {
				return err
			}
			break
		case "trade_token_contract_add_transaction":
			if err := validateTradeAdd(senderAddress, recipientAddress, tokenLabel, amount); err != nil {
				return err
			}
			break
		case "trade_token_contract_fill_config_transaction":
			if err := validateTradeFillConfig(senderAddress, recipientAddress, apparel.ConvertInterfaceToFloat64(txCommentData["commission"]), amount); err != nil {
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
			if err := validateTradeSwap(senderAddress, recipientAddress, tokenLabel, apparel.ConvertInterfaceToString(txCommentData["swap_token_label"]), amount); err != nil {
				return err
			}
			break
		default:
			return api_error.NewApiError(7, "Incorrect transaction comment title")
		}
		break
	case 3:
		switch txCommentTitle {
		case "create_token_transaction":
			standard := apparel.ConvertInterfaceToInt64(txCommentData["standard"])
			tokenType := apparel.ConvertInterfaceToInt64(txCommentData["type"])
			if err := validateCreateToken(senderAddress, recipientAddress, apparel.ConvertInterfaceToString(txCommentData["label"]),
				apparel.ConvertInterfaceToString(txCommentData["name"]),
				apparel.ConvertInterfaceToFloat64(txCommentData["emission"]),
				tokenType,
				standard); err != nil {
				return err
			}

			if standard == 7 && tokenType == 2 {
				commission := apparel.ConvertInterfaceToFloat64(txCommentData["commission"])
				if commission < config.NftTokenMinCommission || commission > config.NftTokenMaxCommission {
					return api_error.NewApiError(112, "invalid commission amount")
				}
			}
			break
		case "change_token_standard_transaction":
			if err := validateChangeTokenStandard(senderAddress, recipientAddress, apparel.ConvertInterfaceToInt64(txCommentData["standard"]), amount); err != nil {
				return err
			}
			break
		case "fill_token_card_transaction":
			if err := validateFillTokenCard(senderAddress, recipientAddress, txCommentDataJson, amount); err != nil {
				return err
			}
			break
		case "fill_token_standard_card_transaction":
			if err := validateFillTokenStandardCard(senderAddress, recipientAddress, txCommentDataJson, amount); err != nil {
				return err
			}
			break
		case "rename_token_transaction":
			if err := validateRenameToken(senderAddress, recipientAddress, apparel.ConvertInterfaceToString(txCommentData["new_name"]), amount); err != nil {
				return err
			}
			break
		default:
			return api_error.NewApiError(7, "Incorrect transaction comment title")
		}
	default:
		return api_error.NewApiError(8, "Incorrect transaction type")
	}

	var amountNull bool

	if apparel.ContainsStringInStringArr(amountNullTrueTxCommentTitles, txCommentTitle) {
		amountNull = true
	} else if apparel.ContainsStringInStringArr(amountNullFalseTxCommentTitles, txCommentTitle) {
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

func validateCreateToken(senderAddress, recipientAddress, label, name string, emission float64, tokenType, standard int64) *api_error.ApiError {
	if !memory.IsMainNode() {
		return api_error.NewApiError(101, "this node is not main")
	}

	if label == "" {
		return api_error.NewApiError(102, "empty label string")
	}

	if strings.Contains(label, " ") || strings.Contains(label, "-") || strings.Contains(label, "_") {
		return api_error.NewApiError(103, "invalid label string")
	}

	if int64(len(label)) > config.MaxLabel {
		return api_error.NewApiError(103, "invalid label string")
	}

	if int64(len(label)) < config.MinLabel {
		return api_error.NewApiError(103, "invalid label string")
	}

	if deep_actions.CheckToken(label) {
		return api_error.NewApiError(104, "token with this label does exist")
	}

	if name == "" {
		return api_error.NewApiError(105, "empty name string")
	}

	if int64(len(name)) > config.MaxName {
		return api_error.NewApiError(106, "invalid name string")
	}

	if !validation.CheckInInt64Array(config.TokenTypes, tokenType) {
		return api_error.NewApiError(107, "invalid type")
	}

	if tokenType != 2 {
		if emission == 0 {
			return api_error.NewApiError(108, "empty emission")
		}

		if emission > config.MaxEmission {
			return api_error.NewApiError(109, "Incorrect emission")
		}
		if emission < config.MinEmission {
			return api_error.NewApiError(109, "Incorrect emission")
		}

		if standard != 0 {
			return api_error.NewApiError(110, "token type does not comply with the standard")
		}
	} else {
		if emission != 0 {
			return api_error.NewApiError(109, "Incorrect emission")
		}

		if standard != 7 {
			return api_error.NewApiError(110, "token type does not comply with the standard")
		}
	}

	balance := storage.GetBalance(senderAddress)
	if balance != nil {
		for _, coin := range balance {
			if coin.TokenLabel == "uwm" {
				if emission == 10000000 {
					if coin.Amount < config.NewTokenCost1 {
						return api_error.NewApiError(9, "Low balance")
					}
				} else if emission > 10000000 {
					if coin.Amount < config.NewTokenCost2 {
						return api_error.NewApiError(9, "Low balance")
					}
				} else {
					if coin.Amount < config.NewTokenCost1 {
						return api_error.NewApiError(9, "Low balance")
					}
				}
			}
		}
	} else {
		return api_error.NewApiError(8, "low balance")
	}

	address := deep_actions.GetAddress(senderAddress)
	token := deep_actions.GetToken(address.TokenLabel)
	if token.Id != 0 {
		return api_error.NewApiError(111, "this address already have a token")
	}

	validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "fill_token_standard_card_transaction", 1)
	if validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateChangeTokenStandard(senderAddress, recipientAddress string, standard int64, amount float64) *api_error.ApiError {
	if !memory.IsMainNode() {
		return api_error.NewApiError(101, "This node is not a main")
	}

	if amount != config.ChangeTokenStandardCost {
		return api_error.NewApiError(102, "Invalid amount")
	}

	if !apparel.SearchInArray([]int64{1, 3, 4, 5}, standard) {
		return api_error.NewApiError(103, "Invalid standard")
	}

	address := deep_actions.GetAddress(senderAddress)
	token := deep_actions.GetToken(address.TokenLabel)
	if token.Id == 0 {
		return api_error.NewApiError(104, "Token does not exist")
	}

	if standard == token.Standard {
		return api_error.NewApiError(105, "Invalid standard")
	}

	if token.Standard == 0 && !apparel.SearchInArray([]int64{1, 3, 4, 5}, standard) {
		return api_error.NewApiError(105, "Invalid standard")
	}

	if token.Standard == 1 && !apparel.SearchInArray([]int64{3, 4, 5}, standard) {
		return api_error.NewApiError(105, "Invalid standard")
	}

	if token.Standard == 3 && !apparel.SearchInArray([]int64{4, 6}, standard) {
		return api_error.NewApiError(105, "Invalid standard")
	}

	if token.Standard == 7 || token.Standard == 2 {
		return api_error.NewApiError(105, "Invalid standard")
	}

	validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "fill_token_standard_card_transaction", 1)
	if validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateFillTokenCard(senderAddress, recipientAddress, tokenCardData string, amount float64) *api_error.ApiError {
	if !memory.IsMainNode() {
		return api_error.NewApiError(101, "This node is not a main")
	}

	if !storage.CheckAddressToken(senderAddress) {
		return api_error.NewApiError(102, "Token does not exist")
	}

	if amount != config.FillTokenCardCost {
		return api_error.NewApiError(103, "Invalid amount")
	}

	tokenCard := deep_actions.PersonalTokenCard{}
	err := json.Unmarshal([]byte(tokenCardData), &tokenCard)
	if err != nil {
		return api_error.NewApiError(104, "fill token card data error")
	}

	if tokenCard.Hashtags != "" {
		hashtags := strings.Split(strings.TrimSpace(tokenCard.Hashtags), "#")
		if len(hashtags)-1 < 3 {
			return api_error.NewApiError(105, "Hashtags count less than 3")
		}

		if len(hashtags)-1 > 10 {
			return api_error.NewApiError(106, "Hashtags count more than 10")
		}
	}

	validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "fill_token_standard_card_transaction", 1)
	if validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateFillTokenStandardCard(senderAddress, recipientAddress, standardCardData string, amount float64) *api_error.ApiError {
	if !memory.IsMainNode() {
		return api_error.NewApiError(101, "This node is not main")
	}

	if amount != config.FillTokenStandardCardCost {
		return api_error.NewApiError(102, "Invalid amount")
	}

	if !storage.CheckAddressToken(senderAddress) {
		return api_error.NewApiError(103, "Token does not exist")
	}

	address := deep_actions.GetAddress(senderAddress)
	token := deep_actions.GetToken(address.TokenLabel)
	switch token.Standard {
	case 1:
		tokenStandardCard := deep_actions.DonateStandardCardData{}
		err := json.Unmarshal([]byte(standardCardData), &tokenStandardCard)
		if err != nil {
			log.Println(err)
			log.Println(standardCardData)
			return api_error.NewApiError(104, "Invalid standard card data")
		}
		break
	case 3:
		tokenStandardCard := deep_actions.StartUpStandardCardData{}
		err := json.Unmarshal([]byte(standardCardData), &tokenStandardCard)
		if err != nil {
			return api_error.NewApiError(104, "Invalid standard card data")
		}
		break
	case 4:
		tokenStandardCard := deep_actions.BusinessStandardCardData{}
		err := json.Unmarshal([]byte(standardCardData), &tokenStandardCard)
		if err != nil {
			return api_error.NewApiError(104, "Invalid standard card data")
		}

		if tokenStandardCard.Partners != nil {
			for _, i := range tokenStandardCard.Partners {
				if !crypt.IsAddressUw(i.Address) {
					return api_error.NewApiError(104, "Invalid standard card data")
				}
			}
		}
		break
	case 7:
		tokenStandardCard := deep_actions.NftStandardCardData{}
		err := json.Unmarshal([]byte(standardCardData), &tokenStandardCard)
		if err != nil {
			return api_error.NewApiError(104, "Invalid standard card data")
		}
		break
	default:
		return api_error.NewApiError(104, "Invalid standard")
	}

	validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "fill_token_standard_card_transaction", 1)
	if validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateRenameToken(senderAddress, recipientAddress, tokenName string, amount float64) *api_error.ApiError {
	if !memory.IsMainNode() {
		return api_error.NewApiError(101, "This node is not main")
	}

	if amount != config.RenameTokenCost {
		return api_error.NewApiError(102, "Invalid amount")
	}

	address := deep_actions.GetAddress(senderAddress)
	if address.TokenLabel == "" {
		return api_error.NewApiError(103, fmt.Sprintf("Address \"%s\" don`t have a token", address.Address))
	}

	token := address.GetToken()
	if token.Id == 0 {
		return api_error.NewApiError(104, fmt.Sprintf("Token \"%s\" does not exist", address.TokenLabel))
	}

	if tokenName == "" {
		return api_error.NewApiError(105, "Empty token name")
	}

	if int64(len(tokenName)) > config.MaxName {
		return api_error.NewApiError(106, "Token name more than max token name")
	}

	if token.Label == "uwm" {
		return api_error.NewApiError(107, "Token is a base token")
	}

	validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "rename_token_transaction", 1)
	if validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
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
		case 417:
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
		case 116:
			return api_error.NewApiError(104, "Incorrect token label")
		case 115:
			return api_error.NewApiError(105, "Token does not exist")
		case 117:
			return api_error.NewApiError(106, "Invalid token standard card data")
		case 118:
		case 119:
			return api_error.NewApiError(107, "Token conversion is null")
		case 1110:
		case 1111:
			return api_error.NewApiError(108, "Smart-contract balance have low balance")
		}
	}

	if validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "donate_token_contract_buy_transaction", 1); validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateTradeAdd(senderAddress, recipientAddress, tokenLabel string, amount float64) *api_error.ApiError {
	if !deep_actions.CheckToken(tokenLabel) {
		return api_error.NewApiError(101, fmt.Sprintf("Token with label \"%s\" does not exist", tokenLabel))
	}

	if validateAdd := trade_token_con.ValidateAdd(trade_token_con.NewTradeArgsForValidate(recipientAddress, senderAddress, amount, tokenLabel)); validateAdd != 0 {
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

func validateTradeFillConfig(senderAddress, recipientAddress string, commission, amount float64) *api_error.ApiError {
	if !crypt.IsAddressUw(senderAddress) {
		return api_error.NewApiError(101, "Sender address is not a uwim address")
	}

	if !crypt.IsAddressSmartContract(recipientAddress) {
		return api_error.NewApiError(102, "Recipient address is not a smart-contract address")
	}

	if validateFillConfig := trade_token_con.ValidateFillConfig(trade_token_con.NewFillConfigArgs(recipientAddress, commission)); validateFillConfig != 0 {
		switch validateFillConfig {
		case 522:
			return api_error.NewApiError(103, "Invalid commission amount")
		case 523:
			return api_error.NewApiError(104, "Token does not exist")
		case 524:
			return api_error.NewApiError(105, "Token standard is not a trade")
		}
	}

	if amount != 1 {
		return api_error.NewApiError(106, "Invalid amount")
	}

	if validateTxInMemory := validateTxInMemory(senderAddress, recipientAddress, "trade_token_contract_fill_config_transaction", 1); validateTxInMemory != 0 {
		return api_error.NewApiError(10, "You have already sent a transaction of this type")
	}

	return nil
}

func validateTradeGetCom(senderAddress, recipientAddress, tokenLabel string) *api_error.ApiError {
	if validateGetLiq := trade_token_con.ValidateGetCom(trade_token_con.NewGetArgsForValidate(recipientAddress, senderAddress, tokenLabel)); validateGetLiq != 0 {
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
		case 5412:
		case 545:
			return api_error.NewApiError(105, "Incorrect token label")
		case 546:
		case 547:
			return api_error.NewApiError(106, "Invalid smart-contract data")
		case 548:
			return api_error.NewApiError(107, "This token don`t have holders")
		case 549:
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
	if validateGetLiq := trade_token_con.ValidateGetLiq(trade_token_con.NewGetArgsForValidate(recipientAddress, senderAddress, tokenLabel)); validateGetLiq != 0 {
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
		case 5312:
		case 535:
			return api_error.NewApiError(105, "Incorrect token label")
		case 536:
		case 537:
			return api_error.NewApiError(106, "Invalid smart-contract data")
		case 538:
			return api_error.NewApiError(107, "This token don`t have holders")
		case 539:
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

	if validateSwap := trade_token_con.ValidateSwap(trade_token_con.NewTradeArgsForValidate(recipientAddress, senderAddress, amount, tokenLabel)); validateSwap != 0 {
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
		case 5210:
			return api_error.NewApiError(107, "Invalid token label")
		case 527:
			return api_error.NewApiError(108, "Invalid smart-contract address pool data")
		case 528:
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
