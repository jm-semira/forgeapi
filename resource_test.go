package forgeapi_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jm-semira/forgeapi"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	forgeapi.Model
	Name  string
	Price float64
}

type CustomProduct struct {
	ProductID string `json:"product_id" gorm:"primaryKey" forgeapi:"id"`
	Name      string `json:"name"`
}

type MissingIDProduct struct {
	Name string
}

type InvalidIDProduct struct {
	ProductID float64 `json:"product_id" gorm:"primaryKey" forgeapi:"id"`
	Name      string  `json:"name"`
}

type CompositeIDProduct struct {
	ProductID string `json:"product_id" gorm:"primaryKey" forgeapi:"id"`
	TenantID  string `json:"tenant_id" gorm:"primaryKey"`
	Name      string `json:"name"`
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(
		sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"),
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

	return db
}

func TestRegisterModelWithUnsupportedIDType(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&InvalidIDProduct{}); err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))

	err := app.Register(&InvalidIDProduct{})
	if !errors.Is(err, forgeapi.ErrUnsupportedIdentifierType) {
		t.Fatalf(
			"expected error %v, got %v",
			forgeapi.ErrUnsupportedIdentifierType,
			err,
		)
	}
}

func TestRegisterModelWithCompositePrimaryKey(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&CompositeIDProduct{}); err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))

	err := app.Register(&CompositeIDProduct{})
	if !errors.Is(err, forgeapi.ErrInvalidModelMetadata) {
		t.Fatalf(
			"expected error %v, got %v",
			forgeapi.ErrInvalidModelMetadata,
			err,
		)
	}
}

func TestRegisterProductList(t *testing.T) {
	db := setupTestDB(t)

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

func TestModelCreatesIDAndTimestamps(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	product := Product{
		Name:  "Keyboard",
		Price: 99.99,
	}

	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}

	if product.ID == 0 {
		t.Fatal("expected ID to be generated")
	}

	if product.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be populated")
	}

	if product.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be populated")
	}
}

func TestRegisterModelWithoutID(t *testing.T) {
	db := setupTestDB(t)
	app := forgeapi.New(forgeapi.WithDatabase(db))

	err := app.Register(&MissingIDProduct{})
	if err == nil {
		t.Fatal("expected error when registering a model without an ID")
	}
}
