package forgeapi

import "errors"

var (
	ErrDatabaseRequired          = errors.New("forgeapi: database is required")
	ErrInvalidModel              = errors.New("forgeapi: model must be a non-nil pointer to a struct")
	ErrInvalidIdentifier         = errors.New("invalid identifier")
	ErrUnsupportedIdentifierType = errors.New("unsupported identifier type")
	ErrInvalidModelMetadata      = errors.New("forgeapi: failed to resolve model metadata")
)
