package forgeapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jm-semira/forgeapi"
	"gorm.io/gorm"
)

func assertPostRejectedWithoutDBChanges(
	t *testing.T,
	db *gorm.DB,
	model any,
	app *forgeapi.API,
	req *http.Request,
) {
	t.Helper()

	var before int64
	if err := db.Model(model).Count(&before).Error; err != nil {
		t.Fatalf("failed to count records before request: %v", err)
	}

	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			rec.Code,
			rec.Body.String(),
		)
	}

	var after int64
	if err := db.Model(model).Count(&after).Error; err != nil {
		t.Fatalf("failed to count records after request: %v", err)
	}
	if after != before {
		t.Errorf("expected record count to remain %d, got %d", before, after)
	}
}

func TestPostProduct(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/products",
		strings.NewReader(`{
			"Name":"Mouse",
			"Price":99.9
		}`),
	)

	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			rec.Code,
			rec.Body.String(),
		)
	}

	var got Product
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got.ID == 0 {
		t.Error("expected created product to have a non-zero ID")
	}

	if got.Name != "Mouse" {
		t.Errorf("expected name %q, got %q", "Mouse", got.Name)
	}

	if got.Price != 99.9 {
		t.Errorf("expected price %.2f, got %.2f", 99.9, got.Price)
	}
}

func TestPostProductTrailingJSON(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/products",
		strings.NewReader(`{"Price":399} {"Name":"Unexpected"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	assertPostRejectedWithoutDBChanges(t, db, &Product{}, app, req)
}

func TestPostProductUnknownField(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/products",
		strings.NewReader(`{"UnknownField":"value"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	assertPostRejectedWithoutDBChanges(t, db, &Product{}, app, req)
}

func TestPostProductReadOnlyField(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/products",
		strings.NewReader(`{"created_at":"2026-01-01T00:00:00Z"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	assertPostRejectedWithoutDBChanges(t, db, &Product{}, app, req)
}

func TestPostProductInvalidFieldType(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/products",
		strings.NewReader(`{"Price":"not-a-number"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	assertPostRejectedWithoutDBChanges(t, db, &Product{}, app, req)
}

func TestPostProductEmptyBody(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/products",
		strings.NewReader(`{}`),
	)

	req.Header.Set("Content-Type", "application/json")
	assertPostRejectedWithoutDBChanges(t, db, &Product{}, app, req)
}

func TestPostCustomProductCanonicalJSONField(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&CustomProduct{}); err != nil {
		t.Fatalf("failed to migrate CustomProduct: %v", err)
	}
	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&CustomProduct{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/customproducts",
		strings.NewReader(`{"name":"Keyboard"}`),
	)
	rec := httptest.NewRecorder()

	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s",
			http.StatusCreated, rec.Code, rec.Body.String())
	}
}

func TestPostCustomProductRejectsNonCanonicalJSONField(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&CustomProduct{}); err != nil {
		t.Fatal(err)
	}
	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&CustomProduct{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/customproducts",
		strings.NewReader(`{"Name":"Keyboard"}`),
	)
	assertPostRejectedWithoutDBChanges(t, db, &CustomProduct{}, app, req)
}
