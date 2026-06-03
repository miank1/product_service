package seed

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	model "product-service/internal/models"
	"runtime"
	"time"

	"gorm.io/gorm"
)

// SeedProducts loads products.json and inserts into DB
func SeedProducts(db *gorm.DB) error {
	start := time.Now()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("resolve seed file path: runtime caller unavailable")
	}

	seedFile := filepath.Join(filepath.Dir(currentFile), "products.json")

	data, err := os.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("read products.json: %w", err)
	}

	var products []model.Product
	if err := json.Unmarshal(data, &products); err != nil {
		return fmt.Errorf("parse products.json: %w", err)
	}

	var existingNames []string
	if err := db.Model(&model.Product{}).Select("name").Find(&existingNames).Error; err != nil {
		return fmt.Errorf("fetch existing product names: %w", err)
	}

	existingSet := make(map[string]bool, len(existingNames))
	for _, name := range existingNames {
		existingSet[name] = true
	}

	var newProducts []model.Product
	for _, p := range products {
		if !existingSet[p.Name] {
			newProducts = append(newProducts, p)
		}
	}

	if len(newProducts) > 0 {
		if err := db.CreateInBatches(newProducts, 100).Error; err != nil {
			return fmt.Errorf("batch insert products: %w", err)
		}
		fmt.Printf("✅ Inserted %d new products (%.2fs)\n", len(newProducts), time.Since(start).Seconds())
	} else {
		fmt.Println("ℹ️ All products already exist — nothing to insert.")
	}

	return nil
}
