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

// TradeTokenContractGetCom method arguments
type TradeTokenContractGetComArgs struct {
	Mnemonic   string `json:"mnemonic"`
	ScAddress  string `json:"sc_address"`
	TokenLabel string `json:"token_label"`
}

func (api *Api) TradeTokenContractGetCom(args *TradeTokenContractGetComArgs, result *string) error {
	args.Mnemonic, args.TokenLabel = apparel.TrimToLower(args.Mnemonic), apparel.TrimToLower(args.TokenLabel)

	uwAddress := crypt.AddressFromMnemonic(args.Mnemonic)
	publicKey, err := crypt.PublicKeyFromAddress(args.ScAddress)
	if err != nil {
		return errors.New(strconv.Itoa(11))
	}
	tokenProposer := crypt.AddressFromPublicKey(metrics.AddressPrefix, publicKey)
	address := deep_actions.GetAddress(tokenProposer)
	token := deep_actions.GetToken(address.TokenLabel)
	if token.Id == 0 {
		return errors.New(strconv.Itoa(10))
	}

	check := validateGetCom(args.Mnemonic, args.TokenLabel, uwAddress, args.ScAddress)
	if check != 0 {
		return errors.New(strconv.FormatInt(check, 10))
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	comment := deep_actions.Comment{
		Title: "trade_token_contract_get_com_transaction",
		Data:  nil,
	}

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))

	tx := deep_actions.Tx{
		Type:       1,
		Nonce:      apparel.GetNonce(timestamp),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       uwAddress,
		To:         args.ScAddress,
		Amount:     0,
		TokenLabel: args.TokenLabel,
		Timestamp:  timestamp,
		Tax:        0,
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

	jsonString, _ = json.Marshal(tx)
	tx.HashTx = crypt.GetHash(jsonString)

	sender.SendTx(tx)

	if memory.IsValidator() {
		storage.TransactionsMemory = append(storage.TransactionsMemory, tx)
	}

	*result = "Transaction send"
	return nil
}

func validateGetCom(mnemonic, tokenLabel, uwAddress, scAddress string) int64 {
	if mnemonic == "" {
		return 1
	}

	validateMnemonic := validateMnemonic(mnemonic, uwAddress)
	if validateMnemonic != 0 {
		return validateMnemonic
	}

	validateGetLiq := trade_token_con.ValidateGetCom(trade_token_con.NewGetArgsForValidate(scAddress, uwAddress, tokenLabel))
	if validateGetLiq != 0 {
		return validateGetLiq
	}

	validateTxInMemory := validateTxInMemory(uwAddress, scAddress, "trade_token_contract_get_com_transaction", 1)
	if validateTxInMemory != 0 {
		return validateTxInMemory
	}

	return 0
}
