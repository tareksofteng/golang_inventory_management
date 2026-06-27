package models

import "gorm.io/gorm"

// AutoMigrate creates or updates the database tables to match our model
// structs. GORM reads the struct fields + tags and issues the needed
// CREATE TABLE / ALTER TABLE statements.
//
// Laravel comparison: this replaces hand-written migration files +
// `php artisan migrate`. Here the struct IS the schema.
//
// Order matters: parent tables (Category, Supplier) are migrated before the
// children that reference them (Product, StockIn), so the foreign-key columns
// have something to point at.
//
// NOTE: AutoMigrate only ADDS tables/columns/indexes — it never drops a column
// or table. That makes it safe to run on every startup, even in production.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&RefreshToken{},
		&Customer{},
		&Category{},
		&Supplier{},
		&Product{},
		&StockIn{},
		&Purchase{},
		&PurchaseItem{},
		&Sale{},
		&SaleItem{},
		&Payment{},
	)
}
