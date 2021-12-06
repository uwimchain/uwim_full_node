package donate_token_con

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/blockchain/contracts"
	"strconv"
)

type FillConfigArgs struct {
	ScAddress   string  `json:"sc_address"`
	Conversion  float64 `json:"conversion"`
	MaxBuy      float64 `json:"max_buy"`
	Changes     bool    `json:"changes"`
	TxHash      string  `json:"tx_hash"`
	BlockHeight int64   `json:"block_height"`
}

func NewFillConfigArgs(scAddress string, conversion float64, maxBuy float64, changes bool, txHash string, blockHeight int64) (*FillConfigArgs, error) {
	return &FillConfigArgs{ScAddress: scAddress, Conversion: conversion, MaxBuy: maxBuy, Changes: changes, TxHash: txHash, BlockHeight: blockHeight}, nil
}

func (args *FillConfigArgs) FillConfig() error {
	if err := fillConfig(args.ScAddress, args.Conversion, args.MaxBuy, args.Changes, args.TxHash, args.BlockHeight); err != nil {
		return errors.New(fmt.Sprintf("fill config error: %v", err))
	}

	return nil
}

func fillConfig(scAddress string, conversion, maxBuy float64, changes bool, txHash string, blockHeight int64) error {
	scAddressConfig := contracts.GetConfig(ConfigDB, scAddress)
	configData := scAddressConfig.GetData()

	if apparel.ConvertInterfaceToBool(configData["changes"]) {
		log.Println("FF")
		return nil
	}

	configData["conversion"] = conversion
	configData["max_buy"] = maxBuy
	configData["changes"] = changes

	log.Println("GG", changes)

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)
	if err := contracts.AddEvent(scAddress, *contracts.NewEvent("Fill config", timestamp, blockHeight, txHash, "", nil), EventDB, ConfigDB); err != nil {
		return err
	}

	scAddressConfig.ConfigData = configData
	scAddressConfig.Update(ConfigDB, scAddress)
	return nil
}
