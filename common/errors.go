package common

import "fmt"

type CollectorError struct {
	Prob string
}

type WriterError struct {
	Params string
	Prob   string
}

func (e *CollectorError) Error() string {
	return fmt.Sprintf("%s", e.Prob)
}

func (e *WriterError) Error() string {
	return fmt.Sprintf("%s - %s", e.Params, e.Prob)
}
