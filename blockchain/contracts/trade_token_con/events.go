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
