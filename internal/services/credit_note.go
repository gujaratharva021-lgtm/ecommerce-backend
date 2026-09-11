package services

import (
	"fmt"
	"time"

	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
	"gorm.io/gorm"
)

// GenerateCreditNoteForSubstitution issues a GST credit note for a price
// reduction caused by a warehouse item substitution (Bug#24: previously
// substitutions updated order totals but never amended an already-issued
// invoice or the sales ledger, leaving the tax invoice and ledger
// out of sync with what was actually delivered/charged).
//
// If no invoice has been generated for this order yet, this is a safe
// no-op - the still-unissued invoice will simply reflect the already-
// adjusted order.TotalAmount when it IS generated later, so no
// correction document is needed in that case.
//
// If an invoice already exists, it is left untouched (a GST tax invoice
// is treated as an immutable legal record once issued) and a separate
// credit note is created referencing it, together with a matching
// sales-ledger adjustment entry - debiting Sales Revenue and GST Payable
// for the reduction, crediting Customer Wallet Liability for the refund
// that CreditWallet already paid out to the customer.
//
// Must be called inside the same transaction as the wallet credit for
// the substitution, using the ORIGINAL product's ID (not the substitute)
// since priceDiff represents value being removed from what was
// originally invoiced.
func GenerateCreditNoteForSubstitution(tx *gorm.DB, orderID, originalProductID uint, priceDiff float64, reason, paymentMethod string) (*models.SubstitutionCreditNote, error) {
	if priceDiff <= 0 {
		return nil, nil
	}

	var invoice models.Invoice
	if err := tx.Where("order_id = ?", orderID).First(&invoice).Error; err != nil {
		// No invoice yet - nothing to amend. The order totals were already
		// adjusted by the caller, so a later invoice generation will pick
		// up the correct, already-substituted amount.
		return nil, nil
	}

	var product models.Product
	if err := tx.Select("gst_percent").First(&product, originalProductID).Error; err != nil {
		return nil, fmt.Errorf("failed to load original product for credit note tax calc: %w", err)
	}

	taxable := priceDiff / (1 + product.GSTPercent/100)
	gstAmount := priceDiff - taxable

	var cgst, sgst, igst float64
	if invoice.IsInterState {
		igst = gstAmount
	} else {
		cgst = gstAmount / 2
		sgst = gstAmount / 2
	}

	var seq int64
	if err := tx.Raw("SELECT nextval('substitution_credit_note_number_seq')").Scan(&seq).Error; err != nil {
		return nil, fmt.Errorf("failed to allocate credit note number: %w", err)
	}

	note := models.SubstitutionCreditNote{
		CreditNoteNumber: fmt.Sprintf("CN-%06d", seq),
		InvoiceID:        invoice.ID,
		OrderID:          orderID,
		Reason:           reason,
		TaxableAmount:    taxable,
		CGSTAmount:       cgst,
		SGSTAmount:       sgst,
		IGSTAmount:       igst,
		TotalAmount:      priceDiff,
		CreatedAt:        time.Now(),
	}
	if err := tx.Create(&note).Error; err != nil {
		return nil, fmt.Errorf("failed to create credit note: %w", err)
	}

	transactionRef := fmt.Sprintf("CREDIT-NOTE-%d-%d", orderID, time.Now().UnixNano())
	type line struct {
		code   string
		lType  string
		amount float64
		desc   string
	}
	lines := []line{
		{"4001", "debit", taxable, fmt.Sprintf("Credit note %s - revenue reduction for order #%d", note.CreditNoteNumber, orderID)},
		{"2002", "debit", gstAmount, fmt.Sprintf("Credit note %s - GST reversal for order #%d", note.CreditNoteNumber, orderID)},
	}
	// Only book a Customer Wallet Liability credit for online (prepaid)
	// orders - the customer already paid, so a cheaper substitute genuinely
	// owes them wallet credit. For COD orders nothing has been paid yet: the
	// order's own TotalAmount was already reduced by the caller, so the
	// customer is simply billed less at the door. Crediting Account 2005
	// here regardless of payment method invented a phantom liability on the
	// balance sheet for money the company was never actually holding
	// (Defect #06).
	if paymentMethod == "online" {
		lines = append(lines, line{"2005", "credit", priceDiff, fmt.Sprintf("Credit note %s - wallet credit for order #%d", note.CreditNoteNumber, orderID)})
	}

	for _, l := range lines {
		if l.amount <= 0 {
			continue
		}
		var account models.Account
		if err := tx.Where("code = ?", l.code).First(&account).Error; err != nil {
			return nil, fmt.Errorf("chart of accounts missing code %s: %w", l.code, err)
		}
		orderIDCopy := orderID
		entry := models.LedgerEntry{
			TransactionRef: transactionRef,
			AccountID:      account.ID,
			Type:           l.lType,
			Amount:         l.amount,
			Description:    l.desc,
			ReferenceType:  "credit_note",
			ReferenceID:    &orderIDCopy,
			EntryDate:      time.Now(),
		}
		if err := tx.Create(&entry).Error; err != nil {
			return nil, fmt.Errorf("failed to create %s ledger entry: %w", l.lType, err)
		}
	}

	return &note, nil
}
