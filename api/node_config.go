package api

import (
	"encoding/json"
	"log"
	"node/config"
	"node/memory"
	"node/storage"
)

// NodeConfig method arguments
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
	jsonString, err := json.Marshal(NewConfigData(
		memory.Proposer,
		config.BlockHeight,
		storage.GetTokenId(),
		config.Tax,
		config.Version,
		memory.ValidatorsMemory,
	))
	if err != nil {
		log.Println("Api node config error 1:", err)
	}
	*result = string(jsonString)
	return nil
}
