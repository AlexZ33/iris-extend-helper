package iris_extend_helper

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
)

func WrapError(err error) error {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		return fmt.Errorf("<%s#%v>%w", file, line, err)
	}
	return nil
}

func ParseErrorSource(str string) (string, string, bool) {
	re := regexp.MustCompile(`<[^<>]+#\d+>`)
	matches := re.FindAllString(str, -1)
	if matches != nil {
		path := strings.Join(matches, "")
		source := strings.ReplaceAll(strings.ReplaceAll(path, "<", ""), ">", "")
		message := strings.Replace(str, path, "", 1)
		return source, message, true
	}
	return "", str, false
}
