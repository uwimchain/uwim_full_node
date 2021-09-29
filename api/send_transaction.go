package api

import (
	"bytes"
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts/delegate_con"
	"node/blockchain/contracts/delegate_con/delegate_validation"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/metrics"
	"node/storage"
	"node/storage/deep_actions"
	"node/websocket/sender"
	"strconv"
	"strings"
)

// SendTransaction method arguments
type SendTransactionArgs struct {
	Mnemonic   string  `json:"mnemonic"`
	From       string  `json:"from"`
	To         string  `json:"to"`
	Amount     float64 `json:"amount"`
	TokenLabel string  `json:"token_label"`
	Type       string  `json:"type"`
	Comment    Comment `json:"comment"`
}

type Comment interface {
}

func (api *Api) SendTransaction(args *SendTransactionArgs, result *string) error {
	args.Mnemonic, args.From, args.To, args.TokenLabel, args.Type =
		apparel.TrimToLower(args.Mnemonic),
		apparel.TrimToLower(args.From),
		apparel.TrimToLower(args.To),
		apparel.TrimToLower(args.TokenLabel),
		apparel.TrimToLower(args.Type)

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))
	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	var tax float64 = 0

	comment := *deep_actions.NewComment(
		args.Type,
		nil,
	)

	amount := args.Amount

	if args.Type != "undelegate_contract_transaction" {
		if args.TokenLabel == config.BaseToken {
			tax = apparel.CalcTax(args.Amount)
		} else {
			tax = 0.01
		}
	}

	switch args.Type {
	case "undelegate_contract_transaction":
		undelegateCommentData := *delegate_con.NewUndelegateCommentData(args.Amount)
		jsonString, _ := json.Marshal(undelegateCommentData)
		comment.Data = jsonString
		amount = 0
		break
	case "smart_contract_abandonment":
		amount = 1
		break
	case "default_transaction":
		if crypt.IsAddressSmartContract(args.To) && args.To != config.DelegateScAddress {
			pubKey, _ := crypt.PublicKeyFromAddress(args.To)
			uwAddress := crypt.AddressFromPublicKey(metrics.AddressPrefix, pubKey)
			token := storage.GetAddressToken(uwAddress)

			if token.Standard != 0 {
				commentJson, _ := json.Marshal(args.Comment)
				comment.Data = commentJson
			}
		}
		break
	}

	tx := deep_actions.Tx{
		Type:       1,
		Nonce:      apparel.GetNonce(timestamp),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       args.From,
		To:         args.To,
		Amount:     amount,
		TokenLabel: args.TokenLabel,
		Timestamp:  timestamp,
		Tax:        tax,
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

	check := validateTransaction(args)
	if check != 0 {
		return errors.New(strconv.FormatInt(check, 10))
	}

	jsonString, _ = json.Marshal(tx)

	sender.SendTx(tx)

	if memory.IsValidator() {
		storage.TransactionsMemory = append(storage.TransactionsMemory, tx)
	}

	*result = "Transaction send"
	return nil
}

func validateTransaction(args *SendTransactionArgs) int64 {
	if args.Mnemonic == "" {
		return 3
	}

	if args.From != crypt.AddressFromMnemonic(strings.TrimSpace(args.Mnemonic)) &&
		args.From != crypt.NodeAddressFromMnemonic(strings.TrimSpace(args.Mnemonic)) &&
		args.From != crypt.ScAddressFromMnemonic(strings.TrimSpace(args.Mnemonic)) ||
		len(bytes.Split([]byte(strings.TrimSpace(args.Mnemonic)), []byte(" "))) != 24 {
		return 4
	}

	if args.From == config.GenesisAddress || crypt.AddressFromMnemonic(args.Mnemonic) == config.GenesisAddress {
		return 5
	}

	if args.To == "" {
		return 6
	}

	publicKey, err := crypt.PublicKeyFromAddress(args.To)
	if err != nil {
		return 7
	}

	if !strings.HasPrefix(args.To, metrics.NodePrefix) &&
		!strings.HasPrefix(args.To, metrics.SmartContractPrefix) &&
		!strings.HasPrefix(args.To, metrics.AddressPrefix) ||
		len(args.To) != 61 {
		return 7
	}

	if args.From == args.To {
		return 8
	}

	if args.Amount < 0 {
		return 91
	}

	if args.Amount == 0 {
		switch args.Type {
		case "undelegate_contract_transaction":
			//pass
			break
		case "default_transaction":
			if !crypt.IsAddressSmartContract(args.To) {
				return 92
			}

			uwAddress := crypt.AddressFromPublicKey(metrics.AddressPrefix, publicKey)
			token := storage.GetAddressToken(uwAddress)
			if token.Standard != 4 {
				return 93
			}

			break
		default:
			return 95
		}
	}

	if args.TokenLabel == "" {
		return 10
	}

	if !storage.CheckToken(args.TokenLabel) {
		return 11
	}

	if args.Type != "undelegate_contract_transaction" {
		validateBalance := validateBalance(args.From, args.Amount, args.TokenLabel, false)
		if validateBalance != 0 {
			return 12
		}
	}

	switch args.Type {
	case "default_transaction":
		// pass
		break
	case "delegate_contract_transaction":
		{
			if args.To != config.DelegateScAddress {
				return 13
			}

			if !crypt.IsAddressUw(args.From) {
				return 14
			}
			break
		}

	case "undelegate_contract_transaction":
		{
			if args.To != config.DelegateScAddress {
				return 13
			}

			check := delegate_validation.UnDelegateValidate(args.From, args.Amount)
			switch check {
			case 1:
				return 15
			case 2:
				return 16
			}

			for _, t := range storage.TransactionsMemory {
				if t.Type == 1 && t.From == args.From &&
					t.Comment.Title == "undelegate_contract_transaction" {
					return 17
				}
			}
			break
		}
	default:
		return 22
	}

	for _, i := range storage.TransactionsMemory {
		if i.From == args.From {
			return 23
		}
	}

	return 0
}
