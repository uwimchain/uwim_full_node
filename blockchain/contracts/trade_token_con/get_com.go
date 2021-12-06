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

func GetCom(args *GetArgs) error {
	err := getCom(args.ScAddress, args.UwAddress, args.TokenLabel, args.TxHash, args.BlockHeight)
	if err != nil {
		return errors.New(fmt.Sprintf("error 1: getPool %v", err))
	}

	return nil
}

func getCom(scAddress, uwAddress, tokenLabel, txHash string, blockHeight int64) error {
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
	var txAmount float64 = 0
	for idx, i := range scAddressHolders {
		if i.Address == uwAddress {
			switch tokenLabel {
			case config.BaseToken:
				if i.Pool.FirstToken.Commission <= 0 {
					return errors.New(fmt.Sprintf("error 5: low balance for token %s", config.BaseToken))
				}
				txAmount = i.Pool.FirstToken.Commission
				scAddressHolders[idx].Pool.FirstToken.Commission = 0
				break
			case token.Label:
				if i.Pool.SecondToken.Commission <= 0 {
					return errors.New(fmt.Sprintf("error 6: low balance for token %s", token.Label))
				}
				txAmount = i.Pool.SecondToken.Commission
				scAddressHolders[idx].Pool.SecondToken.Commission = 0
				break
			default:
				return errors.New(fmt.Sprintf("error 7: unexpected token for this smart-contract \"%s\"", scAddress))
			}

			check = idx
			break
		}
	}

	if check == -1 {
		return errors.New(fmt.Sprintf("error 8: this token holder \"%s\" does not exist on the smart-contract \"%s\"", uwAddress, scAddress))
	}

	if txAmount <= 0 {
		return errors.New("error 9: null tx amount")
	}

	err := contracts.AddEvent(scAddress, *contracts.NewEvent("Get com", timestamp, blockHeight, txHash, uwAddress, newEventGetComTypeData(txAmount, tokenLabel)), EventDB, ConfigDB)
	if err != nil {
		return errors.New(fmt.Sprintf("error 10: %v", err))
	}

	jsonScAddressPool, err := json.Marshal(scAddressPool)
	if err != nil {
		return errors.New(fmt.Sprintf("error 11: %v", err))
	}
	jsonScAddressHolders, err := json.Marshal(scAddressHolders)
	if err != nil {
		return errors.New(fmt.Sprintf("error 12: %v", err))
	}

	PoolDB.Put(scAddress, string(jsonScAddressPool))
	HolderDB.Put(scAddress, string(jsonScAddressHolders))

	contracts.SendNewScTx(scAddress, uwAddress, txAmount, tokenLabel, "default_transaction")

	return nil
}
