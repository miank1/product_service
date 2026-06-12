package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	model "product_service/internal/models"
	"product_service/internal/service"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handlerStubRepo struct {
	getByIDFn     func(string) (*model.Product, error)
	updateFn      func(*model.Product) error
	reduceStockFn func(string, int) error
}

func (s *handlerStubRepo) Create(*model.Product) error               { return nil }
func (s *handlerStubRepo) GetAll() ([]model.Product, error)          { return nil, nil }
func (s *handlerStubRepo) Delete(string) error                       { return nil }
func (s *handlerStubRepo) GetByID(id string) (*model.Product, error) { return s.getByIDFn(id) }
func (s *handlerStubRepo) Update(p *model.Product) error             { return s.updateFn(p) }
func (s *handlerStubRepo) ReduceStock(id string, qty int) error {
	if s.reduceStockFn != nil {
		return s.reduceStockFn(id, qty)
	}
	return nil
}

func TestReduceStockUsesPathParamAndQuantityBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	productID := uuid.New()
	repo := &handlerStubRepo{
		getByIDFn: func(id string) (*model.Product, error) {
			if id != productID.String() {
				t.Fatalf("expected product id %s, got %s", productID.String(), id)
			}
			return &model.Product{
				ID:    productID,
				Name:  "Headphones",
				Price: 199.99,
				Stock: 10,
			}, nil
		},
		updateFn: func(p *model.Product) error {
			if p.Stock != 8 {
				t.Fatalf("expected reduced stock to be 8, got %d", p.Stock)
			}
			return nil
		},
	}

	svc := service.NewProductService(repo)
	h := NewProductHandler(svc)

	router := gin.New()
	router.PATCH("/api/v1/products/:id/reduce-stock", h.ReduceStock)

	body, err := json.Marshal(map[string]int{"quantity": 2})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/products/"+productID.String()+"/reduce-stock", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestReduceStockRejectsMissingQuantity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &handlerStubRepo{
		getByIDFn: func(id string) (*model.Product, error) {
			return &model.Product{ID: uuid.New(), Stock: 10}, nil
		},
		updateFn: func(p *model.Product) error { return nil },
	}

	svc := service.NewProductService(repo)
	h := NewProductHandler(svc)

	router := gin.New()
	router.PATCH("/api/v1/products/:id/reduce-stock", h.ReduceStock)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/products/"+uuid.NewString()+"/reduce-stock", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
