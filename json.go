package iris_extend_helper

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/json-iterator/go"
)

func GetJSON(value interface{}) []byte {
	if value == nil {
		return make([]byte, 0)
	}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	bytes, err := json.Marshal(value)
	if err != nil {
		log.Println(err)
	}
	return bytes
}

func Compare(a, b interface{}) int {
	switch a.(type) {
	case string:
		a := a.(string)
		if number, err := strconv.ParseFloat(a, 64); err != nil {
			if timestamp, ok := ParseTimestamp(a); ok {
				duration := timestamp.Sub(ParseTime(b))
				if duration < 0 {
					return -1
				} else if duration > 0 {
					return 1
				}
				return 0
			}
		} else {
			difference := number - ParseFloat64(b)
			if difference < 0 {
				return -1
			} else if difference > 0 {
				return 1
			}
			return 0
		}
		return strings.Compare(a, ParseString(b))
	case int64:
		a := a.(int64)
		b := ParseInt64(b)
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	case uint64:
		a := a.(uint64)
		b := ParseUint64(b)
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	case float64:
		a := a.(float64)
		b := ParseFloat64(b)
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	}
	return 0
}

func ParseString(value interface{}, values ...string) string {
	switch value.(type) {
	case string:
		return value.(string)
	case int64:
		return strconv.FormatInt(value.(int64), 10)
	case uint64:
		return strconv.FormatUint(value.(uint64), 10)
	case float64:
		return strconv.FormatFloat(value.(float64), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value.(bool))
	case []string:
		return strings.Join(value.([]string), ",")
	case []byte:
		return string(value.([]byte))
	case time.Time:
		return StringifyTime(value.(time.Time))
	case []int64:
		numbers := make([]string, 0)
		for _, number := range value.([]int64) {
			numbers = append(numbers, strconv.FormatInt(number, 10))
		}
		return strings.Join(numbers, ",")
	case []uint64:
		numbers := make([]string, 0)
		for _, number := range value.([]uint64) {
			numbers = append(numbers, strconv.FormatUint(number, 10))
		}
		return strings.Join(numbers, ",")
	case []float64:
		numbers := make([]string, 0)
		for _, number := range value.([]float64) {
			numbers = append(numbers, strconv.FormatFloat(number, 'f', -1, 64))
		}
		return strings.Join(numbers, ",")
	case []interface{}:
		values := make([]string, 0)
		for _, str := range value.([]interface{}) {
			values = append(values, ParseString(str))
		}
		return "[" + strings.Join(values, ",") + "]"
	default:
		if value != nil {
			return string(GetJSON(value))
		}
	}
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func ParseFloat64(value interface{}, values ...float64) float64 {
	switch value.(type) {
	case float64:
		return value.(float64)
	case int64:
		return float64(value.(int64))
	case uint64:
		return float64(value.(uint64))
	case int:
		return float64(value.(int))
	case uint:
		return float64(value.(uint))
	case string:
		str := value.(string)
		if str != "" {
			number, err := strconv.ParseFloat(str, 64)
			if err != nil {
				log.Println(err)
			} else {
				return number
			}
		}
	}
	if len(values) > 0 {
		return values[0]
	}
	return 0.0
}

func ParseInt(value interface{}, values ...int) int {
	switch value.(type) {
	case int:
		return value.(int)
	case int64:
		return int(value.(int64))
	case uint:
		return int(value.(uint))
	case uint64:
		return int(value.(uint64))
	case float64:
		return int(value.(float64))
	case string:
		str := value.(string)
		if str != "" {
			number, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				log.Println(err)
			} else {
				return int(number)
			}
		}
	}
	if len(values) > 0 {
		return values[0]
	}
	return 0
}

func ParseUint(value interface{}, values ...uint) uint {
	switch value.(type) {
	case uint:
		return value.(uint)
	case int:
		return uint(value.(int))
	case int64:
		return uint(value.(int64))
	case uint64:
		return uint(value.(uint64))
	case float64:
		return uint(value.(float64))
	case string:
		str := value.(string)
		if str != "" {
			number, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				log.Println(err)
			} else {
				return uint(number)
			}
		}
	}
	if len(values) > 0 {
		return values[0]
	}
	return 0
}

func ParseInt64(value interface{}, values ...int64) int64 {
	switch value.(type) {
	case int64:
		return value.(int64)
	case uint64:
		return int64(value.(uint64))
	case int:
		return int64(value.(int))
	case uint:
		return int64(value.(uint))
	case float64:
		return int64(value.(float64))
	case string:
		str := value.(string)
		if str != "" {
			number, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				log.Println(err)
			} else {
				return number
			}
		}
	}
	if len(values) > 0 {
		return values[0]
	}
	return 0
}

func ParseUint64(value interface{}, values ...uint64) uint64 {
	switch value.(type) {
	case uint64:
		return value.(uint64)
	case int64:
		return uint64(value.(int64))
	case int:
		return uint64(value.(int))
	case uint:
		return uint64(value.(uint))
	case float64:
		return uint64(value.(float64))
	case string:
		str := value.(string)
		if str != "" {
			number, err := strconv.ParseUint(str, 10, 64)
			if err != nil {
				log.Println(err)
			} else {
				return number
			}
		}
	}
	if len(values) > 0 {
		return values[0]
	}
	return 0
}

func ParseBool(value interface{}, values ...bool) bool {
	switch value.(type) {
	case bool:
		return value.(bool)
	case string:
		str := value.(string)
		if str != "" {
			truth, err := strconv.ParseBool(str)
			if err != nil {
				log.Println(err)
			} else {
				return truth
			}
		}
	}
	if len(values) > 0 {
		return values[0]
	}
	return false
}

func ParseTime(value interface{}, values ...time.Time) time.Time {
	switch value.(type) {
	case time.Time:
		return value.(time.Time)
	case float64:
		return Timestamp(value.(float64))
	case int64:
		return Timestamp(float64(value.(int64)))
	case uint64:
		return Timestamp(float64(value.(uint64)))
	case int:
		return Timestamp(float64(value.(int)))
	case uint:
		return Timestamp(float64(value.(uint64)))
	case string:
		if timestamp, ok := ParseTimestamp(value.(string)); ok {
			return timestamp
		}
	}
	if len(values) > 0 {
		return values[0]
	}
	return time.Unix(0, 0)
}

func ParseMilliseconds(value interface{}) float64 {
	switch value.(type) {
	case float64:
		return value.(float64)
	case int64:
		return float64(value.(int64))
	case uint64:
		return float64(value.(uint64))
	case int:
		return float64(value.(int))
	case uint:
		return float64(value.(uint))
	case string:
		str := value.(string)
		if strings.HasSuffix(str, "ms") {
			return ParseFloat64(strings.TrimSuffix(str, "ms"))
		} else if strings.HasSuffix(str, "µs") || strings.HasSuffix(str, "us") {
			return ParseFloat64(strings.TrimRight(str, "uµs")) / 1000
		} else if strings.HasSuffix(str, "s") {
			return ParseFloat64(strings.TrimSuffix(str, "s")) * 1000
		}
		duration, err := time.ParseDuration(str)
		if err != nil {
			log.Println(err)
		} else {
			return duration.Seconds() * 1000
		}
	}
	return 0.0
}

func ParseMegabytes(value interface{}) int {
	switch value.(type) {
	case int:
		return value.(int) / (1024 * 1024)
	case uint:
		return int(value.(uint)) / (1024 * 1024)
	case int64:
		return int(value.(int64)) / (1024 * 1024)
	case uint64:
		return int(value.(uint64)) / (1024 * 1024)
	case float64:
		return int(value.(float64)) / (1024 * 1024)
	case string:
		str := strings.ToUpper(value.(string))
		if strings.HasSuffix(str, "M") || strings.HasSuffix(str, "MB") {
			return ParseInt(strings.TrimRight(str, "MB"))
		} else if strings.HasSuffix(str, "G") || strings.HasSuffix(str, "GB") {
			return ParseInt(strings.TrimRight(str, "GB")) * 1024
		} else if strings.HasSuffix(str, "K") || strings.HasSuffix(str, "KB") {
			return ParseInt(strings.TrimRight(str, "KB")) / 1024
		} else if strings.HasSuffix(str, "B") {
			return ParseInt(strings.TrimSuffix(str, "B")) / (1024 * 1024)
		}
	}
	return 0
}

func ParseStringArray(value interface{}, values ...[]string) []string {
	switch value.(type) {
	case []string:
		return value.([]string)
	case []interface{}:
		numbers := make([]string, 0)
		for _, number := range value.([]interface{}) {
			numbers = append(numbers, ParseString(number))
		}
		return numbers
	case string:
		value := strings.Trim(value.(string), "[]")
		return strings.Split(value, ",")
	}
	if len(values) > 0 {
		return values[0]
	}
	return make([]string, 0)
}

func ParseUint64Array(value interface{}, values ...[]uint64) []uint64 {
	switch value.(type) {
	case []uint64:
		return value.([]uint64)
	case []string:
		numbers := make([]uint64, 0)
		for _, str := range value.([]string) {
			if str != "" {
				number, err := strconv.ParseUint(str, 10, 64)
				if err != nil {
					log.Println(err)
				} else {
					numbers = append(numbers, number)
				}
			}
		}
		return numbers
	case []interface{}:
		numbers := make([]uint64, 0)
		for _, number := range value.([]interface{}) {
			numbers = append(numbers, ParseUint64(number))
		}
		return numbers
	case string:
		numbers := make([]uint64, 0)
		value := strings.Trim(value.(string), "[]")
		for _, str := range strings.Split(value, ",") {
			if str != "" {
				number, err := strconv.ParseUint(str, 10, 64)
				if err != nil {
					log.Println(err)
				} else {
					numbers = append(numbers, number)
				}
			}
		}
		return numbers
	}
	if len(values) > 0 {
		return values[0]
	}
	return make([]uint64, 0)
}

func StringArrayContains(array []string, value string) bool {
	for _, str := range array {
		if str == value {
			return true
		}
	}
	return false
}

func Uint64ArrayContains(array []uint64, value uint64) bool {
	for _, number := range array {
		if number == value {
			return true
		}
	}
	return false
}

func StringArrayIndexOf(array []string, value string) int {
	for index, str := range array {
		if str == value {
			return index
		}
	}
	return -1
}

func Uint64ArrayIndexOf(array []uint64, value uint64) int {
	for index, number := range array {
		if number == value {
			return index
		}
	}
	return -1
}

func MergeStringValues(value interface{}, values ...string) []string {
	array := make([]string, 0)
	switch value.(type) {
	case []string:
		array = value.([]string)
	case string:
		array = append(array, value.(string))
	}
	if len(values) > 0 {
		for _, value := range values {
			if !StringArrayContains(array, value) {
				array = append(array, value)
			}
		}
	}
	return array
}

func MergeUint64Values(value interface{}, values ...uint64) []uint64 {
	array := make([]uint64, 0)
	switch value.(type) {
	case []uint64:
		array = value.([]uint64)
	case uint64:
		array = append(array, value.(uint64))
	}
	if len(values) > 0 {
		for _, value := range values {
			if !Uint64ArrayContains(array, value) {
				array = append(array, value)
			}
		}
	}
	return array
}

func FilterStringArray(array []string, fn func(string) bool) []string {
	values := make([]string, 0)
	for _, value := range array {
		if fn(value) {
			values = append(values, value)
		}
	}
	return values
}

func FilterUint64Array(array []uint64, fn func(uint64) bool) []uint64 {
	values := make([]uint64, 0)
	for _, value := range array {
		if fn(value) {
			values = append(values, value)
		}
	}
	return values
}
