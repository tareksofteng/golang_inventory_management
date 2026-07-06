package seeder

import (
	"log"

	"inventory-api/internal/models"

	"gorm.io/gorm"
)

// SeedAccounts inserts a standard Chart of Accounts IF the accounts table is
// empty. Idempotent — safe on every startup.
func SeedAccounts(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Account{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	defaults := []models.Account{
		{Code: "1000", Name: "Cash", Type: "asset", IsActive: true},
		{Code: "1010", Name: "Bank", Type: "asset", IsActive: true},
		{Code: "1100", Name: "Accounts Receivable", Type: "asset", IsActive: true},
		{Code: "1200", Name: "Inventory", Type: "asset", IsActive: true},
		{Code: "2000", Name: "Accounts Payable", Type: "liability", IsActive: true},
		{Code: "3000", Name: "Owner's Capital", Type: "equity", IsActive: true},
		{Code: "4000", Name: "Sales Income", Type: "income", IsActive: true},
		{Code: "4100", Name: "Other Income", Type: "income", IsActive: true},
		{Code: "5000", Name: "Cost of Goods Sold", Type: "expense", IsActive: true},
		{Code: "5100", Name: "Salary Expense", Type: "expense", IsActive: true},
		{Code: "5200", Name: "Rent Expense", Type: "expense", IsActive: true},
		{Code: "5300", Name: "Utility Expense", Type: "expense", IsActive: true},
	}
	if err := db.Create(&defaults).Error; err != nil {
		return err
	}
	log.Printf("seeder: %d chart-of-accounts accounts created", len(defaults))
	return nil
}
