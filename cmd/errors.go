package cmd

type UserError struct {
	Message string
	Err     error
}

func (u *UserError) Error() string {
	return u.Err.Error()
}

func (u *UserError) UserError() string {
	return u.Message
}

func (u *UserError) Unwrap() error {
	return u.Err
}

func NewUserError(err error, message string) *UserError {
	return &UserError{
		Message: message,
		Err:     err,
	}
}
