package services

import (
	"math"
	"testing"
)

// computeTotals is shared by purchases and sales, so a mistake here is a
// mistake on every invoice in the system. The discount clamp matters most:
// a discount larger than the subtotal must never produce a negative taxable
// amount, which would flip the sign of the VAT and of the grand total.
func TestComputeTotals(t *testing.T) {
	cases := []struct {
		name                                        string
		subtotal, discount, taxPercent              float64
		wantDiscount, wantTaxAmount, wantGrandTotal float64
	}{
		{
			name:     "standard rate on a plain invoice",
			subtotal: 1000, discount: 0, taxPercent: 19,
			wantDiscount: 0, wantTaxAmount: 190, wantGrandTotal: 1190,
		},
		{
			name:     "tax is charged on the net amount, not the gross",
			subtotal: 1000, discount: 100, taxPercent: 19,
			wantDiscount: 100, wantTaxAmount: 171, wantGrandTotal: 1071,
		},
		{
			name:     "reduced rate",
			subtotal: 200, discount: 0, taxPercent: 7,
			wantDiscount: 0, wantTaxAmount: 14, wantGrandTotal: 214,
		},
		{
			name:     "zero-rated line",
			subtotal: 500, discount: 50, taxPercent: 0,
			wantDiscount: 50, wantTaxAmount: 0, wantGrandTotal: 450,
		},
		{
			name:     "discount larger than the subtotal is clamped, not negative",
			subtotal: 100, discount: 250, taxPercent: 19,
			wantDiscount: 100, wantTaxAmount: 0, wantGrandTotal: 0,
		},
		{
			name:     "negative discount is treated as zero",
			subtotal: 100, discount: -40, taxPercent: 19,
			wantDiscount: 0, wantTaxAmount: 19, wantGrandTotal: 119,
		},
		{
			name:     "discount equal to the subtotal zeroes the invoice",
			subtotal: 750, discount: 750, taxPercent: 19,
			wantDiscount: 750, wantTaxAmount: 0, wantGrandTotal: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotDiscount, gotTax, gotTotal := computeTotals(tc.subtotal, tc.discount, tc.taxPercent)

			assertMoney(t, "discount", gotDiscount, tc.wantDiscount)
			assertMoney(t, "tax amount", gotTax, tc.wantTaxAmount)
			assertMoney(t, "grand total", gotTotal, tc.wantGrandTotal)

			if gotTotal < 0 {
				t.Errorf("grand total went negative: %v", gotTotal)
			}
		})
	}
}

// The grand total must always be the taxable base plus its tax, whatever the
// inputs. This catches a broken clamp that individual cases might miss.
func TestComputeTotalsStaysConsistent(t *testing.T) {
	subtotals := []float64{0, 0.99, 10, 1234.56}
	discounts := []float64{-5, 0, 0.5, 10, 5000}
	rates := []float64{0, 7, 19}

	for _, subtotal := range subtotals {
		for _, discount := range discounts {
			for _, rate := range rates {
				gotDiscount, gotTax, gotTotal := computeTotals(subtotal, discount, rate)

				if gotDiscount < 0 || gotDiscount > subtotal {
					t.Fatalf("discount %v out of range for subtotal %v", gotDiscount, subtotal)
				}
				taxable := subtotal - gotDiscount
				assertMoney(t, "grand total", gotTotal, taxable+gotTax)
			}
		}
	}
}

// Money comparisons need a tolerance: float64 arithmetic on decimal amounts is
// not exact. Half a cent is the same threshold the journal uses to decide
// whether an entry balances.
func assertMoney(t *testing.T, label string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.005 {
		t.Errorf("%s = %v, want %v", label, got, want)
	}
}
