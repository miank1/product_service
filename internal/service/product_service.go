package service

import (
	"errors"
	"fmt"
	model "product_service/internal/models"

	"gorm.io/gorm"
)

var ErrProductNotFound = errors.New("product not found")

type ProductRepository interface {
	Create(*model.Product) error
	GetAll() ([]model.Product, error)
	GetByID(string) (*model.Product, error)
	Update(*model.Product) error
	Delete(string) error
	ReduceStock(string, int) error
}

type ProductService struct {
	Repo ProductRepository
}

func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{Repo: repo}
}

func (s *ProductService) CreateProduct(name, desc string, price float64, stock int) (*model.Product, error) {
	if name == "" || price <= 0 || stock < 0 {
		return nil, errors.New("invalid product details")
	}

	product := &model.Product{
		Name:        name,
		Description: desc,
		Price:       price,
		Stock:       stock,
	}

	if err := s.Repo.Create(product); err != nil {
		return nil, err
	}
	return product, nil
}

// GetAllProducts returns all products
func (s *ProductService) GetAllProducts() ([]model.Product, error) {
	return s.Repo.GetAll()
}

// GetProductByID returns a single product by ID
func (s *ProductService) GetProductByID(id string) (*model.Product, error) {
	product, err := s.Repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	if product == nil {
		return nil, ErrProductNotFound
	}
	return product, nil
}

// Update an existing product
func (s *ProductService) UpdateProduct(id string, name, desc *string, price *float64, stock *int) (*model.Product, error) {
	product, err := s.Repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	if product == nil {
		return nil, ErrProductNotFound
	}

	if name != nil {
		if *name == "" {
			return nil, errors.New("name cannot be empty")
		}
		product.Name = *name
	}
	if desc != nil {
		product.Description = *desc
	}
	if price != nil {
		if *price <= 0 {
			return nil, errors.New("price must be greater than zero")
		}
		product.Price = *price
	}
	if stock != nil {
		if *stock < 0 {
			return nil, errors.New("stock cannot be negative")
		}
		product.Stock = *stock
	}

	if err := s.Repo.Update(product); err != nil {
		return nil, err
	}
	return product, nil
}

// Delete a product
func (s *ProductService) DeleteProduct(id string) error {
	return s.Repo.Delete(id)
}

func (s *ProductService) ReduceStock(productID string, qty int) error {
	if qty <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	if err := s.Repo.ReduceStock(productID, qty); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return err
	}

	return nil
}
