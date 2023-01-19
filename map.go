package iris_extend_helper

import (
	"log"
	"math"
	"math/big"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/json-iterator/go"
	"github.com/kataras/iris/v12"
)

func ParseMap(value interface{}) iris.Map {
	if value == nil {
		return iris.Map{}
	}
	switch value.(type) {
	case iris.Map:
		return value.(iris.Map)
	case string:
		object := iris.Map{}
		err := jsoniter.UnmarshalFromString(value.(string), &object)
		if err != nil {
			log.Println(err)
		}
		return object
	case []byte:
		object := iris.Map{}
		err := jsoniter.Unmarshal(value.([]byte), &object)
		if err != nil {
			log.Println(err)
		}
		return object
	default:
		object := iris.Map{}
		err := jsoniter.Unmarshal(GetJSON(value), &object)
		if err != nil {
			log.Println(err)
		}
		return object
	}
	return iris.Map{}
}

func NormalizeMap(value interface{}) iris.Map {
	object := iris.Map{}
	for key, value := range ParseMap(value) {
		object[NormalizeName(key)] = value
	}
	return object
}

func CheckMapKeys(object iris.Map, keys []string) bool {
	for _, key := range keys {
		if _, ok := object[key]; !ok {
			return false
		}
	}
	return true
}

func ExtendMap(source iris.Map, target interface{}) iris.Map {
	output := iris.Map{}
	for key, value := range source {
		output[key] = value
	}
	for key, value := range ParseMap(target) {
		output[key] = value
	}
	return output
}

func TransformMap(input iris.Map, transform iris.Map) iris.Map {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	output := iris.Map{}
	switch ParseString(transform["name"]) {
	case "map":
		fields := ParseStringArray(transform["fields"])
		mappings := ParseStringArray(transform["mappings"])
		values := ParseStringArray(transform["values"])
		pick := ParseBool(transform["pick"])
		mappingsLength := len(mappings)
		valuesLength := len(values)
		for index, field := range fields {
			value, ok := input[field]
			if !ok && index < valuesLength {
				value = values[index]
			}
			if value != nil {
				if ParseBool(transform["stringify"]) {
					if str, err := json.MarshalToString(value); err != nil {
						log.Println(err)
					} else {
						value = str
					}
				}
				if index < mappingsLength {
					output[mappings[index]] = value
				} else if !pick {
					output[field] = value
				}
			}
		}
	case "normalize":
		for key, value := range input {
			output[NormalizeName(key)] = value
		}
	case "tuple":
		fields := ParseStringArray(transform["fields"])
		mapping := ParseString(transform["mapping"])
		list := make([]string, 0)
		for _, field := range fields {
			list = append(list, ParseString(input[field]))
		}
		output[mapping] = list
	case "expand":
		field := ParseString(transform["field"])
		mappings := ParseStringArray(transform["mappings"])
		list := ParseStringArray(input[field])
		if len(list) > 0 && len(list) <= len(mappings) {
			for index, value := range list {
				output[mappings[index]] = value
			}
		}
	}
	return output
}

func MapIntersects(source iris.Map, target iris.Map) bool {
	for key, value := range target {
		if !reflect.DeepEqual(source[key], value) {
			return false
		}
	}
	return true
}

func GetMapRule(schema iris.Map, field string) (string, iris.Map) {
	if rule, ok := schema[field]; ok {
		return field, ParseMap(rule)
	} else {
		for key, rule := range schema {
			rule := ParseMap(rule)
			if normalize, ok := rule["normalize"]; ok && ParseBool(normalize) {
				field = NormalizeName(field)
			}
			if aliases, ok := rule["aliases"]; ok {
				aliases := ParseStringArray(aliases)
				if StringArrayContains(aliases, field) {
					return key, rule
				}
			}
		}
	}
	return field, iris.Map{}
}

func CheckMapFields(object iris.Map, fields []string) ([]string, bool) {
	missingFields := make([]string, 0)
	valid := true
	for _, field := range fields {
		if _, ok := object[field]; !ok {
			missingFields = append(missingFields, field)
			valid = false
		}
	}
	return missingFields, valid
}

func CheckMapValue(value interface{}, rule iris.Map) (interface{}, bool) {
	if valueType, ok := rule["type"]; ok {
		switch ParseString(valueType) {
		case "null":
			return nil, value == nil
		case "int":
			number := ParseInt64(value)
			if minValue, ok := rule["min_value"]; ok {
				minValue := ParseInt64(minValue)
				if number < minValue {
					if clamp, ok := rule["clamp"]; ok && ParseBool(clamp) {
						return minValue, true
					} else {
						return number, false
					}
				}
			}
			if maxValue, ok := rule["max_value"]; ok {
				maxValue := ParseInt64(maxValue)
				if number > maxValue {
					if clamp, ok := rule["clamp"]; ok && ParseBool(clamp) {
						return maxValue, true
					} else {
						return number, false
					}
				}
			}
			return number, true
		case "float":
			number := ParseFloat64(value)
			if minValue, ok := rule["min_value"]; ok {
				minValue := ParseFloat64(minValue)
				if number < minValue {
					if clamp, ok := rule["clamp"]; ok && ParseBool(clamp) {
						return minValue, true
					} else {
						return number, false
					}
				}
			}
			if maxValue, ok := rule["max_value"]; ok {
				maxValue := ParseFloat64(maxValue)
				if number > maxValue {
					if clamp, ok := rule["clamp"]; ok && ParseBool(clamp) {
						return maxValue, true
					} else {
						return number, false
					}
				}
			}
			return number, true
		case "decimal":
			precision := uint(ParseFloat64(rule["precision"], 16) * math.Log2(10))
			scale := ParseInt(rule["scale"], -1)
			str := ParseString(value)
			number, ok := new(big.Float).SetPrec(precision).SetString(str)
			if !ok {
				return str, false
			}
			if minValue, ok := rule["min_value"]; ok {
				minValue, ok := new(big.Float).SetPrec(precision).SetString(ParseString(minValue))
				if ok && number.Cmp(minValue) == -1 {
					if clamp, ok := rule["clamp"]; ok && ParseBool(clamp) {
						return minValue.Text('f', scale), true
					} else {
						return number.Text('f', scale), false
					}
				}
			}
			if maxValue, ok := rule["max_value"]; ok {
				maxValue, ok := new(big.Float).SetPrec(precision).SetString(ParseString(maxValue))
				if ok && number.Cmp(maxValue) == 1 {
					if clamp, ok := rule["clamp"]; ok && ParseBool(clamp) {
						return maxValue.Text('f', scale), true
					} else {
						return number.Text('f', scale), false
					}
				}
			}
			return number.Text('f', scale), true
		case "bool":
			return ParseBool(value), true
		case "string":
			str := ParseString(value)
			if trimSpace, ok := rule["trim_space"]; ok && ParseBool(trimSpace) {
				str = strings.TrimSpace(str)
			}
			if length, ok := rule["length"]; ok {
				if len(str) != ParseInt(length) {
					return str, false
				}
			}
			if minLength, ok := rule["min_length"]; ok {
				if len(str) < ParseInt(minLength) {
					return str, false
				}
			}
			if maxLength, ok := rule["max_length"]; ok {
				if len(str) > ParseInt(maxLength) {
					return str, false
				}
			}
			if prefix, ok := rule["prefix"]; ok {
				if !strings.HasPrefix(str, ParseString(prefix)) {
					return str, false
				}
			}
			if suffix, ok := rule["suffix"]; ok {
				if !strings.HasSuffix(str, ParseString(suffix)) {
					return str, false
				}
			}
			if pattern, ok := rule["pattern"]; ok {
				if re, err := regexp.Compile(ParseString(pattern)); err != nil {
					log.Println(err)
				} else {
					return str, re.MatchString(str)
				}
			}
			return str, true
		case "uuid":
			id := ParseString(value)
			if _, err := uuid.Parse(id); err != nil {
				return id, false
			}
			return id, true
		case "date":
			date := ParseTime(value)
			layout := "2006-01-02"
			if minValue, ok := rule["min_value"]; ok {
				minValue := ParseTime(minValue)
				if date.Before(minValue) {
					if clamp, ok := rule["clamp"]; ok && ParseBool(clamp) {
						return minValue.Format(layout), true
					} else {
						return date.Format(layout), false
					}
				}
			}
			if maxValue, ok := rule["max_value"]; ok {
				maxValue := ParseTime(maxValue)
				if date.After(maxValue) {
					if clamp, ok := rule["clamp"]; ok && ParseBool(clamp) {
						return maxValue.Format(layout), true
					} else {
						return date.Format(layout), false
					}
				}
			}
			return date.Format(layout), true
		case "datetime":
			datetime := ParseTime(value)
			layout := "2006-01-02 15:04:05"
			if minValue, ok := rule["min_value"]; ok {
				minValue := ParseTime(minValue)
				if datetime.Before(minValue) {
					if clamp, ok := rule["clamp"]; ok && ParseBool(clamp) {
						return minValue.Format(layout), true
					} else {
						return datetime.Format(layout), false
					}
				}
			}
			if maxValue, ok := rule["max_value"]; ok {
				maxValue := ParseTime(maxValue)
				if datetime.After(maxValue) {
					if clamp, ok := rule["clamp"]; ok && ParseBool(clamp) {
						return maxValue.Format(layout), true
					} else {
						return datetime.Format(layout), false
					}
				}
			}
			return datetime.Format(layout), true
		case "enum":
			item := ParseString(value)
			values := ParseStringArray(rule["values"])
			if StringArrayContains(values, item) {
				return item, true
			}
			return item, false
		case "union":
			valueTypes := GetMapValueTypes(value)
			types := ParseRecordset(rule["types"])
			for _, item := range types {
				itemType := ParseString(item["type"])
				if StringArrayContains(valueTypes, itemType) {
					return CheckMapValue(value, item)
				}
			}
			return value, false
		case "array":
			array := []interface{}{}
			element := ParseMap(rule["element"])
			for _, value := range value.([]interface{}) {
				if item, ok := CheckMapValue(value, element); ok {
					array = append(array, item)
				} else {
					return array, false
				}
			}
			if length, ok := rule["length"]; ok {
				if len(array) != ParseInt(length) {
					return array, false
				}
			}
			if minLength, ok := rule["min_length"]; ok {
				if len(array) < ParseInt(minLength) {
					return array, false
				}
			}
			if maxLength, ok := rule["max_length"]; ok {
				if len(array) > ParseInt(maxLength) {
					return array, false
				}
			}
			return array, true
		case "tuple":
			tuple := []interface{}{}
			elements := ParseRecordset(rule["elements"])
			for index, value := range value.([]interface{}) {
				if item, ok := CheckMapValue(value, elements[index]); ok {
					tuple = append(tuple, item)
				} else {
					return tuple, false
				}
			}
			if length, ok := rule["length"]; ok {
				if len(tuple) != ParseInt(length) {
					return tuple, false
				}
			}
			if minLength, ok := rule["min_length"]; ok {
				if len(tuple) < ParseInt(minLength) {
					return tuple, false
				}
			}
			if maxLength, ok := rule["max_length"]; ok {
				if len(tuple) > ParseInt(maxLength) {
					return tuple, false
				}
			}
			return tuple, true
		case "object":
			object := iris.Map{}
			fields := ParseMap(rule["fields"])
			for key, value := range ParseMap(value) {
				key, entry := GetMapRule(fields, key)
				if item, ok := CheckMapValue(value, entry); ok {
					object[key] = item
				} else {
					return object, false
				}
			}
			return object, true
		}
	}
	return value, true
}

func ExportMapValue(value interface{}, rule iris.Map) interface{} {
	if valueType, ok := rule["type"]; ok {
		switch ParseString(valueType) {
		case "null":
			return nil
		case "int":
			number := ParseInt64(value)
			if sensitive, ok := rule["sensitive"]; ok && ParseBool(sensitive) {
				return rand.Int63n(number)
			}
			return number
		case "float":
			number := ParseFloat64(value)
			if sensitive, ok := rule["sensitive"]; ok && ParseBool(sensitive) {
				return number * rand.Float64()
			}
			return number
		case "decimal":
			precision := uint(ParseFloat64(rule["precision"], 16) * math.Log2(10))
			scale := ParseInt(rule["scale"], -1)
			str := ParseString(value)
			number, ok := new(big.Float).SetPrec(precision).SetString(str)
			if !ok {
				return str
			}
			if sensitive, ok := rule["sensitive"]; ok && ParseBool(sensitive) {
				number.Mul(number, big.NewFloat(rand.Float64()))
			}
			return number.Text('f', scale)
		case "bool":
			if sensitive, ok := rule["sensitive"]; ok && ParseBool(sensitive) {
				if rand.Intn(2) == 1 {
					return true
				}
				return false
			}
			return ParseBool(value)
		case "string":
			str := ParseString(value)
			if sensitive, ok := rule["sensitive"]; ok && ParseBool(sensitive) {
				if mask, ok := rule["mask"]; ok {
					if re, err := regexp.Compile(ParseString(mask)); err != nil {
						log.Println(err)
					} else if matches := re.FindStringSubmatchIndex(str); len(matches) > 2 {
						for i, index := range matches {
							if i > 2 && i%2 == 1 {
								start := matches[i-1]
								end := index
								str = str[:start] + strings.Repeat("*", end-start) + str[end:]
							}
						}
					}
				}
			}
			return str
		case "uuid":
			id := ParseString(value)
			if sensitive, ok := rule["sensitive"]; ok && ParseBool(sensitive) {
				id = Id()
			}
			return id
		case "date":
			date := ParseTime(value)
			layout := "2006-01-02"
			if sensitive, ok := rule["sensitive"]; ok && ParseBool(sensitive) {
				return date.AddDate(0, rand.Intn(12), rand.Intn(31)).Format(layout)
			}
			return date.Format(layout)
		case "datetime":
			datetime := ParseTime(value)
			layout := "2006-01-02 15:04:05"
			if sensitive, ok := rule["sensitive"]; ok && ParseBool(sensitive) {
				seconds := datetime.Unix() + rand.Int63n(604800)
				return time.Unix(seconds, 0).Format(layout)
			}
			return datetime.Format(layout)
		case "enum":
			if sensitive, ok := rule["sensitive"]; ok && ParseBool(sensitive) {
				values := ParseStringArray(rule["values"])
				return values[rand.Intn(len(values))]
			}
			return ParseString(value)
		case "union":
			valueTypes := GetMapValueTypes(value)
			types := ParseRecordset(rule["types"])
			for _, item := range types {
				itemType := ParseString(item["type"])
				if StringArrayContains(valueTypes, itemType) {
					return ExportMapValue(value, item)
				}
			}
			return value
		case "array":
			array := []interface{}{}
			element := ParseMap(rule["element"])
			for _, value := range value.([]interface{}) {
				array = append(array, ExportMapValue(value, element))
			}
			return array
		case "tuple":
			tuple := []interface{}{}
			elements := ParseRecordset(rule["elements"])
			for index, value := range value.([]interface{}) {
				tuple = append(tuple, ExportMapValue(value, elements[index]))
			}
			return tuple
		case "object":
			object := iris.Map{}
			fields := ParseMap(rule["fields"])
			for key, value := range ParseMap(value) {
				key, entry := GetMapRule(fields, key)
				object[key] = ExportMapValue(value, entry)
			}
			return object
		}
	}
	return value
}

func GetMapValueTypes(value interface{}) []string {
	types := []string{}
	switch value.(type) {
	case int, uint, int64, uint64:
		types = append(types, "int", "decimal", "float")
	case float64:
		types = append(types, "decimal", "float")
	case bool:
		types = append(types, "bool")
	case string:
		str := value.(string)
		length := len(str)
		if length >= 32 && length <= 36 {
			uuidRegex := regexp.MustCompile(`[0-9a-f\-]{32,36}`)
			if uuidRegex.MatchString(str) {
				if _, err := uuid.Parse(str); err != nil {
					log.Println(err)
				} else {
					types = append(types, "uuid")
				}
			}
		} else if length >= 10 && length <= 35 {
			dateRegex := regexp.MustCompile(`^[1-2][0-9]{3}`)
			if dateRegex.MatchString(str) {
				if length > 10 {
					if _, ok := ParseTimestamp(str); ok {
						types = append(types, "datetime")
					}
				} else {
					if _, ok := ParseTimestamp(str); ok {
						types = append(types, "date", "datetime")
					}
				}
			}
		} else if str == "true" || str == "false" {
			types = append(types, "bool")
		}
		types = append(types, "string", "enum")
	case []int, []uint, []int64, []uint64, []float64, []bool, []string, []iris.Map, []interface{}:
		types = append(types, "array", "tuple")
	case iris.Map:
		types = append(types, "object")
	}
	if len(types) > 0 {
		types = append(types, "union")
	}
	return types
}
