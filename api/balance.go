package api

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/blockchain/contracts/business_token_con"
	"node/blockchain/contracts/delegate_con"
	"node/blockchain/contracts/donate_token_con"
	"node/blockchain/contracts/holder_con"
	"node/blockchain/contracts/my_token_con"
	"node/blockchain/contracts/trade_token_con"
	"node/config"
	"node/crypt"
	"node/metrics"
	"node/storage"
	"node/storage/deep_actions"
	"strconv"
)

type BalanceArgs struct {
	Address string `json:"address"`
}

type Partner struct {
	Address string                 `json:"address"`
	Percent float64                `json:"percent"`
	Balance []deep_actions.Balance `json:"balance"`
}

type BalanceInfo struct {
	Address           string                 `json:"address"`
	Balance           []deep_actions.Balance `json:"balance"`
	Transactions      []deep_actions.Tx      `json:"transactions"`
	Token             deep_actions.Token     `json:"token"`
	ScKeeping         bool                   `json:"sc_keeping"`
	Name              string                 `json:"name"`
	ScBalance         []deep_actions.Balance `json:"sc_balance"`
	DelegateBalance   deep_actions.Balance   `json:"delegate_balance"`
	Percents          interface{}            `json:"percents"`
	TokenContractData interface{}            `json:"token_contract_data"`
	Holder            interface{}            `json:"holder"`
}

func (api *Api) Balance(args *BalanceArgs, result *string) error {
	if args.Address == "" {
		return errors.New(strconv.Itoa(1))
	}

	address := deep_actions.GetAddress(args.Address)
	token := deep_actions.GetToken(address.TokenLabel)
	scBalance := storage.GetScBalance(args.Address)
	delegateBalance := delegate_con.GetBalance(address.Address)
	tokenContractData := make(map[string]interface{})
	percents := make(map[string][]interface{})

	scAddress := crypt.AddressFromPublicKey(metrics.SmartContractPrefix, address.PublicKey)
	if token.Id != 0 {
		switch token.Standard {
		case 0:
			tokenContractData["pool"], _ = my_token_con.GetPool(scAddress)
			break
		case 1:
			tokenContractData["events"], _ = donate_token_con.GetEvents(scAddress)
			tokenContractData["config"] = donate_token_con.GetConfig(scAddress)
			break
		case 4:
			tokenContractData["config"] = business_token_con.GetConfig(scAddress)
			break
		case 5:
			scAddressHolders, err := trade_token_con.GetScHolders(scAddress)
			if err != nil {
				log.Println("Api balance error 3: ", err)
				break
			}
			scAddressPool, err := trade_token_con.GetScPool(scAddress)
			if err != nil {
				log.Println("Api balance error 4: ", err)
				break
			}

			tokenContractData["holders"] = scAddressHolders
			tokenContractData["pool"] = scAddressPool
			tokenContractData["config"] = trade_token_con.GetConfig(scAddress)
			break
		}
	}

	if address.Balance != nil {
		for _, i := range address.Balance {
			if i.TokenLabel != config.BaseToken {
				t := deep_actions.GetToken(i.TokenLabel)
				publicKey, _ := crypt.PublicKeyFromAddress(t.Proposer)
				tokenScAddress := crypt.AddressFromPublicKey(metrics.SmartContractPrefix, publicKey)
				switch t.Standard {
				case 0:
					if i.TokenLabel != token.Label {
						confirmation, addressPercent, err := my_token_con.GetAddressPercent(tokenScAddress,
							address.Address, t.Label, t.Emission, i.Amount)
						if err != nil {
							log.Println("Api Balance error 6: ", err)
							break
						}

						percent := make(map[string]interface{})
						percent["token_label"] = i.TokenLabel
						percent["percent"] = addressPercent
						percent["confirmation"] = confirmation
						percent["token"] = t
						percent["token_sc_address"] = tokenScAddress
						percent["token_sc_balance"] = storage.GetBalance(tokenScAddress)
						percents["my"] = append(percents["my"], percent)
					}
					break
				case 4:
					if i.TokenLabel != token.Label {
						partner := business_token_con.GetPartner(tokenScAddress, address.Address)
						percent := make(map[string]interface{})
						percent["token_label"] = i.TokenLabel
						percent["percent"] = partner.Percent
						percent["token"] = t
						percent["token_sc_address"] = tokenScAddress
						percent["token_sc_balance"] = partner.Balance
						percents["business"] = append(percents["business"], percent)
					}
					break
				case 5:
					holderPool, err := trade_token_con.GetScHolder(tokenScAddress, address.Address)
					if err != nil {
						log.Println("Api balance error 8: ", err)
						break
					}
					scAddressPool, err := trade_token_con.GetScPool(tokenScAddress)
					if err != nil {
						log.Println("Api balance error 9: ", err)
						break
					}

					percent := make(map[string]interface{})
					percent["token_label"] = i.TokenLabel
					percent["token"] = t
					percent["token_sc_address"] = tokenScAddress
					percent["holder_pool"] = holderPool
					percent["sc_pool"] = scAddressPool
					percents["trade"] = append(percents["trade"], percent)
				}
			}
		}
	}

	jsonString, err := json.Marshal(BalanceInfo{
		Address:      args.Address,
		Balance:      address.Balance,
		Transactions: address.GetTxs(),
		Token:        *token,
		ScKeeping:    address.ScKeeping,
		Name:         address.Name,
		ScBalance:    scBalance,
		DelegateBalance: deep_actions.Balance{
			TokenLabel: config.BaseToken,
			Amount:     delegateBalance.Balance,
			UpdateTime: strconv.FormatInt(delegateBalance.UpdateTime, 10),
		},
		Percents:          percents,
		TokenContractData: tokenContractData,
		Holder:            holder_con.GetHolder(args.Address),
	})
	if err != nil {
		log.Println("Api Balance error 8", err)
	}

	*result = string(jsonString)
	return nil
}
