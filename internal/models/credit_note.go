package models

import "time"

// SubstitutionCreditNote documents a downward price correction issued against an
// already-generated Invoice - e.g. because a warehouse substitution
// (Bug#24) swapped in a cheaper product after the invoice was already
// issued. The original Invoice is never rewritten once generated (a GST
// tax invoice is a legal record); this is the compensating document GST
// rules expect for a post-issuance correction, paired with a matching
// sales-ledger adjustment entry (see GenerateCreditNoteForSubstitution).
type SubstitutionCreditNote struct {
ID               uint      `gorm:"primaryKey" json:"id"`
CreditNoteNumber string    `gorm:"uniqueIndex;not null" json:"credit_note_number"`
InvoiceID        uint      `gorm:"not null;index" json:"invoice_id"`
Invoice          Invoice   `gorm:"foreignKey:InvoiceID" json:"-"`
OrderID          uint      `gorm:"not null;index" json:"order_id"`
Reason           string    `json:"reason"`
TaxableAmount    float64   `json:"taxable_amount"`
CGSTAmount       float64   `json:"cgst_amount"`
SGSTAmount       float64   `json:"sgst_amount"`
IGSTAmount       float64   `json:"igst_amount"`
TotalAmount      float64   `json:"total_amount"`
CreatedAt        time.Time `json:"created_at"`
}
