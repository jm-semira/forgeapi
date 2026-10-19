package forgeapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jm-semira/forgeapi"
)

func TestGetProductByID(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	product := Product{
		Name:  "Keyboard",
		Price: 1500,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/products/%d", product.ID),
		nil,
	)
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
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	if got.ID != product.ID {
		t.Errorf("expected ID %d, got %d", product.ID, got.ID)
	}
	if got.Name != product.Name {
		t.Errorf("expected name %q, got %q", product.Name, got.Name)
	}
	if got.Price != product.Price {
		t.Errorf("expected price %.2f, got %.2f", product.Price, got.Price)
	}
}

func TestGetProductNotFound(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/products/%d", 9999),
		nil,
	)
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

func TestGetProductByCustomID(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&CustomProduct{}); err != nil {
		t.Fatal(err)
	}

	product := CustomProduct{
		ProductID: "sku-123",
		Name:      "Keyboard",
	}

	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))

	if err := app.Register(&CustomProduct{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/customproducts/sku-123",
		nil,
	)
	rec := httptest.NewRecorder()

	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			rec.Code,
			rec.Body.String(),
		)
	}

	var got CustomProduct
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	if got.ProductID != "sku-123" {
		t.Errorf("expected ID sku-123, got %q", got.ProductID)
	}
}
