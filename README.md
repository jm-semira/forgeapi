# ForgeAPI

![Go Tests](https://github.com/jm-semira/forgeapi/actions/workflows/go-test.yml/badge.svg)

**Forge your API. Focus on your logic.**

ForgeAPI is a Go library that generates RESTful CRUD endpoints for registered GORM models. It is intended for Go developers who want to reduce repetitive CRUD handler code while keeping their application-specific business logic in their own application.

> **Early-stage / under development:** ForgeAPI is not production-ready. Its generated endpoints currently have no built-in authentication, authorization, request validation, or writable-field restrictions. Read [Security and limitations](#security-and-limitations) before using it.

## Current features

- Register a pointer to a struct model when a GORM database connection is configured.
- Generate list, get-by-ID, create, update, and delete routes.
- Filter, sort, and paginate collection requests using registered model fields.
- Use Go's `net/http` `ServeMux` and expose it as an `http.Handler`.
- Start a server with the provided `Run` method or mount the handler in an application-managed server.
- Use the request context for GORM operations.

Resource paths are derived from the lowercased Go type name with simple English suffix rules (for example, `Product` becomes `/products`). This is not a general-purpose pluralization system.

## Installation

ForgeAPI's module path is `github.com/jm-semira/forgeapi`:

```sh
go get github.com/jm-semira/forgeapi
```

The module currently depends on GORM and includes the SQLite driver. Applications need to import a GORM driver for the database they use.

## Quick start

This complete example uses SQLite. The imports shown are the imports needed by the example application; `gorm.io/driver/sqlite` and `gorm.io/gorm` are direct module dependencies of ForgeAPI.

```go
package main

import (
	"log"

	"github.com/jm-semira/forgeapi"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	forgeapi.Model
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func main() {
	db, err := gorm.Open(sqlite.Open("forgeapi.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	if err := db.AutoMigrate(&Product{}); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		log.Fatalf("register product resource: %v", err)
	}

	log.Fatal(app.Run(":8080"))
}
```

The example uses the public `New`, `WithDatabase`, `Register`, and `Run` APIs. `AutoMigrate` is an application setup step; ForgeAPI does not run migrations automatically. You can instead pass `app.Handler()` to a server or router that accepts an `http.Handler`.

## Generated endpoints

For the `Product` model above, ForgeAPI registers these routes under `/products`:

| Method   | Route            | Purpose                                                           | Responses and notable errors                                                                                                                                                        |
| -------- | ---------------- | ----------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GET`    | `/products`      | Return all records as a JSON array.                               | `200 OK`; `500 Internal Server Error` for a database error.                                                                                                                         |
| `GET`    | `/products/{id}` | Return the first matching record as JSON.                         | `200 OK`; `404 Not Found` when no record exists; `500 Internal Server Error` for another database error.                                                                            |
| `POST`   | `/products`      | Decode a JSON model and create it.                                | `201 Created` with the created record; `400 Bad Request` for invalid JSON; `500 Internal Server Error` for a database error.                                                        |
| `PATCH`  | `/products/{id}` | Apply JSON object fields as GORM updates, then return the record. | `200 OK`; `400 Bad Request` for a non-numeric ID, invalid JSON, or an empty object; `404 Not Found` if the record does not exist; `500 Internal Server Error` for a database error. |
| `DELETE` | `/products/{id}` | Delete the record matching the ID.                                | `204 No Content`; `400 Bad Request` for a non-numeric ID; `404 Not Found` if no row was affected; `500 Internal Server Error` for a database error.                                 |

The ID parsing and update/delete queries assume an `id` primary key compatible with GORM's usual conventions. `DELETE` follows GORM model behavior: models with GORM soft-delete fields are soft-deleted. Unknown methods or paths are handled by `http.ServeMux`.

## Collection queries

Collection query parameters apply to the list route, such as `GET /products`. Queryable fields use the exact JSON name from a model's `json` tag; when a field has no JSON name, ForgeAPI uses its Go field name. Fields excluded from writable model metadata, including primary keys and fields tagged `forgeapi:"readonly"` or `json:"-"`, are not available for filtering or sorting. For the `Product` model above, use lowercase `name` and `price`.

### Filtering

Filters use `filter[field][operator]=value`. The supported operators are `eq` for equality, `gte` for greater than or equal, and `lte` for less than or equal:

```text
GET /products?filter[price][gte]=100
GET /products?filter[name][eq]=Keyboard
```

Add multiple filter parameters to combine conditions; ForgeAPI applies them together (AND):

```text
GET /products?filter[price][gte]=50&filter[price][lte]=100
```

An unknown field or unsupported operator returns `400 Bad Request`. A filter parameter with more than one value also returns `400 Bad Request`.

### Sorting

Use `sort` with a comma-separated list of field names. A field sorts ascending by default; prefix it with `-` for descending. Fields are applied in the order listed:

```text
GET /products?sort=price,-name
```

An unknown sort field returns `400 Bad Request`. Sort fields use the same exact JSON-name or Go-field-name resolution as filters.

### Pagination

Collection routes use a default limit of `20` records and a maximum limit of `100`. Configure these values when creating the API:

```go
app := forgeapi.New(
	forgeapi.WithDatabase(db),
	forgeapi.WithDefaultLimit(25),
	forgeapi.WithMaxLimit(200),
)
```

An omitted `limit` uses the configured default. An explicit `limit` must be a positive integer and cannot exceed the configured maximum. An omitted `offset` starts at the first row; an explicit `offset` must be a non-negative integer that fits the platform's `int` type. Invalid limits and offsets return `400 Bad Request`. `forgeapi.New` panics if either configured limit is non-positive or if the default limit exceeds the maximum.

Filtering, sorting, and pagination can be combined in one collection request:

```text
GET /products?filter[price][gte]=100&sort=-price&limit=10&offset=20
```

This filters to prices of at least 100, sorts by price descending, and returns up to 10 records after skipping the first 20 matching records.

## Database support

ForgeAPI uses GORM for database operations. Compatibility depends on the GORM driver and on the model and queries working with the selected database. The repository's integration tests use GORM's SQLite driver with in-memory SQLite databases. No other database driver or database has been verified by the current test suite; support for other GORM-compatible databases should not be assumed without testing in your application.

## Security and limitations

The generated routes currently do not implement authentication, authorization, or application-specific validation rules. `POST` and `PATCH` reject malformed JSON, invalid field types, unknown fields, and fields excluded from the writable-field metadata, but they do not enforce business rules such as required values or application-specific constraints. Registering a model does **not** make every field in that model safe to expose or modify through a public API.

Do not expose these generated endpoints to untrusted clients. Do not treat ForgeAPI as production-ready until the security controls your application requires have been implemented and tested. Add appropriate controls in your application and carefully choose the models and fields exposed to these routes.

Additional limitations include simple resource-name pluralization and the default behavior and error responses of `http.ServeMux` for unmatched requests.

## Testing

Run the repository tests with:

```sh
go test ./...
```

The existing test suite uses SQLite integration tests to cover collection listing, filtering, sorting, pagination and pagination configuration, fetching a record and a missing record, creating a record, patching a record and several patch error cases, and deleting an existing or missing record (including invalid delete IDs). It does not establish behavior across other database engines or cover every malformed request and failure mode.

## Roadmap

- [x] Initial generated CRUD endpoints
- [x] Basic handling of invalid IDs and missing records
- [x] Basic request validation and writable-field restrictions
- [ ] Application-specific validation rules
- [ ] Authentication and authorization
- [x] Pagination, filtering, and sorting
- [ ] Custom CRUD logic and hooks
- [ ] Custom endpoint registration
- [ ] Expanded database integration tests
- [ ] API documentation and examples

## License

No license file is present in this repository, so no license has been selected or stated here.
