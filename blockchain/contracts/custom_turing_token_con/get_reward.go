package custom_turing_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
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

	txCommentSign := contracts.NewBuyTokenSign(
		config.NodeNdAddress,
	)

	scAddressConfigData["last_get_reward_block_height"] = blockHeight

	scAddressConfig.ConfigData = scAddressConfigData
	jsonScAddressConfig, _ := json.Marshal(scAddressConfig)
	ConfigDB.Put(ScAddress, string(jsonScAddressConfig))

	err := contracts.AddEvent(ScAddress, *contracts.NewEvent("De-delegate another address", timestamp, blockHeight, txHash, UwAddress, ""), EventDB, ConfigDB)
	if err != nil {
		return err
	}

	contracts.SendNewScTx(timestampD, config.BlockHeight, ScAddress, UwAddress, amount, TokenLabel, "default_transaction", txCommentSign)
	return nil
}
