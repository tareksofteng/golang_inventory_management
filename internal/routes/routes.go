package routes

import (
	"inventory-api/internal/controllers"

	"github.com/gin-gonic/gin"
)

// Controllers bundles every controller the router needs. As we add modules we
// add a field here, instead of growing Register's parameter list. main() builds
// this struct after wiring the dependency graph.
type Controllers struct {
	Category *controllers.CategoryController
	// Supplier, Product, StockIn, Dashboard added in later phases.
}

// Register mounts all API routes under a versioned /api/v1 prefix. Versioning
// the API from day one means a future breaking change can ship as /api/v2
// without disturbing existing clients.
func Register(router *gin.Engine, c Controllers) {
	api := router.Group("/api/v1")

	registerCategoryRoutes(api, c.Category)
}

// registerCategoryRoutes wires the 5 RESTful category endpoints. Each module
// gets its own small function like this — clean and easy to scan.
func registerCategoryRoutes(rg *gin.RouterGroup, ctrl *controllers.CategoryController) {
	g := rg.Group("/categories")
	{
		g.POST("", ctrl.Create)       // POST   /api/v1/categories
		g.GET("", ctrl.List)          // GET    /api/v1/categories
		g.GET("/:id", ctrl.Get)       // GET    /api/v1/categories/:id
		g.PUT("/:id", ctrl.Update)    // PUT    /api/v1/categories/:id
		g.DELETE("/:id", ctrl.Delete) // DELETE /api/v1/categories/:id
	}
}
