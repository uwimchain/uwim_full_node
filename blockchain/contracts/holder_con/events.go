package holder_con

func newEventAddTypeData(depositorAddress, recipientAddress, tokenLabel string, amount float64) interface{} {
	eventAddTypeData := make(map[string]interface{})
	eventAddTypeData["depositor_address"] = depositorAddress
	eventAddTypeData["recipient_address"] = recipientAddress
	eventAddTypeData["token_label"] = tokenLabel
	eventAddTypeData["amount"] = amount

	return eventAddTypeData
}
