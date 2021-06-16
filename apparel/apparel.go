package apparel

import (
	"log"
	"math/rand"
	"node/metrics"
	"strconv"
	"strings"
	"time"
)

func Timestamp() string {
	return time.Now().Format(time.RFC3339Nano)
}

func UnixFromStringTimestamp(timestamp string) int64 {
	timestampUnix, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		log.Println("Timestamp Unix error:", err)
	}

	return timestampUnix.UnixNano()
}

func TimestampUnix() int64 {
	timestamp, err := time.Parse(time.RFC3339Nano, time.Now().Format(time.RFC3339Nano))
	if err != nil {
		log.Println("Timestamp Unix error:", err)
	}

	return timestamp.UnixNano()
}

func UnixToString(unixTime int64) string {
	return time.Unix(0, unixTime).Format(time.RFC3339Nano)
}

func CalcTax(tax float64) float64 {
	if tax > metrics.MaxTax {
		return metrics.MaxTax
	}

	if tax < metrics.MinTax {
		return metrics.MinTax
	}

	return tax
}

func ParseInt64(stringForParsing string) int64 {
	result, err := strconv.ParseInt(stringForParsing, 10, 64)
	if err != nil {
		panic(err)
	}

	return result
}

func GetNonce(timestamp string) int64 {
	parsedTime, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		log.Println("Get Nonce error:", err)
	}

	nonce := parsedTime.UnixNano() + rand.Int63()
	if nonce < 0 {
		nonce *= -1
	}

	return nonce
}

func TrimToLower(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func SearchInArray(arr []int64, find int64) bool {
	for _, el := range arr {
		if el == find {
			return true
		}
	}

	return false

}
