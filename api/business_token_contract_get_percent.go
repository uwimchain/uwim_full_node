package api

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/blockchain/contracts/business_token_con"
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
type BusinessTokenContractGetPercentArgs struct {
	Mnemonic   string `json:"mnemonic"`
	TokenLabel string `json:"token_label"`
	GetPercentData map[string]interface{} `json:"get_percent_data"`
	//GetPercentTokenLabel string  `json:"get_percent_token_label"`
	//GetPercentAmount     float64 `json:"get_percent_amount"`
}

func (api *Api) BusinessTokenContractGetPercent(args *BusinessTokenContractGetPercentArgs, result *string) error {
	args.Mnemonic, args.TokenLabel = apparel.TrimToLower(args.Mnemonic), apparel.TrimToLower(args.TokenLabel)
	uwAddress := crypt.AddressFromMnemonic(args.Mnemonic)
	scAddressToken := storage.GetToken(args.TokenLabel)
	scAddressPublicKey, err := crypt.PublicKeyFromAddress(scAddressToken.Proposer)
	if err != nil {
		return err
	}

	scAddress := crypt.AddressFromPublicKey(metrics.SmartContractPrefix, scAddressPublicKey)

	check := validateBusinessGetPercent(args.Mnemonic, args.TokenLabel, uwAddress, scAddress, args.GetPercentData)
	if check != 0 {
		return errors.New(strconv.FormatInt(check, 10))
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	commentData, _ := json.Marshal(args.GetPercentData)

	tx := deep_actions.Tx{
		Type:       1,
		Nonce:      apparel.GetNonce(timestamp),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       uwAddress,
		To:         scAddress,
		Amount:     storage.GetBalanceForToken(uwAddress, scAddressToken.Label).Amount,
		TokenLabel: args.TokenLabel,
		Timestamp:  timestamp,
		Tax:        0,
		Signature:  crypt.SignMessageWithSecretKey(crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic)), []byte(uwAddress)),
		Comment: deep_actions.Comment{
			Title: "business_token_contract_get_percent_transaction",
			Data:  commentData,
		},
	}

	jsonString, err := json.Marshal(tx)
	if err != nil {
		log.Println("Send Transaction error:", err)
	}

	sender.SendTx(jsonString)

	if memory.IsValidator() {
		storage.TransactionsMemory = append(storage.TransactionsMemory, tx)
	}

	*result = "Transaction send"
	return nil
}

func validateBusinessGetPercent(mnemonic, tokenLabel, uwAddress, scAddress string, data map[string]interface{}) int64 {
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

	validateGetPercent := business_token_con.ValidateGetPercent(scAddress, uwAddress, apparel.ConvertInterfaceToString(data["token_label"]), apparel.ConvertInterfaceToFloat64(data["amount"]))
	if validateGetPercent != 0 {
		return validateGetPercent
	}

	validateTxInMemory := validateTxInMemory(uwAddress, scAddress, "business_token_contract_get_percent_transaction", 1)
	if validateTxInMemory != 0 {
		return validateTxInMemory
	}

	return 0
}
