package errno

var (
	// Common errors
	OK                  = &Errno{Code: 200, Message: "Success"}
	InternalServerError = &Errno{Code: 10001, Message: "Internal server error."}
	BindError           = &Errno{Code: 10002, Message: "Error request body."}
	MissParamError      = &Errno{Code: 10003, Message: "Missing param: "}

	DbError = &Errno{Code: 30100, Message: "The database error with: "}
)
