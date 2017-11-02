package util

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	httpClient = &http.Client{}
)

// FormatJSON 格式化输出
func FormatJSON(obj interface{}) {
	var err error
	b, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		log.Println(err)
	}
	log.Println(string(b))
}

// CreateRequest 发起HTTP请求
func CreateRequest(method, url string, body io.Reader, token string) ([]byte, error) {
	request, newRrqErr := http.NewRequest(method, url, body)
	if newRrqErr != nil {
		return nil, newRrqErr
	}

	if len(token) != 0 {
		request.Header.Add("token", token)
	}

	response, doErr := httpClient.Do(request)

	defer func() {
		if err := recover(); err != nil {
			log.Println(method, url, err)
		}
	}()

	statusCode := response.StatusCode
	status := statusCode >= 200 && statusCode <= 299
	b, readErr := ioutil.ReadAll(response.Body)

	if doErr != nil {
		defer response.Body.Close()
		return b, doErr
	}

	if !status {
		return b, errors.New(url + " " + response.Status)
	}

	return b, readErr
}
