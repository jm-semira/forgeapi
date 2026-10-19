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
	tests := []struct {
		name      string
		body      string
		wantName  string
		wantPrice float64
	}{
		{
			name:      "all supplied fields",
			body:      `{"Name":"Mouse","Price":99.9}`,
			wantName:  "Mouse",
			wantPrice: 99.9,
		},
		{
			name:     "omitted optional field",
			body:     `{"Name":"Mouse"}`,
			wantName: "Mouse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			if err := db.AutoMigrate(&Product{}); err != nil {
				t.Fatal(err)
			}
			app := forgeapi.New(forgeapi.WithDatabase(db))
			if err := app.Register(&Product{}); err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			app.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusCreated {
				t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
			}

			var got Product
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if got.ID == 0 {
				t.Error("expected created product to have a non-zero ID")
			}
			if got.Name != tt.wantName {
				t.Errorf("expected name %q, got %q", tt.wantName, got.Name)
			}
			if got.Price != tt.wantPrice {
				t.Errorf("expected price %.2f, got %.2f", tt.wantPrice, got.Price)
			}
		})
	}
}

func TestPostValidationRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name  string
		model any
		path  string
		body  string
	}{
		{
			name:  "trailing JSON",
			model: &Product{},
			path:  "/products",
			body:  `{"Price":399} {"Name":"Unexpected"}`,
		},
		{
			name:  "unknown field",
			model: &Product{},
			path:  "/products",
			body:  `{"UnknownField":"value"}`,
		},
		{
			name:  "read-only canonical field",
			model: &Product{},
			path:  "/products",
			body:  `{"created_at":"2026-01-01T00:00:00Z"}`,
		},
		{
			name:  "invalid field type",
			model: &Product{},
			path:  "/products",
			body:  `{"Price":"not-a-number"}`,
		},
		{
			name:  "empty object",
			model: &Product{},
			path:  "/products",
			body:  `{}`,
		},
		{
			name:  "malformed JSON",
			model: &Product{},
			path:  "/products",
			body:  `{`,
		},
		{
			name:  "noncanonical JSON field name",
			model: &CustomProduct{},
			path:  "/customproducts",
			body:  `{"Name":"Keyboard"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			if err := db.AutoMigrate(tt.model); err != nil {
				t.Fatal(err)
			}
			app := forgeapi.New(forgeapi.WithDatabase(db))
			if err := app.Register(tt.model); err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			assertPostRejectedWithoutDBChanges(t, db, tt.model, app, req)
		})
	}
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
