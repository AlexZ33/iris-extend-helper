package iris_extend_helper

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/inflection"
	"github.com/kataras/iris/v12"
)

func Plural(s string) string {
	return inflection.Plural(strings.Trim(s, " "))
}

func NormalizeName(s string) string {
	re := regexp.MustCompile(`([a-z\d])([A-Z])`)
	s = re.ReplaceAllString(s, "${1}_${2}")
	if strings.ContainsAny(s, "-, ") {
		r := strings.NewReplacer("-", "_", ",", "_", " ", "_")
		s = r.Replace(s)
	}
	if strings.Contains(s, "__") {
		re := regexp.MustCompile(`_+`)
		s = re.ReplaceAllString(s, "_")
	}
	return strings.ToLower(s)
}

func ContainsPrefix(s string, values []string) bool {
	for _, value := range values {
		if strings.HasPrefix(s, value) {
			return true
		}
	}
	return false
}

func ContainsSuffix(s string, values []string) bool {
	for _, value := range values {
		if strings.HasSuffix(s, value) {
			return true
		}
	}
	return false
}

func StripKeywords(s string, values []string) string {
	for _, value := range values {
		if strings.Contains(s, value) {
			s = strings.ReplaceAll(s, value, "")
		}
	}
	return s
}

func FormatString(s string, params iris.Map) string {
	re := regexp.MustCompile(`\$\{[\w\+\-\.]+\}`)
	return re.ReplaceAllStringFunc(s, func(substr string) string {
		key := strings.Trim(substr, "${}")
		if strings.HasPrefix(key, "system.current") && strings.ContainsAny(key, "+-") {
			expr := strings.TrimPrefix(key, "system.current")
			if duration, err := time.ParseDuration(expr); err != nil {
				log.Println(err)
			} else {
				return time.Now().Add(duration).Format("2006-01-02 15:04:05")
			}
		}
		if value, ok := params[key]; ok {
			return ParseString(value)
		}
		return substr
	})
}
