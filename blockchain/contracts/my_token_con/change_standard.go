package my_token_con

import (
	"encoding/json"
	"errors"
	"fmt"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
	"node/memory"
	"strconv"
)

func ChangeStandard(scAddress string) error {
	err := changeStandard(scAddress)
	if err != nil {
		return errors.New(fmt.Sprintf("error 1: %v", err))
	}

	return nil
}

func changeStandard(scAddress string) error {
	scAddressBalance := contracts.GetBalance(scAddress)
	if scAddressBalance == nil {
		return errors.New("error 1: smart-contract balance is empty")
	}

	var scAddressPool []Pool
	scAddressPoolJson := PoolDB.Get(scAddress).Value
	if scAddressPoolJson != "" {
		err := json.Unmarshal([]byte(scAddressPoolJson), &scAddressPool)
		if err != nil {
			return errors.New(fmt.Sprintf("error 2: %v", err))
		}
	}

	scAddressToken := contracts.GetTokenInfoForScAddress(scAddress)
	if scAddressToken.Id == 0 {
		return errors.New("error 3: token does not exist")
	}

	var txs []contracts.Tx
	var allTxsAmount []contracts.Balance
	scAddressBalanceForToken := contracts.GetBalanceForToken(scAddress, scAddressToken.Label)
	if scAddressPool != nil {
		for _, i := range scAddressPool {
			uwAddressBalanceForToken := contracts.GetBalanceForToken(i.Address, scAddressToken.Label)
			uwAddressPercent := uwAddressBalanceForToken.Amount / (scAddressToken.Emission - scAddressBalanceForToken.Amount)

			if uwAddressPercent == 0 {
				continue
			}

			for _, j := range scAddressBalance {
				if j.TokenLabel == scAddressToken.Label {
					continue
				}

				timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

				txs = append(txs, contracts.Tx{
					Type:       5,
					Nonce:      apparel.GetNonce(timestamp),
					HashTx:     "",
					Height:     config.BlockHeight,
					From:       scAddress,
					To:         i.Address,
					Amount:     j.Amount * (uwAddressPercent / 100),
					TokenLabel: j.TokenLabel,
					Timestamp:  timestamp,
					Tax:        0,
					Signature:  crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
					Comment:    *contracts.NewComment("refund_transaction", nil),
				})

				if allTxsAmount != nil {
					check := false
					for _, k := range allTxsAmount {
						if k.TokenLabel == j.TokenLabel {
							k.Amount += j.Amount
							check = true
							break
						}
					}

					if !check {
						allTxsAmount = append(allTxsAmount, contracts.Balance{
							TokenLabel: j.TokenLabel,
							Amount:     j.Amount,
						})
					}
				} else {
					allTxsAmount = append(allTxsAmount, contracts.Balance{
						TokenLabel: j.TokenLabel,
						Amount:     j.Amount,
					})
				}
			}
		}
	}

	if allTxsAmount != nil {
		for _, i := range allTxsAmount {
			for _, j := range scAddressBalance {
				if i.TokenLabel == j.TokenLabel {
					timestampD := strconv.FormatInt(apparel.TimestampUnix(), 10)
					tx := contracts.Tx{
						Type:       5,
						Nonce:      apparel.GetNonce(timestampD),
						HashTx:     "",
						Height:     config.BlockHeight,
						From:       scAddress,
						To:         scAddressToken.Proposer,
						Amount:     j.Amount - i.Amount,
						TokenLabel: j.TokenLabel,
						Timestamp:  timestampD,
						Tax:        0,
						Signature:  crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
						Comment: *contracts.NewComment(
							"refund_transaction",
							nil,
						),
					}
					txs = append(txs, tx)
					break
				}
			}
		}
	} else {
		for _, i := range scAddressBalance {
			timestampD := strconv.FormatInt(apparel.TimestampUnix(), 10)
			tx := contracts.Tx{
				Type:       5,
				Nonce:      apparel.GetNonce(timestampD),
				HashTx:     "",
				Height:     config.BlockHeight,
				From:       scAddress,
				To:         scAddressToken.Proposer,
				Amount:     i.Amount,
				TokenLabel: i.TokenLabel,
				Timestamp:  timestampD,
				Tax:        0,
				Signature:  crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
				Comment: *contracts.NewComment(
					"refund_transaction",
					nil,
				),
			}
			txs = append(txs, tx)
		}
	}

	if txs != nil && memory.IsNodeProposer() {
		for _, j := range txs {

			transaction := contracts.NewTx(
				j.Type,
				j.Nonce,
				"",
				j.Height,
				j.From,
				j.To,
				j.Amount,
				j.TokenLabel,
				j.Timestamp,
				j.Tax,
				j.Signature,
				j.Comment,
			)

			jsonString, _ := json.Marshal(transaction)

			contracts.SendTx(jsonString)
			*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *transaction)
		}
	}

	return nil
}
