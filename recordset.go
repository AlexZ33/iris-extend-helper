package iris_extend_helper

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"log"
	"sort"
	"strings"

	"github.com/json-iterator/go"
	"github.com/kataras/iris/v12"
	"github.com/oliveagle/jsonpath"
)

func GetCSV(value interface{}) []byte {
	keys := []string{}
	records := [][]string{keys}
	for _, record := range ParseRecordset(value) {
		values := make([]string, len(keys))
		for key, value := range record {
			content := ParseString(value)
			index := StringArrayIndexOf(keys, key)
			if index == -1 {
				keys = append(keys, key)
				values = append(values, content)
			} else {
				values[index] = content
			}
		}
		records = append(records, values)
	}
	records[0] = keys

	count := len(keys)
	for index, values := range records {
		padding := count - len(values)
		if padding > 0 {
			records[index] = append(values, make([]string, padding)...)
		}
	}

	buffer := bytes.NewBuffer([]byte{})
	w := csv.NewWriter(buffer)
	w.WriteAll(records)
	if err := w.Error(); err != nil {
		log.Println(err)
	}
	return buffer.Bytes()
}

func ParseRecordset(value interface{}) []iris.Map {
	if value == nil {
		return make([]iris.Map, 0)
	}
	switch value.(type) {
	case []iris.Map:
		return value.([]iris.Map)
	case iris.Map:
		return []iris.Map{value.(iris.Map)}
	case [][]string:
		recordset := make([]iris.Map, 0)
		keys := []string{}
		for index, array := range value.([][]string) {
			if index > 0 {
				object := iris.Map{}
				for i, key := range keys {
					object[key] = array[i]
				}
				recordset = append(recordset, object)
			} else {
				keys = array
			}
		}
		return recordset
	case []string:
		recordset := make([]iris.Map, 0)
		for _, str := range value.([]string) {
			object := iris.Map{}
			err := jsoniter.UnmarshalFromString(str, &object)
			if err != nil {
				log.Println(err)
			}
			recordset = append(recordset, object)
		}
		return recordset
	case string:
		str := strings.TrimSpace(value.(string))
		if strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]") {
			recordset := make([]iris.Map, 0)
			err := jsoniter.UnmarshalFromString(str, &recordset)
			if err != nil {
				log.Println(err)
			}
			return recordset
		} else {
			object := iris.Map{}
			err := jsoniter.UnmarshalFromString(str, &object)
			if err != nil {
				log.Println(err)
			}
			return []iris.Map{object}
		}
	case []interface{}:
		recordset := make([]iris.Map, 0)
		for _, record := range value.([]interface{}) {
			recordset = append(recordset, ParseMap(record))
		}
		return recordset
	default:
		s := bytes.TrimSpace(GetJSON(value))
		if bytes.HasPrefix(s, []byte("[")) && bytes.HasSuffix(s, []byte("]")) {
			recordset := make([]iris.Map, 0)
			err := jsoniter.Unmarshal(s, &recordset)
			if err != nil {
				log.Println(err)
			}
			return recordset
		} else {
			object := iris.Map{}
			err := jsoniter.Unmarshal(s, &object)
			if err != nil {
				log.Println(err)
			}
			return []iris.Map{object}
		}
	}
	return make([]iris.Map, 0)
}

func ParseRows(rows *sql.Rows, offset int, limit int) ([]iris.Map, int) {
	columns, err := rows.Columns()
	if err != nil {
		log.Println(err)
		return make([]iris.Map, 0), 0
	}
	count := 0
	entries := make([]iris.Map, 0)
	values := make([]interface{}, len(columns))
	for index := range values {
		values[index] = new(interface{})
	}
	for rows.Next() {
		entry := iris.Map{}
		if err := rows.Scan(values...); err != nil {
			log.Println(err)
		} else if count >= offset && len(entries) < limit {
			for index, column := range columns {
				value := *(values[index].(*interface{}))
				switch value.(type) {
				case []byte:
					s := value.([]byte)
					object := iris.Map{}
					if err := jsoniter.Unmarshal(s, &object); err != nil {
						if bytes.HasPrefix(s, []byte("{")) && bytes.HasSuffix(s, []byte("}")) {
							str := string(bytes.Trim(s, "{}"))
							entry[column] = strings.Split(str, ",")
						} else {
							entry[column] = string(s)
						}
					} else {
						entry[column] = object
					}
				default:
					entry[column] = value
				}
			}
			entries = append(entries, entry)
		}
		count += 1
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
	}
	rows.Close()
	return entries, count
}

func RecordsetContains(array []iris.Map, record iris.Map) bool {
	for _, object := range array {
		if MapIntersects(object, record) {
			return true
		}
	}
	return false
}

func RecordsetIndexOf(array []iris.Map, record iris.Map) int {
	for index, object := range array {
		if MapIntersects(object, record) {
			return index
		}
	}
	return -1
}

func ExtendRecordset(recordset interface{}, records ...iris.Map) []iris.Map {
	array := make([]iris.Map, 0)
	switch recordset.(type) {
	case []iris.Map:
		array = recordset.([]iris.Map)
	case iris.Map:
		array = append(array, recordset.(iris.Map))
	}
	if len(records) > 0 {
		for _, record := range records {
			if !RecordsetContains(array, record) {
				array = append(array, record)
			}
		}
	}
	return array
}

func FilterRecordset(recordset []iris.Map, fn func(iris.Map) bool) []iris.Map {
	records := make([]iris.Map, 0)
	for _, record := range recordset {
		if fn(record) {
			records = append(records, record)
		}
	}
	return records
}

func TransformRecordset(recordset []iris.Map, transform iris.Map) []iris.Map {
	records := make([]iris.Map, 0)
	switch ParseString(transform["name"]) {
	case "map", "normalize", "tuple", "expand":
		for _, record := range recordset {
			records = append(records, TransformMap(record, transform))
		}
	case "sort":
		fields := ParseStringArray(transform["fields"])
		orders := ParseStringArray(transform["orders"])
		records = recordset
		if len(fields) > 0 {
			sort.SliceStable(records, func(i, j int) bool {
				recordI := records[i]
				recordJ := records[j]
				for index, field := range fields {
					sign := Compare(recordI[field], recordJ[field])
					order := "descending"
					if index < len(orders) {
						order = ParseString(orders[index], "descending")
					}
					if sign == -1 {
						if order == "descending" {
							return false
						} else {
							return true
						}
					} else if sign == 1 {
						if order == "descending" {
							return true
						} else {
							return false
						}
					}
				}
				return false
			})
		}
	case "slice":
		length := len(recordset)
		if length > 0 {
			start := ParseInt(transform["start"], 0)
			end := ParseInt(transform["end"], length)
			if start < 0 {
				start += length
			}
			if end < 0 {
				end += length
			}
			if start < end && start >= 0 && end <= length {
				records = recordset[start:end]
			}
		}
	case "extract":
		path := ParseString(transform["path"])
		if strings.HasPrefix(path, "$.") && len(recordset) > 0 {
			if result, err := jsonpath.JsonPathLookup(recordset, path); err != nil {
				log.Println(err)
			} else {
				records = ParseRecordset(result)
			}
		}
	}
	return records
}

func TransformRecordsetArray(inputs [][]iris.Map, transform iris.Map) [][]iris.Map {
	outputs := [][]iris.Map{}
	switch ParseString(transform["name"]) {
	case "copy":
		copies := ParseUint64Array(transform["copies"])
		length := len(copies)
		for index, input := range inputs {
			if index < length {
				for count := copies[index]; count > 0; count-- {
					outputs = append(outputs, input)
				}
			} else {
				outputs = append(outputs, input)
			}
		}
	case "concat":
		output := []iris.Map{}
		for _, input := range inputs {
			output = append(output, input...)
		}
		outputs = append(outputs, output)
	case "extend":
		output := []iris.Map{}
		for _, input := range inputs {
			output = ExtendRecordset(output, input...)
		}
		outputs = append(outputs, output)
	case "merge":
		output := []iris.Map{}
		for _, input := range inputs {
			length := len(output)
			for index, record := range input {
				if index < length {
					output[index] = ExtendMap(output[index], record)
				} else {
					output = append(output, record)
				}
			}
		}
		outputs = append(outputs, output)
	case "partition":
		field := ParseString(transform["field"])
		values := ParseStringArray(transform["values"])
		if len(values) == 0 {
			for _, input := range inputs {
				for _, record := range input {
					if value, ok := record[field]; ok {
						value := ParseString(value)
						if !StringArrayContains(values, value) {
							values = append(values, value)
						}
					}
				}
			}
		}
		outputs = make([][]iris.Map, len(values))
		for index, value := range values {
			output := []iris.Map{}
			for _, input := range inputs {
				input = FilterRecordset(input, func(record iris.Map) bool {
					return record[field] == value
				})
				if len(input) > 0 {
					output = append(output, input...)
				}
			}
			outputs[index] = output
		}
	case "map", "normalize", "tuple", "expand", "sort", "slice", "extract":
		for _, input := range inputs {
			outputs = append(outputs, TransformRecordset(input, transform))
		}
	}
	return outputs
}
