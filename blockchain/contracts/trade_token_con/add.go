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

func Add(args *TradeArgs) error {
	err := add(args.ScAddress, args.UwAddress, args.TokenLabel, args.TxHash, args.Amount, args.BlockHeight)
	if err != nil {
		refundError := contracts.RefundTransaction(args.ScAddress, args.UwAddress, args.Amount, args.TokenLabel)
		if refundError != nil {
			log.Println(fmt.Sprintf("Refund transaction %v", refundError))
		}
		return errors.New(fmt.Sprintf("error 1: addPool %v", err))
	}

	return nil
}

func add(scAddress, uwAddress, tokenLabel, txHash string, amount float64, blockHeight int64) error {
	scAddressPoolJson := PoolDB.Get(scAddress).Value
	scAddressHoldersJson := HolderDB.Get(scAddress).Value

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		return errors.New(fmt.Sprintf("error 1: token of this smart-contract \"%s\" does not exist", scAddress))
	}

	var (
		scAddressPool Pool
		holders       []Holder
		holder        Holder
	)
	if scAddressPoolJson != "" {
		err := json.Unmarshal([]byte(scAddressPoolJson), &scAddressPool)
		if err != nil {
			return errors.New(fmt.Sprintf("error 2: %v", err))
		}
	}

	if scAddressHoldersJson != "" {
		err := json.Unmarshal([]byte(scAddressHoldersJson), &holders)
		if err != nil {
			return errors.New(fmt.Sprintf("error 4: %v", err))
		}
	}

	check := -1
	if holders != nil {
		for idx, i := range holders {
			if i.Address == uwAddress {
				holder = i
				check = idx
				break
			}
		}
	}

	var course float64 = 0
	if scAddressPool.FirstToken.Amount > 0 && scAddressPool.SecondToken.Amount > 0 {
		course = scAddressPool.FirstToken.Amount / scAddressPool.SecondToken.Amount
	}

	switch tokenLabel {
	case config.BaseToken:
		if holder.Pool.SecondToken.Amount == 0 {
			holder.Pool.FirstToken.Amount += amount
			holder.Pool.FirstToken.UpdateTime = contracts.String(timestamp)
		} else {

			holder.Pool.FirstToken.Amount += amount
			var liq float64 = 0
			if course == 0 {
				scAddressPool.FirstToken.Amount += holder.Pool.FirstToken.Amount
				scAddressPool.FirstToken.UpdateTime = contracts.String(timestamp)

				scAddressPool.SecondToken.Amount += holder.Pool.SecondToken.Amount
				scAddressPool.SecondToken.UpdateTime = contracts.String(timestamp)

				liq = holder.Pool.FirstToken.Amount * holder.Pool.SecondToken.Amount

				holder.Pool.FirstToken.Amount = 0
				holder.Pool.FirstToken.UpdateTime = contracts.String(timestamp)

				holder.Pool.SecondToken.Amount = 0
				holder.Pool.SecondToken.UpdateTime = contracts.String(timestamp)
			} else {

				if holder.Pool.FirstToken.Amount > holder.Pool.SecondToken.Amount*course {
					var t1 float64 = holder.Pool.SecondToken.Amount * course

					scAddressPool.FirstToken.Amount += t1
					scAddressPool.FirstToken.UpdateTime = contracts.String(timestamp)

					scAddressPool.SecondToken.Amount += holder.Pool.SecondToken.Amount
					scAddressPool.SecondToken.UpdateTime = contracts.String(timestamp)

					liq = t1 * holder.Pool.SecondToken.Amount

					holder.Pool.FirstToken.Amount -= t1
					holder.Pool.FirstToken.UpdateTime = contracts.String(timestamp)

					holder.Pool.SecondToken.Amount = 0
					holder.Pool.SecondToken.UpdateTime = contracts.String(timestamp)
					if holder.Pool.FirstToken.Amount < 0 {
						return errors.New("GG 1")
					}
				} else {
					var t2 float64 = holder.Pool.FirstToken.Amount / course

					scAddressPool.FirstToken.Amount += holder.Pool.FirstToken.Amount
					scAddressPool.FirstToken.UpdateTime = contracts.String(timestamp)

					scAddressPool.SecondToken.Amount += t2
					scAddressPool.SecondToken.UpdateTime = contracts.String(timestamp)

					liq = holder.Pool.FirstToken.Amount * t2

					holder.Pool.FirstToken.Amount = 0
					holder.Pool.FirstToken.UpdateTime = contracts.String(timestamp)

					holder.Pool.SecondToken.Amount -= t2
					holder.Pool.SecondToken.UpdateTime = contracts.String(timestamp)
					if holder.Pool.SecondToken.Amount < 0 {
						return errors.New("GG 2")
					}
				}
			}

			scAddressPool.Liq.Amount += liq
			holder.Pool.Liq.Amount += liq
		}
		break
	case token.Label:
		if holder.Pool.FirstToken.Amount == 0 {
			holder.Pool.SecondToken.Amount += amount
			holder.Pool.SecondToken.UpdateTime = contracts.String(timestamp)
		} else {
			holder.Pool.SecondToken.Amount += amount
			holder.Pool.SecondToken.UpdateTime = contracts.String(timestamp)

			var liq float64 = 0

			if course == 0 {
				scAddressPool.FirstToken.Amount += holder.Pool.FirstToken.Amount
				scAddressPool.FirstToken.UpdateTime = contracts.String(timestamp)

				scAddressPool.SecondToken.Amount += holder.Pool.SecondToken.Amount
				scAddressPool.SecondToken.UpdateTime = contracts.String(timestamp)

				liq = holder.Pool.FirstToken.Amount * holder.Pool.SecondToken.Amount

				holder.Pool.FirstToken.Amount = 0
				holder.Pool.FirstToken.UpdateTime = contracts.String(timestamp)

				holder.Pool.SecondToken.Amount = 0
				holder.Pool.SecondToken.UpdateTime = contracts.String(timestamp)
			} else {
				if holder.Pool.FirstToken.Amount > holder.Pool.SecondToken.Amount*course {
					var t1 float64 = holder.Pool.SecondToken.Amount * course

					scAddressPool.FirstToken.Amount += t1
					scAddressPool.FirstToken.UpdateTime = contracts.String(timestamp)

					scAddressPool.SecondToken.Amount += holder.Pool.SecondToken.Amount

					scAddressPool.SecondToken.UpdateTime = contracts.String(timestamp)

					liq = t1 * holder.Pool.SecondToken.Amount

					holder.Pool.FirstToken.Amount -= t1
					holder.Pool.FirstToken.UpdateTime = contracts.String(timestamp)

					holder.Pool.SecondToken.Amount = 0
					holder.Pool.SecondToken.UpdateTime = contracts.String(timestamp)
					if holder.Pool.FirstToken.Amount < 0 {
						return errors.New("GG 1")
					}
				} else {
					var t2 float64 = holder.Pool.FirstToken.Amount / course

					scAddressPool.FirstToken.Amount += holder.Pool.FirstToken.Amount
					scAddressPool.FirstToken.UpdateTime = contracts.String(timestamp)

					scAddressPool.SecondToken.Amount += t2
					scAddressPool.SecondToken.UpdateTime = contracts.String(timestamp)

					liq = holder.Pool.FirstToken.Amount * t2

					holder.Pool.FirstToken.Amount = 0
					holder.Pool.FirstToken.UpdateTime = contracts.String(timestamp)

					holder.Pool.SecondToken.Amount -= t2
					holder.Pool.SecondToken.UpdateTime = contracts.String(timestamp)
					if holder.Pool.SecondToken.Amount < 0 {
						return errors.New("GG 2")
					}
				}
			}

			scAddressPool.Liq.Amount += liq
			scAddressPool.Liq.UpdateTime = contracts.String(timestamp)

			holder.Pool.Liq.Amount += liq
			scAddressPool.Liq.UpdateTime = contracts.String(timestamp)
		}
		break
	default:
		return errors.New(fmt.Sprintf("error 5: unexpected token for this smart-contract \"%s\"", scAddress))
	}

	if check != -1 {
		holders[check] = holder
	} else {
		holders = append(holders, Holder{
			Address: uwAddress,
			Pool:    holder.Pool,
		})
	}

	jsonScAddressPool, err := json.Marshal(scAddressPool)
	if err != nil {
		return errors.New(fmt.Sprintf("error 6: %v %v", err, scAddressPool))
	}

	jsonHolders, err := json.Marshal(holders)
	if err != nil {
		return errors.New(fmt.Sprintf("error 7: %v %v", err, holders))
	}

	err = contracts.AddEvent(scAddress, *contracts.NewEvent("add", timestamp, blockHeight, txHash, uwAddress, newEventAddTypeData(tokenLabel, amount)), EventDB, ConfigDB)
	if err != nil {
		return errors.New(fmt.Sprintf("error 8: add event %v", err))
	}

	PoolDB.Put(scAddress, string(jsonScAddressPool))
	HolderDB.Put(scAddress, string(jsonHolders))

	return nil
}
