package trade_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/blockchain/contracts"
)

func FillConfig(args *FillConfigArgs) error {
	err := fillConfig(args.ScAddress, args.Commission)
	if err != nil {
		return errors.New(fmt.Sprintf("error 1: fillConfig %v", err))
	}

	return nil
}

func fillConfig(scAddress string, commission float64) error {
	scAddressConfigJson := ConfigDB.Get(scAddress).Value
	var (
		scAddressConfig     contracts.Config
		scAddressConfigData TradeConfig
	)
	if scAddressConfigJson != "" {
		err := json.Unmarshal([]byte(scAddressConfigJson), &scAddressConfig)
		if err != nil {
			return errors.New(fmt.Sprintf("erorr 1: %v", err))
		}

		if scAddressConfig.ConfigData != nil {
			scAddressConfigDataJson, err := json.Marshal(scAddressConfig.ConfigData)
			if err != nil {
				return errors.New(fmt.Sprintf("erorr 2: %v", err))
			}

			if scAddressConfigDataJson != nil {
				err := json.Unmarshal(scAddressConfigDataJson, &scAddressConfigData)
				if err != nil {
					return errors.New(fmt.Sprintf("erorr 3: %v", err))
				}
			}
		}
	}

	if commission < 0 || commission > 2 {
		return errors.New("error 4: commission error")
	}

	scAddressConfigData.Commission = commission

	jsonScAddressConfig, err := json.Marshal(scAddressConfig)
	if err != nil {
		return errors.New(fmt.Sprintf("error 5: %v", err))
	}

	ConfigDB.Put(scAddress, string(jsonScAddressConfig))

	return nil
}
