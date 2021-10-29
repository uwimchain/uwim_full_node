package trade_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"strconv"
)

// function for swap uwm on scToken
func Swap(args *TradeArgs) error {
	err := swap(args.ScAddress, args.UwAddress, args.TokenLabel, args.TxHash, args.Amount, args.BlockHeight)
	if err != nil {
		refundError := contracts.RefundTransaction(args.ScAddress, args.UwAddress, args.Amount, args.TokenLabel)
		if refundError != nil {
			log.Println(fmt.Sprintf("Refund transaction %v", refundError))
		}
		return errors.New(fmt.Sprintf("error 1: swapPool %v", err))
	}

	return nil
}

func swap(scAddress, uwAddress, tokenLabel, txHash string, amount float64, blockHeight int64) error {
	scAddressPoolJson := PoolDB.Get(scAddress).Value
	scAddressConfigJson := ConfigDB.Get(scAddress).Value
	scAddressHoldersJson := HolderDB.Get(scAddress).Value

	var (
		scAddressPool       Pool
		scAddressConfig     contracts.Config
		scAddressConfigData TradeConfig
		scAddressHolders    []Holder
	)
	if scAddressPoolJson != "" {
		err := json.Unmarshal([]byte(scAddressPoolJson), &scAddressPool)
		if err != nil {
			return errors.New(fmt.Sprintf("error 1: %v", err))
		}
	}
	if scAddressConfigJson != "" {
		err := json.Unmarshal([]byte(scAddressConfigJson), &scAddressConfig)
		if err != nil {
			return errors.New(fmt.Sprintf("error 2: %v", err))
		}

		if scAddressConfig.ConfigData != nil {
			scAddressConfigDataJson, err := json.Marshal(scAddressConfig.ConfigData)
			if err != nil {
				return errors.New(fmt.Sprintf("error 3: %v", err))
			}

			if scAddressConfigDataJson != nil {
				err := json.Unmarshal(scAddressConfigDataJson, &scAddressConfigData)
				if err != nil {
					return errors.New(fmt.Sprintf("error 4: %v", err))
				}
			}
		}
	}
	if scAddressHoldersJson != "" {
		err := json.Unmarshal([]byte(scAddressHoldersJson), &scAddressHolders)
		if err != nil {
			return errors.New(fmt.Sprintf("error 5: %v", err))
		}
	}

	if scAddressConfigData.Commission < 0 {
		return errors.New("error 6: token commission is null")
	}

	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		return errors.New(fmt.Sprintf("error 7: token of this smart-contract \"%s\" does not exist", scAddress))
	}

	timestamp := apparel.TimestampUnix()
	timestampD := strconv.FormatInt(timestamp, 10)

	var (
		txAmount float64 = 0
		course   float64 = 0
		tax      float64 = 0
	)

	if scAddressPool.SecondToken.Amount > 0 && scAddressPool.FirstToken.Amount > 0 {
		if scAddressPool.FirstToken.Amount > scAddressPool.SecondToken.Amount {
			course = scAddressPool.SecondToken.Amount / scAddressPool.FirstToken.Amount
		} else {
			course = scAddressPool.FirstToken.Amount / scAddressPool.SecondToken.Amount
		}
	}

	txTokenLabel := ""
	switch tokenLabel {
	case config.BaseToken:
		txTokenLabel = token.Label
		txAmount = apparel.Round(amount * course)
		//log.Println(fmt.Sprintf("cource amount if you send uwm for swap: %g", course))
		//log.Println(fmt.Sprintf("tx amount if you send uwm for swap: %g %s", txAmount, txTokenLabel))
		tax = txAmount * (scAddressConfigData.Commission / 100)

		scAddressPool.FirstToken.Amount += amount
		scAddressPool.FirstToken.UpdateTime = timestamp

		scAddressPool.SecondToken.Amount -= txAmount
		if scAddressPool.SecondToken.Amount < 1 {
			return errors.New(fmt.Sprintf("error 8: low balance for token %s,  txAmount: %g,  amount: %g,  cource: %g",
				txTokenLabel, txAmount, amount, course))
		}
		scAddressPool.SecondToken.UpdateTime = timestamp
		break
	case token.Label:
		txTokenLabel = config.BaseToken
		txAmount = apparel.Round(amount / course)
		tax = txAmount * (scAddressConfigData.Commission / 100)

		scAddressPool.FirstToken.Amount -= txAmount
		if scAddressPool.FirstToken.Amount < 1 {
			return errors.New(fmt.Sprintf("error 9: low balance for token %s,  txAmount: %g,  amount: %g,  cource: %g",
				txTokenLabel, txAmount, amount, course))
		}
		scAddressPool.FirstToken.UpdateTime = timestamp

		scAddressPool.SecondToken.Amount += amount - tax
		scAddressPool.SecondToken.UpdateTime = timestamp
		break
	}
	if tax != 0 && scAddressHolders != nil {
		var holdersReward float64 = 0
		for _, i := range scAddressHolders {
			if i.Pool.Liq.Amount == 0 {
				continue
			}

			holderPercent := ((i.Pool.Liq.Amount * 100) / scAddressPool.Liq.Amount) / 100
			holdersReward += tax * holderPercent
			break
		}

		if holdersReward <= tax {
			for idx, i := range scAddressHolders {
				if i.Pool.Liq.Amount == 0 {
					continue
				}

				holderPercent := ((i.Pool.Liq.Amount * 100) / scAddressPool.Liq.Amount) / 100
				holderReward := tax * holderPercent

				switch tokenLabel {
				case config.BaseToken:
					scAddressHolders[idx].Pool.SecondToken.Commission += holderReward
					break
				case token.Label:
					scAddressHolders[idx].Pool.FirstToken.Commission += holderReward
					break
				default:
					return errors.New(fmt.Sprintf("error 10: unexpected token for this smart-contract \"%s\"", scAddress))
				}
			}
		}
	}

	err := contracts.AddEvent(scAddress, *contracts.NewEvent("Swap", timestamp, blockHeight, txHash, uwAddress, newEventSwapTypeData(amount, course, tokenLabel, txTokenLabel)), EventDB, ConfigDB)
	if err != nil {
		return errors.New(fmt.Sprintf("error 7: %v", err))
	}

	jsonScAddressPool, err := json.Marshal(scAddressPool)
	if err != nil {
		return errors.New(fmt.Sprintf("error 12: %v", err))
	}
	jsonScAddressConfig, err := json.Marshal(scAddressConfig)
	if err != nil {
		return errors.New(fmt.Sprintf("error 13: %v", err))
	}
	jsonScAddressHolders, err := json.Marshal(scAddressHolders)
	if err != nil {
		return errors.New(fmt.Sprintf("error 14: %v", err))
	}

	PoolDB.Put(scAddress, string(jsonScAddressPool))
	ConfigDB.Put(scAddress, string(jsonScAddressConfig))
	HolderDB.Put(scAddress, string(jsonScAddressHolders))

	txCommentSign := contracts.NewBuyTokenSign(
		config.NodeNdAddress,
	)

	contracts.SendNewScTx(timestampD, config.BlockHeight, scAddress, uwAddress, txAmount-tax, txTokenLabel, "default_transaction", txCommentSign)

	return nil
}
