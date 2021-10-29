package bridge_con

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/blockchain/contracts"
	"node/config"
)

type BuyFromEthArgs struct {
	Address string
	Amount  float64
}

func (args *BuyFromEthArgs) BuyFromEth() error {
	if err := buyFromEth(args.Amount, args.Address); err != nil {
		if refundError := contracts.RefundTransaction(ScAddress, args.Address, args.Amount,
			config.BaseToken); refundError != nil {
			log.Println(fmt.Sprintf("Refund transaction %v", refundError))
		}

		return errors.New(fmt.Sprintf("error 1: buy from eth %v", err))
	}

	return nil
}

func buyFromEth(amount float64, address string) error {

	return nil
}
