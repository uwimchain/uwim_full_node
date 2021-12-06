package api

import (
	"encoding/json"
	"node/api/api_error"
	"node/blockchain/contracts/business_token_con"
	"node/blockchain/contracts/default_con"
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
)

type BalanceArgs struct {
	Address           string `json:"address"`
}

func (api *Api) Balance(args *BalanceArgs, result *string) error {
	if args.Address == "" {
		return api_error.NewApiError(1, "Empty address").AddError()
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
			scAddressHolders, _ := trade_token_con.GetScHolders(scAddress)

			scAddressPool, _ := trade_token_con.GetScPool(scAddress)

			tokenContractData["holders"] = scAddressHolders
			tokenContractData["pool"] = scAddressPool
			tokenContractData["config"] = trade_token_con.GetConfig(scAddress)
			break
		case 7:
			tokenContractData["config"]=default_con.GetConfig(scAddress)
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
						confirmation, addressPercent, _ := my_token_con.GetAddressPercent(tokenScAddress,
							address.Address, t.Label, t.Emission, i.Amount)

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
					holderPool, _ := trade_token_con.GetScHolder(tokenScAddress, address.Address)

					scAddressPool, _ := trade_token_con.GetScPool(tokenScAddress)

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

	memoryTransactions := deep_actions.Txs{}
	if storage.TransactionsMemory != nil {
		for _, i := range storage.TransactionsMemory {
			if i.From == args.Address || i.To == args.Address {
				memoryTransactions = append(memoryTransactions, i)
			}
		}
	}

	holder := holder_con.GetHolder(args.Address)
	holderGet := false
	if holder != nil {
		for _, i := range holder {
			if i.GetBlockHeight <= config.BlockHeight {
				holderGet = true
				break
			}
		}
	}

	info := make(map[string]interface{})
	info["address"] = args.Address
	info["sc_address"] = crypt.AddressFromAnotherAddress(metrics.SmartContractPrefix, args.Address)
	info["nd_address"] = crypt.AddressFromAnotherAddress(metrics.NodePrefix, args.Address)
	info["balance"] = address.Balance
	info["transactions"] = address.GetTxs()
	info["token"] = token
	info["sc_keeping"] = address.ScKeeping
	info["name"] = address.Name
	info["sc_balance"] = scBalance
	info["delegate_balance"] = deep_actions.Balance{
		TokenLabel: config.DelegateToken,
		Amount:     delegateBalance.Balance,
		UpdateTime: string(delegateBalance.UpdateTime),
	}
	info["percents"] = percents
	info["token_contract_data"] = tokenContractData
	info["holder"] = holder
	info["holder_get"] = holderGet
	info["memory_transactions"] = memoryTransactions

	jsonString, _ := json.Marshal(info)

	*result = string(jsonString)
	return nil
}
