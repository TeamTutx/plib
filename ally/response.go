package ally

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/g-harshit/plib/conf"
	"gitlab.com/g-harshit/plib/constant"
	"gitlab.com/g-harshit/plib/perror"
)

//NewResp : Create New Response Objext
func NewResp() *Resp {
	return &Resp{}
}

//BuildResponse : creates the response of API
func BuildResponse(data interface{}, err error) (int, interface{}) {
	out := NewResp()
	httpCode, status := GetErrorHTTPCode(err)
	out.Set(data, status, err)
	return httpCode, out
}

//GinResponse : creates the response of gin API
func GinResponse(c *gin.Context, data interface{}, err error) {
	httpCode, out := BuildResponse(data, err)
	c.JSON(httpCode, out)
}

//Set Response Values
func (r *Resp) Set(data interface{}, status bool, err error) {
	var (
		respData interface{}
		msg      string
	)
	if val, ok := data.(MsgResp); ok {
		respData = val.Data
		msg = val.Msg
	} else {
		respData = data
	}
	if err == nil {
		r.Data = respData
		r.Msg = msg
	} else {
		r.Error = err
	}
	r.Status = status
	if debugMsg := conf.Bool("error.debug_msg", false); debugMsg == false {
		r.Error = unsetDebugMsg(r.Error)
	}
}

//unsetDebugMsg : Unset debugMsg of *Error
func unsetDebugMsg(err error) error {
	switch err.(type) {
	case *perror.Error:
		err.(*perror.Error).DebugMsg = ""
		err.(*perror.Error).Trace = ""
	}
	return err
}

//GetErrorHTTPCode : Get HTTP code from error code. Please check for httpCode = 0.
func GetErrorHTTPCode(err error) (httpCode int, status bool) {
	errCode := perror.GetErrorCode(err)
	httpCode = constant.ErrorHTTPCode[errCode]
	if err == nil {
		status = true
	}
	return
}

//BuildResponseData : creates the response of API
func BuildResponseData(data interface{}, err error, dataCheck bool) (int, interface{}) {
	out := NewResp()
	httpCode, status := GetErrorHTTPCode(err)
	out.SetData(data, status, err, dataCheck)
	return httpCode, out
}

//GinResponseData : creates the response of gin API and resturns data even if there is err
func GinResponseData(c *gin.Context, data interface{}, err error, dataCheck bool) {
	httpCode, out := BuildResponseData(data, err, dataCheck)
	c.JSON(httpCode, out)
}

//Set Response Values
func (r *Resp) SetData(data interface{}, status bool, err error, dataCheck bool) {
	var (
		respData interface{}
		msg      string
	)
	if val, ok := data.(MsgResp); ok {
		respData = val.Data
		msg = val.Msg
	} else {
		respData = data
	}
	if err == nil {
		r.Data = respData
		r.Msg = msg
	} else {
		r.Error = err
		if dataCheck {
			r.Data = respData
		}
	}
	r.Status = status
	if debugMsg := conf.Bool("error.debug_msg", false); debugMsg == false {
		r.Error = unsetDebugMsg(r.Error)
	}
}
