package perror

import (
	"database/sql"
	"strings"

	"github.com/go-pg/pg"
	"github.com/jinzhu/gorm"
	"gitlab.com/g-harshit/plib/constant"
)

//VError : validation error
func VError(debugMsg ...string) *Error {
	return newError(constant.InvalidReqMsg, nil, constant.ValidateError, debugMsg...)
}

//ScoutRunningError : scout running error
func ScoutRunningError(debugMsg ...string) *Error {
	return newError(constant.ScoutRunningMsg, nil, constant.ScoutRunningError, debugMsg...)
}

//NoUpdateError : error occured when blank input is given to update
func NoUpdateError(debugMsg ...string) *Error {
	return newError(constant.BlandReqMsg, nil, constant.ValidateError, debugMsg...)
}

//CustomError : custom error
func CustomError(debugMsg ...string) *Error {
	msg := strings.Join(debugMsg, " ")
	return newError(msg, nil, constant.CustomError, debugMsg...)
}

//UnmarshalError : error occured while unmarshal
func UnmarshalError(err error, debugMsg ...string) *Error {
	return newError(constant.InvalidReqMsg, err, constant.ValidateError, debugMsg...)
}

//MarshalError : error occured while unmarshal
func MarshalError(err error, debugMsg ...string) *Error {
	return newError(constant.InvalidReqMsg, err, constant.ValidateError, debugMsg...)
}

//MiscError : error occured while processing
func MiscError(err error, debugMsg ...string) *Error {
	return newError(constant.ServerErrorMsg, err, constant.MiscError, debugMsg...)
}

//ConnError : error occured while creating connection
func ConnError(err error, debugMsg ...string) *Error {
	return newError(constant.ServerErrorMsg, err, constant.DBError, debugMsg...)
}

//SelectError : error occured while select query
func SelectError(err error, debugMsg ...string) *Error {
	if err == pg.ErrNoRows {
		return NotFoundError(debugMsg...)
	}
	return newError(constant.ServerErrorMsg, err, constant.SelectQueryError, debugMsg...)
}

//SelectIgnoreNoRow : error occured while select query by ignoring no row error
func SelectIgnoreNoRow(err error, debugMsg ...string) error {
	if err != pg.ErrNoRows && err != sql.ErrNoRows && err != gorm.ErrRecordNotFound {
		return newError(constant.ServerErrorMsg, err, constant.SelectQueryError, debugMsg...)
	}
	return nil
}

//SearchError : error occured while elastic search query
func SearchError(err error, debugMsg ...string) *Error {
	return newError(constant.ServerErrorMsg, err, constant.SearchQueryError, debugMsg...)
}

//InsertError : error occured while insert query
func InsertError(err error, debugMsg ...string) *Error {
	return newError(constant.ServerErrorMsg, err, constant.InsertQueryError, debugMsg...)
}

//UpdateError : error occured while update query
func UpdateError(err error, debugMsg ...string) *Error {
	return newError(constant.ServerErrorMsg, err, constant.UpdateQueryError, debugMsg...)
}

//DeleteError : error occured while delete query
func DeleteError(err error, debugMsg ...string) *Error {
	return newError(constant.ServerErrorMsg, err, constant.DeleteQueryError, debugMsg...)
}

//TxError : error occured while starting transaction
func TxError(err error, debugMsg ...string) *Error {
	return newError(constant.ServerErrorMsg, err, constant.TransactionError, debugMsg...)
}

//NotFoundError : error occured when id not found
func NotFoundError(debugMsg ...string) *Error {
	if len(debugMsg) == 0 {
		debugMsg = append(debugMsg, "Record not exists")
	}
	return newError(constant.NotFoundMsg, nil, constant.NotFoundError, debugMsg...)
}

//BadReqError : error occured while validating request
//like while typecasting request, fk in request dosn't exists
func BadReqError(err error, debugMsg ...string) *Error {
	return newError(constant.InvalidReqMsg, err, constant.ValidateError, debugMsg...)
}

//ForbiddenErr : unauthorized access
func ForbiddenErr(debugMsg ...string) *Error {
	return newError(constant.ForbiddenMsg, nil, constant.ForbiddenError, debugMsg...)
}

//InvalidParamError : Error due to request validation fail
func InvalidParamError(err error) *Error {
	val := strings.Split(err.Error(), "\n")
	msg := ""
	for _, v := range val {
		eVal := strings.Split(v, "Error:Field validation for '")
		if len(eVal) > 1 {
			field := strings.Split(eVal[1], "'")[0]
			field = strings.Replace(field, "CID", "Company", -1)
			field = strings.Replace(field, "ID", "", -1)
			v = "Invalid " + snakeName(field)
			msg += v + ", "
		}
	}
	msg = strings.Trim(msg, ", ")
	return VError(err.Error()).SetMsg(msg)
}

//MappingError : error occured deleting entity without removing it's mapping
func MappingError(debugMsg ...string) *Error {
	return newError(constant.RemoveMappingMsg, nil, constant.ValidateError, debugMsg...)
}

//HttpError : error occured while making http request
func HTTPError(err error, debugMsg ...string) *Error {
	return newError(constant.RequestError, err, constant.HttpRequestFailed, debugMsg...)
}

//ExecError : error occured while exec query
func ExecError(err error, debugMsg ...string) *Error {
	return newError(constant.ServerErrorMsg, err, constant.HttpRequestFailed, debugMsg...)
}

//UnauthorizedErr : unauthorized access
func UnauthorizedErr(debugMsg ...string) *Error {
	return newError(constant.ForbiddenMsg, nil, constant.UnauthorizedError, debugMsg...)
}
