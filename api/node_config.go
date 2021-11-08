package api

import (
	"encoding/json"
	"log"
	"node/config"
	"node/memory"
	"node/storage"
	"regexp"
	"strings"
)

type NodeConfigArgs struct{}

type ConfigData struct {
	NowProposer string             `json:"nowProposer"`
	BlockHeight int64              `json:"blockHeight"`
	TokensCount int64              `json:"tokensCount"`
	Tax         float64            `json:"tax"`
	Version     string             `json:"version"`
	Validators  []memory.Validator `json:"validators"`
}

func NewConfigData(nowProposer string, blockHeight int64, tokensCount int64, tax float64, version string, validators []memory.Validator) *ConfigData {
	return &ConfigData{
		NowProposer: nowProposer,
		BlockHeight: blockHeight,
		TokensCount: tokensCount,
		Tax:         tax,
		Version:     version,
		Validators:  validators,
	}
}

func (api *Api) NodeConfig(args *NodeConfigArgs, result *string) error {
	var validators []memory.Validator
	for _, v := range memory.ValidatorsMemory {
		ipArr1 := strings.Split(v.Ip, ":")
		ipArr2 := strings.Split(ipArr1[0], ".")

		for idx := range ipArr2 {
			regex := regexp.MustCompile("[0-9]$")
			ipArr2[idx] = regex.ReplaceAllString(ipArr2[idx], "*")
		}

		ip := strings.Join(ipArr2, ".")

		validators = append(validators, memory.Validator{
			Idx:     v.Idx,
			Address: v.Address,
			Ip:      ip,
		})
	}

	jsonString, err := json.Marshal(NewConfigData(
		memory.Proposer,
		config.BlockHeight,
		storage.GetTokenId(),
		config.Tax,
		config.Version,
		validators,
	))
	if err != nil {
		log.Println("Api node config error 1:", err)
	}
	*result = string(jsonString)
	return nil
}
