package forgeapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestGetProductsFilterByNameNoJsonTag(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}

	products := []Product{
		{Name: "Keyboard", Price: 100},
		{Name: "Mouse", Price: 50},
		{Name: "Monitor", Price: 300},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&Product{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/products?filter[Name][eq]=Keyboard",
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

	var got []Product
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 product, got %d", len(got))
	}

	if got[0].Name != "Keyboard" {
		t.Errorf("expected Keyboard, got %q", got[0].Name)
	}
}

func TestGetProductsFilterByName(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AutoMigrate(&JsonTaggedProduct{}); err != nil {
		t.Fatal(err)
	}

	products := []JsonTaggedProduct{
		{Name: "Keyboard", Price: 100},
		{Name: "Mouse", Price: 50},
		{Name: "Monitor", Price: 300},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatal(err)
	}

	app := forgeapi.New(forgeapi.WithDatabase(db))
	if err := app.Register(&JsonTaggedProduct{}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/jsontaggedproducts?filter[name][eq]=Keyboard",
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

	var got []JsonTaggedProduct
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 product, got %d", len(got))
	}

	if got[0].Name != "Keyboard" {
		t.Errorf("expected Keyboard, got %q", got[0].Name)
	}
}

func setupJsonTaggedProductList(t *testing.T, products []JsonTaggedProduct, opts ...forgeapi.Option) http.Handler {
	t.Helper()

	db := setupTestDB(t)
	if err := db.AutoMigrate(&JsonTaggedProduct{}); err != nil {
		t.Fatal(err)
	}
	if len(products) > 0 {
		if err := db.Create(&products).Error; err != nil {
			t.Fatal(err)
		}
	}

	appOpts := append([]forgeapi.Option{forgeapi.WithDatabase(db)}, opts...)
	app := forgeapi.New(appOpts...)
	if err := app.Register(&JsonTaggedProduct{}); err != nil {
		t.Fatal(err)
	}
	return app.Handler()
}

func TestGetProductsFilterInvalidFieldOrOperator(t *testing.T) {
	handler := setupJsonTaggedProductList(t, nil)
	tests := []struct {
		name  string
		query string
	}{
		{name: "unknown field", query: "filter[unknown][eq]=value"},
		{name: "unsupported operator", query: "filter[price][contains]=100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/jsontaggedproducts?"+tt.query, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestGetProductsFilterPriceComparisons(t *testing.T) {
	products := []JsonTaggedProduct{
		{Name: "Keyboard", Price: 100},
		{Name: "Mouse", Price: 50},
		{Name: "Monitor", Price: 300},
	}
	handler := setupJsonTaggedProductList(t, products)
	tests := []struct {
		name      string
		query     string
		wantNames []string
		minPrice  float64
		maxPrice  float64
	}{
		{
			name:      "greater than or equal",
			query:     "filter[price][gte]=100",
			wantNames: []string{"Keyboard", "Monitor"},
			minPrice:  100,
		},
		{
			name:      "less than or equal",
			query:     "filter[price][lte]=100",
			wantNames: []string{"Keyboard", "Mouse"},
			maxPrice:  100,
		},
		{
			name:      "combined range",
			query:     "filter[price][gte]=50&filter[price][lte]=100",
			wantNames: []string{"Keyboard", "Mouse"},
			minPrice:  50,
			maxPrice:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/jsontaggedproducts?"+tt.query, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
			}

			var got []JsonTaggedProduct
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.wantNames) {
				t.Fatalf("expected %d products, got %d", len(tt.wantNames), len(got))
			}

			gotNames := make(map[string]bool, len(got))
			for _, product := range got {
				gotNames[product.Name] = true
				if tt.minPrice > 0 && product.Price < tt.minPrice {
					t.Errorf("expected price >= %v, got %v", tt.minPrice, product.Price)
				}
				if tt.maxPrice > 0 && product.Price > tt.maxPrice {
					t.Errorf("expected price <= %v, got %v", tt.maxPrice, product.Price)
				}
			}
			for _, name := range tt.wantNames {
				if !gotNames[name] {
					t.Errorf("expected product %q in results, got %v", name, gotNames)
				}
			}
		})
	}
}

func TestGetProductsFilterSQLInjectionValue(t *testing.T) {
	products := []JsonTaggedProduct{
		{Name: "Keyboard", Price: 100},
		{Name: "Mouse", Price: 50},
	}
	handler := setupJsonTaggedProductList(t, products)

	req := httptest.NewRequest(
		http.MethodGet,
		"/jsontaggedproducts?filter[name][eq]="+url.QueryEscape("' OR 1=1 --"),
		nil,
	)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var got []JsonTaggedProduct
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 products, got %d", len(got))
	}
}

func TestGetProductsSort(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantCode  int
		wantNames []string
		products  []JsonTaggedProduct
	}{
		{
			name:      "ascending",
			query:     "sort=price",
			wantCode:  http.StatusOK,
			wantNames: []string{"Mouse", "Keyboard", "Monitor"},
			products: []JsonTaggedProduct{
				{Name: "Keyboard", Price: 100},
				{Name: "Mouse", Price: 50},
				{Name: "Monitor", Price: 300},
			},
		},
		{
			name:      "descending",
			query:     "sort=-price",
			wantCode:  http.StatusOK,
			wantNames: []string{"Monitor", "Keyboard", "Mouse"},
			products: []JsonTaggedProduct{
				{Name: "Keyboard", Price: 100},
				{Name: "Mouse", Price: 50},
				{Name: "Monitor", Price: 300},
			},
		},
		{
			name:      "multiple fields",
			query:     "sort=-price,name",
			wantCode:  http.StatusOK,
			wantNames: []string{"Keyboard", "Monitor", "Mouse"},
			products: []JsonTaggedProduct{
				{Name: "Keyboard", Price: 100},
				{Name: "Mouse", Price: 50},
				{Name: "Monitor", Price: 100},
			},
		},
		{name: "unknown field", query: "sort=unknown", wantCode: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupJsonTaggedProductList(t, tt.products)
			req := httptest.NewRequest(http.MethodGet, "/jsontaggedproducts?"+tt.query, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("expected status %d, got %d: %s", tt.wantCode, rec.Code, rec.Body.String())
			}
			if tt.wantCode != http.StatusOK {
				return
			}

			var got []JsonTaggedProduct
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.wantNames) {
				t.Fatalf("expected %d products, got %d", len(tt.wantNames), len(got))
			}
			for i, product := range got {
				if product.Name != tt.wantNames[i] {
					t.Errorf("product %d: expected %q, got %q", i, tt.wantNames[i], product.Name)
				}
			}
		})
	}
}

func TestGetProductsPagination(t *testing.T) {
	products := []JsonTaggedProduct{
		{Name: "Keyboard", Price: 100},
		{Name: "Mouse", Price: 50},
		{Name: "Monitor", Price: 300},
	}
	handler := setupJsonTaggedProductList(t, products)
	tests := []struct {
		name       string
		query      string
		wantNames  []string
		wantPrices []float64
	}{
		{name: "limit", query: "sort=price&limit=2", wantNames: []string{"Mouse", "Keyboard"}, wantPrices: []float64{50, 100}},
		{name: "offset", query: "sort=price&offset=1", wantNames: []string{"Keyboard", "Monitor"}, wantPrices: []float64{100, 300}},
		{name: "limit and offset", query: "sort=price&limit=1&offset=1", wantNames: []string{"Keyboard"}, wantPrices: []float64{100}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/jsontaggedproducts?"+tt.query, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
			}
			var got []JsonTaggedProduct
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.wantNames) {
				t.Fatalf("expected %d products, got %d", len(tt.wantNames), len(got))
			}
			for i, product := range got {
				if product.Name != tt.wantNames[i] {
					t.Errorf("product %d: expected name %q, got %q", i, tt.wantNames[i], product.Name)
				}
				if product.Price != tt.wantPrices[i] {
					t.Errorf("product %d: expected price %v, got %v", i, tt.wantPrices[i], product.Price)
				}
			}
		})
	}
}

func TestGetProductsInvalidPagination(t *testing.T) {
	handler := setupJsonTaggedProductList(t, nil)
	tests := []struct {
		name  string
		query string
	}{
		{name: "nonnumeric limit", query: "limit=abc"},
		{name: "zero limit", query: "limit=0"},
		{name: "negative limit", query: "limit=-1"},
		{name: "nonnumeric offset", query: "offset=abc"},
		{name: "negative offset", query: "offset=-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/jsontaggedproducts?"+tt.query, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestGetProductsDefaultPaginationLimit(t *testing.T) {
	products := make([]JsonTaggedProduct, 25)
	for i := range products {
		products[i] = JsonTaggedProduct{
			Name:  fmt.Sprintf("Product %02d", i+1),
			Price: float64(i + 1),
		}
	}
	tests := []struct {
		name           string
		query          string
		opts           []forgeapi.Option
		expectedCount  int
		expectedStatus int
	}{
		{
			name:           "no limit uses default",
			expectedCount:  20,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "custom default limit",
			opts:           []forgeapi.Option{forgeapi.WithDefaultLimit(5)},
			expectedCount:  5,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupJsonTaggedProductList(t, products, tt.opts...)
			req := httptest.NewRequest(http.MethodGet, "/jsontaggedproducts?"+tt.query, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			var got []JsonTaggedProduct
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if len(got) != tt.expectedCount {
				t.Fatalf("expected %d products, got %d", tt.expectedCount, len(got))
			}
		})
	}
}

func TestGetProductsMaxPaginationLimit(t *testing.T) {
	products := make([]JsonTaggedProduct, 5)
	for i := range products {
		products[i] = JsonTaggedProduct{
			Name:  fmt.Sprintf("Product %02d", i+1),
			Price: float64(i + 1),
		}
	}

	tests := []struct {
		name           string
		query          string
		opts           []forgeapi.Option
		expectedCount  int
		expectedStatus int
	}{
		{
			name:  "limit equals maximum",
			query: "?limit=3",
			opts: []forgeapi.Option{
				forgeapi.WithDefaultLimit(3),
				forgeapi.WithMaxLimit(3),
			},
			expectedCount:  3,
			expectedStatus: http.StatusOK,
		},
		{
			name:  "limit exceeds maximum",
			query: "?limit=4",
			opts: []forgeapi.Option{
				forgeapi.WithDefaultLimit(3),
				forgeapi.WithMaxLimit(3),
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupJsonTaggedProductList(t, products, tt.opts...)
			req := httptest.NewRequest(
				http.MethodGet,
				"/jsontaggedproducts"+tt.query,
				nil,
			)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d: %s",
					tt.expectedStatus,
					rec.Code,
					rec.Body.String(),
				)
			}

			if rec.Code != http.StatusOK {
				return
			}

			var got []JsonTaggedProduct
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if len(got) != tt.expectedCount {
				t.Fatalf("expected %d products, got %d", tt.expectedCount, len(got))
			}
		})
	}
}

func TestNewPanicsForInvalidPaginationConfig(t *testing.T) {
	tests := []struct {
		name string
		opts []forgeapi.Option
	}{
		{
			name: "non-positive default",
			opts: []forgeapi.Option{forgeapi.WithDefaultLimit(0)},
		},
		{
			name: "non-positive maximum",
			opts: []forgeapi.Option{forgeapi.WithMaxLimit(0)},
		},
		{
			name: "default exceeds maximum",
			opts: []forgeapi.Option{
				forgeapi.WithDefaultLimit(4),
				forgeapi.WithMaxLimit(3),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected New to panic for invalid pagination configuration")
				}
			}()

			forgeapi.New(tt.opts...)
		})
	}
}
