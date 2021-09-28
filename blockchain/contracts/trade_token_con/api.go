package trade_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
	"strings"
)

func GetTokens() ([]byte, error) {
	result := make(map[string]map[string]interface{})

	tokensJson := PoolDB.GetAll("")
	if tokensJson == nil {
		return nil, nil
	}

	for _, i := range tokensJson {
		var pool Pool
		if i.Value != "" {
			err := json.Unmarshal([]byte(i.Value), &pool)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("error 1: %v", err))
			}
		}

		info := make(map[string]interface{})

		token := contracts.GetTokenInfoForScAddress(i.Key)
		if token.Id == 0 {
			return nil, errors.New("error 2: this token does not exist")
		}

		tokenInfo := make(map[string]interface{})
		tokenInfo["name"] = token.Name
		tokenInfo["label"] = token.Label

		var price float64 = 0
		var tvl float64 = 0
		if pool.FirstToken.Amount > 0 && pool.SecondToken.Amount > 0 {
			price = pool.FirstToken.Amount / pool.SecondToken.Amount

			tvl = pool.FirstToken.Amount + (pool.SecondToken.Amount * price)
		}

		info["token"] = tokenInfo
		info["price"] = price
		info["tvl"] = tvl

		result[token.Label] = info
	}

	jsonString, err := json.Marshal(result)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error 3: %v %v", err, result))
	}

	return jsonString, nil
}

func GetToken(scAddress string) ([]byte, error) {
	info := make(map[string]interface{})

	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		return nil, errors.New("error 1: token does not exist")
	}

	tokenInfo := make(map[string]interface{})
	tokenInfo["name"] = token.Name
	tokenInfo["label"] = token.Label

	var (
		pool   Pool
		events []contracts.Event
		price  float64 = 0
		tvl    float64 = 0
	)
	poolJson := PoolDB.Get(scAddress).Value
	if poolJson != "" {
		err := json.Unmarshal([]byte(poolJson), &pool)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error 2: %v", err))
		}
	}

	eventsJson := EventDB.Get(scAddress).Value
	if eventsJson != "" {
		err := json.Unmarshal([]byte(eventsJson), &events)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error 3: %v", err))
		}
	}

	if pool.FirstToken.Amount > 0 && pool.SecondToken.Amount > 0 {
		price = pool.FirstToken.Amount / pool.SecondToken.Amount

		tvl = pool.FirstToken.Amount + (pool.SecondToken.Amount * price)
	}

	info["token"] = tokenInfo
	info["pool"] = pool
	info["price"] = price
	info["tvl"] = tvl
	info["events"] = events

	jsonString, err := json.Marshal(info)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error 4: %v", err))
	}

	return jsonString, nil
}

func GetTokensForCrontab() ([]byte, error) {
	poolsJson := PoolDB.GetAll("")
	if poolsJson == nil {
		return nil, errors.New("error 1: pools is null")
	}

	tokens := make(map[string]interface{})
	for _, i := range poolsJson {
		eventsJson := EventDB.Get(i.Key).Value
		info := make(map[string]interface{})
		poolJson := i.Value
		var (
			pool   Pool
			events []contracts.Event
			price  float64 = 0
			tvl    float64 = 0
			volume float64 = 0
		)
		if poolJson != "" {
			err := json.Unmarshal([]byte(poolJson), &pool)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("error 2: %v", err))
			}
		}
		if eventsJson != "" {
			err := json.Unmarshal([]byte(eventsJson), &events)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("error 3: %v", err))
			}
		}

		token := contracts.GetTokenInfoForScAddress(i.Key)
		if token.Id == 0 {
			return nil, errors.New("error 4: token does not exist")
		}

		if pool.FirstToken.Amount > 0 && pool.SecondToken.Amount > 0 {
			price = pool.FirstToken.Amount / pool.SecondToken.Amount

			tvl = pool.FirstToken.Amount + (pool.SecondToken.Amount * price)
		}

		for _, i := range events {
			if strings.ToLower(i.Type) != "swap" {
				continue
			}

			typeData, err := getInterfaceData(i.TypeData)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("error 5: get interface data %v", err))
			}
			if typeData["first_token"] == nil || typeData["second_token"] == nil {
				continue
			}

			typeDataFirstToken, err := getInterfaceData(typeData["first_token"])
			if err != nil {
				return nil, errors.New(fmt.Sprintf("error 6: get interface data %v", err))
			}
			typeDataSecondToken, err := getInterfaceData(typeData["second_token"])
			if err != nil {
				return nil, errors.New(fmt.Sprintf("error 7: get interface data %v", err))
			}

			firstTokenAmount := apparel.ConvertInterfaceToFloat64(typeDataFirstToken["amount"])
			secondTokenAmount := apparel.ConvertInterfaceToFloat64(typeDataSecondToken["amount"])
			var course float64 = 0
			if typeDataFirstToken["token_label"] == config.BaseToken {
				if firstTokenAmount > 0 && secondTokenAmount > 0 {
					course = firstTokenAmount / secondTokenAmount
				}

				amount := firstTokenAmount + (secondTokenAmount * course)
				volume += amount

			} else if typeDataSecondToken["token_label"] == config.BaseToken {
				if firstTokenAmount > 0 && secondTokenAmount > 0 {
					course = secondTokenAmount / firstTokenAmount
				}

				amount := (firstTokenAmount * course) + secondTokenAmount
				volume += amount
			}

		}

		info["timestamp"] = apparel.TimestampUnix()
		info["price"] = price
		info["tvl"] = tvl
		info["volume"] = volume
		log.Println(info)
		tokens[token.Label] = info
	}

	jsonString, err := json.Marshal(tokens)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error 5: %v", err))
	}

	return jsonString, nil
}

func getInterfaceData(typeData interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	typeDataJson, err := json.Marshal(typeData)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error 1: %v", err))
	}
	err = json.Unmarshal(typeDataJson, &result)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error 2: %v", err))
	}
	return result, nil
}

func GetScHolder(scAddress, uwAddress string) (interface{}, error) {
	scAddressHoldersJson := HolderDB.Get(scAddress).Value
	var scAddressHolders []Holder
	if scAddressHoldersJson != "" {
		err := json.Unmarshal([]byte(scAddressHoldersJson), &scAddressHolders)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error 1: %v", err))
		}
	}

	if scAddressHolders == nil {
		return nil, errors.New("error 2: smart-contract address holders in null")
	}

	for _, i := range scAddressHolders {
		if i.Address == uwAddress {
			return i, nil
		}
	}

	return nil, errors.New("error 4: this uwim address ")
}

func GetScPool(scAddress string) (interface{}, error) {
	scAddressPoolJson := PoolDB.Get(scAddress).Value
	var scAddressPool Pool
	if scAddressPoolJson != "" {
		err := json.Unmarshal([]byte(scAddressPoolJson), &scAddressPool)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error 1: %v", err))
		}
	}
	return scAddressPool, nil
}

func GetScHolders(scAddress string) (interface{}, error) {
	scAddressHoldersJson := HolderDB.Get(scAddress).Value
	var scAddressHolders []Holder
	if scAddressHoldersJson != "" {
		err := json.Unmarshal([]byte(scAddressHoldersJson), &scAddressHolders)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error 1: %v", err))
		}
	}

	return scAddressHolders, nil
}

func GetScConfig(scAddress string) (interface{}, error) {
	scAddressConfigJson := ConfigDB.Get(scAddress).Value
	var scAddressConfig contracts.Config
	if scAddressConfigJson != "" {
		err := json.Unmarshal([]byte(scAddressConfigJson), &scAddressConfig)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("erorr 1: %v", err))
		}
	}

	return scAddressConfig, nil
}

func ValidateAdd(args *TradeArgs) int64 {
	if !crypt.IsAddressSmartContract(args.ScAddress) {
		return 511
	}

	if !crypt.IsAddressUw(args.UwAddress) {
		return 512
	}

	if args.Amount <= 0 {
		return 513
	}

	token := contracts.GetTokenInfoForScAddress(args.ScAddress)
	if token.Id == 0 {
		return 514
	}

	if token.Standard != 5 {
		return 515
	}

	if args.TokenLabel != config.BaseToken && args.TokenLabel != token.Label {
		return 516
	}

	return 0
}

func ValidateSwap(args *TradeArgs) int64 {
	if !crypt.IsAddressSmartContract(args.ScAddress) {
		return 521
	}

	if !crypt.IsAddressUw(args.UwAddress) {
		return 522
	}

	if args.Amount <= 0 {
		return 523
	}

	token := contracts.GetTokenInfoForScAddress(args.ScAddress)
	if token.Id == 0 {
		return 524
	}

	if token.Standard != 5 {
		return 525
	}

	if args.TokenLabel != config.BaseToken && args.TokenLabel != token.Label {
		return 526
	}

	scAddressPoolJson := PoolDB.Get(args.ScAddress).Value
	var scAddressPool Pool
	if scAddressPoolJson != "" {
		err := json.Unmarshal([]byte(scAddressPoolJson), &scAddressPool)
		if err != nil {
			log.Println(fmt.Sprintf("validate swap error 1: %v", err))
			return 527
		}
	}

	switch args.TokenLabel {
	case config.BaseToken:
		if scAddressPool.FirstToken.Amount-args.Amount < 1 {
			return 528
		}
		break
	case token.Label:
		if scAddressPool.SecondToken.Amount-args.Amount < 1 {
			return 529
		}
		break
	default:
		return 5210
	}

	return 0
}

func ValidateGetLiq(args *GetArgs) int64 {
	if !crypt.IsAddressSmartContract(args.ScAddress) {
		return 531
	}

	if !crypt.IsAddressUw(args.UwAddress) {
		return 532
	}

	token := contracts.GetTokenInfoForScAddress(args.ScAddress)
	if token.Id == 0 {
		return 533
	}

	if token.Standard != 5 {
		return 534
	}

	if args.TokenLabel != config.BaseToken && args.TokenLabel != token.Label {
		return 535
	}

	scAddressPoolJson := PoolDB.Get(args.ScAddress).Value
	scAddressHoldersJson := HolderDB.Get(args.ScAddress).Value
	var (
		scAddressPool    Pool
		scAddressHolders []Holder
	)
	if scAddressPoolJson != "" {
		err := json.Unmarshal([]byte(scAddressPoolJson), &scAddressPool)
		if err != nil {
			log.Println(fmt.Sprintf("validate get liq error 1: %v", err))
			return 536
		}
	}
	if scAddressHoldersJson != "" {
		err := json.Unmarshal([]byte(scAddressHoldersJson), &scAddressHolders)
		if err != nil {
			log.Println(fmt.Sprintf("validate get liq error 2: %v", err))
			return 537
		}
	}

	if scAddressHolders == nil {
		return 538
	}

	check := -1
	for idx, i := range scAddressHolders {
		if i.Address == args.UwAddress {
			switch args.TokenLabel {
			case config.BaseToken:
				if i.Pool.Liq.Amount <= 0 && i.Pool.SecondToken.Amount != 0 {
					return 539
				}

				break
			case token.Label:
				if i.Pool.Liq.Amount <= 0 && i.Pool.FirstToken.Amount != 0 {
					return 5310
				}

				break
			default:
				return 5311
			}

			check = idx
		}
	}

	if check == -1 {
		return 5312
	}

	return 0
}

func ValidateGetCom(args *GetArgs) int64 {
	if !crypt.IsAddressSmartContract(args.ScAddress) {
		return 541
	}

	if !crypt.IsAddressUw(args.UwAddress) {
		return 542
	}

	token := contracts.GetTokenInfoForScAddress(args.ScAddress)
	if token.Id == 0 {
		return 543
	}

	if token.Standard != 5 {
		return 544
	}

	if args.TokenLabel != config.BaseToken && args.TokenLabel != token.Label {
		return 545
	}

	scAddressPoolJson := PoolDB.Get(args.ScAddress).Value
	scAddressHoldersJson := HolderDB.Get(args.ScAddress).Value
	var (
		scAddressPool    Pool
		scAddressHolders []Holder
	)
	if scAddressPoolJson != "" {
		err := json.Unmarshal([]byte(scAddressPoolJson), &scAddressPool)
		if err != nil {
			log.Println(fmt.Sprintf("validate get liq error 1: %v", err))
			return 546
		}
	}

	if scAddressHoldersJson != "" {
		err := json.Unmarshal([]byte(scAddressHoldersJson), &scAddressHolders)
		if err != nil {
			log.Println(fmt.Sprintf("validate get liq error 2: %v json: %s", err, scAddressHoldersJson))
			return 547
		}
	}

	if scAddressHolders == nil {
		return 548
	}

	check := -1
	for idx, i := range scAddressHolders {
		if i.Address == args.UwAddress {
			switch args.TokenLabel {
			case config.BaseToken:
				if i.Pool.FirstToken.Commission <= 0 {
					return 549
				}

				break
			case token.Label:
				if i.Pool.SecondToken.Commission <= 0 {
					return 5410
				}

				break
			default:
				return 5411
			}

			check = idx
		}
	}

	if check == -1 {
		return 5412
	}

	return 0
}

func ValidateFillConfig(args *FillConfigArgs) int64 {
	if !crypt.IsAddressSmartContract(args.ScAddress) {
		return 551
	}

	scAddressConfigJson := ConfigDB.Get(args.ScAddress).Value
	var (
		scAddressConfig     contracts.Config
		scAddressConfigData TradeConfig
	)

	if scAddressConfigJson != "" {
		err := json.Unmarshal([]byte(scAddressConfigJson), &scAddressConfig)
		if err != nil {
			log.Println(fmt.Sprintf("validate fiil config error 1: %v", err))
		}
	}

	if scAddressConfig.ConfigData != nil {
		scAddressConfigDataJson, err := json.Marshal(scAddressConfig.ConfigData)
		if err != nil {
			log.Println(fmt.Sprintf("validate fiil config error 2: %v", err))
		}

		if scAddressConfigDataJson != nil {
			err = json.Unmarshal(scAddressConfigDataJson, &scAddressConfigData)
			if err != nil {
				log.Println(fmt.Sprintf("validate fiil config error 3: %v", err))
			}
		}
	}

	if scAddressConfigData.Commission < 0 || scAddressConfigData.Commission > 2 {
		return 552
	}

	token := contracts.GetTokenInfoForScAddress(args.ScAddress)
	if token.Id == 0 {
		return 553
	}

	if token.Standard != 5 {
		return 554
	}

	return 0
}
