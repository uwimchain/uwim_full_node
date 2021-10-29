package holder_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
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
		refundError := contracts.RefundTransaction(config.HolderScAddress, args.RecipientAddress, config.HolderGetCost,
			config.BaseToken)
		if refundError != nil {
			log.Println(fmt.Sprintf("Refund transaction %v", refundError))
		}
		return errors.New(fmt.Sprintf("error 1: getHolder %v", err))
	}

	return nil
}

func get(recipientAddress, txHash string, blockHeight int64) error {
	if !crypt.IsAddressUw(recipientAddress) && !crypt.IsAddressSmartContract(recipientAddress) &&
		!crypt.IsAddressNode(recipientAddress) {
		return errors.New("error 1: incorrect recipient address")
	}

	var holder []Holder
	holderJson := HolderDB.Get(recipientAddress).Value
	if holderJson == "" {
		return errors.New("error 2: recipient address haven`t a deposits")
	}

	err := json.Unmarshal([]byte(holderJson), &holder)
	if err != nil {
		return errors.New(fmt.Sprintf("erorr 3: %v", err))
	}

	if holder == nil {
		return errors.New("error 4: recipient address haven`t a deposits")
	}

	var (
		newHolder    []Holder
		getTxs       []contracts.Tx
		getAllAmount float64 = 0
	)

	for _, i := range holder {
		if i.RecipientAddress == recipientAddress && i.GetBlockHeight <= blockHeight {
			timestamp := apparel.TimestampUnix()

			getTxs = append(getTxs, contracts.Tx{
				Type:       5,
				Nonce:      apparel.GetNonce(strconv.FormatInt(timestamp, 10)),
				HashTx:     "",
				Height:     config.BlockHeight,
				From:       config.HolderScAddress,
				To:         recipientAddress,
				Amount:     i.Amount,
				TokenLabel: i.TokenLabel,
				Timestamp:  strconv.FormatInt(timestamp, 10),
				Tax:        0,
				Signature:  nil,
				Comment:    *contracts.NewComment("default_transaction", nil),
			})

			getAllAmount += i.Amount

			continue
		}

		newHolder = append(newHolder, i)
	}

	if getTxs == nil {
		return errors.New("error 5: recipient address haven`t a deposits")
	}

	scAddressBalance := contracts.GetBalanceForToken(config.HolderScAddress, config.BaseToken)
	if scAddressBalance.Amount < getAllAmount {
		return errors.New("error 6: Holder smart-contract address has low balance for send transactions")
	}

	timestamp := apparel.TimestampUnix()
	err = contracts.AddEvent(config.HolderScAddress, *contracts.NewEvent("Get", timestamp, config.BlockHeight,
		txHash, recipientAddress, getTxs), EventDB, ConfigDB)
	if err != nil {
		return errors.New(fmt.Sprintf("error 7: %v", err))
	}

	jsonHolder, err := json.Marshal(newHolder)
	if err != nil {
		return errors.New(fmt.Sprintf("error 8: %v", err))
	}

	HolderDB.Put(recipientAddress, string(jsonHolder))

	for _, i := range getTxs {
		txCommentSign := contracts.NewBuyTokenSign(
			config.NodeNdAddress,
		)

		contracts.SendNewScTx(i.Timestamp, i.Height, i.From, i.To, i.Amount, i.TokenLabel, i.Comment.Title, txCommentSign)
	}

	return nil
}
