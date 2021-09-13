package api

import (
	"encoding/json"
	"fmt"
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

type BusinessTokenContractBuyArgs struct {
	Mnemonic   string  `json:"mnemonic"`
	Amount     float64 `json:"amount"`
	TokenLabel string  `json:"token_label"`
}

func (api *Api) BusinessTokenContractBuy(args *BusinessTokenContractBuyArgs, result *string) error {
	args.Mnemonic, args.TokenLabel = apparel.TrimToLower(args.Mnemonic), apparel.TrimToLower(args.TokenLabel)
	args.Amount, _ = apparel.Round(args.Amount)

	scAddressToken := storage.GetToken(args.TokenLabel)
	if scAddressToken.Id == 0 {
		return errors.New(fmt.Sprintf("error 1: %s", args.TokenLabel))
	}

	scAddressPublicKey, err := crypt.PublicKeyFromAddress(scAddressToken.Proposer)
	if err != nil {
		return errors.New(fmt.Sprintf("error 2: %s", args.TokenLabel))
	}
	scAddress := crypt.AddressFromPublicKey(metrics.SmartContractPrefix, scAddressPublicKey)

	uwAddress := crypt.AddressFromMnemonic(args.Mnemonic)

	check := validateBusinessBuy(args.Mnemonic, args.TokenLabel, scAddress, uwAddress, args.Amount)
	if check != 0 {
		return errors.New(strconv.FormatInt(check, 10))
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	var tax float64 = 0.01

	tx := deep_actions.Tx{
		Type:       1,
		Nonce:      apparel.GetNonce(timestamp),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       uwAddress,
		To:         scAddress,
		Amount:     args.Amount,
		TokenLabel: config.BaseToken,
		Timestamp:  timestamp,
		Tax:        tax,
		Signature:  crypt.SignMessageWithSecretKey(crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic)), []byte(uwAddress)),
		Comment: deep_actions.Comment{
			Title: "business_token_contract_buy_transaction",
			Data:  nil,
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

func validateBusinessBuy(mnemonic, tokenLAbel, scAddress, uwAddress string, amount float64) int64 {
	validateMnemonic := validateMnemonic(mnemonic, uwAddress)
	if validateMnemonic != 0 {
		return validateMnemonic
	}

	validateBuy := business_token_con.ValidateBuy(scAddress, uwAddress, tokenLAbel, amount)
	if validateBuy != 0 {
		return validateBuy
	}

	validateBalance := validateBalance(uwAddress, amount, config.BaseToken, true)
	if validateBalance != 0 {
		return validateBalance
	}

	validateTxInMemory := validateTxInMemory(uwAddress, scAddress, "business_token_contract_buy_transaction", 1)
	if validateTxInMemory != 0 {
		return validateTxInMemory
	}

	return 0
}