package services

import (
	"errors"
	"testing"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"

	"gorm.io/gorm"
)

// stubProductRepo records the AdjustStock call and returns a scripted result.
// Only AdjustStock is exercised here; the rest come from the embedded interface.
type stubProductRepo struct {
	repositories.ProductRepository
	called    bool
	gotChange int
	retErr    error
}

func (s *stubProductRepo) AdjustStock(productID, userID uint, change int, reason, note string) (*models.StockAdjustment, error) {
	s.called = true
	s.gotChange = change
	if s.retErr != nil {
		return nil, s.retErr
	}
	return &models.StockAdjustment{ProductID: productID, UserID: userID, Change: change, Reason: reason, Note: note}, nil
}

// AdjustStock must reject nonsensical requests BEFORE touching the database: a
// zero change (nothing to do) or a missing reason (no audit trail). A valid
// request is delegated unchanged; a not-found product is surfaced as the
// service's own ErrProductNotFound.
func TestAdjustStockValidation(t *testing.T) {
	t.Run("zero change is rejected before hitting the repo", func(t *testing.T) {
		repo := &stubProductRepo{}
		svc := NewProductService(repo, nil, nil)

		_, err := svc.AdjustStock(1, 1, 0, "recount", "")
		if !errors.Is(err, ErrInvalidAdjustment) {
			t.Fatalf("got %v, want ErrInvalidAdjustment", err)
		}
		if repo.called {
			t.Error("repo.AdjustStock should not be called for an invalid request")
		}
	})

	t.Run("empty reason is rejected", func(t *testing.T) {
		repo := &stubProductRepo{}
		svc := NewProductService(repo, nil, nil)

		_, err := svc.AdjustStock(1, 1, -5, "", "")
		if !errors.Is(err, ErrInvalidAdjustment) {
			t.Fatalf("got %v, want ErrInvalidAdjustment", err)
		}
		if repo.called {
			t.Error("repo.AdjustStock should not be called for an invalid request")
		}
	})

	t.Run("valid request is delegated to the repo", func(t *testing.T) {
		repo := &stubProductRepo{}
		svc := NewProductService(repo, nil, nil)

		adj, err := svc.AdjustStock(7, 3, -2, "damaged", "broke in storage")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !repo.called || repo.gotChange != -2 {
			t.Errorf("repo got change=%d called=%v, want -2 called", repo.gotChange, repo.called)
		}
		if adj.Change != -2 || adj.Reason != "damaged" {
			t.Errorf("unexpected adjustment: %+v", adj)
		}
	})

	t.Run("a missing product maps to ErrProductNotFound", func(t *testing.T) {
		repo := &stubProductRepo{retErr: gorm.ErrRecordNotFound}
		svc := NewProductService(repo, nil, nil)

		_, err := svc.AdjustStock(99, 1, 3, "found", "")
		if !errors.Is(err, ErrProductNotFound) {
			t.Fatalf("got %v, want ErrProductNotFound", err)
		}
	})
}
