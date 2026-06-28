package main

import (
	"log"
	"os"

	"inventory-api/config"
	_ "inventory-api/docs" // generated swagger docs
	"inventory-api/internal/controllers"
	"inventory-api/internal/middleware"
	"inventory-api/internal/models"
	"inventory-api/internal/repositories"
	"inventory-api/internal/routes"
	"inventory-api/internal/seeder"
	"inventory-api/internal/services"
	"inventory-api/pkg/auth"
	"inventory-api/pkg/database"
	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Inventory Management API
// @version         1.0
// @description     Production-style inventory + POS REST API built with Gin, GORM and MySQL.
// @description     Authenticate via POST /auth/login, then click "Authorize" and enter: Bearer &lt;access_token&gt;.
// @host            localhost:9000
// @BasePath        /api/v1
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
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

	// 2c. Seed the first super-admin if the users table is empty.
	if err := seeder.SeedSuperAdmin(db, cfg.Seed); err != nil {
		log.Fatalf("startup: seeding failed: %v", err)
	}

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
	router.Use(middleware.CORS()) // allow the Vue frontend to call the API

	// Serve uploaded images at /uploads/... (created on startup if missing).
	_ = os.MkdirAll("uploads", 0o755)
	router.Static("/uploads", "./uploads")

	// Swagger UI at /swagger/index.html (interactive API documentation).
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
	tokenManager := auth.NewTokenManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTTLMinutes,
		cfg.JWT.RefreshTTLHours,
	)

	userRepo := repositories.NewUserRepository(db)
	refreshRepo := repositories.NewRefreshTokenRepository(db)
	userService := services.NewUserService(userRepo, refreshRepo)
	authService := services.NewAuthService(userRepo, refreshRepo, tokenManager)
	authController := controllers.NewAuthController(authService, userService)
	userController := controllers.NewUserController(userService)

	categoryRepo := repositories.NewCategoryRepository(db)
	categoryService := services.NewCategoryService(categoryRepo)
	categoryController := controllers.NewCategoryController(categoryService)

	supplierRepo := repositories.NewSupplierRepository(db)
	supplierService := services.NewSupplierService(supplierRepo)
	supplierController := controllers.NewSupplierController(supplierService)

	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo, categoryRepo, supplierRepo)
	productController := controllers.NewProductController(productService)

	customerRepo := repositories.NewCustomerRepository(db)
	customerService := services.NewCustomerService(customerRepo)
	customerController := controllers.NewCustomerController(customerService)

	dashboardRepo := repositories.NewDashboardRepository(db)
	dashboardService := services.NewDashboardService(dashboardRepo)
	dashboardController := controllers.NewDashboardController(dashboardService)

	purchaseRepo := repositories.NewPurchaseRepository(db)
	purchaseService := services.NewPurchaseService(purchaseRepo, supplierRepo, productRepo)
	purchaseController := controllers.NewPurchaseController(purchaseService)

	saleRepo := repositories.NewSaleRepository(db)
	saleService := services.NewSaleService(saleRepo, customerRepo, productRepo)
	saleController := controllers.NewSaleController(saleService)

	reportRepo := repositories.NewReportRepository(db)
	reportService := services.NewReportService(reportRepo)
	reportController := controllers.NewReportController(reportService)

	paymentRepo := repositories.NewPaymentRepository(db)
	paymentService := services.NewPaymentService(paymentRepo, customerRepo, supplierRepo)
	paymentController := controllers.NewPaymentController(paymentService)

	returnRepo := repositories.NewReturnRepository(db)
	returnService := services.NewReturnService(returnRepo, supplierRepo, customerRepo, productRepo)
	returnController := controllers.NewReturnController(returnService)

	ledgerRepo := repositories.NewLedgerRepository(db)
	ledgerService := services.NewLedgerService(ledgerRepo, customerRepo, supplierRepo)
	ledgerController := controllers.NewLedgerController(ledgerService)

	uploadController := controllers.NewUploadController()

	// 8. Register all API routes.
	routes.Register(router, routes.Controllers{
		Auth:      authController,
		User:      userController,
		Category:  categoryController,
		Supplier:  supplierController,
		Product:   productController,
		Customer:  customerController,
		Dashboard: dashboardController,
		Purchase:  purchaseController,
		Sale:      saleController,
		Report:    reportController,
		Payment:   paymentController,
		Return:    returnController,
		Ledger:    ledgerController,
		Upload:    uploadController,
	}, tokenManager)

	// 9. Start the server. router.Run blocks forever (until the process is
	//    killed), serving requests on the configured port.
	addr := ":" + cfg.AppPort
	log.Printf("startup: server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("startup: server failed: %v", err)
	}
}
