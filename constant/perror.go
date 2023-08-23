package constant

import (
	"net/http"
)

//Error Messages
const (
	RemoveMappingMsg  = "Please remove it's all mapping before delete"
	InvalidReqMsg     = "Invalid Request, Please contact system administrator for further clarification."
	ForbiddenMsg      = "You are not allowed to perform this operation. Please contact system administrator."
	BlandReqMsg       = "Blank request, Please provide input to process"
	NotFoundMsg       = "Record not found"
	ServerErrorMsg    = "Sorry unable to process this request. Please Try Again"
	DuplicateEntryMsg = "Record Already Exists"
	ScoutRunningMsg   = "Scout Already Running"
	RequestError      = "HTTP request could not intialized"
)

// Error Code. PLEASE Map New Error Code To The HTTP Code Map Below.
const (
	NoError           = 0
	ValidateError     = 101  //Primary Validation fail
	DBError           = 103  //Database Connection error
	SelectQueryError  = 104  //Select Query error
	CreateQueryError  = 105  //Create Query error
	InsertQueryError  = 106  //Insert Query error
	UpdateQueryError  = 107  //Update Query error
	DeleteQueryError  = 108  //Delete Query error
	CustomError       = 109  //Custom error
	TransactionError  = 110  //Transaction error
	NotFoundError     = 111  //Request ID not found error
	ForbiddenError    = 112  //Unauthorized access
	DuplicateError    = 113  //Duplicate entry error
	MiscError         = 114  //Misselaineous error
	ScoutRunningError = 115  //Elasticsearch scout already running error
	SearchQueryError  = 116  //Search Query error
	UnauthorizedError = 117  //Unauthorized access
	HttpRequestFailed = 9001 //Http request failed error

)

var (
	//ErrorHTTPCode : Error Code to Http Code map
	ErrorHTTPCode = map[int]int{
		NoError:           http.StatusOK,
		ValidateError:     http.StatusBadRequest,
		DBError:           http.StatusInternalServerError,
		SelectQueryError:  http.StatusBadRequest,
		CreateQueryError:  http.StatusBadRequest,
		InsertQueryError:  http.StatusBadRequest,
		UpdateQueryError:  http.StatusBadRequest,
		DeleteQueryError:  http.StatusBadRequest,
		CustomError:       http.StatusBadRequest,
		TransactionError:  http.StatusInternalServerError,
		NotFoundError:     http.StatusOK,
		ForbiddenError:    http.StatusForbidden,
		DuplicateError:    http.StatusConflict,
		MiscError:         http.StatusInternalServerError,
		ScoutRunningError: http.StatusServiceUnavailable,
		SearchQueryError:  http.StatusBadRequest,
		HttpRequestFailed: http.StatusInternalServerError,
		UnauthorizedError: http.StatusUnauthorized,
	}
)
