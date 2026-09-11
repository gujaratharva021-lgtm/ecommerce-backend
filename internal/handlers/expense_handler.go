package handlers

import (
"errors"
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

// ListExpenses godoc
// GET /api/v1/admin/finance/expenses?category=&warehouse_id=&from=&to=&page=&limit=
func ListExpenses(c *gin.Context) {
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

db := database.DB.Model(&models.Expense{}).Preload("Warehouse")
if category := c.Query("category"); category != "" {
db = db.Where("category = ?", category)
}
if warehouseID := c.Query("warehouse_id"); warehouseID != "" {
db = db.Where("warehouse_id = ?", warehouseID)
}
if from := c.Query("from"); from != "" {
if t, err := time.Parse("2006-01-02", from); err == nil {
db = db.Where("expense_date >= ?", t)
}
}
if to := c.Query("to"); to != "" {
if t, err := time.Parse("2006-01-02", to); err == nil {
db = db.Where("expense_date < ?", t.AddDate(0, 0, 1))
}
}

var total int64
db.Count(&total)

var expenses []models.Expense
offset := (page - 1) * limit
if err := db.Order("expense_date DESC").Offset(offset).Limit(limit).Find(&expenses).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch expenses"})
return
}

var totalAmount float64
db.Session(&gorm.Session{}).Select("COALESCE(SUM(amount),0)").Scan(&totalAmount)

c.JSON(http.StatusOK, gin.H{
"expenses":     expenses,
"page":         page,
"limit":        limit,
"total":        total,
"total_pages":  int((total + int64(limit) - 1) / int64(limit)),
"total_amount": totalAmount,
})
}

// CreateExpense godoc
// POST /api/v1/admin/finance/expenses
func CreateExpense(c *gin.Context) {
var req models.ExpenseRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if !models.ValidExpenseCategories[req.Category] {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
return
}
expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense_date, use YYYY-MM-DD"})
return
}

adminID := c.MustGet("user_id").(uint)
expense := models.Expense{
Amount:      req.Amount,
Category:    req.Category,
ExpenseDate: expenseDate,
WarehouseID: req.WarehouseID,
Note:        req.Note,
ReceiptURL:  req.ReceiptURL,
AddedByID:   adminID,
ApprovalStatus: "draft",
}
if err := database.DB.Create(&expense).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create expense"})
return
}

adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "create_expense", "expense", strconv.Itoa(int(expense.ID)), "created")
c.JSON(http.StatusCreated, expense)
}

// UpdateExpense godoc
// PUT /api/v1/admin/finance/expenses/:id
func UpdateExpense(c *gin.Context) {
id := c.Param("id")
var expense models.Expense
if err := database.DB.First(&expense, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
return
}

var req models.ExpenseRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if !models.ValidExpenseCategories[req.Category] {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
return
}
expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense_date, use YYYY-MM-DD"})
return
}

previousAmount := expense.Amount
previousStatus := expense.ApprovalStatus

// If an already-approved (but not yet paid) expense has its amount
// changed, the approval that was given no longer means anything - it
// was granted for the OLD amount. Silently keeping ApprovalStatus as
// "approved" would let PayExpense pay out the new amount with zero
// re-review, bypassing the entire point of maker-checker (Bug #07:
// approved expenses could be mutated post-approval with no re-check).
// Reset to "submitted" so a (different) admin must approve the new
// amount before it can be paid. A paid expense's amount can still be
// corrected (handled below via the ledger adjustment), since the money
// has already moved and blocking the correction would prevent fixing
// a genuine data-entry mistake after the fact.
if previousStatus == "approved" && req.Amount != previousAmount {
expense.ApprovalStatus = "submitted"
expense.ApprovedByID = nil
expense.ApprovedAt = nil
}

expense.Amount = req.Amount
expense.Category = req.Category
expense.ExpenseDate = expenseDate
expense.WarehouseID = req.WarehouseID
expense.Note = req.Note
expense.ReceiptURL = req.ReceiptURL

if err := database.DB.Save(&expense).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update expense"})
return
}

// If this expense was already paid (and therefore already posted to
// the general ledger), post an adjusting entry for the amount delta so
// the ledger doesn't go stale relative to the updated Expense record.
if previousStatus == "paid" && req.Amount != previousAmount {
if err := services.PostExpenseAdjustmentLedgerEntry(expense.ID, req.Amount-previousAmount); err != nil {
log.Printf("failed to post expense adjustment ledger entry for expense %d: %v", expense.ID, err)
}
}

adminID := c.MustGet("user_id").(uint)
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "update_expense", "expense", id, "updated")

c.JSON(http.StatusOK, expense)
}

// DeleteExpense godoc
// DELETE /api/v1/admin/finance/expenses/:id
// A paid expense has already been posted to the general ledger
// (PostExpenseLedgerEntry, called from PayExpense) - hard-deleting it
// would leave that ledger entry orphaned, pointing at an expense record
// that no longer exists, with no way to trace or reconcile it (Bug
// FINANCE-06/#06: violates the audit trail the same way an unvalidated
// vendor-bill void would). Same rule VoidVendorBill already enforces for
// paid bills: a paid financial record must have its effect reversed
// through a proper accounting entry first, never silently deleted out
// from under posted ledger activity.
func DeleteExpense(c *gin.Context) {
id := c.Param("id")

var expense models.Expense
if err := database.DB.First(&expense, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
return
}
if expense.ApprovalStatus == "paid" {
c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete a paid expense - it has already been posted to the general ledger. Reverse it with an adjusting entry instead."})
return
}

if err := database.DB.Delete(&expense).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete expense"})
return
}

adminID := c.MustGet("user_id").(uint)
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "delete_expense", "expense", id, "deleted")

c.JSON(http.StatusOK, gin.H{"success": true})
}

// SubmitExpense godoc
// POST /api/v1/admin/finance/expenses/:id/submit
func SubmitExpense(c *gin.Context) {
id := c.Param("id")
var expense models.Expense
if err := database.DB.First(&expense, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
return
}
if expense.ApprovalStatus != "draft" {
c.JSON(http.StatusBadRequest, gin.H{"error": "Only draft expenses can be submitted"})
return
}
expense.ApprovalStatus = "submitted"
if err := database.DB.Save(&expense).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit expense"})
return
}
adminID := c.MustGet("user_id").(uint)
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "submit_expense", "expense", id, "draft->submitted")
c.JSON(http.StatusOK, expense)
}

// ApproveExpense godoc
// POST /api/v1/admin/finance/expenses/:id/approve
// Maker-checker (12.13, 12.25): the approver must be a different admin
// than the one who created the expense.
func ApproveExpense(c *gin.Context) {
id := c.Param("id")
var expense models.Expense
if err := database.DB.First(&expense, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
return
}
if expense.ApprovalStatus != "submitted" {
c.JSON(http.StatusBadRequest, gin.H{"error": "Only submitted expenses can be approved"})
return
}
adminID := c.MustGet("user_id").(uint)
if adminID == expense.AddedByID {
c.JSON(http.StatusForbidden, gin.H{"error": "Maker-checker: the expense creator cannot approve their own expense"})
return
}
now := time.Now()
expense.ApprovalStatus = "approved"
expense.ApprovedByID = &adminID
expense.ApprovedAt = &now
if err := database.DB.Save(&expense).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve expense"})
return
}
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "approve_expense", "expense", id, "submitted->approved")
c.JSON(http.StatusOK, expense)
}

// RejectExpense godoc
// POST /api/v1/admin/finance/expenses/:id/reject
func RejectExpense(c *gin.Context) {
id := c.Param("id")
var expense models.Expense
if err := database.DB.First(&expense, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
return
}
if expense.ApprovalStatus != "submitted" {
c.JSON(http.StatusBadRequest, gin.H{"error": "Only submitted expenses can be rejected"})
return
}
var req models.ExpenseRejectRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
adminID := c.MustGet("user_id").(uint)
if adminID == expense.AddedByID {
c.JSON(http.StatusForbidden, gin.H{"error": "Maker-checker: the expense creator cannot reject their own expense"})
return
}
expense.ApprovalStatus = "rejected"
expense.RejectionReason = req.Reason
if err := database.DB.Save(&expense).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject expense"})
return
}
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "reject_expense", "expense", id, req.Reason)
c.JSON(http.StatusOK, expense)
}

// PayExpense godoc
// POST /api/v1/admin/finance/expenses/:id/pay
// Only an approved expense can be paid. This is the point where the
// expense is actually recorded to the ledger (moved out of CreateExpense,
// which now only creates a draft - see 12.13/12.25 maker-checker workflow).
var errNotApproved = errors.New("expense is not in approved status")

func PayExpense(c *gin.Context) {
id := c.Param("id")

// The status check, "approved" -> "paid" transition, and ledger posting
// must all be serialized under a single row lock (Bug #12, same class
// as FINANCE-09's vendor-bill fix and Bug #10's ApproveReturn fix).
// Without this, two concurrent PayExpense calls on the same expense can
// both read ApprovalStatus == "approved" before either commits, both
// flip it to "paid", and both attempt to post to the ledger -
// PostExpenseLedgerEntry's own existing-entry check is a plain read
// with no lock behind it, so it does not by itself prevent two
// concurrent callers from both passing that check before either's
// insert commits.
var expense models.Expense
txErr := database.DB.Transaction(func(tx *gorm.DB) error {
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&expense, id).Error; err != nil {
return err
}
if expense.ApprovalStatus != "approved" {
return errNotApproved
}
now := time.Now()
expense.ApprovalStatus = "paid"
expense.PaidAt = &now
return tx.Save(&expense).Error
})

if txErr != nil {
if txErr == errNotApproved {
c.JSON(http.StatusBadRequest, gin.H{"error": "Only approved expenses can be paid"})
return
}
if txErr == gorm.ErrRecordNotFound {
c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
return
}
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark expense paid"})
return
}

// Only ever reached once per expense, since the row lock above
// serializes concurrent callers and only the one that actually
// performed the approved->paid transition gets here.
if err := services.PostExpenseLedgerEntry(expense.ID); err != nil {
log.Printf("failed to post expense ledger entry for expense %d: %v", expense.ID, err)
}
adminID := c.MustGet("user_id").(uint)
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "pay_expense", "expense", id, "approved->paid")
c.JSON(http.StatusOK, expense)
}
