package forgeapi

import (
	"errors"
	"net/http"

	"gorm.io/gorm"
)

var (
	ErrDatabaseRequired = errors.New("forgeapi: database is required")
	ErrInvalidModel     = errors.New("forgeapi: model must be a non-nil pointer to a struct")
)

type Option func(*API)

type API struct {
	db  *gorm.DB
	mux *http.ServeMux
}

func New(opts ...Option) *API {
	app := &API{
		mux: http.NewServeMux(),
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

func WithDatabase(db *gorm.DB) Option {
	return func(app *API) {
		app.db = db
	}
}

func (app *API) Handler() http.Handler {
	return app.mux
}

func (app *API) Run(addr string) error {
	return http.ListenAndServe(addr, app.Handler())
}
