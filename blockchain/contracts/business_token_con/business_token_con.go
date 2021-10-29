package business_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/memory"
	"strconv"
)

var (
	db = contracts.Database{}

	ContractsDB = db.NewConnection("blockchain/contracts/business_token_con/storage/business_token_contract_contracts")
	EventDB     = db.NewConnection("blockchain/contracts/business_token_con/storage/business_token_contract_event")
	ConfigDB    = db.NewConnection("blockchain/contracts/business_token_con/storage/business_token_contract_config")
)

type Partner struct {
	Address string
	Balance []contracts.Balance
}

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
							timestampD := strconv.FormatInt(timestamp, 10)

							amount := j.Amount

							refundPartners[idx].Balance[jdx].Amount = 0
							refundPartners[idx].Balance[jdx].UpdateTime = strconv.FormatInt(timestamp, 10)

							txCommentSign := contracts.NewBuyTokenSign(
								config.NodeNdAddress,
							)

							contracts.SendNewScTx(timestampD, config.BlockHeight, scAddress, i.Address, amount, j.TokenLabel, "default_transaction", txCommentSign)
						}
					}

					for _, j := range scBalance {
						amount := j.Amount * (i.Percent / 100)

						timestamp := apparel.TimestampUnix()
						timestampD := strconv.FormatInt(timestamp, 10)

						txCommentSign:=contracts.NewBuyTokenSign(
							config.NodeNdAddress,
						)

						contracts.SendNewScTx(timestampD, config.BlockHeight, scAddress, i.Address, amount, j.TokenLabel, "default_transaction", txCommentSign)
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
					Signature:  nil,
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
						Signature:  nil,
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
				Signature:  nil,
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
			contracts.SendNewScTx(i.Timestamp, i.Height, i.From, i.To, i.Amount, i.TokenLabel, i.Comment.Title, i.Comment.Data)
		}
	}

	return nil
}
