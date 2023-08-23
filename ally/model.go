package ally

import (
	"net/http"

	"github.com/TeamTutx/plib/perror"
)

//Resp : Response Model
type Resp struct {
	Data   interface{} `json:"data,omitempty"`
	Msg    interface{} `json:"message,omitempty"`
	Error  error       `json:"error,omitempty"`
	Status bool        `json:"status"`
}

//MsgResp : Message response for create/update/delete api
type MsgResp struct {
	Data interface{}
	Msg  string
}

//TestResp : Response Model for test files
type TestResp struct {
	Data   interface{}  `json:"data"`
	Msg    interface{}  `json:"message"`
	Error  perror.Error `json:"error"`
	Status bool         `json:"status"`
}

//statusWriter implemented http ResponseWriter
type statusWriter struct {
	http.ResponseWriter
	status int
}

//WriteHeader implementing http ResponseWriter method
func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

//FilterDet : select query filter alias name of field, operator to use and value
type FilterDet struct {
	Display         string
	Key             string
	Alias           string
	Operator        string
	Value           string
	ValidTag        string
	DefaultOperator string
	DefaultValue    string
	Sort            bool
	Skip            bool
	Or              bool
	Unescape        bool
	ESFilterType    string
	Exclude         []string
}

//AdvFilter : advance select filter
type AdvFilter struct {
	Alias          string
	Filter         map[string]FilterDet
	Offset         int
	Limit          int
	IncludeFields  string
	ExcludeFields  string
	SplitSearchStr string
}
