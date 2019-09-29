package zfs

type Error struct {
	errorCode 	int
	message	string
}

func (self *Error) Error() string {
	return self.message
}

func (self *Error) GetErrorCode() int {
	return self.errorCode
}

func NewError(errcode int, msg string) error {
	return &Error{
		errorCode: errcode,
		message: msg,
	}
}
