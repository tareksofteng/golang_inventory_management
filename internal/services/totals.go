package services

// computeTotals turns a line-items subtotal plus a discount and a tax percent
// into the final money breakdown shared by purchases and sales.
//
//	taxable     = subtotal - discount        (discount clamped to [0, subtotal])
//	taxAmount   = taxable * taxPercent / 100  (VAT/tax charged on the net amount)
//	grandTotal  = taxable + taxAmount
//
// It returns the clamped discount, the tax amount, and the grand total.
func computeTotals(subtotal, discount, taxPercent float64) (float64, float64, float64) {
	if discount < 0 {
		discount = 0
	}
	if discount > subtotal {
		discount = subtotal
	}
	taxable := subtotal - discount
	taxAmount := taxable * taxPercent / 100
	grandTotal := taxable + taxAmount
	return discount, taxAmount, grandTotal
}
