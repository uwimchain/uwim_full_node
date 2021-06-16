package api

import (
	"bytes"
	"encoding/json"
	"log"
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

	"github.com/syndtr/goleveldb/leveldb/errors"
)

// SendTransactions method arguments
type SendTransactionArgs struct {
	Mnemonic   string  `json:"mnemonic"`
	From       string  `json:"from"`
	To         string  `json:"to"`
	Amount     float64 `json:"amount"`
	TokenLabel string  `json:"tokenLabel"`
	Type       string  `json:"type"`
}

func (api *Api) SendTransaction(args *SendTransactionArgs, result *string) error {
	args.Mnemonic, args.From, args.To, args.TokenLabel, args.Type =
		apparel.TrimToLower(args.Mnemonic),
		apparel.TrimToLower(args.From),
		apparel.TrimToLower(args.To),
		apparel.TrimToLower(args.TokenLabel),
		apparel.TrimToLower(args.Type)

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))
	signature := crypt.SignMessageWithSecretKey(secretKey, []byte(args.From))
	timestamp := apparel.Timestamp()

	var tax float64 = 0
	if args.Type != "undelegate_contract_transaction" {
		tax = apparel.CalcTax(args.Amount * config.TaxConversion * config.Tax)
	}

	comment := *deep_actions.NewComment(
		args.Type,
		nil,
	)

	amount := args.Amount

	if args.Type == "undelegate_contract_transaction" {
		undelegateCommentData := *delegate_con.NewUndelegateCommentData(args.Amount)
		jsonString, _ := json.Marshal(undelegateCommentData)
		comment.Data = jsonString

		amount = 0
	}

	if args.Type == "smart_contract_abandonment" {
		amount = 1
	}

	transaction := *deep_actions.NewTx(
		1,
		apparel.GetNonce(timestamp),
		"",
		config.BlockHeight,
		args.From,
		args.To,
		amount,
		args.TokenLabel,
		timestamp,
		tax,
		signature,
		comment,
	)

	check := validateTransaction(args)
	if check == 0 {
		jsonString, err := json.Marshal(transaction)
		if err != nil {
			log.Println("Send Transaction error:", err)
		}

		sender.SendTx(jsonString)

		if memory.IsValidator() {
			storage.TransactionsMemory = append(storage.TransactionsMemory, transaction)
		}

		*result = "Transaction send"
		return nil
	}

	return errors.New(strconv.FormatInt(check, 10))
}

func validateTransaction(args *SendTransactionArgs) int64 {
	if args.From == "" {
		return 1
	}

	if !strings.HasPrefix(args.From, metrics.NodePrefix) &&
		!strings.HasPrefix(args.From, metrics.SmartContractPrefix) &&
		!strings.HasPrefix(args.From, metrics.AddressPrefix) ||
		len(args.From) != 61 {
		return 2
	}

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

	_, err := crypt.PublicKeyFromAddress(args.To)
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

	if args.Amount <= 0 && args.Type != "undelegate_contract_transaction" {
		return 9
	}

	if args.TokenLabel == "" {
		return 10
	}

	if !storage.CheckToken(args.TokenLabel) {
		return 11
	}

	if args.Type != "undelegate_contract_transaction" {
		sendToken := deep_actions.Balance{}
		taxToken := deep_actions.Balance{}

		for _, token := range storage.GetBalance(args.From) {
			if token.TokenLabel == args.TokenLabel {
				sendToken = token
			}

			if token.TokenLabel == "uwm" {
				taxToken = token
			}
		}

		if sendToken.Amount < args.Amount {
			return 12
		}

		if taxToken.Amount < apparel.CalcTax(args.Amount*config.TaxConversion*config.Tax) {
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

	return 0
}
