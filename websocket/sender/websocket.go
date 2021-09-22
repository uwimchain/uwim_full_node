package sender

import (
	"encoding/json"
	"fmt"
	"log"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/storage"
	"node/storage/deep_actions"
	"node/websocket"
)

func SendTx(tx deep_actions.Tx) {
	jsonString, _ := json.Marshal(tx)

	message, _ := json.Marshal(websocket.RequestSign{
		SenderIp: config.Ip,
		Sign:     crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
		Address:  config.NodeNdAddress,
		Data:     jsonString,
	})

	requestsSender("NewTx", message)
}

func GetProposer() {
	message, _ := json.Marshal(websocket.RequestSign{
		SenderIp: config.Ip,
		Sign:     crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
		Address:  config.NodeNdAddress,
		Data:     nil,
	})

	if memory.ValidatorsMemory != nil {
		if err := Client(memory.ValidatorsMemory[0].Ip, Request{
			DataType: "GetProposer",
			Body:     string(message),
		}, "/ws"); err != nil {
			log.Println("Client Error: " + fmt.Sprintf("%v", err))
		} else {
			return
		}
	}
}

func SendProposer(ip string) {
	if memory.IsMainNode() {
		message, _ := json.Marshal(websocket.RequestSign{
			SenderIp: config.Ip,
			Sign:     crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
			Address:  config.NodeNdAddress,
			Data:     []byte(memory.Proposer),
		})

		if err := Client(ip, Request{
			DataType: "Proposer",
			Body:     string(message),
		}, "/ws"); err != nil {
			log.Println("Client Error: " + fmt.Sprintf("%v", err))
		} else {
			return
		}
	}
}

func SendVersion(version []byte, ip string) {
	message, _ := json.Marshal(websocket.RequestSign{
		SenderIp: config.Ip,
		Sign:     crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
		Address:  config.NodeNdAddress,
		Data:     version,
	})

	_ = Client(ip, Request{
		DataType: "GetVersion",
		Body:     string(message),
	}, "/ws")
}

func SendBlockVote(vote deep_actions.Vote) {
	jsonString, _ := json.Marshal(vote)

	requestsSender("BlockVote", jsonString)
}

func SendNewBlock() {
	//body, err := json.Marshal(storage.BlockMemory)
	block, err := json.Marshal(storage.BlockMemory)
	if err != nil {
		log.Println("Send New Block error:", err)
	}

	message, _ := json.Marshal(websocket.RequestSign{
		SenderIp: config.Ip,
		Sign:     crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
		Address:  config.NodeNdAddress,
		Data:     block,
	})

	//requestsSender("NewBlock", body)
	requestsSender("NewBlock", message)
}

type GetBlocksRequest struct {
	SenderIp     string `json:"senderIp"`
	SenderHeight int64  `json:"senderHeight"`
}

func NewGetBlocksRequest(senderIp string, senderHeight int64) *GetBlocksRequest {
	return &GetBlocksRequest{SenderIp: senderIp, SenderHeight: senderHeight}
}

func DownloadBlocksFromNodes() {

	body, err := json.Marshal(NewGetBlocksRequest(config.Ip, config.BlockHeight))
	if err != nil {
		log.Println("Download Blocks From Nodes error:", err)
	}

	if memory.ValidatorsMemory != nil {
		for _, validator := range memory.ValidatorsMemory {
			if validator.Ip != config.Ip {
				if err := Client(validator.Ip, Request{
					DataType: "DownloadBlocks",
					Body:     string(body),
				}, "/ws"); err != nil {
					log.Println("Download blocks error:", err)
				} else {
					return
				}
			}
		}
	}
}

func requestsSender(dataType string, body []byte) {
	if len(memory.ValidatorsMemory) != 0 {
		for _, node := range memory.ValidatorsMemory {
			if node.Address != config.NodeNdAddress && node.Ip != config.Ip {
				if err := Client(node.Ip, Request{
					DataType: dataType,
					Body:     string(body),
				}, "/ws"); err != nil {
					log.Println(dataType, "error: client Error:", err)
				}
			}
		}
	} else {
		log.Println("Client Error: NodesMemory array is empty.")
	}
}
