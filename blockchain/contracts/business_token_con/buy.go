package business_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/storage"
	"strconv"
)

type BuyArgs struct {
	ScAddress   string  `json:"sc_address"`
	UwAddress   string  `json:"uw_address"`
	TokenLabel  string  `json:"token_label"`
	Amount      float64 `json:"amount"`
	TxHash      string  `json:"tx_hash"`
	BlockHeight int64   `json:"block_height"`
}

func NewBuyArgs(scAddress string, uwAddress, tokenLabel string, amount float64, txHash string, blockHeight int64) (*BuyArgs, error) {
	scAddress, uwAddress, txHash = apparel.TrimToLower(scAddress), apparel.TrimToLower(uwAddress), apparel.TrimToLower(txHash)

	if !crypt.IsAddressSmartContract(scAddress) {
		return nil, errors.New(fmt.Sprintf("error 1: this address \"%s\" is not a smart-contract address", scAddress))
	}

	if !crypt.IsAddressUw(uwAddress) {
		return nil, errors.New(fmt.Sprintf("error 2: this address \"%s\" is not a uwim address", scAddress))
	}

	if txHash == "" || storage.GetTxForHash(txHash) == "" {
		return nil, errors.New(fmt.Sprintf("error 3: transaction with this hash \"%s\" does not exist", txHash))
	}

	amount, _ = apparel.Round(amount)
	if amount <= 0 {
		return nil, errors.New("error 4: zero or negative amount")
	}

	return &BuyArgs{ScAddress: scAddress, UwAddress: uwAddress, TokenLabel: tokenLabel, Amount: amount, TxHash: txHash, BlockHeight: blockHeight}, nil
}

func Buy(args *BuyArgs) error {
	err := buy(args.ScAddress, args.UwAddress, args.TokenLabel, args.TxHash, args.Amount, args.BlockHeight)
	if err != nil {
		refundError := contracts.RefundTransaction(args.ScAddress, args.UwAddress, args.Amount, args.TokenLabel)
		if refundError != nil {
			log.Println(fmt.Sprintf("Refund transaction %v", refundError))
		}
		return errors.New(fmt.Sprintf("error 1: buy %v", err))
	}

	return nil
}

func buy(scAddress, uwAddress, tokenLabel, txHash string, amount float64, blockHeight int64) error {
	if !crypt.IsAddressSmartContract(scAddress) {
		return errors.New(fmt.Sprintf("error 1: this address \"%s\" is not a smart-contract", scAddress))
	}

	if !crypt.IsAddressUw(uwAddress) {
		return errors.New(fmt.Sprintf("error 2: this address \"%s\" is not a uwim address", uwAddress))
	}

	if amount <= 0 {
		return errors.New("error 3: zero or negative amount")
	}

	if tokenLabel != config.BaseToken {
		return errors.New(fmt.Sprintf("error 4: token label is not a \"%v\"", config.BaseToken))
	}

	scAddressToken := contracts.GetTokenInfoForScAddress(scAddress)
	if scAddressToken.Id == 0 {
		return errors.New(fmt.Sprintf("error 5: this token \"%s\" does not exist", tokenLabel))
	}

	scAddressBalanceForToken := contracts.GetBalanceForToken(scAddress, scAddressToken.Label)
	scAddressBalanceForTokenUwm := contracts.GetBalanceForToken(scAddress, config.BaseToken)

	if scAddressBalanceForToken.Amount < amount || scAddressBalanceForToken.Amount-amount < 1 {
		return errors.New(fmt.Sprintf("error 6: smart-contract low balance for token \"%s\"", scAddressToken.Label))
	}

	scAddressTokenStandardCard := contracts.BusinessStandardCardData{}
	err := json.Unmarshal([]byte(scAddressToken.StandardCard), &scAddressTokenStandardCard)
	if err != nil {
		return errors.New(fmt.Sprintf("error 7: %v", err))
	}

	txAmount, _ := apparel.Round(amount * scAddressTokenStandardCard.Conversion)

	txTax, _ := apparel.Round(apparel.CalcTax(txAmount))
	if scAddressBalanceForTokenUwm.Amount < amount || scAddressBalanceForTokenUwm.Amount-amount < 1 {
		return errors.New(fmt.Sprintf("error 8: smart-contract low balance for token \"%s\"", scAddressToken.Label))
	}

	timestamp := apparel.TimestampUnix()

	txCommentSign, err := json.Marshal(contracts.NewBuyTokenSign(
		config.NodeNdAddress,
	))

	tx := contracts.NewTx(
		5,
		apparel.GetNonce(strconv.FormatInt(timestamp, 10)),
		"",
		config.BlockHeight,
		scAddress,
		uwAddress,
		txAmount,
		scAddressToken.Label,
		strconv.FormatInt(timestamp, 10),
		txTax,
		crypt.SignMessageWithSecretKey(
			config.NodeSecretKey,
			[]byte(config.NodeNdAddress),
		),
		*contracts.NewComment(
			"default_transaction",
			txCommentSign,
		),
	)

	jsonString, _ := json.Marshal(contracts.Tx{
		Type:       tx.Type,
		Nonce:      tx.Nonce,
		From:       tx.From,
		To:         tx.To,
		Amount:     tx.Amount,
		TokenLabel: tx.TokenLabel,
		Comment:    tx.Comment,
	})
	tx.Signature = crypt.SignMessageWithSecretKey(config.NodeSecretKey, jsonString)

	if scAddressTokenStandardCard.Partners != nil {
		scAddressPartnersJson := ContractsDB.Get(scAddress).Value
		var scAddressPartners []Partner
		if scAddressPartnersJson != "" {
			err := json.Unmarshal([]byte(scAddressPartnersJson), &scAddressPartners)
			if err != nil {
				return errors.New(fmt.Sprintf("error 10: %v", err))
			}
		}

		if scAddressPartners != nil {
			for _, i := range scAddressTokenStandardCard.Partners {
				partnerReward, _ := apparel.Round(amount * (i.Percent / 100))
				if partnerReward <= 0 {
					continue
				}
				for jdx, j := range scAddressPartners {
					if j.Address == i.Address {
						if j.Balance != nil {
							for kdx, k := range j.Balance {
								if k.TokenLabel == config.BaseToken {
									scAddressPartners[jdx].Balance[kdx].Amount += partnerReward
									scAddressPartners[jdx].Balance[jdx].UpdateTime += strconv.FormatInt(timestamp, 10)
								}
								break
							}
							break
						}
					}
				}
			}
		}

		jsonScAddressPartners, err := json.Marshal(scAddressPartners)
		if err != nil {
			return errors.New(fmt.Sprintf("error 11: %v", err))
		}

		ContractsDB.Put(scAddress, string(jsonScAddressPartners))
	}

	err = contracts.AddEvent(scAddress, *contracts.NewEvent("Buy", timestamp, blockHeight, txHash, uwAddress, newEventBuyTypeData(txAmount, scAddressTokenStandardCard.Conversion, scAddressToken.Label)), EventDB, ConfigDB)
	if err != nil {
		return errors.New(fmt.Sprintf("error 12: %v", err))
	}

	if memory.IsNodeProposer() {
		contracts.SendTx(*tx)
		*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *tx)
	}

	return nil
}
