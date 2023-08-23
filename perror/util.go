package perror

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/TeamTutx/plib/conf"
	"github.com/TeamTutx/plib/constant"
)

var debug bool

func init() {
	debug = true
}

//SetDebug will set debug value for error
func SetDebug(flag bool) {
	debug = flag
}

//newError : Create new *Error object
func newError(msg string, err error, code int, debugMsg ...string) *Error {
	tracePath := ""
	//If ty debug is enable then only stackTrace i.e. header is sending DEBUG-GLACIER:true
	if debug {
		stackDepth := conf.Int("error.stack_depth", 3)
		funcName, fileName, line := StackTrace(stackDepth)
		tracePath = fileName + " -> " + funcName + ":" + strconv.Itoa(line)
	}

	dMsg := strings.Join(debugMsg, " ")
	if err != nil {
		msg, code = checkMySQLError(msg, code, err)
		dMsg += " " + err.Error()
	}
	return &Error{
		Msg:      msg,
		DebugMsg: dMsg,
		Trace:    tracePath,
		Code:     code,
		Info:     make(map[string]interface{}),
	}
}

//checkMySQLError : check mysql error codes
func checkMySQLError(msg string, code int, err error) (string, int) {
	errMsg := err.Error()
	if strings.Contains(errMsg, "#23505 ") == true {
		msgCode := strings.Split(errMsg, " ")
		msgCodeLen := len(msgCode)
		if msgCodeLen > 1 {
			msg = constant.DuplicateEntryMsg
			code = constant.DuplicateError
		}
	}
	return msg, code
}

//Error : Implement Error method of error interface
func (e *Error) Error() string {
	return fmt.Sprintf("\nCode:\t\t[%d]\nMessage:\t[%v]\nStackTrace:\t[%v]\nDebugMsg:\t[%v]\n", e.Code, e.Msg, e.Trace, e.DebugMsg)
}

//SetMsg will overwrite msg in error
func (e *Error) SetMsg(msg string) *Error {
	if msg != "" {
		e.Msg = msg
	}
	return e
}

//IfCodeSetMsg will set msg if error code matches
func (e *Error) IfCodeSetMsg(errCode int, msg string) *Error {
	if e.Code == errCode {
		e.SetMsg(msg)
	}
	return e
}

//IfDuplicate will set msg if error code is duplicate error
func (e *Error) IfDuplicate(msg string) *Error {
	if e.Code == constant.DuplicateError {
		e.SetMsg(msg)
	}
	return e
}

//StackTrace : Get function name, file name and line no of the caller function
//Depth is the value from which it will start searching in the stack
func StackTrace(depth int) (funcName string, file string, line int) {
	var (
		ok bool
		pc uintptr
	)
	for i := depth; ; i++ {
		if pc, file, line, ok = runtime.Caller(i); ok {
			if strings.Contains(file, constant.PackageName) {
				continue
			}
			fileName := strings.Split(file, "github.com")
			if len(fileName) > 1 {
				file = fileName[1]
			}
			_, funcName = packageFuncName(pc)
			break
		} else {
			break
		}
	}
	return
}

//packageFuncName : Package and function name from package counter
func packageFuncName(pc uintptr) (packageName string, funcName string) {
	if f := runtime.FuncForPC(pc); f != nil {
		funcName = f.Name()
		if ind := strings.LastIndex(funcName, "/"); ind > 0 {
			packageName += funcName[:ind+1]
			funcName = funcName[ind+1:]
		}
		if ind := strings.Index(funcName, "."); ind > 0 {
			packageName += funcName[:ind]
			funcName = funcName[ind+1:]
		}
	}
	return
}

//GetErrorCode : Get Error Code from *Error
func GetErrorCode(err error) (code int) {
	if err == nil {
		code = constant.NoError
	} else {
		switch err.(type) {
		case *Error:
			code = err.(*Error).Code
		default:
			code = constant.NoError
		}
	}
	return
}

//AppendDebug will append debug msg
func AppendDebug(err error, dMsg string) error {
	switch err.(type) {
	case *Error:
		nErr := err.(*Error)
		nErr.DebugMsg += " " + dMsg
		err = nErr
	}
	return err
}

//GetMsg will return error msg
func GetMsg(err error) (msg string) {
	switch err.(type) {
	case *Error:
		msg = err.(*Error).Msg
	default:
		msg = err.Error()
	}
	return
}

//GetInfo will return error info
func GetInfo(err error) (info map[string]interface{}) {
	switch err.(type) {
	case *Error:
		info = err.(*Error).Info
	default:
	}
	return
}

//GetDebug will return debug msg and trace
func GetDebug(err error) (dMsg, trace string) {
	switch err.(type) {
	case *Error:
		e := err.(*Error)
		dMsg, trace = e.DebugMsg, e.Trace
	default:
		dMsg = err.Error()
	}
	return
}

func snakeName(name string) (capName string) {
	capName = ""
	for i := range name {
		v := name[i]
		if name[i] <= 'Z' {
			if i != 0 {
				capName += " "
			}
		}
		capName += string(v)
	}
	return
}
