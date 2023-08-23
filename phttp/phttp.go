package phttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"gitlab.com/g-harshit/plib/perror"
)

//HTTPReq Parameters required to make http call
type HTTPReq struct {
	URL       string
	Method    string
	Header    map[string]string
	Body      interface{}
	FormValue io.Reader
	Log       Logger
}

//Logger interface which need to be implemented to start logging
type Logger interface {
	Printf(req *HTTPReq, resp *HTTPRes)
}

//HTTPRes HTTP call response
type HTTPRes struct {
	Status     bool
	StatusCode int
	Body       string
}

//NewReq will create new req
func NewReq(method string, url string) HTTPReq {
	return HTTPReq{
		URL:    url,
		Method: method,
	}
}

//setHeader : Set HTTP Request Headers
func setHeader(req *http.Request, header map[string]string) *http.Request {
	for headerK, headerV := range header {
		req.Header.Add(headerK, headerV)
	}
	return req
}

//readResonse : Read HTTP Request Response
func readResonse(resp *http.Response) (httpRes HTTPRes, err error) {
	respCode := resp.StatusCode
	httpRes.StatusCode = respCode
	if respCode == http.StatusOK {
		httpRes.Status = true
	}

	//check whether request is timed or not
	if respCode == http.StatusGatewayTimeout {
		err = perror.CustomError("Gateway Timeout")
	} else {
		var bodyBytes []byte
		if bodyBytes, err = ioutil.ReadAll(resp.Body); err == nil {
			httpRes.Body = fmt.Sprintf("%s", string(bodyBytes[:]))
		} else {
			err = perror.MiscError(err)
		}
	}
	return
}

//RequestHTTP : Make HTTP Call
func (httpReq HTTPReq) RequestHTTP() (httpRes HTTPRes, err error) {
	var (
		req        *http.Request
		resp       *http.Response
		reqBody    io.Reader
		reqBodyStr []byte
	)
	if httpReq.Body != nil {
		if reqBodyStr, err = json.Marshal(httpReq.Body); err != nil {
			err = perror.MarshalError(err)
		} else {
			reqBody = bytes.NewBuffer(reqBodyStr)
		}
	} else if httpReq.Method == "POST" || httpReq.Method == "PUT" {
		reqBody = httpReq.FormValue
	}

	if err == nil {
		if req, err = http.NewRequest(httpReq.Method, httpReq.URL, reqBody); err == nil {
			setHeader(req, httpReq.Header)
			client := &http.Client{}
			if httpReq.Log != nil {
				httpReq.Log.Printf(&httpReq, nil)
			}
			if resp, err = client.Do(req); err == nil {
				defer resp.Body.Close()
				httpRes, err = readResonse(resp)
				if httpReq.Log != nil {
					httpReq.Log.Printf(&httpReq, &httpRes)
				}
			} else {
				err = perror.MiscError(err)
			}
		} else {
			err = perror.MiscError(err)
		}
	}
	return
}
