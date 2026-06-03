package main

import (
	"fmt"
	"log"
	"os"
	"product-service/internal/handler"
	model "product-service/internal/models"
	"product-service/internal/repository"
	"product-service/internal/service"
	"product-service/pkg/config"
	"product-service/pkg/db"
	"product-service/pkg/logger"
	"product-service/seed"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Get database configuration from environment
	logger.Init()
	defer logger.Sync()

	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Println("⚠️ No .env file found")
	}

	dsn := os.Getenv("DATABASE_DSN")

	gormDB, err := db.InitDB(dsn)
	if err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}

	if err = gormDB.AutoMigrate(&model.Product{}); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	} else {
		log.Println("✅ Product table migration successful!")
	}

	var count int64
	if err := gormDB.Model(&model.Product{}).Count(&count).Error; err != nil {
		log.Fatalf("failed to count existing products: %v", err)
	}

	if count == 0 {
		log.Println("Seeding products for the first time ... ")
		if err := seed.SeedProducts(gormDB); err != nil {
			log.Fatalf("failed to seed products: %v", err)
		}
		log.Println("Products seeded successfully ")
	} else {
		log.Println("Products alreadys existed, skipping seed.")
	}

	// Repository Pattern

	productRepo := repository.NewProductRepository(gormDB)
	productSvc := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productSvc)

	// Set up HTTP server
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "productservice up"})
	})

	api := r.Group("/products")

	api.POST("/create", productHandler.Create)
	api.GET("/list", productHandler.List)
	api.GET("/:id", productHandler.GetByID)
	api.PUT("/:id", productHandler.Update)
	api.DELETE("/:id", productHandler.Delete)
	api.PATCH("/:id/reduce-stock", productHandler.ReduceStock)

	port := config.GetEnv("PORT", "8082")
	fmt.Println("✅ ProductService running on port", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start product service: %v", err)
	}
}
