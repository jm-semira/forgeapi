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

type JsonTaggedProduct struct {
	forgeapi.Model
	Name  string  `json:"name"`
	Price float64 `json:"price"`
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

func TestRegisterModelValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		model   any
		wantErr error
	}{
		{
			name:    "unsupported ID type",
			model:   &InvalidIDProduct{},
			wantErr: forgeapi.ErrUnsupportedIdentifierType,
		},
		{
			name:    "composite primary key",
			model:   &CompositeIDProduct{},
			wantErr: forgeapi.ErrInvalidModelMetadata,
		},
		{
			name:    "missing ID metadata",
			model:   &MissingIDProduct{},
			wantErr: forgeapi.ErrInvalidModelMetadata,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			if err := db.AutoMigrate(tt.model); err != nil {
				t.Fatal(err)
			}
			app := forgeapi.New(forgeapi.WithDatabase(db))

			err := app.Register(tt.model)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Register() error = %v, want %v", err, tt.wantErr)
			}
		})
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
