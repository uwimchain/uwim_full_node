package business_token_con

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
	"strconv"
)

type GetPercentArgs struct {
	ScAddress   string  `json:"sc_address"`
	UwAddress   string  `json:"uw_address"`
	Amount      float64 `json:"amount"`
	TokenLabel  string  `json:"token_label"`
	TxHash      string  `json:"tx_hash"`
	BlockHeight int64   `json:"block_height"`
}

func NewGetPercentArgs(scAddress string, uwAddress string, amount float64, tokenLabel string, txHash string, blockHeight int64) (*GetPercentArgs, error) {
	return &GetPercentArgs{ScAddress: scAddress, UwAddress: uwAddress, Amount: amount, TokenLabel: tokenLabel, TxHash: txHash, BlockHeight: blockHeight}, nil
}

func (args *GetPercentArgs) GetPercent() error {
	err := getPercent(args.ScAddress, args.UwAddress, args.TokenLabel, args.TxHash, args.Amount, args.BlockHeight)
	if err != nil {
		return errors.New(fmt.Sprintf("error 1: %v", err))
	}

	return nil
}

func getPercent(scAddress, uwAddress, tokenLabel, txHash string, amount float64, blockHeight int64) error {
	if !crypt.IsAddressSmartContract(scAddress) {
		return errors.New(fmt.Sprintf("error 1: this address \"%s\" is not a smart-contract address", scAddress))
	}

	if !crypt.IsAddressUw(uwAddress) {
		return errors.New(fmt.Sprintf("error 2: this address \"%s\" is not a uwim address", scAddress))
	}

	if amount <= 0 {
		return errors.New("error 3: zero or negative amount")
	}

	partners := GetPartners(scAddress)
	if partners == nil {
		return errors.New("empty partners list")
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	partnerExist := false
	for idx, i := range partners {
		if i.Address == uwAddress {
			if i.Balance == nil {
				return errors.New(fmt.Sprintf("error 5: balance of \"%s\" is null", uwAddress))
			}

			tokenExist := false
			for jdx, j := range i.Balance {
				if j.TokenLabel == tokenLabel {
					if j.Amount < amount {
						return errors.New("error 6: low balance")
					}

					partners[idx].Balance[jdx].Amount -= amount
					partners[idx].Balance[jdx].UpdateTime = timestamp

					tokenExist = true
					break
				}
			}

			if !tokenExist {
				return errors.New(fmt.Sprintf("error 7: token \"%s\" does not exist on partner \"%s\" balance", tokenLabel, uwAddress))
			}
			partnerExist = true
			break
		}
	}

	if !partnerExist {
		return errors.New(fmt.Sprintf("error 8: partner \"%s\" does not exist on pertners list of \"%s\"", uwAddress, scAddress))
	}

	scAddressBalanceForToken := contracts.GetBalanceForToken(scAddress, tokenLabel)
	scAddressBalanceForTaxToken := contracts.GetBalanceForToken(scAddress, config.BaseToken)

	if scAddressBalanceForToken.TokenLabel != scAddressBalanceForTaxToken.TokenLabel {
		if scAddressBalanceForToken.Amount < amount || scAddressBalanceForToken.Amount-amount < 1 {
			return errors.New(fmt.Sprintf("error 9: smart-contract has low balance for token \"%s\"", tokenLabel))
		}
	}

	err := contracts.AddEvent(scAddress, *contracts.NewEvent("Get percent", timestamp, blockHeight, txHash, uwAddress, newEventGetPercentTypeData(amount, tokenLabel)), EventDB, ConfigDB)
	if err != nil {
		return errors.New(fmt.Sprintf("error 13: %v", err))
	}

	partners.Update(scAddress)

	contracts.SendNewScTx(scAddress, uwAddress, amount, tokenLabel, "default_transaction")

	return nil
}
