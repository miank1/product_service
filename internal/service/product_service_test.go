package service

import (
	"errors"
	model "product-service/internal/models"
	"testing"

	"github.com/google/uuid"
)

type stubProductRepo struct {
	createFn      func(*model.Product) error
	getAllFn      func() ([]model.Product, error)
	getByIDFn     func(string) (*model.Product, error)
	updateFn      func(*model.Product) error
	deleteFn      func(string) error
	reduceStockFn func(string, int) error
}

func (s *stubProductRepo) Create(p *model.Product) error {
	if s.createFn != nil {
		return s.createFn(p)
	}
	return nil
}

func (s *stubProductRepo) GetAll() ([]model.Product, error) {
	if s.getAllFn != nil {
		return s.getAllFn()
	}
	return nil, nil
}

func (s *stubProductRepo) GetByID(id string) (*model.Product, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(id)
	}
	return nil, nil
}

func (s *stubProductRepo) Update(p *model.Product) error {
	if s.updateFn != nil {
		return s.updateFn(p)
	}
	return nil
}

func (s *stubProductRepo) Delete(id string) error {
	if s.deleteFn != nil {
		return s.deleteFn(id)
	}
	return nil
}

func (s *stubProductRepo) ReduceStock(id string, qty int) error {
	if s.reduceStockFn != nil {
		return s.reduceStockFn(id, qty)
	}
	return nil
}

func TestCreateProductRejectsInvalidInput(t *testing.T) {
	svc := NewProductService(&stubProductRepo{})

	tests := []struct {
		name  string
		input struct {
			name  string
			price float64
			stock int
		}
	}{
		{
			name: "empty name",
			input: struct {
				name  string
				price float64
				stock int
			}{name: "", price: 10, stock: 5},
		},
		{
			name: "non-positive price",
			input: struct {
				name  string
				price float64
				stock int
			}{name: "Laptop", price: 0, stock: 5},
		},
		{
			name: "negative stock",
			input: struct {
				name  string
				price float64
				stock int
			}{name: "Laptop", price: 10, stock: -1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := svc.CreateProduct(tt.input.name, "desc", tt.input.price, tt.input.stock)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if product != nil {
				t.Fatalf("expected nil product on invalid input")
			}
		})
	}
}

func TestReduceStockRejectsNonPositiveQuantity(t *testing.T) {
	svc := NewProductService(&stubProductRepo{})

	err := svc.ReduceStock(uuid.NewString(), 0)
	if err == nil {
		t.Fatalf("expected error for zero quantity")
	}

	err = svc.ReduceStock(uuid.NewString(), -2)
	if err == nil {
		t.Fatalf("expected error for negative quantity")
	}
}

func TestReduceStockRejectsInsufficientStock(t *testing.T) {
	repo := &stubProductRepo{
		reduceStockFn: func(id string, qty int) error {
			return errors.New("insufficient stock")
		},
	}

	svc := NewProductService(repo)

	err := svc.ReduceStock(uuid.NewString(), 3)
	if err == nil {
		t.Fatalf("expected insufficient stock error")
	}
}

func TestReduceStockUpdatesStockWhenQuantityIsValid(t *testing.T) {
	updated := false
	repo := &stubProductRepo{
		reduceStockFn: func(id string, qty int) error {
			updated = true
			if qty != 3 {
				t.Fatalf("expected quantity 3, got %d", qty)
			}
			return nil
		},
	}

	svc := NewProductService(repo)

	if err := svc.ReduceStock(uuid.NewString(), 3); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !updated {
		t.Fatalf("expected repository ReduceStock to be called")
	}
}

func TestReduceStockReturnsRepoError(t *testing.T) {
	repo := &stubProductRepo{
		reduceStockFn: func(id string, qty int) error {
			return errors.New("db unavailable")
		},
	}

	svc := NewProductService(repo)

	err := svc.ReduceStock(uuid.NewString(), 1)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
