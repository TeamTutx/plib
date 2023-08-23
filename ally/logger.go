package ally

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

var (
	ignoreHeader map[string]struct{}
)

func init() {
	ignoreHeader = map[string]struct{}{
		"Accept-Encoding":          {},
		"Connection":               {},
		"Accept-Language":          {},
		"User-Agent":               {},
		"Postman-Token":            {},
		"Accept":                   {},
		"X-Postman-Interceptor-Id": {},
		"Cookie":                   {},
		"Content-Length":           {},
		"Origin":                   {},
		"Cache-Control":            {},
		"X-Newrelic-Transaction":   {},
		"X-Newrelic-Id":            {},
	}
}

//LogReq will log request
func LogReq(r *http.Request) {
	var body interface{}
	r.Body, body = CopyReqBody(r.Body)
	fmt.Println()
	fmt.Println("----REQUEST----")

	fmt.Println(r.Method, r.URL)
	if r.Method != http.MethodGet {
		fmt.Println(body)
		fmt.Println()
	}
	fmt.Println("HEADER:")
	for k, v := range r.Header {
		if _, exists := ignoreHeader[k]; exists {
			continue
		}
		fmt.Println(k, v)
	}
	fmt.Println("----------------------------------------------------------------")
}

//HeaderString will return header's key value as string
func HeaderString(h http.Header) (msg string) {
	for k, v := range h {
		if _, exists := ignoreHeader[k]; exists {
			continue
		}

		msg += fmt.Sprintf("%v %v\n", k, v)
	}
	return
}

//CopyReqBody request body
func CopyReqBody(reqBody io.ReadCloser) (originalBody io.ReadCloser, copyBody interface{}) {
	bodyByte, _ := ioutil.ReadAll(reqBody)
	// if err := json.Unmarshal(bodyByte, &copyBody); err != nil {
	copyBody = string(bodyByte)
	// }
	originalBody = ioutil.NopCloser(bytes.NewBuffer(bodyByte))
	return
}
