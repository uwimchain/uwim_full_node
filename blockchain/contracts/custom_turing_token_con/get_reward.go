package custom_turing_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
	"node/memory"
	"strconv"
)

type GetRewardArgs struct {
	BlockHeight int64  `json:"block_height"`
	TxHash      string `json:"tx_hash"`
}

func NewGetRewardArgs(blockHeight int64, txHash string) *GetRewardArgs {
	return &GetRewardArgs{BlockHeight: blockHeight, TxHash: txHash}
}

func GetReward(args *GetRewardArgs) error {
	if err := getReward(args.TxHash, args.BlockHeight); err != nil {
		return errors.New(fmt.Sprintf("get reward error 1: %v", err))
	}

	return nil
}

func getReward(txHash string, blockHeight int64) error {
	scAddressConfigJson := ConfigDB.Get(ScAddress).Value
	scAddressConfig := contracts.Config{}

	_ = json.Unmarshal([]byte(scAddressConfigJson), &scAddressConfig)

	scAddressConfigDataJson, _ := json.Marshal(scAddressConfig.ConfigData)
	scAddressConfigData := make(map[string]interface{})

	_ = json.Unmarshal(scAddressConfigDataJson, &scAddressConfigData)

	lastGetRewardBlockHeight := apparel.ConvertInterfaceToInt64(scAddressConfigData["last_get_reward_block_height"])
	if lastGetRewardBlockHeight == 0 {
		lastGetRewardBlockHeight = blockHeight
	}

	timestamp := apparel.TimestampUnix()
	timestampD := strconv.FormatInt(timestamp, 10)

	amount := ((24 * 60 * 60 * 60) * float64(lastGetRewardBlockHeight)) * 0.1

	txCommentSign, _ := json.Marshal(contracts.NewBuyTokenSign(
		config.NodeNdAddress,
	))

	tx := contracts.NewTx(
		5,
		apparel.GetNonce(timestampD),
		"",
		blockHeight,
		ScAddress,
		UwAddress,
		amount,
		TokenLabel,
		timestampD,
		0,
		nil,
		*contracts.NewComment("default_transaction", txCommentSign))

	jsonString, _ := json.Marshal(contracts.Tx{
		Type:       tx.Type,
		Nonce:      tx.Nonce,
		From:       tx.From,
		To:         tx.To,
		Amount:     tx.Amount,
		TokenLabel: tx.TokenLabel,
		Comment:    tx.Comment,
	})
	tx.Signature = crypt.SignMessageWithSecretKey(config.NodeSecretKey, jsonString)

	jsonString, _ = json.Marshal(tx)
	tx.HashTx = crypt.GetHash(jsonString)

	scAddressConfigData["last_get_reward_block_height"] = blockHeight

	scAddressConfig.ConfigData = scAddressConfigData
	jsonScAddressConfig, _ := json.Marshal(scAddressConfig)
	ConfigDB.Put(ScAddress, string(jsonScAddressConfig))

	err := contracts.AddEvent(ScAddress, *contracts.NewEvent("De-delegate another address", timestamp, blockHeight, txHash, UwAddress, ""), EventDB, ConfigDB)
	if err != nil {
		return err
	}

	if memory.IsNodeProposer() {
		contracts.SendTx(*tx)
		*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *tx)
	}
	return nil
}
