package holder_con

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
	"strconv"
)

type GetArgs struct {
	RecipientAddress string `json:"recipient_address"`
	TxHash           string `json:"tx_hash"`
	BlockHeight      int64  `json:"block_height"`
}

func NewGetArgs(recipientAddress string, txHash string, blockHeight int64) (*GetArgs, error) {
	return &GetArgs{RecipientAddress: recipientAddress, TxHash: txHash, BlockHeight: blockHeight}, nil
}

func Get(args *GetArgs) error {
	err := get(args.RecipientAddress, args.TxHash, args.BlockHeight)
	if err != nil {
		return errors.New(fmt.Sprintf("error 1: getHolder %v", err))
	}

	return nil
}

func get(recipientAddress, txHash string, blockHeight int64) error {
	if !crypt.IsAddressUw(recipientAddress) && !crypt.IsAddressSmartContract(recipientAddress) &&
		!crypt.IsAddressNode(recipientAddress) {
		return errors.New("error 1: incorrect recipient address")
	}

	holders := GetHolder(recipientAddress)

	if holders == nil {
		return errors.New("error 4: recipient address haven`t a deposits")
	}

	var getAmount float64 = 0

	var deleteHoldersIds []int
	for idx, i := range holders {
		if i.RecipientAddress == recipientAddress && i.GetBlockHeight <= blockHeight {
			getAmount += i.Amount
			deleteHoldersIds = append(deleteHoldersIds, idx)
		}
	}

	if deleteHoldersIds == nil {
		return errors.New("error 5: recipient address haven`t a deposits")
	}

	newHolders := Holders{}
	for idx := range holders {
		check := false
		for _, i := range deleteHoldersIds {
			if idx == i {
				check = true
			}
		}

		if !check {
			newHolders = append(newHolders, holders[idx])
		}
	}

	holders = newHolders

	scAddressBalance := contracts.GetBalanceForToken(config.HolderScAddress, config.BaseToken)
	if scAddressBalance.Amount < getAmount {
		return errors.New("error 6: Holder smart-contract address has low balance for send transactions")
	}

	timestamp := apparel.TimestampUnix()
	err := contracts.AddEvent(config.HolderScAddress, *contracts.NewEvent("Get", timestamp, config.BlockHeight,
		txHash, recipientAddress, nil), EventDB, ConfigDB)
	if err != nil {
		return errors.New(fmt.Sprintf("error 7: %v", err))
	}


	holders.Update(recipientAddress)

	txCommentSign := contracts.NewBuyTokenSign(
		config.NodeNdAddress,
	)
	contracts.SendNewScTx(strconv.FormatInt(timestamp, 10), config.BlockHeight, config.HolderScAddress, recipientAddress, getAmount, config.BaseToken, "default_transaction", txCommentSign)

	return nil
}
