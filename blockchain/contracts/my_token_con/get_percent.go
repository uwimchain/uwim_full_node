package my_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
	"node/memory"
	"strconv"
)

type GetPercentArgs struct {
	ScAddress   string `json:"sc_address"`
	UwAddress   string `json:"uw_address"`
	TokenLabel  string `json:"token_label"`
	BlockHeight int64  `json:"block_height"`
	TxHash      string `json:"tx_hash"`
}

func NewGetPercentArgs(scAddress string, uwAddress string, tokenLabel string, blockHeight int64, txHash string) (*GetPercentArgs, error) {
	return &GetPercentArgs{ScAddress: scAddress, UwAddress: uwAddress, TokenLabel: tokenLabel, BlockHeight: blockHeight, TxHash: txHash}, nil
}

func GetPercent(args *GetPercentArgs) error {
	err := getPercent(args.ScAddress, args.UwAddress, args.TokenLabel, args.TxHash, args.BlockHeight)
	if err != nil {
		return errors.New(fmt.Sprintf("error 1: get_percent %v", err))
	}

	return nil
}

func getPercent(scAddress, uwAddress, tokenLabel, txHash string, blockHeight int64) error {
	timestamp := apparel.TimestampUnix()

	if !crypt.IsAddressSmartContract(scAddress) || scAddress == "" {
		return errors.New("error 1: sc address is null or not sc address")
	}

	if !crypt.IsAddressUw(uwAddress) || uwAddress == "" {
		return errors.New("error 2: sender address is null or not uwim address")
	}

	/*	if amount <= 0 {
		return errors.New("error 3: null or negative amount")
	}*/

	if tokenLabel == config.BaseToken {
		return nil
	}

	scAddressToken := contracts.GetTokenInfoForScAddress(scAddress)
	if scAddressToken.Id == 0 {
		return errors.New("error 4: this token does not exist")
	}

	var scAddressPool []Pool
	scAddressPoolJson := PoolDB.Get(scAddress).Value
	if scAddressPoolJson != "" {
		err := json.Unmarshal([]byte(scAddressPoolJson), &scAddressPool)
		if err != nil {
			return errors.New(fmt.Sprintf("error 5: %v", err))
		}
	}

	if scAddressPool == nil {
		return errors.New("error 6: pool of this token is empty")
	}

	if tokenLabel == scAddressToken.Label {
		var uwAddressPercent float64 = 0
		for _, i := range scAddressPool {
			if i.Address == uwAddress {
				scAddressBalanceForToken := contracts.GetBalanceForToken(scAddress, scAddressToken.Label)
				uwAddressBalanceForToken := contracts.GetBalanceForToken(i.Address, scAddressToken.Label)
				uwAddressPercent = uwAddressBalanceForToken.Amount / (scAddressToken.Emission - scAddressBalanceForToken.Amount)
				break
			}
		}

		if uwAddressPercent <= 0 {
			return errors.New("error 7: address percent <= 0")
		}

		scAddressBalance := contracts.GetBalance(scAddress)
		if scAddressBalance == nil {
			return errors.New("error 8: smart-contract balance is empty")
		}

		var transactions []contracts.Tx
		for _, i := range scAddressBalance {
			if i.TokenLabel != scAddressToken.Label {
				amount, _ := apparel.Round(i.Amount * (uwAddressPercent / 100))
				tax, _ := apparel.Round(apparel.CalcTax(amount))
				nonce := apparel.GetNonce(strconv.FormatInt(timestamp, 10))
				txCommentSign, _ := json.Marshal(contracts.NewBuyTokenSign(
					config.NodeNdAddress,
				))

				transaction := contracts.Tx{
					Type:       5,
					Nonce:      nonce,
					HashTx:     "",
					Height:     config.BlockHeight,
					From:       scAddress,
					To:         uwAddress,
					Amount:     amount,
					TokenLabel: i.TokenLabel,
					Timestamp:  strconv.FormatInt(timestamp, 10),
					Tax:        tax,
					Signature: crypt.SignMessageWithSecretKey(
						config.NodeSecretKey,
						[]byte(config.NodeNdAddress),
					),
					Comment: *contracts.NewComment(
						"default_transaction",
						txCommentSign,
					),
				}

				transactions = append(transactions, transaction)
			}
		}

		if transactions == nil {
			return nil
		}

		var allTax float64 = 0
		for _, t := range transactions {
			allTax += t.Tax

			for _, i := range scAddressBalance {
				if i.TokenLabel == t.TokenLabel {
					if i.Amount <= t.Amount {
						return errors.New("error 9: low balance on sc address")
					}
					break
				}
			}
		}

		for _, i := range scAddressBalance {
			if i.TokenLabel == config.BaseToken {
				if i.Amount <= allTax {
					return errors.New("error 10: sc address low balance of uwim for get address percent transactions taxes")
				}
				break
			}
		}

		err := contracts.AddEvent(scAddress, *contracts.NewEvent("buy", timestamp, blockHeight, txHash, uwAddress, nil), EventDB, ConfigDB)
		if err != nil {
			return errors.New(fmt.Sprintf("error 11: add event %v", err))
		}

		if memory.IsNodeProposer() {
			for _, i := range transactions {
				tx:= contracts.NewTx(
					i.Type,
					i.Nonce,
					i.HashTx,
					i.Height,
					i.From,
					i.To,
					i.Amount,
					i.TokenLabel,
					i.Timestamp,
					i.Tax,
					i.Signature,
					i.Comment,
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

				contracts.SendTx(*tx)
				*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *tx)
			}
		}
	}

	return nil
}
