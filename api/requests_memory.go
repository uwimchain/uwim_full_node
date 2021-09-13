package api

import (
	"encoding/json"
	"net/http"
	"node/apparel"
	"node/config"
	"node/crypt"
	"time"
)

type RequestsMemoryData struct {
	TimestampUnix int64
	Hash          string
}

var RequestsMemory []RequestsMemoryData

type requestArgs struct {
	Header http.Header
	Result string
}

func cleanRequestsMemory() {
	for {
		time.Sleep(time.Duration(config.RequestsMemoryLifeTime) * time.Second)

		clean()
	}
}

func requestInRequestsMemory(header http.Header, result string) bool {
	request, _ := json.Marshal(requestArgs{header, result})
	hash := crypt.GetHash(request)

	if searchRequestInRequestsMemory(hash) {
		return true
	} else {
		RequestsMemory = append(RequestsMemory, RequestsMemoryData{
			TimestampUnix: apparel.TimestampUnix(),
			Hash:          hash,
		})

		return false
	}

}

func searchRequestInRequestsMemory(request string) bool {

	var requestCount int64 = 0
	for i := range RequestsMemory {
		if RequestsMemory[i].Hash == request {
			requestCount++
		}
	}

	return requestCount >= config.RequestsMemoryCount
}

func clean() {
	var tmp []RequestsMemoryData
	for i := range RequestsMemory {
		if RequestsMemory[i].TimestampUnix+(time.Second.Nanoseconds()*config.RequestsMemoryLifeTime) >= apparel.TimestampUnix() {
			tmp = append(tmp, RequestsMemory[i])
		}
	}

	RequestsMemory = tmp
}
