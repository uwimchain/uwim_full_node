package custom_turing_token_con

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

type DeDelegateArgs struct {
	UwAddress   string  `json:"uw_address"`
	Amount      float64 `json:"amount"`
	BlockHeight int64   `json:"block_height"`
	TxHash      string  `json:"tx_hash"`
}

func NewDeDelegateArgs(uwAddress string, amount float64, blockHeight int64, txHash string) *DeDelegateArgs {
	return &DeDelegateArgs{UwAddress: uwAddress, Amount: amount, BlockHeight: blockHeight, TxHash: txHash}
}

func DeDelegate(args *DeDelegateArgs) error {
	if err := deDelegate(args.UwAddress, args.TxHash, args.Amount, args.BlockHeight); err != nil {
		if err := contracts.RefundTransaction(ScAddress, args.UwAddress, args.Amount, TokenLabel); err != nil {
			return errors.New(fmt.Sprintf("de-delegate refund error 1: %v", err))
		}

		return errors.New(fmt.Sprintf("de-delegate error 1: %v", err))
	}

	return nil
}

func deDelegate(uwAddress, txHash string, amount float64, blockHeight int64) error {
	holderJson := HolderDB.Get(uwAddress).Value

	timestamp := apparel.TimestampUnix()
	timestampD := strconv.FormatInt(timestamp, 10)

	holder := Holder{}

	if holderJson == "" {
		return errors.New(fmt.Sprintf("De-delegate error 1: %s does not exist in smart-contract holders list", uwAddress))
	}

	_ = json.Unmarshal([]byte(holderJson), &holder)

	holder.Amount -= amount
	holder.UpdateTime = timestampD

	var txs []contracts.Tx

	if uwAddress == UwAddress {
		txs = append(txs, contracts.Tx{
			Type:       5,
			Nonce:      apparel.GetNonce(timestampD),
			HashTx:     "",
			Height:     blockHeight,
			From:       ScAddress,
			To:         uwAddress,
			Amount:     amount,
			TokenLabel: TokenLabel,
			Timestamp:  timestampD,
			Tax:        0,
			Signature:  nil,
			Comment: *contracts.NewComment(
				"default_transaction",
				nil,
			),
		})
	} else {
		//amount1 := amount - ((100-ScAddressPercent)*100)/amount
		amount1 := amount - (amount * (ScAddressPercent / 100))
		amount2 := amount - amount1

		txs = append(txs, contracts.Tx{
			Type:       5,
			Nonce:      apparel.GetNonce(timestampD),
			HashTx:     "",
			Height:     blockHeight,
			From:       ScAddress,
			To:         uwAddress,
			Amount:     amount1,
			TokenLabel: TokenLabel,
			Timestamp:  timestampD,
			Tax:        0,
			Signature:  nil,
			Comment: *contracts.NewComment(
				"default_transaction",
				nil,
			),
		}, contracts.Tx{
			Type:       5,
			Nonce:      apparel.GetNonce(timestampD),
			HashTx:     "",
			Height:     blockHeight,
			From:       ScAddress,
			To:         UwAddress,
			Amount:     amount2,
			TokenLabel: TokenLabel,
			Timestamp:  timestampD,
			Tax:        0,
			Signature:  nil,
			Comment: *contracts.NewComment(
				"default_transaction",
				nil,
			),
		})
	}

	if txs == nil {
		return errors.New("De-delegate error 2: empty transactions list")
	}

	/*for i := range txs {
		jsonString, _ := json.Marshal(contracts.Tx{
			Type:       txs[i].Type,
			Nonce:      txs[i].Nonce,
			From:       txs[i].From,
			To:         txs[i].To,
			Amount:     txs[i].Amount,
			TokenLabel: txs[i].TokenLabel,
			Comment:    txs[i].Comment,
		})
		txs[i].Signature = crypt.SignMessageWithSecretKey(config.NodeSecretKey, jsonString)

		jsonString, _ = json.Marshal(txs[i])
		txs[i].HashTx = crypt.GetHash(jsonString)
	}*/

	err := contracts.AddEvent(ScAddress, *contracts.NewEvent("De-delegate", timestamp, blockHeight, txHash, uwAddress, ""), EventDB, ConfigDB)
	if err != nil {
		return err
	}

	jsonHolder, err := json.Marshal(holder)
	if err != nil {
		return err
	}
	HolderDB.Put(uwAddress, string(jsonHolder))

	if memory.IsNodeProposer() {
		for _, i := range txs {
			txCommentSign:=contracts.NewBuyTokenSign(
				config.NodeNdAddress,
			)
			contracts.SendNewScTx(i.Timestamp, i.Height, i.From, i.To, i.Amount, i.TokenLabel, i.Comment.Title, txCommentSign)
		}
	}

	return nil
}
