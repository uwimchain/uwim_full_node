package business_token_con

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
)

type FillConfigArgs struct {
	ScAddress   string   `json:"sc_address"`
	Conversion  float64  `json:"conversion"`
	SalesValue  float64  `json:"sales_value"`
	Partners    Partners `json:"partners"`
	Changes     bool     `json:"changes"`
	TxHash      string   `json:"tx_hash"`
	BlockHeight int64    `json:"block_height"`
}

func NewFillConfigArgs(scAddress string, conversion float64, salesValue float64, partners Partners, changes bool, txHash string, blockHeight int64) (*FillConfigArgs, error) {
	return &FillConfigArgs{ScAddress: scAddress, Conversion: conversion, SalesValue: salesValue, Partners: partners, Changes: changes, TxHash: txHash, BlockHeight: blockHeight}, nil
}

func (args *FillConfigArgs) FillConfig() error {
	if err := fillConfig(args.ScAddress, args.Conversion, args.SalesValue, args.Partners, args.Changes, args.TxHash, args.BlockHeight); err != nil {
		return errors.New(fmt.Sprintf("fill config error: %v", err))
	}

	return nil
}

func fillConfig(scAddress string, conversion, salesValue float64, partners Partners, changes bool, txHash string, blockHeight int64) error {
	scAddressConfig := contracts.GetConfig(ConfigDB, scAddress)
	configData := scAddressConfig.GetData()

	if apparel.ConvertInterfaceToBool(configData["changes"]) {
		return nil
	}

	configData["conversion"] = conversion
	configData["sales_value"] = salesValue
	configData["changes"] = changes

	timestamp := apparel.TimestampUnix()
	if err := contracts.AddEvent(scAddress, *contracts.NewEvent("Fill config", timestamp, blockHeight, txHash, "", nil), EventDB, ConfigDB); err != nil {
		return err
	}

	scAddressPartners := GetPartners(scAddress)

	addPartners := Partners{}
	var delPartnersIds []int
	scAddressBalance := contracts.GetBalance(scAddress)

	if partners != nil {
		for idx, i := range scAddressPartners {
			check := false
			for _, j := range partners {
				if i.Address == j.Address {
					check = true
					break
				}
			}

			if !check {
				delPartnersIds = append(delPartnersIds, idx)
			}
		}

		for _, i := range partners {
			check := false

			if scAddressPartners != nil {
				for jdx, j := range scAddressPartners {
					if i.Address == j.Address {
						scAddressPartners[jdx].Percent = i.Percent
						check = true
					}
				}
			}

			if !check {
				addPartners = append(addPartners, i)
			}
		}

		if delPartnersIds != nil && scAddressPartners != nil {
			if scAddressBalance == nil {
				return errors.New("empty smart-contract balance")
			}

			var allRefundAmount []contracts.Balance
			var refundTransactions []contracts.Tx
			for _, i := range delPartnersIds {
				if scAddressPartners[i].Balance != nil {
					for _, j := range scAddressPartners[i].Balance {
						if j.Amount != 0 {
							if allRefundAmount != nil {
								check := false
								for kdx, k := range allRefundAmount {
									if k.TokenLabel == j.TokenLabel {
										allRefundAmount[kdx].Amount += j.Amount
										check = true
										break
									}
								}

								if !check {
									allRefundAmount = append(allRefundAmount, contracts.Balance{
										TokenLabel: j.TokenLabel,
										Amount:     j.Amount,
									})
								}
							} else {
								allRefundAmount = append(allRefundAmount, contracts.Balance{
									TokenLabel: j.TokenLabel,
									Amount:     j.Amount,
								})
							}

							refundTransactions = append(refundTransactions, contracts.Tx{
								To:         scAddressPartners[i].Address,
								Amount:     j.Amount,
								TokenLabel: j.TokenLabel,
							})
						}
					}
				}

				if len(scAddressPartners) != 1 {
					scAddressPartners = append(scAddressPartners[:i], scAddressPartners[i+1:]...)
				} else {
					scAddressPartners = nil
					break
				}
			}

			if refundTransactions != nil && allRefundAmount != nil {
				for _, i := range allRefundAmount {
					check := false
					for _, j := range scAddressBalance {
						if i.TokenLabel == j.TokenLabel {
							if i.Amount > j.Amount {
								return errors.New("smart-contract address has low balance for send refund transaction")
							}

							check = true
						}
					}

					if !check {
						return errors.New("smart-contract address has low balance for send refund transaction")
					}
				}

				for _, i := range refundTransactions {
					_ = contracts.RefundTransaction(scAddress, i.To, i.Amount, i.TokenLabel)
				}
			}
		}

		if addPartners != nil {
			scAddressPartners = append(scAddressPartners, addPartners...)
		}
	} else {
		scAddressPartners = nil
	}

	if scAddressPartners != nil {
		scAddressPartners.Update(scAddress)
	} else {
		ClearPartners(scAddress)
	}

	scAddressConfig.ConfigData = configData
	scAddressConfig.Update(ConfigDB, scAddress)
	return nil
}
