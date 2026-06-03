package repository

import (
	"errors"
	model "product-service/internal/models"

	"gorm.io/gorm"
)

type ProductRepository struct {
	DB *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{DB: db}
}

func (r *ProductRepository) Create(p *model.Product) error {
	return r.DB.Create(p).Error
}

func (r *ProductRepository) GetAll() ([]model.Product, error) {
	var products []model.Product
	if err := r.DB.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepository) GetByID(id string) (*model.Product, error) {
	var product model.Product
	if err := r.DB.First(&product, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

// Update product
func (r *ProductRepository) Update(p *model.Product) error {
	return r.DB.Save(p).Error
}

func (r *ProductRepository) ReduceStock(id string, qty int) error {
	res := r.DB.Model(&model.Product{}).
		Where("id = ? AND stock >= ?", id, qty).
		UpdateColumn("stock", gorm.Expr("stock - ?", qty))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		var count int64
		if err := r.DB.Model(&model.Product{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}
		return errors.New("insufficient stock")
	}
	return nil
}

// Delete product
func (r *ProductRepository) Delete(id string) error {
	res := r.DB.Delete(&model.Product{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
