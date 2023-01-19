package iris_extend_helper

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/json-iterator/go"
	"github.com/kataras/iris/v12"
	"github.com/pelletier/go-toml"
)

var client *http.Client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
	},
}

func GetMimeType(rawurl string) string {
	res, err := http.Get(rawurl)
	if err != nil {
		log.Println(err)
	}
	mediatype, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		log.Println(err)
	}
	res.Body.Close()
	return mediatype
}

func ParseURL(rawurl string) (*url.URL, bool) {
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Println(err)
	} else {
		return u, true
	}
	return u, false
}

func AliyunRequestSign(request iris.Map) iris.Map {
	method := ParseString(request["method"], "GET")
	content := method + "\n"
	headers := ParseMap(request["headers"])
	if contentMD5, ok := headers["Content-MD5"]; ok {
		content += ParseString(contentMD5) + "\n"
	}
	contentType := ParseString(headers["Content-Type"], "application/json")
	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	content += contentType + "\n" + date + "\n"

	name := ParseString(request["aliyun_product_name"])
	headerPrefix := "x-" + strings.ToLower(name) + "-"
	customHeaders := []string{}
	for key := range headers {
		header := strings.ToLower(key)
		if strings.HasPrefix(header, headerPrefix) {
			customHeaders = append(customHeaders, header)
		}
	}
	sort.Strings(customHeaders)
	for _, header := range customHeaders {
		content += header + ":" + ParseString(headers[header]) + "\n"
	}

	rawurl := ParseString(request["url"])
	if u, err := url.Parse(rawurl); err != nil {
		log.Println(err)
	} else {
		content += u.Path
	}

	accessId := ParseString(request["aliyun_access_id"])
	accessKey := ParseString(request["aliyun_access_key"])
	mac := hmac.New(sha1.New, []byte(accessKey))
	mac.Write([]byte(content))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	headers["Content-Type"] = contentType
	headers["Date"] = date
	headers["Authorization"] = strings.ToUpper(name) + " " + accessId + ":" + signature
	request["headers"] = headers
	return request
}

func GetData(request iris.Map) ([]byte, bool) {
	rawurl := ParseString(request["url"])
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Println(err)
	}
	if u.Scheme == "" {
		return make([]byte, 0), false
	}
	if params, ok := request["url_params"]; ok {
		params := ParseMap(params)
		values := u.Query()
		for key, value := range params {
			values.Set(key, ParseString(value))
		}
		u.RawQuery = values.Encode()
	}
	header := http.Header{}
	if headers, ok := request["headers"]; ok {
		headers := ParseMap(headers)
		for key, value := range headers {
			header.Set(key, ParseString(value))
		}
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Println(err)
		return make([]byte, 0), false
	}
	req.Header = header

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return make([]byte, 0), false
	}
	defer res.Body.Close()
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	} else {
		code := res.StatusCode
		return content, code >= 200 && code < 400
	}
	return make([]byte, 0), false
}

func PostData(request iris.Map) ([]byte, bool) {
	rawurl := ParseString(request["url"])
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Println(err)
	}
	if u.Scheme == "" {
		return make([]byte, 0), false
	}
	if params, ok := request["url_params"]; ok {
		params := ParseMap(params)
		values := u.Query()
		for key, value := range params {
			values.Set(key, ParseString(value))
		}
		u.RawQuery = values.Encode()
	}
	header := http.Header{}
	header.Set("Content-Type", "application/x-www-form-urlencoded")
	if headers, ok := request["headers"]; ok {
		headers := ParseMap(headers)
		for key, value := range headers {
			header.Set(key, ParseString(value))
		}
	}
	data := make([]byte, 0)
	if body, ok := request["body"]; ok {
		switch header.Get("Content-Type") {
		case "application/x-www-form-urlencoded":
			body := ParseMap(body)
			values := url.Values{}
			for key, value := range body {
				values.Set(key, ParseString(value))
			}
			data = []byte(values.Encode())
		default:
			data = []byte(ParseString(body))
		}
	}
	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(data))
	if err != nil {
		log.Println(err)
		return make([]byte, 0), false
	}
	req.Header = header

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return make([]byte, 0), false
	}
	defer res.Body.Close()
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	} else {
		code := res.StatusCode
		return content, code >= 200 && code < 400
	}
	return make([]byte, 0), false
}

func RetryGetData(request iris.Map, config *toml.Tree) ([]byte, bool) {
	mutex := new(sync.Mutex)
	mode := GetString(config, "fail-mode", "failtry")
	retries := GetInt(config, "max-retries")
	if mode == "failfast" || retries == 0 {
		return make([]byte, 0), false
	}
	endpoints := GetStringArray(config, "service-endpoints")
	strategy := GetString(config, "backoff-strategy")
	duration := GetDuration(config, "initial-backoff")
	if duration < time.Millisecond {
		duration = time.Millisecond
	}
	count := 0
	ticks := 0
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			ready := true
			ticks += 1
			if strategy == "linear" {
				ready = ticks >= (count+1)*(count+2)/2
			} else if strategy == "exponential" {
				ready = ticks >= (1 << uint(count))
			}
			if ready {
				if count < retries {
					length := len(endpoints)
					if length > 0 && mode == "failover" {
						request["url"] = endpoints[count%length]
					}
					if result, ok := GetData(request); ok {
						if mode == "failover" {
							mutex.Lock()
							config.Set("service-url", request["url"])
							mutex.Unlock()
						}
						return result, true
					}
					count += 1
				} else {
					return make([]byte, 0), false
				}
			}
		}
	}
	return make([]byte, 0), false
}

func RetryPostData(request iris.Map, config *toml.Tree) ([]byte, bool) {
	mutex := new(sync.Mutex)
	mode := GetString(config, "fail-mode", "failtry")
	retries := GetInt(config, "max-retries")
	if mode == "failfast" || retries == 0 {
		return make([]byte, 0), false
	}
	endpoints := GetStringArray(config, "service-endpoints")
	strategy := GetString(config, "backoff-strategy")
	duration := GetDuration(config, "initial-backoff")
	if duration < time.Millisecond {
		duration = time.Millisecond
	}
	count := 0
	ticks := 0
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			ready := true
			ticks += 1
			if strategy == "linear" {
				ready = ticks >= (count+1)*(count+2)/2
			} else if strategy == "exponential" {
				ready = ticks >= (1 << uint(count))
			}
			if ready {
				if count < retries {
					length := len(endpoints)
					if length > 0 && mode == "failover" {
						request["url"] = endpoints[count%length]
					}
					if result, ok := PostData(request); ok {
						if mode == "failover" {
							mutex.Lock()
							config.Set("service-url", request["url"])
							mutex.Unlock()
						}
						return result, true
					}
					count += 1
				} else {
					return make([]byte, 0), false
				}
			}
		}
	}
	return make([]byte, 0), false
}

func CheckResponseResult(result []byte) ([]byte, bool) {
	data := []byte{}
	content := jsoniter.Get(result, "data")
	if content.ValueType() == jsoniter.ObjectValue {
		data = []byte(content.ToString())
	} else {
		data = result
	}

	success := jsoniter.Get(result, "success")
	switch success.ValueType() {
	case jsoniter.BoolValue:
		return data, success.ToBool() == true
	case jsoniter.StringValue:
		return data, success.ToString() == "true"
	}
	if status := jsoniter.Get(result, "status").ToString(); status != "" {
		return data, strings.ToLower(status) == "success"
	}
	if code := jsoniter.Get(result, "code").ToInt(); code >= 100 && code <= 600 {
		return data, code >= 200 && code < 400
	}
	return data, false
}

func CheckRequestParams(request iris.Map, params []string) ([]string, bool) {
	invalidParams := make([]string, 0)
	valid := true
	for _, param := range params {
		if value, ok := request[param]; ok {
			switch value.(type) {
			case string:
				if strings.TrimSpace(value.(string)) == "" {
					invalidParams = append(invalidParams, param)
					valid = false
				}
			}
		} else {
			invalidParams = append(invalidParams, param)
			valid = false
		}
	}
	return invalidParams, valid
}

func ConcatRequestParams(request iris.Map, params []string) string {
	values := url.Values{}
	for _, param := range params {
		if value, ok := request[param]; ok {
			values.Set(param, ParseString(value))
		}
	}
	return values.Encode()
}
