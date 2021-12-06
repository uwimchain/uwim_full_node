package contracts

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/crypt"
)

type Event struct {
	Hash     string `json:"hash"`
	PrevHash string `json:"prev_hash"`
	Type     string `json:"type"`
	Timestamp   String      `json:"timestamp"`
	BlockHeight int64       `json:"block_height"`
	TxHash      string      `json:"tx_hash"`
	UwAddress   string      `json:"uw_address"`
	TypeData    interface{} `json:"type_data"`
}

type Events []Event

func NewEvent(eventType string, timestamp string, blockHeight int64, txHash string, uwAddress string, typeData interface{}) *Event {
	return &Event{Type: eventType, Timestamp: String(timestamp), BlockHeight: blockHeight, TxHash: txHash, UwAddress: uwAddress, TypeData: typeData}
}

func GetEvents(eventsDb *Database, scAddress string) Events {
	eventsJson := eventsDb.Get(scAddress).Value
	var events Events
	_ = json.Unmarshal([]byte(eventsJson), &events)

	return events
}

func (es *Events) Update(eventsDb *Database, scAddress string) {
	eventsJson, _ := json.Marshal(es)
	eventsDb.Put(scAddress, string(eventsJson))
}

func (e *Event) Add(configDb, eventsDb *Database, scAddress string) error {
	config := GetConfig(configDb, scAddress)
	events := GetEvents(eventsDb, scAddress)

	if config.LastEventHash != "" {
		e.PrevHash = config.LastEventHash
	}

	jsonEvent, err := json.Marshal(Event{
		Hash:        "",
		PrevHash:    e.PrevHash,
		Type:        e.Type,
		BlockHeight: e.BlockHeight,
		TxHash:      e.TxHash,
		UwAddress:   e.UwAddress,
		TypeData:    e.TypeData,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("error 1: %v", err))
	}

	e.Hash = crypt.GetHash(jsonEvent)

	events = append(events, *e)

	config.LastEventHash = e.Hash

	events.Update(eventsDb, scAddress)
	config.Update(configDb, scAddress)
	return nil
}

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
