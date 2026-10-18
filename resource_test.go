package forgeapi_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jm-semira/forgeapi"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name  string
	Price float64
}

func TestRegisterProductList(t *testing.T) {
	db, err := gorm.Open(
		sqlite.Open("file:forgeapi_test?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	if err := db.Create(&Product{
		Name:  "Keyboard",
		Price: 1500,
	}).Error; err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))

	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()

	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var products []Product
	if err := json.Unmarshal(rec.Body.Bytes(), &products); err != nil {
		t.Fatal(err)
	}

	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}

	if products[0].Name != "Keyboard" {
		t.Errorf("expected Keyboard, got %q", products[0].Name)
	}
}

func TestGetProductByID(t *testing.T) {
	db, err := gorm.Open(
		sqlite.Open("file:get_product_by_id?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

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
	db, err := gorm.Open(
		sqlite.Open("file:get_product_not_found?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

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

func TestPostProduct(t *testing.T) {
	db, err := gorm.Open(
		sqlite.Open("file:post_product?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

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

func TestPatchProduct(t *testing.T) {
	db, err := gorm.Open(
		sqlite.Open("file:patch_product?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

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

func TestPatchProductNotFound(t *testing.T) {
	db, err := gorm.Open(
		sqlite.Open("file:patch_product_not_found?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

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
	db, err := gorm.Open(
		sqlite.Open("file:patch_product_invalid_id?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

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
	db, err := gorm.Open(
		sqlite.Open("file:patch_product_empty_body?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

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

func TestDeleteProduct(t *testing.T) {
	db, err := gorm.Open(
		sqlite.Open("file:delete_product?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

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
	err = db.First(&got, product.ID).Error

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected product to be deleted, got error: %v", err)
	}
}

func TestDeleteProductNotExist(t *testing.T) {
	db, err := gorm.Open(
		sqlite.Open("file:delete_product_not_exist?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

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
	db, err := gorm.Open(
		sqlite.Open("file:delete_product_invalid_id?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

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
