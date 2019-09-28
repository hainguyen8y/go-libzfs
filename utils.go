package zfs

type Error struct {
	errno 	int
	message	string
}

func (self *Error) Error() string {
	return self.message
}

func (self *Error) Errno() int {
	return self.errno
}

func NewError(errcode int, msg string) error {
	return &Error{
		errno: errcode,
		message: msg,
	}
}
