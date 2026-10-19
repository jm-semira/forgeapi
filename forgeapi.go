package forgeapi

import (
	"net/http"

	"gorm.io/gorm"
)

type Option func(*API)

const (
	defaultPaginationLimit = 20
	defaultMaxLimit        = 100
)

type API struct {
	db           *gorm.DB
	mux          *http.ServeMux
	defaultLimit int
	maxLimit     int
}

func New(opts ...Option) *API {
	app := &API{
		mux:          http.NewServeMux(),
		defaultLimit: defaultPaginationLimit,
		maxLimit:     defaultMaxLimit,
	}

	for _, opt := range opts {
		opt(app)
	}

	if app.defaultLimit <= 0 {
		panic("forgeapi: default pagination limit must be positive")
	}

	if app.maxLimit <= 0 {
		panic("forgeapi: maximum pagination limit must be positive")
	}

	if app.defaultLimit > app.maxLimit {
		panic("forgeapi: default pagination limit cannot exceed maximum pagination limit")
	}

	return app
}

func WithDatabase(db *gorm.DB) Option {
	return func(app *API) {
		app.db = db
	}
}
func WithDefaultLimit(limit int) Option {
	return func(app *API) {
		app.defaultLimit = limit
	}
}

func WithMaxLimit(limit int) Option {
	return func(app *API) {
		app.maxLimit = limit
	}
}

func (app *API) Handler() http.Handler {
	return app.mux
}

func (app *API) Run(addr string) error {
	return http.ListenAndServe(addr, app.Handler())
}
