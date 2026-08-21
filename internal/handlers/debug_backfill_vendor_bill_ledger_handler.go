package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
)

// TEMPORARY DEBUG ENDPOINT - remove after use.
// Backfills PostVendorBillLedgerEntry for any vendor bill created before
// that posting existed (i.e. every bill created before this fix), so
// Vendor Payable finally has the missing credit leg to net against the
// debits already posted on payment.
// POST /api/v1/admin/debug/backfill-vendor-bill-ledger
func DebugBackfillVendorBillLedger(c *gin.Context) {
var bills []models.VendorBill
database.DB.Find(&bills)

type result struct {
BillID  uint   `json:"bill_id"`
BillNo  string `json:"bill_number"`
Posted  bool   `json:"posted"`
Skipped bool   `json:"skipped_already_posted"`
Error   string `json:"error,omitempty"`
}
var results []result

for _, bill := range bills {
var existing models.LedgerEntry
err := database.DB.Where("reference_type = ? AND reference_id = ?", "vendor_bill", bill.ID).First(&existing).Error
if err == nil {
results = append(results, result{BillID: bill.ID, BillNo: bill.BillNumber, Skipped: true})
continue
}
if postErr := services.PostVendorBillLedgerEntry(bill.ID); postErr != nil {
results = append(results, result{BillID: bill.ID, BillNo: bill.BillNumber, Error: postErr.Error()})
continue
}
results = append(results, result{BillID: bill.ID, BillNo: bill.BillNumber, Posted: true})
}

c.JSON(http.StatusOK, gin.H{"results": results})
}
