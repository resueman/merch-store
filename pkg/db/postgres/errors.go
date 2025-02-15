package postgres

import "errors"

var (
	ErrTxRetriesExceeded = errors.New("error executing transaction: retries exceeded")
)
