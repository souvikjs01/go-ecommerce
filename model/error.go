package model

type ErrMsg struct {
	Err  error
	Code int
}

func (e ErrMsg) Error() string {
	return e.Err.Error()
}
