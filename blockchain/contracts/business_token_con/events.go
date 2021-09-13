package business_token_con

func newEventBuyTypeData(amount, conversion float64, tokenLabel string) interface{} {
	eventBuyTypeData := make(map[string]interface{})
	eventBuyTypeData["amount"] = amount
	eventBuyTypeData["conversion"] = conversion
	eventBuyTypeData["token_label"] = tokenLabel

	return eventBuyTypeData
}

func newEventGetPercentTypeData(amount float64, tokenLabel string) interface{} {
	eventTakePercentTypeData := make(map[string]interface{})
	eventTakePercentTypeData["amount"] = amount
	eventTakePercentTypeData["token_label"] = tokenLabel

	return eventTakePercentTypeData
}
