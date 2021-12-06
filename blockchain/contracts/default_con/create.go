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

type Create struct {
	Owner       string  `json:"owner"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Data        string  `json:"data"`
	TxHash      string  `json:"tx_hash"`
	BlockHeight int64   `json:"block_height"`
}

func NewCreate(owner, name string, price float64, data string, txHash string, blockHeight int64) (*Create, error) {
	return &Create{
		Owner:       owner,
		Name:        name,
		Price:       price,
		Data:        data,
		TxHash:      txHash,
		BlockHeight: blockHeight,
	}, nil
}

func (args *Create) Create() error {
	if err := create(args.Owner, args.Name, args.Price, args.Data, args.TxHash, args.BlockHeight); err != nil {
		if refundError := contracts.RefundTransaction(config.MainNodeAddress, args.Owner, config.NftCreateCost,
			config.BaseToken); refundError != nil {
			log.Println(fmt.Sprintf("Refund transaction %v", refundError))
		}

		return errors.New(fmt.Sprintf("error 1: buy %v", err))
	}
	return nil
}

func create(owner, name string, price float64, data, txHash string, blockHeight int64) error {
	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	tokenEl := NewNftTokenEl(name, owner, price, data, timestamp)

	scAddress := crypt.AddressFromAnotherAddress(metrics.SmartContractPrefix, owner)
	if err := contracts.AddEvent(scAddress, *contracts.NewEvent("Create", timestamp, blockHeight, txHash, owner, ""), EventDB, ConfigDB); err != nil {
		return err
	}

	tokenEl.Create()
	return nil
}
