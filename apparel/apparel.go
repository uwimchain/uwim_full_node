package apparel

import (
	"fmt"
	"log"
	"math/rand"
	"node/config"
	"node/metrics"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func TimestampUnix() int64 {
	timestamp, err := time.Parse(time.RFC3339Nano, time.Now().Format(time.RFC3339Nano))
	if err != nil {
		log.Println("Timestamp unix error:", err)
	}

	return timestamp.UnixNano()
}

func CalcTax(amount float64) float64 {
	amount = amount * config.TaxConversion * config.Tax
	if amount > metrics.MaxTax {
		return metrics.MaxTax
	}

	if amount < metrics.MinTax {
		return metrics.MinTax
	}

	return amount
}

func ParseInt64(stringForParsing string) int64 {
	result, err := strconv.ParseInt(stringForParsing, 10, 64)
	if err != nil {
		panic(err)
	}

	return result
}

func GetNonce(timestampString string) int64 {
	timestamp, _ := strconv.ParseInt(timestampString, 10, 64)

	nonce := timestamp + rand.Int63()
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

func Round(number float64) float64 {
	if number < 0 {
		//return 0, errors.New("error 1: round number less than zero")
		return 0
	}

	roundNumber, err := strconv.ParseFloat(fmt.Sprintf("%.12f", number), 64)
	if err != nil {
		//return 0, errors.New(fmt.Sprintf("error 2: %v", err))
		return 0
	}

	return roundNumber
}

func ConvertInterfaceToFloat64(float64_ interface{}) float64 {
	if float64_ == nil {
		return 0
	}
	var (
		float64Type reflect.Type = reflect.TypeOf(float64(0))
		result      float64      = 0
	)

	v2 := reflect.ValueOf(float64_)
	v2 = reflect.Indirect(v2)
	if v2.Type().ConvertibleTo(float64Type) {
		result = v2.Convert(float64Type).Float()
	}

	return result
}

func ConvertInterfaceToInt64(int64_ interface{}) int64 {
	if int64_ == nil {
		return 0
	}
	var (
		Int64Type reflect.Type = reflect.TypeOf(int64(0))
		result    int64        = 0
	)

	v2 := reflect.ValueOf(int64_)
	v2 = reflect.Indirect(v2)
	if v2.Type().ConvertibleTo(Int64Type) {
		result = v2.Convert(Int64Type).Int()
	}

	return result
}

func ConvertInterfaceToString(string_ interface{}) string {
	if string_ == nil {
		return ""
	}
	var (
		stringType reflect.Type = reflect.TypeOf(string(""))
		result     string       = ""
	)

	v2 := reflect.ValueOf(string_)
	v2 = reflect.Indirect(v2)
	if v2.Type().ConvertibleTo(stringType) {
		result = v2.Convert(stringType).String()
	}

	return result
}

func ConvertInterfaceToBool(bool_ interface{}) bool {
	if bool_ == nil {
		return false
	}
	var (
		boolType reflect.Type = reflect.TypeOf(false)
		result   bool         = false
	)

	v2 := reflect.ValueOf(bool_)
	v2 = reflect.Indirect(v2)
	if v2.Type().ConvertibleTo(boolType) {
		result = v2.Convert(boolType).Bool()
	}

	return result
}

func ConvertInterfaceToMapStringInterface(arr_ interface{}) map[string]interface{} {
	if arr_ == nil {
		return nil
	}

	result, err := arr_.(map[string]interface{})
	if !err {
		return nil
	}

	return result
}

func InArray(arr interface{}, element interface{}) bool {
	if arr == nil || element == nil {
		return false
	}

	arrType := reflect.TypeOf(arr).Elem()
	elementType := reflect.TypeOf(element)
	if arrType != elementType {
		return false
	}

	arrObj := reflect.ValueOf(arr)
	arrObjLen := arrObj.Len()
	if arrObjLen == 0 {
		return false
	}

	for i := 0; i < arrObjLen; i++ {
		if arrObj.Index(i).Interface() == element {
			return true
		}
	}

	return false
}
