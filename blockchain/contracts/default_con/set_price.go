package default_con

import (
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/crypt"
	"node/metrics"
	"strconv"
)

type SetPrice struct {
	TokenElId   int64   `json:"token_el_id"`
	NewPrice    float64 `json:"new_price"`
	TxHash      string  `json:"tx_hash"`
	BlockHeight int64   `json:"block_height"`
}

func NewSetPrice(tokenElId int64, newPrice float64, txHash string, blockHeight int64) (*SetPrice, error) {
	return &SetPrice{TokenElId: tokenElId, NewPrice: newPrice, TxHash: txHash, BlockHeight: blockHeight}, nil
}

func (args *SetPrice) SetPrice() error {
	if err := setPrice(args.TokenElId, args.NewPrice, args.TxHash, args.BlockHeight); err != nil {
		return err
	}

	return nil
}

func setPrice(tokenElId int64, newPrice float64, txHash string, blockHeight int64) error {
	tokenEl := GetNftTokenElForId(tokenElId)
	if tokenEl == nil {
		return errors.New("token does not exist")
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	parentToken := contracts.GetToken(tokenEl.ParentLabel)
	scAddress := crypt.AddressFromAnotherAddress(metrics.SmartContractPrefix, parentToken.Proposer)

	if err := contracts.AddEvent(scAddress, *contracts.NewEvent("Set price", timestamp, blockHeight,
		txHash, tokenEl.Owner, ""), EventDB, ConfigDB); err != nil {
		return err
	}

	tokenEl.Price = newPrice
	tokenEl.Update()

	return nil
}
