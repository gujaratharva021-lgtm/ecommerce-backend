package handlers

import (
"log"
"net/http"
"strconv"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
"gorm.io/gorm"
"gorm.io/gorm/clause"
)

// ListPayroll godoc
// GET /api/v1/admin/finance/payroll?staff_id=&status=&month=&year=&page=&limit=
func ListPayroll(c *gin.Context) {
page := 1
limit := 20
if p := c.Query("page"); p != "" {
if v, err := strconv.Atoi(p); err == nil && v > 0 {
page = v
}
}
if l := c.Query("limit"); l != "" {
if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
limit = v
}
}

// baseQuery carries only the filters that should scope the summary cards
// too - staff_id/month/year narrow down "which payroll records", so the
// pending/paid totals should respect them. status is deliberately NOT
// applied here: status is exactly what distinguishes the pending total
// from the paid total, so baking a status filter into this shared base
// would make one of the two summary cards always read zero.
baseQuery := database.DB.Model(&models.Payroll{})
if staffID := c.Query("staff_id"); staffID != "" {
baseQuery = baseQuery.Where("staff_id = ?", staffID)
}
if month := c.Query("month"); month != "" {
baseQuery = baseQuery.Where("month = ?", month)
}
if year := c.Query("year"); year != "" {
baseQuery = baseQuery.Where("year = ?", year)
}

db := baseQuery.Session(&gorm.Session{}).Preload("Staff")
if status := c.Query("status"); status != "" {
db = db.Where("status = ?", status)
}

var total int64
db.Count(&total)

var payrolls []models.Payroll
offset := (page - 1) * limit
if err := db.Order("year DESC, month DESC, created_at DESC").Offset(offset).Limit(limit).Find(&payrolls).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payroll"})
return
}

// Defect #10: these two totals now reuse baseQuery (staff_id/month/year),
// via a fresh Session() each time so GORM doesn't carry over conditions
// between the two calls - previously these ran against unfiltered
// database.DB, so filtering payroll by staff/month/year left the summary
// cards showing lifetime, company-wide numbers instead of the filtered
// view the admin was actually looking at.
var totalPending float64
baseQuery.Session(&gorm.Session{}).Where("status = ?", "pending").Select("COALESCE(SUM(amount),0)").Scan(&totalPending)
var totalPaid float64
baseQuery.Session(&gorm.Session{}).Where("status = ?", "paid").Select("COALESCE(SUM(amount),0)").Scan(&totalPaid)

c.JSON(http.StatusOK, gin.H{
"payroll":       payrolls,
"page":          page,
"limit":         limit,
"total":         total,
"total_pages":   int((total + int64(limit) - 1) / int64(limit)),
"total_pending": totalPending,
"total_paid":    totalPaid,
})
}

// CreatePayroll godoc
// POST /api/v1/admin/finance/payroll
func CreatePayroll(c *gin.Context) {
var req models.PayrollRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var staff models.WarehouseStaff
if err := database.DB.First(&staff, req.StaffID).Error; err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Staff member not found"})
return
}

status := req.Status
if status == "" {
status = "pending"
}
if !models.ValidPayrollStatuses[status] {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
return
}
if req.PaymentMethod != "" && !models.ValidPaymentMethods[req.PaymentMethod] {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment_method"})
return
}

payroll := models.Payroll{
StaffID:       req.StaffID,
Amount:        req.Amount,
Month:         req.Month,
Year:          req.Year,
Status:        status,
PaymentMethod: req.PaymentMethod,
Note:          req.Note,
}

adminID := c.MustGet("user_id").(uint)
if status == "paid" {
now := time.Now()
payroll.PaidByID = &adminID
payroll.PaidAt = &now
}

// Create + ledger post now share one transaction (Defect #12 consistency -
// PostPayrollLedgerEntry's signature changed to require an active tx).
if err := database.DB.Transaction(func(tx *gorm.DB) error {
if err := tx.Create(&payroll).Error; err != nil {
return err
}
// If this payroll record was created already marked "paid" (rather than
// going through the normal pending->paid transition in UpdatePayroll),
// it still needs a ledger entry - otherwise a payroll payment recorded
// this way would silently never reach the general ledger (Bug
// FINANCE-07).
if status == "paid" {
if err := services.PostPayrollLedgerEntry(tx, payroll.ID); err != nil {
log.Printf("failed to post payroll ledger entry for payroll %d: %v", payroll.ID, err)
}
}
return nil
}); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payroll record"})
return
}

adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "create_payroll", "payroll", strconv.Itoa(int(payroll.ID)), "created")

database.DB.Preload("Staff").First(&payroll, payroll.ID)
c.JSON(http.StatusCreated, payroll)
}

// UpdatePayroll godoc
// PUT /api/v1/admin/finance/payroll/:id
func UpdatePayroll(c *gin.Context) {
id := c.Param("id")

var req models.PayrollRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var staff models.WarehouseStaff
if err := database.DB.First(&staff, req.StaffID).Error; err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Staff member not found"})
return
}

status := req.Status
if status == "" {
status = "pending"
}
if !models.ValidPayrollStatuses[status] {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
return
}
if req.PaymentMethod != "" && !models.ValidPaymentMethods[req.PaymentMethod] {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment_method"})
return
}

adminID := c.MustGet("user_id").(uint)

// The entire read-modify-write cycle (lock, status-transition decision,
// save, ledger post) now happens inside one transaction with the payroll
// row locked for its duration (Defect #12) - without this, two concurrent
// UpdatePayroll requests could both read the same pre-update "pending"
// status, both decide status=="paid" && wasPending, and both post a
// separate salary ledger entry for the same payroll record.
var payroll models.Payroll
var notFound bool
txErr := database.DB.Transaction(func(tx *gorm.DB) error {
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&payroll, id).Error; err != nil {
notFound = true
return err
}

wasPending := payroll.Status != "paid"
wasPaid := payroll.Status == "paid"
previousAmount := payroll.Amount
payroll.StaffID = req.StaffID
payroll.Amount = req.Amount
payroll.Month = req.Month
payroll.Year = req.Year
payroll.Status = status
payroll.PaymentMethod = req.PaymentMethod
payroll.Note = req.Note

if status == "paid" && wasPending {
now := time.Now()
payroll.PaidByID = &adminID
payroll.PaidAt = &now
}

if err := tx.Save(&payroll).Error; err != nil {
return err
}

// Post to the general ledger exactly once, the moment this record
// actually becomes "paid" - PostPayrollLedgerEntry is itself
// idempotent (keyed on reference_type="payroll", reference_id) as a
// second safety net, but gating on wasPending here avoids even
// attempting redundant calls on every unrelated edit to an
// already-paid record (Bug FINANCE-07). Now runs inside this locked
// transaction rather than against database.DB directly (Defect #12).
if status == "paid" && wasPending {
if err := services.PostPayrollLedgerEntry(tx, payroll.ID); err != nil {
log.Printf("failed to post payroll ledger entry for payroll %d: %v", payroll.ID, err)
}
}

// If this payroll record was ALREADY paid (and therefore already
// posted to the general ledger) and its amount just changed, post a
// compensating adjustment entry so the ledger reflects the corrected
// figure instead of silently going stale (Bug #09). Also now inside
// this locked transaction (Defect #12).
if wasPaid && req.Amount != previousAmount {
if err := services.PostPayrollAdjustmentLedgerEntry(tx, payroll.ID, req.Amount-previousAmount); err != nil {
log.Printf("failed to post payroll adjustment ledger entry for payroll %d: %v", payroll.ID, err)
}
}

return nil
})

if txErr != nil {
if notFound {
c.JSON(http.StatusNotFound, gin.H{"error": "Payroll record not found"})
return
}
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payroll record"})
return
}

adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "update_payroll", "payroll", id, "updated")

database.DB.Preload("Staff").First(&payroll, payroll.ID)
c.JSON(http.StatusOK, payroll)
}

// DeletePayroll godoc
// DELETE /api/v1/admin/finance/payroll/:id
// A paid payroll record has already been posted to the general ledger
// (PostPayrollLedgerEntry, called when status transitions to "paid" in
// CreatePayroll/UpdatePayroll) - hard-deleting it would leave that
// salary ledger entry orphaned, pointing at a payroll record that no
// longer exists (Bug #08, same class as FINANCE-06's expense-delete
// guard). A pending (unpaid) payroll record has no ledger entry yet and
// remains safe to delete outright.
func DeletePayroll(c *gin.Context) {
id := c.Param("id")

var payroll models.Payroll
if err := database.DB.First(&payroll, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Payroll record not found"})
return
}
if payroll.Status == "paid" {
c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete a paid payroll record - it has already been posted to the general ledger."})
return
}

if err := database.DB.Delete(&payroll).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete payroll record"})
return
}

adminID := c.MustGet("user_id").(uint)
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "delete_payroll", "payroll", id, "deleted")

c.JSON(http.StatusOK, gin.H{"success": true})
}
