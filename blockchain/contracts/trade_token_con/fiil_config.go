package trade_token_con

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"strconv"
)

type FillConfigArgs struct {
	ScAddress   string  `json:"sc_address"`
	Commission  float64 `json:"commission"`
	Changes     bool    `json:"changes"`
	TxHash      string  `json:"tx_hash"`
	BlockHeight int64   `json:"block_height"`
}

func NewFillConfigArgs(scAddress string, commission float64, changes bool, txHash string, blockHeight int64) (*FillConfigArgs, error) {
	return &FillConfigArgs{ScAddress: scAddress, Commission: commission, Changes: changes, TxHash: txHash, BlockHeight: blockHeight}, nil
}

func (args *FillConfigArgs) FillConfig() error {
	err := fillConfig(args.ScAddress, args.Commission, args.Changes, args.TxHash, args.BlockHeight)
	if err != nil {
		return errors.New(fmt.Sprintf("fill config error: %v", err))
	}

	return nil
}

func fillConfig(scAddress string, commission float64, changes bool, txHash string, blockHeight int64) error {
	scAddressConfig := contracts.GetConfig(ConfigDB, scAddress)
	configData := scAddressConfig.GetData()

	if apparel.ConvertInterfaceToBool(configData["changes"]) {
		return nil
	}

	if commission < 0 || commission > 2 {
		return errors.New("error 4: commission error")
	}

	configData["commission"] = commission
	configData["changes"] = changes

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)
	if err := contracts.AddEvent(scAddress, *contracts.NewEvent("Fill config", timestamp, blockHeight, txHash, "", nil), EventDB, ConfigDB); err != nil {
		return err
	}

	scAddressConfig.ConfigData = configData
	scAddressConfig.Update(ConfigDB, scAddress)
	return nil
}
