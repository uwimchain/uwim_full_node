package default_con

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
	"node/metrics"
	"strconv"
)

type Buy struct {
	TokenElId   int64   `json:"token_el_id"`
	TxHash      string  `json:"tx_hash"`
	Buyer       string  `json:"buyer"`
	Amount      float64 `json:"amount"`
	BlockHeight int64   `json:"block_height"`
}

func NewBuy(tokenElId int64, txHash string, buyer string, amount float64, blockHeight int64) (*Buy, error) {
	return &Buy{TokenElId: tokenElId, TxHash: txHash, Buyer: buyer, Amount: amount, BlockHeight: blockHeight}, nil
}

func (args *Buy) Buy() error {
	if err := buy(args.TokenElId, args.Buyer, args.TxHash, args.BlockHeight); err != nil {
		tokenEl := GetNftTokenElForId(args.TokenElId)
		scAddress := crypt.AddressFromAnotherAddress(metrics.SmartContractPrefix, tokenEl.Owner)
		if refundError := contracts.RefundTransaction(scAddress, args.Buyer, args.Amount,
			config.BaseToken); refundError != nil {
			log.Println(fmt.Sprintf("Refund transaction %v", refundError))
		}

		return errors.New(fmt.Sprintf("error 1: buy %v", err))
	}

	return nil
}

func buy(tokenElId int64, buyer, txHash string, blockHeight int64) error {
	tokenEl := GetNftTokenElForId(tokenElId)

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	parentToken := contracts.GetToken(tokenEl.ParentLabel)
	scAddress := crypt.AddressFromAnotherAddress(metrics.SmartContractPrefix, parentToken.Proposer)

	if err := contracts.AddEvent(scAddress, *contracts.NewEvent("Buy", timestamp, blockHeight,
		txHash, buyer, nil), EventDB, ConfigDB); err != nil {
		return err
	}

	parentTokenStandardCard := parentToken.GetStandardCard()
	commission := apparel.ConvertInterfaceToFloat64(parentTokenStandardCard["commission"])

	if commission != 0 && tokenEl.Owner != parentToken.Proposer {
		commissionAmount := tokenEl.Price * (commission / 100)
		contracts.SendNewScTx(scAddress, parentToken.Proposer, commissionAmount, config.BaseToken, "default_transaction")

		contracts.SendNewScTx(scAddress, tokenEl.Owner, tokenEl.Price-commissionAmount, config.BaseToken, "default_transaction")
	} else {
		contracts.SendNewScTx(scAddress, tokenEl.Owner, tokenEl.Price, config.BaseToken, "default_transaction")
	}

	tokenEl.Owner = buyer
	tokenEl.Price = 0
	tokenEl.Update()
	return nil
}
