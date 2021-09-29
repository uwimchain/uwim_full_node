package memory

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"node/apparel"
	"node/config"
	"time"
)

type Validator struct {
	Idx     int64  `json:"idx"`
	Address string `json:"address"`
	Ip      string `json:"ip"`
}

var (
	Proposer         string
	ValidatorsMemory []Validator
)

var DownloadBlocks bool = false

func Init() {
	DownloadValidators()
	Proposer = GetNextProposer()

	if IsMainNode() {
		log.Println("This node is main.")
	}

	if IsValidator() {
		log.Println("This node is validator.")
	}
}

func IsNodeProposer() bool {
	return config.NodeNdAddress == Proposer
}

func IsMainNode() bool {
	mainNode := GetMainNode()
	return config.Ip == mainNode.Ip && config.NodeNdAddress == mainNode.Address
}

func GetMainNode() Validator {
	mainNode := Validator{}
	if ValidatorsMemory != nil {
		for _, validator := range ValidatorsMemory {
			if validator.Idx == 0 {
				mainNode = validator
			}
		}
	}

	return mainNode
}

func GetNextProposer() string {
	if ValidatorsMemory != nil {
		idx := (apparel.TimestampUnix() / (config.CalculateBlockWriteTime() * time.Second.Nanoseconds())) % int64(len(ValidatorsMemory))
		for _, validator := range ValidatorsMemory {
			if validator.Idx == idx {
				return validator.Address
			}
		}
	} else {
		log.Println("Error: empty validators list.")
	}

	return ""
}

func GetValidators() []Validator {
	var validators []Validator
	err := json.Unmarshal(GetJsonData("validators"), &validators)
	if err != nil {
		log.Println("Get Validators error:", err)
		return nil
	} else {
		return validators
	}
}

func GetJsonData(jsonFile string) []byte {
	res, err := http.Get(config.JsonDownloadIp + "/" + jsonFile + ".json")
	if err != nil {
		log.Println("Get JSON Data error:", err)
	} else {
		defer res.Body.Close()
		if res.StatusCode == 200 {
			answer, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Println("Get JSON Data ReadAll error:", err)
			} else {
				return answer
			}
		}
	}
	return nil
}

func DownloadValidators() {
	validators := GetValidators()
	if ValidatorsMemory = validators; ValidatorsMemory == nil {
		if config.JsonDownloadIp == "" {
			ValidatorsMemory = nil
			ValidatorsMemory = append(ValidatorsMemory, Validator{
				Idx:     config.FirstPeerIdx,
				Address: config.FirstPeerAddress,
				Ip:      config.FirstPeerIp,
			})

		}
	}
}

func IsValidator() bool {
	if ValidatorsMemory != nil {
		for _, validator := range ValidatorsMemory {
			if validator.Address == config.NodeNdAddress && validator.Ip == config.Ip {
				return true
			}
		}
	} else {
		log.Println("Error: empty validators list.")
	}
	return false
}

func IsNodeValidator(address string) bool {
	if ValidatorsMemory != nil {
		for _, validator := range ValidatorsMemory {
			if validator.Address == address {
				return true
			}
		}
	} else {
		log.Println("Error: empty validators list.")
	}
	return false
}