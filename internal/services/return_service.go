package services

import (
	"errors"
	"fmt"
	"time"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"

	"gorm.io/gorm"
)

var ErrReturnExceedsAvailable = errors.New("return quantity exceeds available quantity")

// ReturnItemInput is one requested return line.
type ReturnItemInput struct {
	ProductID uint
	Quantity  int
	UnitValue float64
}

// CreatePurchaseReturnInput / CreateSaleReturnInput are invoice-based: the
// return is made against a specific source invoice.
type CreatePurchaseReturnInput struct {
	PurchaseID uint
	UserID     uint
	Note       string
	Items      []ReturnItemInput
}

type CreateSaleReturnInput struct {
	SaleID uint
	UserID uint
	Note   string
	Items  []ReturnItemInput
}

// ReturnableItem is one row of a return-lookup: how much was on the invoice,
// how much was already returned, and how much is still returnable.
type ReturnableItem struct {
	ProductID       uint    `json:"product_id"`
	Name            string  `json:"name"`
	SKU             string  `json:"sku"`
	Image           string  `json:"image"`
	Ordered         int     `json:"ordered"`
	AlreadyReturned int     `json:"already_returned"`
	Available       int     `json:"available"`
	UnitValue       float64 `json:"unit_value"`
}

// ReturnLookup is the full picture the return form needs after searching an
// invoice number.
type ReturnLookup struct {
	SourceID     uint             `json:"source_id"`
	InvoiceNo    string           `json:"invoice_no"`
	Date         time.Time        `json:"date"`
	PartyName    string           `json:"party_name"`
	PartyPhone   string           `json:"party_phone"`
	PartyAddress string           `json:"party_address"`
	Items        []ReturnableItem `json:"items"`
}

type ReturnService interface {
	LookupPurchase(invoiceNo string) (*ReturnLookup, error)
	LookupSale(invoiceNo string) (*ReturnLookup, error)
	CreatePurchaseReturn(input CreatePurchaseReturnInput) (*models.PurchaseReturn, error)
	CreateSaleReturn(input CreateSaleReturnInput) (*models.SaleReturn, error)
	ListPurchaseReturns(page, perPage int) ([]models.PurchaseReturn, int64, error)
	ListSaleReturns(page, perPage int) ([]models.SaleReturn, int64, error)
}

type returnService struct {
	repo         repositories.ReturnRepository
	purchaseRepo repositories.PurchaseRepository
	saleRepo     repositories.SaleRepository
}

func NewReturnService(
	repo repositories.ReturnRepository,
	purchaseRepo repositories.PurchaseRepository,
	saleRepo repositories.SaleRepository,
) ReturnService {
	return &returnService{repo: repo, purchaseRepo: purchaseRepo, saleRepo: saleRepo}
}

// ---- Lookups ----

func (s *returnService) LookupPurchase(invoiceNo string) (*ReturnLookup, error) {
	p, err := s.purchaseRepo.FindByInvoiceNo(invoiceNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPurchaseNotFound
		}
		return nil, err
	}
	returned, err := s.repo.ReturnedQtyByPurchase(p.ID)
	if err != nil {
		return nil, err
	}

	out := &ReturnLookup{SourceID: p.ID, InvoiceNo: p.InvoiceNo, Date: p.CreatedAt}
	if p.Supplier != nil {
		out.PartyName, out.PartyPhone, out.PartyAddress = p.Supplier.Name, p.Supplier.Phone, p.Supplier.Address
	}
	for _, it := range p.Items {
		out.Items = append(out.Items, returnableItem(it.Product, it.ProductID, it.Quantity, returned[it.ProductID], it.UnitCost))
	}
	return out, nil
}

func (s *returnService) LookupSale(invoiceNo string) (*ReturnLookup, error) {
	sale, err := s.saleRepo.FindByInvoiceNo(invoiceNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSaleNotFound
		}
		return nil, err
	}
	returned, err := s.repo.ReturnedQtyBySale(sale.ID)
	if err != nil {
		return nil, err
	}

	out := &ReturnLookup{SourceID: sale.ID, InvoiceNo: sale.InvoiceNo, Date: sale.CreatedAt}
	if sale.Customer != nil {
		out.PartyName, out.PartyPhone, out.PartyAddress = sale.Customer.Name, sale.Customer.Phone, sale.Customer.Address
	}
	for _, it := range sale.Items {
		out.Items = append(out.Items, returnableItem(it.Product, it.ProductID, it.Quantity, returned[it.ProductID], it.UnitPrice))
	}
	return out, nil
}

func returnableItem(p *models.Product, productID uint, ordered, already int, unit float64) ReturnableItem {
	ri := ReturnableItem{
		ProductID:       productID,
		Ordered:         ordered,
		AlreadyReturned: already,
		Available:       ordered - already,
		UnitValue:       unit,
	}
	if p != nil {
		ri.Name, ri.SKU, ri.Image = p.Name, p.SKU, p.Image
	}
	return ri
}

// ---- Create ----

func (s *returnService) CreatePurchaseReturn(input CreatePurchaseReturnInput) (*models.PurchaseReturn, error) {
	p, err := s.purchaseRepo.FindByID(input.PurchaseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPurchaseNotFound
		}
		return nil, err
	}
	returned, err := s.repo.ReturnedQtyByPurchase(p.ID)
	if err != nil {
		return nil, err
	}
	available := availabilityMap(p.Items, returned)

	items := make([]models.PurchaseReturnItem, 0, len(input.Items))
	var total float64
	for _, in := range input.Items {
		if in.Quantity <= 0 {
			continue // skip untouched lines
		}
		if in.Quantity > available[in.ProductID] {
			return nil, ErrReturnExceedsAvailable
		}
		sub := float64(in.Quantity) * in.UnitValue
		total += sub
		items = append(items, models.PurchaseReturnItem{ProductID: in.ProductID, Quantity: in.Quantity, UnitCost: in.UnitValue, Subtotal: sub})
	}
	if len(items) == 0 {
		return nil, ErrNoItems
	}

	count, err := s.repo.CountPurchaseReturns()
	if err != nil {
		return nil, err
	}
	ret := &models.PurchaseReturn{
		InvoiceNo:   fmt.Sprintf("PRET-%06d", count+1),
		PurchaseID:  p.ID,
		SupplierID:  p.SupplierID,
		UserID:      input.UserID,
		TotalAmount: total,
		Note:        input.Note,
		Items:       items,
	}
	if err := s.repo.CreatePurchaseReturn(ret); err != nil {
		if errors.Is(err, repositories.ErrInsufficientStock) {
			return nil, ErrInsufficientStock
		}
		return nil, err
	}
	return ret, nil
}

func (s *returnService) CreateSaleReturn(input CreateSaleReturnInput) (*models.SaleReturn, error) {
	sale, err := s.saleRepo.FindByID(input.SaleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSaleNotFound
		}
		return nil, err
	}
	returned, err := s.repo.ReturnedQtyBySale(sale.ID)
	if err != nil {
		return nil, err
	}
	available := availabilityMap(saleItemsToReturnable(sale.Items), returned)

	items := make([]models.SaleReturnItem, 0, len(input.Items))
	var total float64
	for _, in := range input.Items {
		if in.Quantity <= 0 {
			continue
		}
		if in.Quantity > available[in.ProductID] {
			return nil, ErrReturnExceedsAvailable
		}
		sub := float64(in.Quantity) * in.UnitValue
		total += sub
		items = append(items, models.SaleReturnItem{ProductID: in.ProductID, Quantity: in.Quantity, UnitPrice: in.UnitValue, Subtotal: sub})
	}
	if len(items) == 0 {
		return nil, ErrNoItems
	}

	count, err := s.repo.CountSaleReturns()
	if err != nil {
		return nil, err
	}
	ret := &models.SaleReturn{
		InvoiceNo:   fmt.Sprintf("SRET-%06d", count+1),
		SaleID:      sale.ID,
		CustomerID:  sale.CustomerID,
		UserID:      input.UserID,
		TotalAmount: total,
		Note:        input.Note,
		Items:       items,
	}
	if err := s.repo.CreateSaleReturn(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// availabilityMap = ordered-per-product minus already-returned.
func availabilityMap(items []models.PurchaseItem, returned map[uint]int) map[uint]int {
	ordered := make(map[uint]int)
	for _, it := range items {
		ordered[it.ProductID] += it.Quantity
	}
	avail := make(map[uint]int, len(ordered))
	for pid, q := range ordered {
		avail[pid] = q - returned[pid]
	}
	return avail
}

// saleItemsToReturnable adapts sale items to the shared availability helper.
func saleItemsToReturnable(items []models.SaleItem) []models.PurchaseItem {
	out := make([]models.PurchaseItem, len(items))
	for i, it := range items {
		out[i] = models.PurchaseItem{ProductID: it.ProductID, Quantity: it.Quantity}
	}
	return out
}

func (s *returnService) ListPurchaseReturns(page, perPage int) ([]models.PurchaseReturn, int64, error) {
	return s.repo.FindPurchaseReturns((page-1)*perPage, perPage)
}

func (s *returnService) ListSaleReturns(page, perPage int) ([]models.SaleReturn, int64, error) {
	return s.repo.FindSaleReturns((page-1)*perPage, perPage)
}
