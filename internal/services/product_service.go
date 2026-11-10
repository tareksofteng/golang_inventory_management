package services

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"

	"gorm.io/gorm"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrProductSKUTaken   = errors.New("product SKU already exists")
	ErrInvalidAdjustment = errors.New("stock adjustment must be a non-zero quantity with a reason")
	// Reused for foreign-key validation: the referenced category/supplier
	// does not exist. (ErrCategoryNotFound / ErrSupplierNotFound are declared
	// in the category/supplier services — same package, so we reuse them.)
)

type ProductService interface {
	Create(product *models.Product) (*models.Product, error)
	List(search string, lowStockOnly bool, page, perPage int) ([]models.Product, int64, error)
	Export(search string, lowStockOnly bool) ([]models.Product, error)
	ImportCSV(r io.Reader) (*ImportResult, error)
	Get(id uint) (*models.Product, error)
	AdjustStock(productID, userID uint, change int, reason, note string) (*models.StockAdjustment, error)
	StockHistory(productID uint, page, perPage int) ([]models.StockAdjustment, int64, error)
	Update(id uint, data *models.Product) (*models.Product, error)
	Delete(id uint) error
}

// productService also depends on the category & supplier repositories so it can
// verify that a product's CategoryID / SupplierID actually point at real rows
// BEFORE inserting — otherwise the DB would reject it with a raw FK error, or
// (worse, since we don't add hard FK constraints) accept an orphan reference.
type productService struct {
	repo         repositories.ProductRepository
	categoryRepo repositories.CategoryRepository
	supplierRepo repositories.SupplierRepository
}

func NewProductService(
	repo repositories.ProductRepository,
	categoryRepo repositories.CategoryRepository,
	supplierRepo repositories.SupplierRepository,
) ProductService {
	return &productService{
		repo:         repo,
		categoryRepo: categoryRepo,
		supplierRepo: supplierRepo,
	}
}

// validateRefs ensures the referenced category and supplier exist.
func (s *productService) validateRefs(categoryID, supplierID uint) error {
	if _, err := s.categoryRepo.FindByID(categoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	if _, err := s.supplierRepo.FindByID(supplierID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSupplierNotFound
		}
		return err
	}
	return nil
}

func (s *productService) Create(product *models.Product) (*models.Product, error) {
	if err := s.validateRefs(product.CategoryID, product.SupplierID); err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsBySKU(product.SKU, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProductSKUTaken
	}

	if err := s.repo.Create(product); err != nil {
		return nil, err
	}
	// Re-fetch so the response includes the preloaded Category + Supplier.
	return s.repo.FindByID(product.ID)
}

func (s *productService) List(search string, lowStockOnly bool, page, perPage int) ([]models.Product, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.FindAll(search, lowStockOnly, offset, perPage)
}

// Export returns every matching product (no pagination) for a CSV download,
// honouring the same search and low-stock filters as the list.
func (s *productService) Export(search string, lowStockOnly bool) ([]models.Product, error) {
	products, _, err := s.repo.FindAll(search, lowStockOnly, 0, -1)
	return products, err
}

// ImportError reports why one CSV row could not be imported. Row is 1-based and
// counts data rows (the header is row 0, not reported).
type ImportError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

// ImportResult summarises a bulk import: how many products were created, how
// many rows failed, and why each one failed.
type ImportResult struct {
	Created int           `json:"created"`
	Failed  int           `json:"failed"`
	Errors  []ImportError `json:"errors"`
}

// ImportCSV bulk-creates products from a CSV stream. Columns are matched by
// header name (case-insensitive), so order does not matter; the expected
// headers mirror the export: Name, SKU, Category, Supplier, Price, Cost,
// Quantity, Reorder, Unit. Category and Supplier are resolved by name. Each row
// is validated independently: a bad row is skipped and reported, and the rest
// still import.
func (s *productService) ImportCSV(r io.Reader) (*ImportResult, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1 // tolerate ragged rows; we validate per cell
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not parse CSV: %w", err)
	}
	if len(records) < 2 {
		return nil, errors.New("CSV has no data rows")
	}

	// Map header name -> column index (lower-cased).
	col := make(map[string]int)
	for i, h := range records[0] {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}
	for _, required := range []string{"name", "sku", "category", "supplier"} {
		if _, ok := col[required]; !ok {
			return nil, fmt.Errorf("missing required column: %s", required)
		}
	}

	categories := s.nameIndex("category")
	suppliers := s.nameIndex("supplier")

	cell := func(row []string, key string) string {
		if idx, ok := col[key]; ok && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	result := &ImportResult{}
	for i, row := range records[1:] {
		rowNum := i + 1
		fail := func(msg string) {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Message: msg})
		}

		name := cell(row, "name")
		sku := cell(row, "sku")
		if name == "" || sku == "" {
			fail("name and sku are required")
			continue
		}

		categoryID, ok := categories[strings.ToLower(cell(row, "category"))]
		if !ok {
			fail(fmt.Sprintf("unknown category %q", cell(row, "category")))
			continue
		}
		supplierID, ok := suppliers[strings.ToLower(cell(row, "supplier"))]
		if !ok {
			fail(fmt.Sprintf("unknown supplier %q", cell(row, "supplier")))
			continue
		}

		price, err := parseAmount(cell(row, "price"))
		if err != nil {
			fail("invalid price")
			continue
		}
		cost, err := parseAmount(cell(row, "cost"))
		if err != nil {
			fail("invalid cost")
			continue
		}
		qty, err := parseCount(cell(row, "quantity"))
		if err != nil {
			fail("invalid quantity")
			continue
		}
		reorder := 10
		if v := cell(row, "reorder"); v != "" {
			reorder, err = parseCount(v)
			if err != nil {
				fail("invalid reorder level")
				continue
			}
		}

		taken, err := s.repo.ExistsBySKU(sku, 0)
		if err != nil {
			return nil, err
		}
		if taken {
			fail(fmt.Sprintf("SKU %q already exists", sku))
			continue
		}

		unit := cell(row, "unit")
		if unit == "" {
			unit = "pcs"
		}
		product := &models.Product{
			Name: name, SKU: sku, CategoryID: categoryID, SupplierID: supplierID,
			Price: price, CostPrice: cost, Quantity: qty, ReorderLevel: reorder,
			Unit: unit, IsActive: true,
		}
		if err := s.repo.Create(product); err != nil {
			fail("could not save product")
			continue
		}
		result.Created++
	}

	return result, nil
}

// nameIndex loads all categories or suppliers into a lower-cased name -> id map
// so each CSV row can be resolved without a query per row.
func (s *productService) nameIndex(kind string) map[string]uint {
	out := make(map[string]uint)
	if kind == "category" {
		cats, _, _ := s.categoryRepo.FindAll("", 0, 100000)
		for _, c := range cats {
			out[strings.ToLower(c.Name)] = c.ID
		}
	} else {
		sups, _, _ := s.supplierRepo.FindAll("", 0, 100000)
		for _, su := range sups {
			out[strings.ToLower(su.Name)] = su.ID
		}
	}
	return out
}

// parseAmount parses a money cell; blank means 0.
func parseAmount(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

// parseCount parses an integer cell; blank means 0.
func parseCount(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}

func (s *productService) Get(id uint) (*models.Product, error) {
	product, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return product, nil
}

// AdjustStock validates the request, then delegates the atomic change to the
// repository. change must be non-zero and reason must be present.
func (s *productService) AdjustStock(productID, userID uint, change int, reason, note string) (*models.StockAdjustment, error) {
	if change == 0 || reason == "" {
		return nil, ErrInvalidAdjustment
	}
	adj, err := s.repo.AdjustStock(productID, userID, change, reason, note)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return adj, nil
}

func (s *productService) StockHistory(productID uint, page, perPage int) ([]models.StockAdjustment, int64, error) {
	return s.repo.StockAdjustments(productID, (page-1)*perPage, perPage)
}

// Update changes product details but NOT Quantity — stock levels only move
// through the Stock In module, never via a plain product edit. We preserve
// existing.Quantity by simply not overwriting it.
func (s *productService) Update(id uint, data *models.Product) (*models.Product, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	if err := s.validateRefs(data.CategoryID, data.SupplierID); err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsBySKU(data.SKU, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProductSKUTaken
	}

	existing.Name = data.Name
	existing.SKU = data.SKU
	existing.Image = data.Image
	existing.CategoryID = data.CategoryID
	existing.SupplierID = data.SupplierID
	existing.Price = data.Price
	existing.CostPrice = data.CostPrice
	existing.Unit = data.Unit
	existing.IsActive = data.IsActive
	existing.ReorderLevel = data.ReorderLevel
	// Quantity intentionally NOT updated here.

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return s.repo.FindByID(existing.ID)
}

func (s *productService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}
