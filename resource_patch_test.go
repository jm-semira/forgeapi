package forgeapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jm-semira/forgeapi"
)

func TestPatchProduct(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	product := Product{
		Name:  "Mouse",
		Price: 99.9,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/products/%d", product.ID),
		strings.NewReader(`{
			"Price":399
		}`),
	)

	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}

	var got Product
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got.ID != product.ID {
		t.Errorf("expected udpted product to have %d, got %d ID", product.ID, got.ID)
	}

	if got.Name != "Mouse" {
		t.Errorf("expected name %q, got %q", "Mouse", got.Name)
	}

	if got.Price != 399 {
		t.Errorf("expected price %.2f, got %.2f", 399.0, got.Price)
	}
}

func TestPatchProductTrailingJSON(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	product := Product{Name: "Mouse", Price: 99.9}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/products/%d", product.ID),
		strings.NewReader(`{"Price":399} {"Name":"Unexpected"}`),
	)
	req.Header.Set("Content-Type", "application/json")

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
}

func TestPatchProductUnknownField(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	product := Product{Name: "Mouse", Price: 99.9}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/products/%d", product.ID),
		strings.NewReader(`{"UnknownField":"value"}`),
	)
	req.Header.Set("Content-Type", "application/json")

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
}

func TestPatchProductReadOnlyField(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	product := Product{Name: "Mouse", Price: 99.9}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/products/%d", product.ID),
		strings.NewReader(`{"CreatedAt":"2026-01-01T00:00:00Z"}`),
	)
	req.Header.Set("Content-Type", "application/json")

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
}

func TestPatchProductInvalidFieldType(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	product := Product{Name: "Mouse", Price: 99.9}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/products/%d", product.ID),
		strings.NewReader(`{"Price":"not-a-number"}`),
	)
	req.Header.Set("Content-Type", "application/json")

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
}

func TestPatchProductNotFound(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/products/%d", 999),
		strings.NewReader(`{
			"Price":399
		}`),
	)

	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestPatchProductInvalidID(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPatch,
		"/products/abc",
		strings.NewReader(`{
			"Price":399
		}`),
	)

	req.Header.Set("Content-Type", "application/json")
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
}

func TestPatchProductEmptyBody(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPatch,
		"/products/1",
		strings.NewReader(`{}`),
	)

	req.Header.Set("Content-Type", "application/json")
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
}
func TestPatchProductByCustomID(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&CustomProduct{}); err != nil {
		t.Fatal(err)
	}

	product := CustomProduct{
		ProductID: "sku-123",
		Name:      "Mouse",
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&CustomProduct{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/customproducts/%s", product.ProductID),
		strings.NewReader(`{
			"name":"Keyboard"
		}`),
	)

	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}

	var got CustomProduct
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got.ProductID != product.ProductID {
		t.Errorf("expected udpted product to have %s, got %s ID", product.ProductID, got.ProductID)
	}

	if got.Name != "Keyboard" {
		t.Errorf("expected name %q, got %q", "Keyboard", got.Name)
	}
}

type RenamedJSONProduct struct {
	forgeapi.Model
	Name string `json:"display_name" gorm:"column:product_title"`
}

func TestPatchProductWithJSONNameDifferentFromColumn(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&RenamedJSONProduct{}); err != nil {
		t.Fatal(err)
	}

	product := RenamedJSONProduct{Name: "Keyboard"}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&RenamedJSONProduct{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/renamedjsonproducts/%d", product.ID),
		strings.NewReader(`{"display_name":"Mouse"}`),
	)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var got RenamedJSONProduct
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.Name != "Mouse" {
		t.Errorf("expected updated name %q, got %q", "Mouse", got.Name)
	}
}
