package validation

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts/business_token_con"
	"node/blockchain/contracts/custom_turing_token_con"
	"node/blockchain/contracts/default_con"
	"node/blockchain/contracts/delegate_con/delegate_validation"
	"node/blockchain/contracts/donate_token_con"
	"node/blockchain/contracts/holder_con"
	"node/blockchain/contracts/my_token_con"
	"node/blockchain/contracts/trade_token_con"
	"node/blockchain/contracts/vote_con"
	"node/config"
	"node/storage"
	"node/storage/deep_actions"
)

func validateTransactionType1(t deep_actions.Tx) error {
	if t.From == config.GenesisAddress {
		return errors.New("this address haven`t permission for send transactions of this type")
	}

	switch t.Comment.Title {
	case "default_transaction":
		//pass
		break
	case "delegate_contract_transaction":
		if t.To != config.DelegateScAddress {
			return errors.New("delegate transaction send not to this smart-contract address")
		}

		break
	case "undelegate_contract_transaction":
		if t.To != config.DelegateScAddress {
			return errors.New("undelegate transaction send not to this smart-contract address")
		}

		if t.Amount != 0 {
			return errors.New("undelegate transaction amount must be 0")
		}

		check := delegate_validation.UnDelegateValidate(t.From, t.Amount)
		switch check {
		case 1:
			return errors.New("low uwm-delegate balance")
		}

		for _, t := range storage.TransactionsMemory {
			if t.Type == 1 && t.From == t.From &&
				t.Comment.Title == "undelegate_contract_transaction" &&
				t.Nonce != t.Nonce {
				return errors.New("only one undelegate transaction can be send")
			}
		}

		break
	case "smart_contract_abandonment":
		if t.Amount != 1 {
			return errors.New("amount of transaction of this type must be 1")
		}

		if t.TokenLabel != config.BaseToken {
			return errors.New("token label of transaction of this type must be uwm")
		}

		if t.To != config.GenesisAddress {
			return errors.New("transactions of this type must be send to genesis address")
		}

		//if storage.CheckAddressScKeeping(t.From) {
		if !deep_actions.GetAddress(t.From).ScKeeping {
			return errors.New("address haven`t smart-contract")
		}

		break
	case "my_token_contract_confirmation_transaction":
		validateConfirmation := my_token_con.ValidateConfirmation(t.To, t.From)
		if validateConfirmation != 0 {
			return errors.New(
				fmt.Sprintf("error for validate my token contract confirmation transaction: %d",
					validateConfirmation))
		}
		break
	case "my_token_contract_get_percent_transaction":
		validateGetPercent := my_token_con.ValidateGetPercent(t.To, t.From)
		if validateGetPercent != 0 {
			return errors.New(
				fmt.Sprintf("error for validate my token contract get percent transaction: %d", validateGetPercent))
		}
		break
	case "donate_token_contract_buy_transaction":
		validateBuyPercent := donate_token_con.ValidateBuy(t.To, t.From, t.TokenLabel,
			t.Amount)
		if validateBuyPercent != 0 {
			return errors.New(
				fmt.Sprintf("error for validate donate token contract buy transaction: %d", validateBuyPercent))
		}
		break
	case "business_token_contract_buy_transaction":
		validateBuy := business_token_con.ValidateBuy(t.To, t.From, t.TokenLabel,
			t.Amount)
		if validateBuy != 0 {
			return errors.New(
				fmt.Sprintf("error for validate business token contract buy transaction: %d", validateBuy))
		}
		break
	case "business_token_contract_get_percent_transaction":
		data := make(map[string]interface{})
		err := json.Unmarshal(t.Comment.Data, &data)
		if err != nil {
			return errors.New(
				fmt.Sprintf("error for validate business token contract get percent transaction: %v", err))
		}

		validateBuy := business_token_con.ValidateGetPercent(t.To, t.From,
			apparel.ConvertInterfaceToString(data["token_label"]),
			apparel.ConvertInterfaceToFloat64(data["amount"]))
		if validateBuy != 0 {
			return errors.New(
				fmt.Sprintf("error for validate business token contract get percent transaction: %d",
					validateBuy))
		}
		break
	case "trade_token_contract_add_transaction":
		validateAdd := trade_token_con.ValidateAdd(
			trade_token_con.NewTradeArgsForValidate(t.To, t.From, t.Amount,
				t.TokenLabel))
		if validateAdd != 0 {
			return errors.New(
				fmt.Sprintf("error for validate trade token contract add transaction: %d", validateAdd))
		}
		break
	case "trade_token_contract_swap_transaction":
		validateSwap := trade_token_con.ValidateSwap(trade_token_con.NewTradeArgsForValidate(t.To,
			t.From, t.Amount, t.TokenLabel))
		if validateSwap != 0 {
			return errors.New(
				fmt.Sprintf("error for validate trade token contract swap transaction: %d", validateSwap))
		}
		break
	case "trade_token_contract_get_liq_transaction":
		validateGetLiq := trade_token_con.ValidateGetLiq(
			trade_token_con.NewGetArgsForValidate(t.To, t.From, t.TokenLabel))
		if validateGetLiq != 0 {
			return errors.New(
				fmt.Sprintf("error for validate trade token contract get liq transaction: %d", validateGetLiq))
		}
		break
	case "trade_token_contract_get_com_transaction":
		validateGetCom := trade_token_con.ValidateGetCom(trade_token_con.NewGetArgsForValidate(t.To,
			t.From, t.TokenLabel))
		if validateGetCom != 0 {
			return errors.New(
				fmt.Sprintf("error for validate trade token contract get com transaction: %d", validateGetCom))
		}
		break
	case "trade_token_contract_fill_config_transaction":
		var scAddressConfigData trade_token_con.TradeConfig

		if t.Comment.Data != nil {
			err := json.Unmarshal(t.Comment.Data, &scAddressConfigData)
			if err != nil {
				return errors.New(
					fmt.Sprintf("error 1 for validate trade token contract get com transaction: %v", err))
			}
			validateGetCom := trade_token_con.ValidateFillConfig(
				trade_token_con.NewFillConfigArgs(t.To, scAddressConfigData.Commission))
			if validateGetCom != 0 {
				return errors.New(
					fmt.Sprintf("error 2 for validate trade token contract get com transaction: %d",
						validateGetCom))
			}
		}
		break
	case "holder_contract_add_transaction":
		if t.Tax != config.HolderAddCost {
			return errors.New("error for validate holder contract add transaction 1: invalid tax amount")
		}
		commentData := make(map[string]interface{})
		err := json.Unmarshal(t.Comment.Data, &commentData)
		if err != nil {
			return errors.New(fmt.Sprintf("error for validate holder contract add transaction 2: %v", err))
		}

		//amount, _ := apparel.Round(t.Amount)
		amount := apparel.Round(t.Amount)
		validateAdd := holder_con.ValidateAdd(t.From,
			apparel.ConvertInterfaceToString(commentData["recipient_address"]),
			apparel.ConvertInterfaceToString(commentData["token_label"]),
			amount, apparel.ConvertInterfaceToInt64(commentData["get_block_height"]))
		if validateAdd != 0 {
			return errors.New(fmt.Sprintf("error for validate holder contract add transaction 3: %d", validateAdd))
		}
		break
	case "holder_contract_get_transaction":
		if t.Tax != config.HolderGetCost {
			return errors.New("error for validate holder contract get transaction 1: invalid tax amount")
		}
		validateGet := holder_con.ValidateGet(t.From)
		if validateGet != 0 {
			return errors.New(fmt.Sprintf("error for validate holder contract get transaction 2: %d",
				validateGet))
		}
		break
	case "vote_contract_start_transaction":
		commentData := make(map[string]interface{})
		err := json.Unmarshal(t.Comment.Data, &commentData)
		if err != nil {
			return errors.New(fmt.Sprintf("error for validate vote contract start transaction 1: %v", err))
		}

		validateStart := vote_con.ValidateStart(
			apparel.ConvertInterfaceToString(commentData["title"]),
			apparel.ConvertInterfaceToString(commentData["description"]),
			t.From,
			commentData["answer_options"],
			apparel.ConvertInterfaceToInt64(commentData["end_block_height"]))
		if validateStart != 0 {
			return errors.New(fmt.Sprintf("error for validate vote contract start transaction 2: %v", validateStart))
		}

		break
	case "vote_contract_hard_stop_transaction":
		commentData := make(map[string]interface{})
		if err := json.Unmarshal(t.Comment.Data, &commentData); err != nil {
			return errors.New(fmt.Sprintf("error for validate vote contract hard stop transaction 1: %v", err))
		}

		if validateHardStop := vote_con.ValidateHardStop(t.From,
			apparel.ConvertInterfaceToString(commentData["vote_nonce"])); validateHardStop != 0 {
			return errors.New(fmt.Sprintf("error for validate vote contract hard stop transaction 2: %v",
				validateHardStop))
		}

		break
	case "vote_contract_answer_transaction":
		commentData := make(map[string]interface{})
		if err := json.Unmarshal(t.Comment.Data, &commentData); err != nil {
			return errors.New(fmt.Sprintf("error for validate vote contract answer transaction 1: %v", err))
		}

		if validateAnswer := vote_con.ValidateAnswer(t.From, apparel.ConvertInterfaceToString(commentData["vote_nonce"]),
			apparel.ConvertInterfaceToString(commentData["possible_answer_nonce"])); validateAnswer != 0 {
			return errors.New(fmt.Sprintf("error for validate vote contract answer transaction 2: %v",
				validateAnswer))
		}
		break
	case "custom_turing_token_add_emission_transaction":
		commentData := make(map[string]interface{})
		_ = json.Unmarshal(t.Comment.Data, &commentData)

		if err := custom_turing_token_con.ValidateAddEmission(t.From, t.To,
			apparel.ConvertInterfaceToFloat64(commentData["add_emission_amount"])); err != 0 {
			return errors.New(fmt.Sprintf("error for validate custom turing token contract add emission transaction 2: %v", err))
		}
		break
	case "custom_turing_token_de_delegate_transaction":
		commentData := make(map[string]interface{})
		_ = json.Unmarshal(t.Comment.Data, &commentData)

		if err := custom_turing_token_con.ValidateDeDelegate(t.From, t.To,
			apparel.ConvertInterfaceToFloat64(commentData["de_delegate_amount"])); err != 0 {
			return errors.New(fmt.Sprintf("error for validate custom turing token contract de-delegate transaction 2: %v", err))
		}
		break
	case "custom_turing_token_de_delegate_another_address_transaction":
		commentData := make(map[string]interface{})
		_ = json.Unmarshal(t.Comment.Data, &commentData)

		if err := custom_turing_token_con.ValidateDeDelegateAnotherAddress(t.From, t.To,
			apparel.ConvertInterfaceToFloat64(commentData["de_delegate_amount"])); err != 0 {
			return errors.New(fmt.Sprintf("error for validate custom turing token contract de-delegate another address transaction 2: %v", err))
		}
		break
	case "custom_turing_token_get_reward_transaction":
		if err := custom_turing_token_con.ValidateGetReward(t.From, t.To); err != 0 {
			return errors.New(fmt.Sprintf("error for validate custom turing token contract get reward transaction 2: %v", err))
		}
		break
	case "custom_turing_token_re_delegate_transaction":
		commentData := make(map[string]interface{})
		_ = json.Unmarshal(t.Comment.Data, &commentData)

		if err := custom_turing_token_con.ValidateReDelegate(t.From,
			apparel.ConvertInterfaceToString(commentData["re_delegate_recipient_address"]), t.To,
			apparel.ConvertInterfaceToFloat64(commentData["re_delegate_amount"])); err != 0 {
			return errors.New(fmt.Sprintf("error for validate custom turing token contract re-delegate transaction 2: %v", err))
		}
		break
	case "custom_turing_token_delegate_transaction":
		if err := custom_turing_token_con.ValidateDelegate(t.To, t.TokenLabel); err != 0 {
			return errors.New(fmt.Sprintf("error for validate custom turing token contract delegate transaction 2: %v", err))
		}
		break
	case "default_contract_create_transaction":
		//commentData := make(map[string]interface{})
		var commentData []interface{}
		_ = json.Unmarshal(t.Comment.Data, &commentData)

		if commentData == nil {
			return errors.New("error for validate default contract create transaction 1: empty fields")
		}

		if len(commentData) > config.NftTokenElCreateLimit {
			return errors.New("error for validate default contract create transaction 2: limit token elements to one transaction")
		}

		for i := range commentData {
			// convert txCommentData element to map string interface for validating
			el := apparel.ConvertInterfaceToMapStringInterface(commentData[i])
			if err := default_con.ValidateCreate(apparel.ConvertInterfaceToString(el["name"]), t.From, t.To,
				t.TokenLabel, apparel.ConvertInterfaceToFloat64(el["price"]), t.Amount, apparel.ConvertInterfaceToString(el["data"])); err != 0 {
				return errors.New(fmt.Sprintf("error for validate default contract create transaction 3: %v", err))
			}
		}
		break
	case "default_contract_buy_transaction":
		commentData := make(map[string]interface{})
		_ = json.Unmarshal(t.Comment.Data, &commentData)

		if err := default_con.ValidateBuy(apparel.ConvertInterfaceToInt64(commentData["id"]), t.To,
			t.From, t.TokenLabel, t.Amount); err != 0 {
			return errors.New(fmt.Sprintf("error for validate default contract buy transaction 1: %v", err))
		}
		break
	case "default_contract_set_price_transaction":
		commentData := make(map[string]interface{})
		_ = json.Unmarshal(t.Comment.Data, &commentData)

		if err := default_con.ValidateSetPrice(apparel.ConvertInterfaceToInt64(commentData["id"]), t.From, t.To, t.TokenLabel); err != 0 {
			return errors.New(fmt.Sprintf("error for validate default contract buy transaction 1: %v", err))
		}
		break
	default:
		return errors.New("transaction type does not match the comment title 1:" + t.Comment.Title)
	}

	return nil
}
