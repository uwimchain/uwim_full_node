package trade_token_con

func newEventAddTypeData(tokenLabel string, amount float64) interface{} {
	eventAddTypeData := make(map[string]interface{})
	eventAddTypeData["token_label"] = tokenLabel
	eventAddTypeData["amount"] = amount

	return eventAddTypeData
}

func newEventSwapTypeData(amount, course float64, tokenLabel, swapTokenLabel string, ) interface{} {
	eventSwapTypeData := make(map[string]interface{})
	firstToken := make(map[string]interface{})
	firstToken["amount"] = amount
	firstToken["token_label"] = swapTokenLabel
	eventSwapTypeData["first_token"] = firstToken
	secondToken := make(map[string]interface{})
	secondToken["amount"] = amount * course
	secondToken["token_label"] = tokenLabel
	eventSwapTypeData["second_token"] = secondToken

	return eventSwapTypeData
}

func newEventGetLiqTypeData(amount float64, tokenLabel string) interface{} {
	eventGetLiqTypeData := make(map[string]interface{})
	eventGetLiqTypeData["token_label"] = tokenLabel
	eventGetLiqTypeData["amount"] = amount

	return eventGetLiqTypeData
}

func newEventGetComTypeData(amount float64, tokenLabel string) interface{} {
	eventGetComTypeData := make(map[string]interface{})
	eventGetComTypeData["token_label"] = tokenLabel
	eventGetComTypeData["amount"] = amount

	return eventGetComTypeData
}

/*// function for add new event
func AddEvent(scAddress string, event contracts.Event) error {
	scAddressEventsJson := EventDb.Get(scAddress).Value
	scAddressConfigJson := ConfigDb.Get(scAddress).Value
	var (
		scAddressEvents []contracts.Event
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
	jsonEvent, err := json.Marshal(contracts.Event{
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

	EventDb.Put(scAddress, string(jsonScAddressEvents))
	ConfigDb.Put(scAddress, string(jsonScAddressConfig))

	return nil
}
*/