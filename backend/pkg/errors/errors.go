package errors


import "errors"

var (
	ErrNotFound       = errors.New("record not found")
	ErrDuplicate      = errors.New("record already exists")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInvalidRequest = errors.New("invalid request")
	ErrInternal       = errors.New("internal server error")
	ErrNoTenantCtx    = errors.New("tenant not found in context")
)
