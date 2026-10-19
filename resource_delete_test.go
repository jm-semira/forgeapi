package forgeapi_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jm-semira/forgeapi"
	"gorm.io/gorm"
)

func TestDeleteProduct(t *testing.T) {
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
		http.MethodDelete,
		fmt.Sprintf("/products/%d", product.ID),
		nil,
	)

	rec := httptest.NewRecorder()

	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNoContent,
			rec.Code,
			rec.Body.String(),
		)
	}

	var got Product
	err := db.First(&got, product.ID).Error

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected product to be deleted, got error: %v", err)
	}
}

func TestDeleteProductNotExist(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodDelete,
		"/products/1",
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

func TestDeleteProductInvalidID(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodDelete,
		"/products/abc",
		nil,
	)

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
func TestDeleteProductByCustomID(t *testing.T) {
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
		http.MethodDelete,
		fmt.Sprintf("/customproducts/%s", product.ProductID),
		nil,
	)

	rec := httptest.NewRecorder()

	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNoContent,
			rec.Code,
			rec.Body.String(),
		)
	}

	var got CustomProduct
	err := db.Where("product_id = ?", product.ProductID).First(&got).Error

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected product to be deleted, got error: %v", err)
	}
}
