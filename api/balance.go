package api

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/blockchain/contracts/business_token_con"
	"node/blockchain/contracts/delegate_con"
	"node/blockchain/contracts/donate_token_con"
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

type MyPercent struct {
	TokenLabel     string                 `json:"token_label"`
	Percent        float64                `json:"percent"`
	TokenScBalance []deep_actions.Balance `json:"token_sc_balance"`
	Confirmation   bool                   `json:"confirmation"`
	Token          deep_actions.Token     `json:"token"`
	TokenScAddress string                 `json:"token_sc_address"`
}

type BusinessPercent struct {
	TokenLabel     string                 `json:"token_label"`
	Percent        float64                `json:"percent"`
	TokenScBalance []deep_actions.Balance `json:"token_sc_balance"`
	Token          deep_actions.Token     `json:"token"`
	TokenScAddress string                 `json:"token_sc_address"`
}

type TradePercent struct {
	TokenLabel     string               `json:"token_label"`
	HolderPool     trade_token_con.Pool `json:"holder_pool"`
	ScPool         trade_token_con.Pool `json:"sc_pool"`
	Token          deep_actions.Token   `json:"token"`
	TokenScAddress string               `json:"token_sc_address"`
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
}

func (api *Api) Balance(args *BalanceArgs, result *string) error {
	if args.Address != "" {
		address := storage.GetAddress(args.Address)
		token := storage.GetToken(address.TokenLabel)
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
				break
			case 4:
				partnersFromCon := business_token_con.GetPartners(scAddress)
				tokenStandardCard := deep_actions.BusinessStandardCardData{}
				if token.StandardCard != "" {
					err := json.Unmarshal([]byte(token.StandardCard), &tokenStandardCard)
					if err != nil {
						log.Println("Api Balance error 1:", err)
						break
					}

					if partnersFromCon != nil && tokenStandardCard.Partners != nil {

						if len(partnersFromCon) != len(tokenStandardCard.Partners) {
							log.Println("Api Balance error 2")
						} else {
							var partners []Partner

							for _, i := range partnersFromCon {
								partner := Partner{
									Address: i.Address,
								}

								var balance []deep_actions.Balance
								if i.Balance != nil {
									for _, j := range i.Balance {
										balance = append(balance, deep_actions.Balance{
											TokenLabel: j.TokenLabel,
											Amount:     j.Amount,
											UpdateTime: j.UpdateTime,
										})
									}
								}
								partner.Balance = balance

								for _, j := range tokenStandardCard.Partners {
									if i.Address == j.Address {
										partner.Percent = j.Percent
										break
									}
								}

								partners = append(partners, partner)
							}

							if partners != nil {
								tokenContractData["partners"] = partners
							}
						}
					}
				}
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
				scAddressConfig, err := trade_token_con.GetScConfig(scAddress)
				if err != nil {
					log.Println("Api balance error 5: ", err)
					break
				}

				tokenContractData["holders"] = scAddressHolders
				tokenContractData["pool"] = scAddressPool
				tokenContractData["config"] = scAddressConfig
				break
			}
		}

		if address.Balance != nil {
			for _, i := range address.Balance {
				if i.TokenLabel != config.BaseToken {
					t := storage.GetToken(i.TokenLabel)
					publicKey, _ := crypt.PublicKeyFromAddress(t.Proposer)
					tokenScAddress := crypt.AddressFromPublicKey(metrics.SmartContractPrefix, publicKey)
					switch t.Standard {
					case 0:
						if i.TokenLabel != token.Label {
							confirmation, addressPercent, err := my_token_con.GetAddressPercent(tokenScAddress, address.Address, t.Label, t.Emission, i.Amount)
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
							var tokenStandardCard deep_actions.BusinessStandardCardData
							if t.StandardCard != "" {
								err := json.Unmarshal([]byte(t.StandardCard), &tokenStandardCard)
								if err != nil {
									log.Println("Api Balance error 7: ", err)
									break
								}
								var businessPercentBalance []deep_actions.Balance
								if partner.Balance != nil {
									for _, i := range partner.Balance {
										businessPercentBalance = append(businessPercentBalance, deep_actions.Balance{
											TokenLabel: i.TokenLabel,
											Amount:     i.Amount,
											UpdateTime: i.UpdateTime,
										})
									}
								}

								var businessPercentAmount float64 = 0
								for _, i := range tokenStandardCard.Partners {
									if i.Address == address.Address {
										businessPercentAmount = i.Percent
										break
									}
								}

								percent := make(map[string]interface{})
								percent["token_label"] = i.TokenLabel
								percent["percent"] = businessPercentAmount
								percent["token"] = t
								percent["token_sc_address"] = tokenScAddress
								percent["token_sc_balance"] = businessPercentBalance
								percents["business"] = append(percents["business"], percent)

							}
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
			Transactions: storage.GetTransactions(args.Address),
			Token:        token,
			ScKeeping:    address.ScKeeping,
			Name:         address.Name,
			ScBalance:    scBalance,
			DelegateBalance: deep_actions.Balance{
				TokenLabel: config.BaseToken,
				Amount:     delegateBalance.Balance,
				UpdateTime: apparel.UnixToString(delegateBalance.UpdateTime),
			},
			Percents: percents,
			TokenContractData: tokenContractData,
		})
		if err != nil {
			log.Println("Api Balance error 8", err)
		}
		*result = string(jsonString)
		return nil
	} else {
		return errors.New(strconv.Itoa(1))
	}
}
