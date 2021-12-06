package donate_token_con

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
	"strconv"
)

type BuyArgs struct {
	ScAddress   string  `json:"sc_address"`
	UwAddress   string  `json:"uw_address"`
	Amount      float64 `json:"amount"`
	TxHash      string  `json:"tx_hash"`
	BlockHeight int64   `json:"block_height"`
}

func NewBuyArgs(scAddress string, uwAddress string, amount float64, txHash string, blockHeight int64) (*BuyArgs, error) {
	amount = apparel.Round(amount)
	return &BuyArgs{ScAddress: scAddress, UwAddress: uwAddress, Amount: amount, TxHash: txHash, BlockHeight: blockHeight}, nil
}

func Buy(args *BuyArgs) error {
	err := buy(args.ScAddress, args.UwAddress, args.TxHash, args.Amount, args.BlockHeight)
	if err != nil {
		refundError := contracts.RefundTransaction(args.ScAddress, args.UwAddress, args.Amount, config.BaseToken)
		if refundError != nil {
			log.Println(fmt.Sprintf("Refund transaction %v", refundError))
		}
		return errors.New(fmt.Sprintf("error 1: buy %v", err))
	}

	return nil
}

func buy(scAddress, uwAddress, txHash string, amount float64, blockHeight int64) error {
	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	if !crypt.IsAddressSmartContract(scAddress) {
		return errors.New("error 1: invalid smart-contract address")
	}

	if !crypt.IsAddressUw(uwAddress) {
		return errors.New("error 2: invalid uw address")
	}

	if amount <= 0 {
		return errors.New("error 3: zero or negative amount")
	}

	scAddressToken := contracts.GetTokenInfoForScAddress(scAddress)
	if scAddressToken.Id == 0 {
		return errors.New("error 4: token does not exist")
	}

	if scAddressToken.Proposer == uwAddress {
		return errors.New("error 6: sender address is a token proposer")
	}

	scAddressBalanceForToken := contracts.GetBalanceForToken(scAddress, scAddressToken.Label)
	if scAddressBalanceForToken.Amount <= 0 {
		return errors.New(fmt.Sprintf("error 7: smart-contract balance for token %s", scAddressToken.Label))
	}

	scAddressConfig := contracts.GetConfig(ConfigDB, scAddress)
	configData := scAddressConfig.GetData()
	conversion := apparel.ConvertInterfaceToFloat64(configData["conversion"])
	maxBuy := apparel.ConvertInterfaceToFloat64(configData["max_buy"])

	txAmount := conversion * amount
	if txAmount > maxBuy {
		return errors.New("error 8: tx amount more than max buy")
	}

	if scAddressBalanceForToken.Amount < txAmount {
		return errors.New(fmt.Sprintf("erorr 9: smart-contract balance for token %s", scAddressToken.Label))
	}

	scAddressBalanceForTokenUwm := contracts.GetBalanceForToken(scAddress, config.BaseToken)
	txTax := apparel.CalcTax(txAmount)
	if scAddressBalanceForTokenUwm.Amount < txTax {
		return errors.New("error 10: smart-contract balance for token uwm")
	}

	err := contracts.AddEvent(scAddress, *contracts.NewEvent("Buy", timestamp, blockHeight, txHash, uwAddress, nil), EventDB, ConfigDB)
	if err != nil {
		return errors.New(fmt.Sprintf("error 12: %v", err))
	}

	contracts.SendNewScTx(scAddress, uwAddress, txAmount, scAddressToken.Label, "default_transaction")
	return nil
}
