package business_token_con

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

var (
	db = contracts.Database{}

	//TxDB        = db.NewConnection("blockchain/contracts/business_token_con/storage/business_token_contract_tx")
	ContractsDB = db.NewConnection("blockchain/contracts/business_token_con/storage/business_token_contract_contracts")
	EventDB     = db.NewConnection("blockchain/contracts/business_token_con/storage/business_token_contract_event")
	ConfigDB    = db.NewConnection("blockchain/contracts/business_token_con/storage/business_token_contract_config")
	//LogsDb      = db.NewConnection("blockchain/contracts/business_token_con/storage/business_token_contract_logs")
)

/*type BusinessSmartData struct {
	ReqType string `json:"req_type"`
	Data    []byte `json:"data"`
}*/

/*type BuyTokensData struct {
	TxFrom   string  `json:"tx_from"`
	TxTo     string  `json:"tx_to"`
	TxAmount float64 `json:"tx_amount"`
}*/

/*type TakePercentAmountData struct {
	TxFrom                             string  `json:"tx_from"`
	TxTo                               string  `json:"tx_to"`
	TxCommentDataTakePercentAmount     float64 `json:"tx_comment_data_take_percent_amount"`
	TxCommentDataTakePercentTokenLabel string  `json:"tx_comment_data_take_percent_token_label"`
}*/

/*func Smart(smartData BusinessSmartData) error {
	timestamp := apparel.TimestampUnix()
	switch smartData.ReqType {
	case "buy_tokens":
	buyData := BuyTokensData{}
	err := json.Unmarshal(smartData.Data, &buyData)
	if err != nil {
		return err
	}

	if crypt.IsAddressSmartContract(buyData.TxTo) {
		publicKey, err := crypt.PublicKeyFromAddress(buyData.TxTo)
		if err != nil {
			return err
		}

		uwAddress := crypt.AddressFromPublicKey(metrics.AddressPrefix, publicKey)
		if contracts.CheckAddressToken(uwAddress) {
			token := contracts.GetAddressToken(uwAddress)

			if token.Standard == 4 {
				address := contracts.GetAddress(uwAddress)

				for _, el := range address.Balance {
					if el.TokenLabel == token.Label {
						//tokenStandardCard := contracts.BusinessStandardCardData
						tokenStandardCard := contracts.BusinessStandardCardData{}
						err := json.Unmarshal([]byte(token.StandardCard), &tokenStandardCard)
						if err != nil {
							return err
						}

						amount := tokenStandardCard.Conversion * buyData.TxAmount
						tax := apparel.CalcTax(amount * config.Tax)
						nonce := apparel.GetNonce(apparel.UnixToString(timestamp))

						commentSign, err := json.Marshal(contracts.NewBuyTokenSign(
							config.NodeNdAddress,
						))
						if err != nil {
							return err
						}

						transaction := contracts.NewTx(
							1,
							nonce,
							"",
							config.BlockHeight,
							uwAddress,
							buyData.TxFrom,
							amount,
							el.TokenLabel,
							apparel.UnixToString(timestamp),
							tax,
							crypt.SignMessageWithSecretKey(
								config.NodeSecretKey,
								[]byte(config.NodeNdAddress),
							),
							*contracts.NewComment(
								"default_transaction",
								commentSign,
							),
						)

						jsonString, err := json.Marshal(transaction)
						if err != nil {
							return err
						}

						newTxs(nonce, jsonString)

						if tokenStandardCard.Partners != nil {
							for _, partner := range tokenStandardCard.Partners {
								amount := buyData.TxAmount * (partner.Percent / 100)
								updateBalance(
									buyData.TxTo,
									partner.Address,
									amount,
									config.BaseToken,
									timestamp,
									true,
								)
							}
						}

						if memory.IsNodeProposer() {
							contracts.SendTx(jsonString)
							*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *transaction)
						}
						break
					}
				}
			}
		}
	}
	break
	case "take_percent_amount":
		takePercentAmountData := TakePercentAmountData{}
		err := json.Unmarshal(smartData.Data, &takePercentAmountData)
		if err != nil {
			return err
		}

		if crypt.IsAddressSmartContract(takePercentAmountData.TxTo) {
			publicKey, err := crypt.PublicKeyFromAddress(takePercentAmountData.TxTo)
			if err != nil {
				return err
			}

			uwAddress := crypt.AddressFromPublicKey(metrics.AddressPrefix, publicKey)
			if contracts.CheckAddressToken(uwAddress) {
				token := contracts.GetAddressToken(uwAddress)

				if token.Standard == 4 {
					jsonString := ContractsDB.Get(takePercentAmountData.TxTo).Value
					var partners []Partner
					err := json.Unmarshal([]byte(jsonString), &partners)
					if err != nil {
						return err
					}

					for _, partner := range partners {
						if partner.Address == takePercentAmountData.TxFrom {
							for _, coin := range partner.Balance {
								if coin.TokenLabel == takePercentAmountData.TxCommentDataTakePercentTokenLabel {
									if coin.Amount < takePercentAmountData.TxCommentDataTakePercentAmount {
										return errors.New("low balance")
									}

									updateBalance(takePercentAmountData.TxTo, takePercentAmountData.TxFrom, takePercentAmountData.TxCommentDataTakePercentAmount, takePercentAmountData.TxCommentDataTakePercentTokenLabel, timestamp, false)

									amount := takePercentAmountData.TxCommentDataTakePercentAmount
									tax := apparel.CalcTax(amount * config.Tax)
									nonce := apparel.GetNonce(apparel.UnixToString(timestamp))

									commentSign, err := json.Marshal(contracts.NewBuyTokenSign(
										config.NodeNdAddress,
									))
									if err != nil {
										return err
									}

									transaction := contracts.NewTx(
										5,
										nonce,
										"",
										config.BlockHeight,
										uwAddress,
										takePercentAmountData.TxFrom,
										amount,
										takePercentAmountData.TxCommentDataTakePercentTokenLabel,
										apparel.UnixToString(timestamp),
										tax,
										crypt.SignMessageWithSecretKey(
											config.NodeSecretKey,
											[]byte(config.NodeNdAddress),
										),
										*contracts.NewComment(
											"refund_transaction",
											commentSign,
										),
									)

									jsonString, err := json.Marshal(transaction)
									if err != nil {
										return err
									}

									newTxs(nonce, jsonString)

									if memory.IsNodeProposer() {
										contracts.SendTx(jsonString)
										*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *transaction)
									}

									break
								}
							}

							break
						}
					}
				}
			}
		}

		break
	}

	return nil
}*/

/*func newTxs(nonce int64, jsonString []byte) {
	TxDB.Put(strconv.FormatInt(nonce, 10), string(jsonString))
}*/

type Partner struct {
	Address string
	Balance []contracts.Balance
}

/*func updateBalance(scAddress string, address string, amount float64, tokenLabel string, timestamp int64, side bool) {
	jsonString := ContractsDB.Get(scAddress).Value
	if jsonString == "" {
		log(scAddress, contracts.Log{
			Timestamp: timestamp,
			Record:    "Update balance error: record is null",
		})
	} else {
		var partners []Partner
		err := json.Unmarshal([]byte(jsonString), &partners)
		if err != nil {
			log(scAddress, contracts.Log{
				Timestamp: timestamp,
				Record:    "Update balance error: json unmarshal",
			})
		} else {

			for _, partner := range partners {
				if partner.Address == address {
					if side == true {
						if searchTokenInBalance(partner.Balance, tokenLabel) {
							for _, token := range partner.Balance {
								if token.TokenLabel == tokenLabel {
									token.Amount += amount
								}
							}
							break
						} else {
							partner.Balance = append(partner.Balance, contracts.Balance{
								TokenLabel: tokenLabel,
								Amount:     amount,
								UpdateTime: apparel.UnixToString(timestamp),
							})
							break
						}
					} else {
						if searchTokenInBalance(partner.Balance, tokenLabel) {
							for _, token := range partner.Balance {
								if token.TokenLabel == tokenLabel {
									token.Amount += amount
								}
							}
							break
						}
					}
				}
			}

			jsonString, err := json.Marshal(partners)
			if err != nil {
				log(scAddress, contracts.Log{
					Timestamp: timestamp,
					Record:    "Update balance error: json marshal",
				})
			} else {
				ContractsDB.Put(scAddress, string(jsonString))
			}
		}
	}

}*/

/*func log(scAddress string, log contracts.Log) {
	logString, _ := json.Marshal(log)

	LogsDb.Put(scAddress, string(logString))
}*/

/*func searchTokenInBalance(balance []contracts.Balance, tokenLabel string) bool {
	if balance == nil {
		return false
	}

	for _, token := range balance {
		if token.TokenLabel == tokenLabel {
			return true
		}
	}

	return false
}*/

func UpdatePartners(scAddress string) error {
	// get partners list on business smart-contract
	var partnersOnScAddress []Partner
	partnersOnScAddressJson := ContractsDB.Get(scAddress).Value
	if partnersOnScAddressJson != "" {
		err := json.Unmarshal([]byte(partnersOnScAddressJson), &partnersOnScAddress)
		if err != nil {
			return errors.New(fmt.Sprintf("error 1: %v", err))
		}
	}

	// get partners on token standard card data
	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		return errors.New("error 2: token does not exist")
	}

	if token.Standard != 4 {
		return errors.New(fmt.Sprintf("error 3: token standard does not a business. token standard is %d", token.Standard))
	}

	if token.StandardCard == "" {
		return errors.New("error 4: token standard card dont filled")
	}

	var tokenStandardCard contracts.BusinessStandardCardData
	err := json.Unmarshal([]byte(token.StandardCard), &tokenStandardCard)
	if err != nil {
		return errors.New(fmt.Sprintf("error 5: %v", err))
	}

	partnersOnTokenStandard := tokenStandardCard.Partners

	var newPartnersOnScAddress []Partner
	for _, i := range partnersOnTokenStandard {
		newPartnersOnScAddress = append(newPartnersOnScAddress, Partner{
			Address: i.Address,
			Balance: nil,
		})
	}

	newPartnersOnScAddressJson, err := json.Marshal(newPartnersOnScAddress)
	if err != nil {
		return errors.New(fmt.Sprintf("error 6: %v", err))
	}

	// refund old partners percent
	if partnersOnScAddress != nil {
		type RefundPartner struct {
			Address string              `json:"address"`
			Percent float64             `json:"percent"`
			Balance []contracts.Balance `json:"balance"`
		}

		var refundPartners []RefundPartner

		for _, i := range partnersOnTokenStandard {
			for _, j := range partnersOnScAddress {
				if i.Address == j.Address {
					refundPartners = append(refundPartners, RefundPartner{
						Address: i.Address,
						Percent: i.Percent,
						Balance: j.Balance,
					})
					break
				}
			}
		}

		if refundPartners != nil {
			scBalance := contracts.GetBalance(scAddress)

			if scBalance != nil {


				for idx, i := range refundPartners {
					if i.Balance != nil {
						for jdx, j := range i.Balance {

							timestamp := apparel.TimestampUnix()

							amount := j.Amount

							refundPartners[idx].Balance[jdx].Amount = 0
							refundPartners[idx].Balance[jdx].UpdateTime = strconv.FormatInt(timestamp, 10)

							tax := apparel.CalcTax(amount)

							commentSign, err := json.Marshal(contracts.NewBuyTokenSign(
								config.NodeNdAddress,
							))
							if err != nil {
								return errors.New(fmt.Sprintf("error 7: %v", err))
							}

							transaction := contracts.NewTx(
								5,
								apparel.GetNonce(strconv.FormatInt(timestamp, 10)),
								"",
								config.BlockHeight,
								scAddress,
								i.Address,
								amount,
								j.TokenLabel,
								strconv.FormatInt(timestamp, 10),
								tax,
								crypt.SignMessageWithSecretKey(
									config.NodeSecretKey,
									[]byte(config.NodeNdAddress),
								),
								*contracts.NewComment(
									"refund_transaction",
									commentSign,
								),
							)

							jsonString, err := json.Marshal(transaction)
							if err != nil {
								return errors.New(fmt.Sprintf("error 8: %v", err))
							}

							if memory.IsNodeProposer() {
								contracts.SendTx(jsonString)
								*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *transaction)
							}
						}
					}

					for _, j := range scBalance {
						amount := j.Amount * (i.Percent / 100)

						timestamp := apparel.TimestampUnix()

						tax := apparel.CalcTax(amount)
						nonce := apparel.GetNonce(apparel.UnixToString(timestamp))

						commentSign, err := json.Marshal(contracts.NewBuyTokenSign(
							config.NodeNdAddress,
						))
						if err != nil {
							return errors.New(fmt.Sprintf("error 7: %v", err))
						}

						transaction := contracts.NewTx(
							5,
							nonce,
							"",
							config.BlockHeight,
							scAddress,
							i.Address,
							amount,
							j.TokenLabel,
							strconv.FormatInt(timestamp, 10),
							tax,
							crypt.SignMessageWithSecretKey(
								config.NodeSecretKey,
								[]byte(config.NodeNdAddress),
							),
							*contracts.NewComment(
								"refund_transaction",
								commentSign,
							),
						)

						jsonString, err := json.Marshal(transaction)
						if err != nil {
							return errors.New(fmt.Sprintf("error 8: %v", err))
						}

						if memory.IsNodeProposer() {
							contracts.SendTx(jsonString)
							*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *transaction)
						}
					}
				}
			}
		}
	}

	ContractsDB.Put(scAddress, string(newPartnersOnScAddressJson))

	return nil
}

func ChangeStandard(scAddress string) error {
	var partners []Partner
	jsonPartners := ContractsDB.Get(scAddress).Value
	if jsonPartners != "" {
		err := json.Unmarshal([]byte(jsonPartners), &partners)
		if err != nil {
			return errors.New(fmt.Sprintf("Business token contract error 1: %v", err))
		}
	}

	if partners == nil {
		return nil
	}

	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		return errors.New("Business token contract error 2: token does not exist")
	}

	var tokenStandard contracts.BusinessStandardCardData
	if token.StandardCard != "" {
		err := json.Unmarshal([]byte(token.StandardCard), &tokenStandard)
		if err != nil {
			return errors.New(fmt.Sprintf("Business token contract error 3: %v", err))
		}
	}

	var txs []contracts.Tx
	scBalance := contracts.GetBalance(scAddress)
	var allTxsAmount []contracts.Balance

	if partners != nil {
		for _, i := range partners {
			if i.Balance == nil {
				continue
			}

			for _, j := range i.Balance {

				var addressPercent float64 = 0

				for _, n := range tokenStandard.Partners {
					if i.Address == n.Address {
						addressPercent = n.Percent
						break
					}
				}

				var percentAmount float64 = 0
				for _, n := range scBalance {
					if n.TokenLabel == j.TokenLabel {
						percentAmount = n.Amount * (addressPercent / 100)
						break
					}
				}

				timestampD := strconv.FormatInt(apparel.TimestampUnix(), 10)

				tx := contracts.Tx{
					Type:       5,
					Nonce:      apparel.GetNonce(timestampD),
					HashTx:     "",
					Height:     config.BlockHeight,
					From:       scAddress,
					To:         i.Address,
					Amount:     j.Amount + percentAmount,
					TokenLabel: j.TokenLabel,
					Timestamp:  timestampD,
					Tax:        0,
					Signature:  crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
					Comment: *contracts.NewComment(
						"refund_transaction",
						nil,
					),
				}

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

				txs = append(txs, tx)
			}
		}
	}

	if allTxsAmount != nil {
		for _, i := range allTxsAmount {
			for _, j := range scBalance {
				if i.TokenLabel == j.TokenLabel {
					timestampD := strconv.FormatInt(apparel.TimestampUnix(), 10)

					tx := contracts.Tx{
						Type:       5,
						Nonce:      apparel.GetNonce(timestampD),
						HashTx:     "",
						Height:     config.BlockHeight,
						From:       scAddress,
						To:         token.Proposer,
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
				}
			}
		}
	} else {
		for _, i := range scBalance {
			timestampD := strconv.FormatInt(apparel.TimestampUnix(), 10)

			tx := contracts.Tx{
				Type:       5,
				Nonce:      apparel.GetNonce(timestampD),
				HashTx:     "",
				Height:     config.BlockHeight,
				From:       scAddress,
				To:         token.Proposer,
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
		for _, i := range txs {
			transaction := contracts.NewTx(
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

			jsonString, _ := json.Marshal(transaction)

			contracts.SendTx(jsonString)
			*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *transaction)
		}
	}

	return nil
}
