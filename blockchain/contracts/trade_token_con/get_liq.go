package trade_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"strconv"
)

func GetLiq(args *GetArgs) error {
	err := getLiq(args.ScAddress, args.UwAddress, args.TokenLabel, args.TxHash, args.BlockHeight)
	if err != nil {
		return errors.New(fmt.Sprintf("error 1: getLiq %v", err))
	}

	return nil
}

func getLiq(scAddress, uwAddress, tokenLabel, txHash string, blockHeight int64) error {
	scAddressPoolJson := PoolDB.Get(scAddress).Value
	scAddressHoldersJson := HolderDB.Get(scAddress).Value

	var (
		scAddressPool    Pool
		scAddressHolders []Holder
	)
	if scAddressPoolJson != "" {
		err := json.Unmarshal([]byte(scAddressPoolJson), &scAddressPool)
		if err != nil {
			return errors.New(fmt.Sprintf("error 1: %v", err))
		}
	}
	if scAddressHoldersJson != "" {
		err := json.Unmarshal([]byte(scAddressHoldersJson), &scAddressHolders)
		if err != nil {
			return errors.New(fmt.Sprintf("error 2: %v", err))
		}
	}

	if scAddressHolders == nil {
		return errors.New(fmt.Sprintf("error 3: not enouth token holders of this smart-contract \"%s\"", scAddress))
	}

	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		return errors.New("error 4: token does not exist")
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	check := -1
	for idx, i := range scAddressHolders {
		if i.Address == uwAddress {
			if i.Pool.Liq.Amount != 0 {
				getPercent := ((i.Pool.Liq.Amount * 100) / scAddressPool.Liq.Amount) / 100
				getFirstTokenAmount := scAddressPool.FirstToken.Amount * getPercent
				getSecondTokenAmount := scAddressPool.SecondToken.Amount * getPercent

				scAddressPool.Liq.Amount -= i.Pool.Liq.Amount

				scAddressPool.FirstToken.Amount -= getFirstTokenAmount
				scAddressPool.FirstToken.UpdateTime = contracts.String(timestamp)

				scAddressPool.SecondToken.Amount -= getSecondTokenAmount
				scAddressPool.SecondToken.UpdateTime = contracts.String(timestamp)

				scAddressHolders[idx].Pool.FirstToken.Amount += getFirstTokenAmount
				scAddressHolders[idx].Pool.FirstToken.UpdateTime = contracts.String(timestamp)

				scAddressHolders[idx].Pool.SecondToken.Amount += getSecondTokenAmount
				scAddressHolders[idx].Pool.SecondToken.UpdateTime = contracts.String(timestamp)

				scAddressHolders[idx].Pool.Liq.Amount = 0
			}

			check = idx
			break
		}
	}

	if check == -1 {
		return errors.New(fmt.Sprintf("error 5: this token holder \"%s\" does not exist on the smart-contract \"%s\"", uwAddress, scAddress))
	}

	var txAmount float64 = 0
	switch tokenLabel {
	case config.BaseToken:
		txAmount = scAddressHolders[check].Pool.FirstToken.Amount
		scAddressHolders[check].Pool.FirstToken.Amount = 0
		break
	case token.Label:
		txAmount = scAddressHolders[check].Pool.SecondToken.Amount
		scAddressHolders[check].Pool.SecondToken.Amount = 0
		break
	default:
		return errors.New(fmt.Sprintf("error 6: unexpected token for this smart-contract \"%s\"", scAddress))
	}

	err := contracts.AddEvent(scAddress, *contracts.NewEvent("Get liq", timestamp, blockHeight, txHash, uwAddress, newEventGetLiqTypeData(txAmount, tokenLabel)), EventDB, ConfigDB)
	if err != nil {
		return errors.New(fmt.Sprintf("error 7: %v", err))
	}

	jsonScAddressPool, err := json.Marshal(scAddressPool)
	if err != nil {
		return errors.New(fmt.Sprintf("error 8: %v", err))
	}
	jsonScAddressHolders, err := json.Marshal(scAddressHolders)
	if err != nil {
		return errors.New(fmt.Sprintf("error 9: %v", err))
	}

	PoolDB.Put(scAddress, string(jsonScAddressPool))
	HolderDB.Put(scAddress, string(jsonScAddressHolders))

	contracts.SendNewScTx(scAddress, uwAddress, txAmount, tokenLabel, "default_transaction")

	return nil
}
