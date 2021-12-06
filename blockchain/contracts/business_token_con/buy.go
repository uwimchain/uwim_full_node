package business_token_con

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
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

	amount = apparel.Round(amount)
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
	// validation
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

	scAddressConfig := contracts.GetConfig(ConfigDB, scAddress)
	configData := scAddressConfig.GetData()
	configDataConversion := apparel.ConvertInterfaceToFloat64(configData["conversion"])
	configDataSalesValue := apparel.ConvertInterfaceToFloat64(configData["sales_value"])
	txAmount := apparel.Round(amount * configDataConversion)

	if txAmount > configDataSalesValue {
		return errors.New("error 9: amount more than sales value")
	}

	if scAddressBalanceForTokenUwm.Amount < amount || scAddressBalanceForTokenUwm.Amount-amount < 1 {
		return errors.New(fmt.Sprintf("error 8: smart-contract low balance for token \"%s\"", scAddressToken.Label))
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	partners := GetPartners(scAddress)

	if partners != nil {
		for idx, i := range partners {
			if i.Percent <= 0 {
				continue
			}

			partnerReward := apparel.Round(amount * (i.Percent / 100))
			if partnerReward <= 0 {
				continue
			}

			if i.Balance != nil {
				for kdx, k := range i.Balance {
					if k.TokenLabel == config.BaseToken {
						partners[idx].Balance[kdx].Amount += partnerReward
						partners[idx].Balance[kdx].UpdateTime = timestamp
					}
					break
				}
			} else {
				partners[idx].Balance = append(partners[idx].Balance, contracts.Balance{
					TokenLabel: config.BaseToken,
					Amount:     partnerReward,
					UpdateTime: timestamp,
				})
			}
		}

		partners.Update(scAddress)
	}

	if err := contracts.AddEvent(scAddress, *contracts.NewEvent("Buy", timestamp, blockHeight, txHash,
		uwAddress, newEventBuyTypeData(txAmount, configDataConversion, scAddressToken.Label)), EventDB, ConfigDB); err != nil {
		return errors.New(fmt.Sprintf("error 12: %v", err))
	}

	contracts.SendNewScTx(scAddress, uwAddress, txAmount, scAddressToken.Label, "default_transaction")
	return nil
}
