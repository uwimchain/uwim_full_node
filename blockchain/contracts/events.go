package contracts

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/crypt"
)

type Event struct {
	Hash        string      `json:"hash"`
	PrevHash    string      `json:"prev_hash"`
	Type        string      `json:"type"`
	Timestamp   int64       `json:"timestamp"`
	BlockHeight int64       `json:"block_height"`
	TxHash      string      `json:"tx_hash"`
	UwAddress   string      `json:"uw_address"`
	TypeData    interface{} `json:"type_data"`
}

func NewEvent(eventType string, timestamp int64, blockHeight int64, txHash string, uwAddress string, typeData interface{}) *Event {
	return &Event{Type: eventType, Timestamp: timestamp, BlockHeight: blockHeight, TxHash: txHash, UwAddress: uwAddress, TypeData: typeData}
}

// function for add new event
func AddEvent(scAddress string, event Event, eventDb, configDb *Database) error {
	scAddressEventsJson := eventDb.Get(scAddress).Value
	scAddressConfigJson := configDb.Get(scAddress).Value
	var (
		scAddressEvents []Event
		scAddressConfig Config
	)
	if scAddressEventsJson != "" {
		err := json.Unmarshal([]byte(scAddressEventsJson), &scAddressEvents)
		if err != nil {
			return errors.New(fmt.Sprintf("error 1: %v", err))
		}
	}
	if scAddressConfigJson != "" {
		err := json.Unmarshal([]byte(scAddressConfigJson), &scAddressConfig)
		if err != nil {
			return errors.New(fmt.Sprintf("error 2: %v", err))
		}
	}

	if scAddressConfig.LastEventHash != "" {
		event.PrevHash = scAddressConfig.LastEventHash
	}

	// записываю эвент в json, чтобы получить хэш, без времени
	jsonEvent, err := json.Marshal(Event{
		Hash:        "",
		PrevHash:    event.PrevHash,
		Type:        event.Type,
		BlockHeight: event.BlockHeight,
		TxHash:      event.TxHash,
		UwAddress:   event.UwAddress,
		TypeData:    event.TypeData,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("error 3: %v", err))
	}

	event.Hash = crypt.GetHash(jsonEvent)

	scAddressEvents = append(scAddressEvents, event)

	scAddressConfig.LastEventHash = event.Hash

	jsonScAddressEvents, err := json.Marshal(scAddressEvents)
	if err != nil {
		return errors.New(fmt.Sprintf("error 4: %v", err))
	}

	jsonScAddressConfig, err := json.Marshal(scAddressConfig)
	if err != nil {
		return errors.New(fmt.Sprintf("error 5: %v", err))
	}

	eventDb.Put(scAddress, string(jsonScAddressEvents))
	configDb.Put(scAddress, string(jsonScAddressConfig))

	return nil
}
