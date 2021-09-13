package vote_con

func newEventStartTypeData(title, description, starterAddress string, answerOptions []AnswerOption, endBlockHeight int64) interface{} {
	eventStartTypeData := make(map[string]interface{})
	eventStartTypeData["title"] = title
	eventStartTypeData["description"] = description
	eventStartTypeData["answer_options"] = answerOptions
	eventStartTypeData["end_block_height"] = endBlockHeight
	eventStartTypeData["starter_address"] = starterAddress

	return eventStartTypeData
}

func newEventHardStopTypeData(stopperAddress string, hardStopBlockHeight int64) interface{} {
	eventHardStopTypeData := make(map[string]interface{})
	eventHardStopTypeData["stopper_address"] = stopperAddress
	eventHardStopTypeData["hard_stop_block_height"] = hardStopBlockHeight

	return eventHardStopTypeData
}
