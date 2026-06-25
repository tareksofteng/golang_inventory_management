package routes

import (
	"inventory-api/internal/controllers"
	"inventory-api/internal/middleware"
	"inventory-api/internal/rbac"
	"inventory-api/pkg/auth"

	"github.com/gin-gonic/gin"
)

// Controllers bundles every controller the router needs.
type Controllers struct {
	Auth     *controllers.AuthController
	User     *controllers.UserController
	Category *controllers.CategoryController
	Supplier *controllers.SupplierController
	Product  *controllers.ProductController
	Customer *controllers.CustomerController
}

// Register mounts all API routes under /api/v1. Auth endpoints are public;
// everything else sits behind the JWT Auth middleware, and each resource group
// further requires a specific RBAC permission.
func Register(router *gin.Engine, c Controllers, tm *auth.TokenManager) {
	api := router.Group("/api/v1")

	// ---- Public auth endpoints ----
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", c.Auth.Login)
		authGroup.POST("/refresh", c.Auth.Refresh)
		authGroup.POST("/logout", c.Auth.Logout)
	}

	// ---- Protected: requires a valid access token ----
	protected := api.Group("")
	protected.Use(middleware.Auth(tm))

	protected.GET("/auth/me", c.Auth.Me)

	registerUserRoutes(protected, c.User)
	registerCategoryRoutes(protected, c.Category)
	registerSupplierRoutes(protected, c.Supplier)
	registerProductRoutes(protected, c.Product)
	registerCustomerRoutes(protected, c.Customer)
}

func registerCustomerRoutes(rg *gin.RouterGroup, ctrl *controllers.CustomerController) {
	g := rg.Group("/customers")
	g.Use(middleware.RequirePermission(rbac.PermSalesManage))
	{
		g.POST("", ctrl.Create)
		g.GET("", ctrl.List)
		g.GET("/:id", ctrl.Get)
		g.PUT("/:id", ctrl.Update)
		g.DELETE("/:id", ctrl.Delete)
	}
}

func registerUserRoutes(rg *gin.RouterGroup, ctrl *controllers.UserController) {
	g := rg.Group("/users")
	g.Use(middleware.RequirePermission(rbac.PermUserManage))
	{
		g.POST("", ctrl.Create)
		g.GET("", ctrl.List)
		g.GET("/:id", ctrl.Get)
		g.PUT("/:id", ctrl.Update)
		g.PUT("/:id/password", ctrl.ChangePassword)
		g.PATCH("/:id/disable", ctrl.Disable)
	}
}

func registerCategoryRoutes(rg *gin.RouterGroup, ctrl *controllers.CategoryController) {
	g := rg.Group("/categories")
	g.Use(middleware.RequirePermission(rbac.PermProductManage))
	{
		g.POST("", ctrl.Create)
		g.GET("", ctrl.List)
		g.GET("/:id", ctrl.Get)
		g.PUT("/:id", ctrl.Update)
		g.DELETE("/:id", ctrl.Delete)
	}
}

func registerSupplierRoutes(rg *gin.RouterGroup, ctrl *controllers.SupplierController) {
	g := rg.Group("/suppliers")
	g.Use(middleware.RequirePermission(rbac.PermProductManage))
	{
		g.POST("", ctrl.Create)
		g.GET("", ctrl.List)
		g.GET("/:id", ctrl.Get)
		g.PUT("/:id", ctrl.Update)
		g.DELETE("/:id", ctrl.Delete)
	}
}

func registerProductRoutes(rg *gin.RouterGroup, ctrl *controllers.ProductController) {
	g := rg.Group("/products")
	g.Use(middleware.RequirePermission(rbac.PermProductManage))
	{
		g.POST("", ctrl.Create)
		g.GET("", ctrl.List)
		g.GET("/:id", ctrl.Get)
		g.PUT("/:id", ctrl.Update)
		g.DELETE("/:id", ctrl.Delete)
	}
}
