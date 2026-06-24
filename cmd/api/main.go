package main

import (
	"log"

	"inventory-api/config"
	"inventory-api/internal/controllers"
	"inventory-api/internal/models"
	"inventory-api/internal/repositories"
	"inventory-api/internal/routes"
	"inventory-api/internal/services"
	"inventory-api/pkg/database"
	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration from .env / environment.
	cfg := config.Load()
	isProduction := cfg.AppEnv == "production"

	// 2. Connect to the database (fail fast if it is unreachable).
	db, err := database.Connect(cfg.DB.DSN(), isProduction)
	if err != nil {
		// log.Fatal prints the error and exits with status 1. In main() this
		// is correct: if we cannot reach the DB, there is no point starting.
		log.Fatalf("startup: %v", err)
	}
	log.Println("startup: database connected")

	// 2b. Run schema migrations (create/update tables from our models).
	if err := models.AutoMigrate(db); err != nil {
		log.Fatalf("startup: migration failed: %v", err)
	}
	log.Println("startup: migrations applied")

	// 3. Configure Gin's mode. Release mode disables debug logs and is the
	//    production default; development keeps the verbose debug output.
	if isProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	// 4. Make validation error keys use json field names (e.g. category_id).
	response.RegisterValidatorJSONTags()

	// 5. Build the HTTP router. gin.Default() includes Logger + Recovery
	//    middleware (Recovery turns a panic into a 500 instead of crashing
	//    the whole server — like Laravel's exception handler).
	router := gin.Default()

	// 6. A health-check endpoint. Load balancers / uptime monitors hit this
	//    to confirm the service is alive. We also ping the DB here.
	router.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(503, gin.H{"status": "unhealthy", "database": "down"})
			return
		}
		c.JSON(200, gin.H{"status": "ok", "database": "up"})
	})

	// 7. Wire the dependency graph: repository -> service -> controller.
	//    This is manual dependency injection — explicit and easy to follow.
	categoryRepo := repositories.NewCategoryRepository(db)
	categoryService := services.NewCategoryService(categoryRepo)
	categoryController := controllers.NewCategoryController(categoryService)

	supplierRepo := repositories.NewSupplierRepository(db)
	supplierService := services.NewSupplierService(supplierRepo)
	supplierController := controllers.NewSupplierController(supplierService)

	// 8. Register all API routes.
	routes.Register(router, routes.Controllers{
		Category: categoryController,
		Supplier: supplierController,
	})

	// 9. Start the server. router.Run blocks forever (until the process is
	//    killed), serving requests on the configured port.
	addr := ":" + cfg.AppPort
	log.Printf("startup: server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("startup: server failed: %v", err)
	}
}
