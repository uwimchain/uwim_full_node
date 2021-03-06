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

	var (
		pool Pool
		price float64 = 0
		tvl   float64 = 0
	)
	poolJson := PoolDB.Get(scAddress).Value
	if poolJson != "" {
		err := json.Unmarshal([]byte(poolJson), &pool)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error 2: %v", err))
		}
	}

	if pool.FirstToken.Amount > 0 && pool.SecondToken.Amount > 0 {
		price = pool.FirstToken.Amount / pool.SecondToken.Amount

		tvl = pool.FirstToken.Amount + (pool.SecondToken.Amount * price)
	}

	info["pool"] = pool
	info["price"] = price
	info["tvl"] = tvl
	info["events"] = contracts.GetEvents(EventDB, scAddress)

	jsonString, err := json.Marshal(info)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error 4: %v", err))
	}

	return jsonString, nil
}

func GetConfig(scAddress string) map[string]interface{} {
	scAddressConfig := contracts.GetConfig(ConfigDB, scAddress)
	return scAddressConfig.GetData()
}

func GetScHolder(scAddress, uwAddress string) (interface{}, error) {
	scAddressHoldersJson := HolderDB.Get(scAddress).Value
	var scAddressHolders Holders
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
	var scAddressHolders Holders
	if scAddressHoldersJson != "" {
		err := json.Unmarshal([]byte(scAddressHoldersJson), &scAddressHolders)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error 1: %v", err))
		}
	}

	return scAddressHolders, nil
}

func ValidateAdd(scAddress, uwAddress string, amount float64, tokenLabel string) int64 {
	if !crypt.IsAddressSmartContract(scAddress) {
		return 511
	}

	if !crypt.IsAddressUw(uwAddress) {
		return 512
	}

	if amount <= 0 {
		return 513
	}

	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		return 514
	}

	if token.Standard != 5 {
		return 515
	}

	if tokenLabel != config.BaseToken && tokenLabel != token.Label {
		return 516
	}

	return 0
}

func ValidateSwap(scAddress, uwAddress string, amount float64, tokenLabel string) int64 {
	if !crypt.IsAddressSmartContract(scAddress) {
		return 521
	}

	if !crypt.IsAddressUw(uwAddress) {
		return 522
	}

	if amount <= 0 {
		return 523
	}

	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		return 524
	}

	if token.Standard != 5 {
		return 525
	}

	if tokenLabel != config.BaseToken && tokenLabel != token.Label {
		return 526
	}

	scAddressPoolJson := PoolDB.Get(scAddress).Value
	var scAddressPool Pool
	if scAddressPoolJson != "" {
		err := json.Unmarshal([]byte(scAddressPoolJson), &scAddressPool)
		if err != nil {
			log.Println(fmt.Sprintf("validate swap error 1: %v", err))
			return 527
		}
	}

	var course float64 = 0
	if scAddressPool.FirstToken.Amount > scAddressPool.SecondToken.Amount {
		course = scAddressPool.SecondToken.Amount / scAddressPool.FirstToken.Amount
	} else {
		course = scAddressPool.FirstToken.Amount / scAddressPool.SecondToken.Amount
	}

	txAmount := apparel.Round(amount * course)

	switch tokenLabel {
	case config.BaseToken:
		if scAddressPool.FirstToken.Amount-txAmount < 1 {
			return 528
		}
		break
	case token.Label:
		if scAddressPool.SecondToken.Amount-txAmount < 1 {
			return 529
		}
		break
	default:
		return 5210
	}

	return 0
}

func ValidateGetLiq(scAddress, uwAddress, tokenLabel string) int64 {
	if !crypt.IsAddressSmartContract(scAddress) {
		return 531
	}

	if !crypt.IsAddressUw(uwAddress) {
		return 532
	}

	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		return 533
	}

	if token.Standard != 5 {
		return 534
	}

	if tokenLabel != config.BaseToken && tokenLabel != token.Label {
		return 535
	}

	scAddressPoolJson := PoolDB.Get(scAddress).Value
	scAddressHoldersJson := HolderDB.Get(scAddress).Value
	var (
		scAddressPool    Pool
		scAddressHolders Holders
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
		if i.Address == uwAddress {
			switch tokenLabel {
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

func ValidateGetCom(scAddress, uwAddress, tokenLabel string) int64 {
	if !crypt.IsAddressSmartContract(scAddress) {
		return 541
	}

	if !crypt.IsAddressUw(uwAddress) {
		return 542
	}

	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		return 543
	}

	if token.Standard != 5 {
		return 544
	}

	if tokenLabel != config.BaseToken && tokenLabel != token.Label {
		return 545
	}

	scAddressPoolJson := PoolDB.Get(scAddress).Value
	scAddressHoldersJson := HolderDB.Get(scAddress).Value
	var (
		scAddressPool    Pool
		scAddressHolders Holders
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
		if i.Address == uwAddress {
			switch tokenLabel {
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

func ValidateFillConfig(senderAddress, recipientAddress string, commission, amount float64, tokenLabel string) int {
	if recipientAddress != config.MainNodeAddress {
		return 551
	}

	if !crypt.IsAddressUw(senderAddress) {
		return 552
	}

	if amount != config.FillTokenConfigCost {
		return 553
	}

	if tokenLabel != config.BaseToken {
		return 554
	}

	if commission < 0 || commission > 2 {
		return 555
	}

	address := contracts.GetAddress(senderAddress)
	token := address.GetToken()
	if token.Id == 0 {
		return 556
	}

	if token.Standard != 5 {
		return 557
	}

	return 0
}
