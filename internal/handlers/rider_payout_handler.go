package handlers

import (
"errors"
"fmt"
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
"gorm.io/gorm"
"gorm.io/gorm/clause"
)

// ---- Rider COD Deposits (SRS 12.9) ----

// CreateRiderCODDeposit godoc
// POST /api/v1/admin/finance/rider-cod-deposits
func CreateRiderCODDeposit(c *gin.Context) {
var req models.RiderCODDepositRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
depositDate, err := time.Parse("2006-01-02", req.DepositDate)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deposit_date, use YYYY-MM-DD"})
return
}
adminID := c.MustGet("user_id").(uint)
deposit := models.RiderCODDeposit{
DeliveryPartnerID: req.DeliveryPartnerID,
Amount:            req.Amount,
DepositDate:       depositDate,
Status:            "pending",
Note:              req.Note,
CreatedByID:       adminID,
}
if err := database.DB.Create(&deposit).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record COD deposit"})
return
}
services.CreateDeliveryNotification(
deposit.DeliveryPartnerID,
"COD settlement verified",
fmt.Sprintf("Your cash deposit of Rs.%.2f has been verified.", deposit.Amount),
"cod_settlement_verified",
nil,
)
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "create_rider_cod_deposit", "rider_cod_deposit", "", "pending")
c.JSON(http.StatusCreated, deposit)
}

// ListRiderCODDeposits godoc
// GET /api/v1/admin/finance/rider-cod-deposits
func ListRiderCODDeposits(c *gin.Context) {
var deposits []models.RiderCODDeposit
db := database.DB.Order("created_at DESC")
if status := c.Query("status"); status != "" {
db = db.Where("status = ?", status)
}
if partnerID := c.Query("delivery_partner_id"); partnerID != "" {
db = db.Where("delivery_partner_id = ?", partnerID)
}
db.Find(&deposits)
c.JSON(http.StatusOK, gin.H{"rider_cod_deposits": deposits})
}

// VerifyRiderCODDeposit godoc
// POST /api/v1/admin/finance/rider-cod-deposits/:id/verify
var errDepositNotPending = errors.New("only pending deposits can be verified")

func VerifyRiderCODDeposit(c *gin.Context) {
id := c.Param("id")
adminID := c.MustGet("user_id").(uint)

// The status check, pending->verified transition, and ledger posting all
// run inside one locked transaction (Bug #14). This closes two related
// problems at once: (1) atomicity - if PostRiderCODDepositLedgerEntry
// fails, the whole transaction rolls back, so the deposit is never left
// stranded as "verified" with no matching ledger entry; and (2) the same
// concurrent-double-verify race already fixed elsewhere this session
// (Bug #10/#11/#12) - the row lock means two simultaneous verify calls
// on the same deposit can't both pass the pending check and both post to
// the ledger.
var deposit models.RiderCODDeposit
txErr := database.DB.Transaction(func(tx *gorm.DB) error {
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&deposit, id).Error; err != nil {
return err
}
if deposit.Status != "pending" {
return errDepositNotPending
}
now := time.Now()
deposit.Status = "verified"
deposit.VerifiedByID = &adminID
deposit.VerifiedAt = &now
if err := tx.Save(&deposit).Error; err != nil {
return err
}
return services.PostRiderCODDepositLedgerEntry(tx, deposit.ID)
})

if txErr != nil {
if txErr == errDepositNotPending {
c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending deposits can be verified"})
return
}
if txErr == gorm.ErrRecordNotFound {
c.JSON(http.StatusNotFound, gin.H{"error": "COD deposit not found"})
return
}
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify deposit: " + txErr.Error()})
return
}

adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "verify_rider_cod_deposit", "rider_cod_deposit", id, "verified")
c.JSON(http.StatusOK, deposit)
}

// ---- Rider Payouts (SRS 12.11) ----

const perDeliveryEarningRate = 30.0

// CreateRiderPayout godoc
// POST /api/v1/admin/finance/rider-payouts
// Computes the payout amount from delivered-order count in the period,
// same rate used by the Rider Payable report.
func CreateRiderPayout(c *gin.Context) {
var req models.RiderPayoutRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
periodFrom, err := time.Parse("2006-01-02", req.PeriodFrom)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period_from, use YYYY-MM-DD"})
return
}
periodTo, err := time.Parse("2006-01-02", req.PeriodTo)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period_to, use YYYY-MM-DD"})
return
}
periodToExclusive := periodTo.AddDate(0, 0, 1)

// Block duplicate payouts: if any payout already exists for this rider
// whose period overlaps the requested range, refuse rather than create a
// second one that would double-count and double-accrue the same deliveries.
var overlapCount int64
database.DB.Model(&models.RiderPayout{}).
Where("delivery_partner_id = ? AND period_from < ? AND period_to >= ?",
req.DeliveryPartnerID, periodToExclusive, periodFrom).
Count(&overlapCount)
if overlapCount > 0 {
c.JSON(http.StatusConflict, gin.H{"error": "A payout already exists for this rider covering an overlapping period"})
return
}

var deliveredCount int64
database.DB.Model(&models.Order{}).
Where("delivery_partner_id = ? AND status = ? AND updated_at >= ? AND updated_at < ?",
req.DeliveryPartnerID, "delivered", periodFrom, periodToExclusive).
Count(&deliveredCount)

adminID := c.MustGet("user_id").(uint)
payout := models.RiderPayout{
DeliveryPartnerID: req.DeliveryPartnerID,
PeriodFrom:        periodFrom,
PeriodTo:          periodTo,
DeliveredCount:    int(deliveredCount),
Amount:            float64(deliveredCount) * perDeliveryEarningRate,
Status:            "pending",
CreatedByID:       adminID,
}
if err := database.DB.Create(&payout).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rider payout"})
return
}
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "create_rider_payout", "rider_payout", "", "pending")
c.JSON(http.StatusCreated, payout)
}

// ListRiderPayouts godoc
// GET /api/v1/admin/finance/rider-payouts
func ListRiderPayouts(c *gin.Context) {
var payouts []models.RiderPayout
db := database.DB.Order("created_at DESC")
if status := c.Query("status"); status != "" {
db = db.Where("status = ?", status)
}
if partnerID := c.Query("delivery_partner_id"); partnerID != "" {
db = db.Where("delivery_partner_id = ?", partnerID)
}
db.Find(&payouts)
c.JSON(http.StatusOK, gin.H{"rider_payouts": payouts})
}

// ApproveRiderPayout godoc
// POST /api/v1/admin/finance/rider-payouts/:id/approve
// Accrues the payout to the ledger (Debit Rider Delivery Expense, Credit
// Rider Payable) - the amount is now formally recognized as owed.
func ApproveRiderPayout(c *gin.Context) {
id := c.Param("id")
adminID := c.MustGet("user_id").(uint)

// Defect #12: status check, pending->approved transition, and ledger
// accrual now run inside one locked transaction - without this, two
// concurrent approve requests on the same payout could both read
// Status == "pending" before either commits, both flip it to "approved",
// and both accrue the rider's owed amount to the ledger a second time.
var payout models.RiderPayout
var notFound bool
txErr := database.DB.Transaction(func(tx *gorm.DB) error {
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&payout, id).Error; err != nil {
notFound = true
return err
}
if payout.Status != "pending" {
return errors.New("Only pending payouts can be approved")
}
now := time.Now()
payout.Status = "approved"
payout.ApprovedByID = &adminID
payout.ApprovedAt = &now
if err := tx.Save(&payout).Error; err != nil {
return err
}
return services.PostRiderPayoutAccrualLedgerEntry(tx, payout.ID)
})

if txErr != nil {
if notFound {
c.JSON(http.StatusNotFound, gin.H{"error": "Rider payout not found"})
return
}
if txErr.Error() == "Only pending payouts can be approved" {
c.JSON(http.StatusBadRequest, gin.H{"error": txErr.Error()})
return
}
c.JSON(http.StatusInternalServerError, gin.H{"error": "Payout approval failed: " + txErr.Error()})
return
}

adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "approve_rider_payout", "rider_payout", id, "pending->approved")
c.JSON(http.StatusOK, payout)
}

// PayRiderPayout godoc
// POST /api/v1/admin/finance/rider-payouts/:id/pay
// Settles an approved payout (Debit Rider Payable, Credit Bank) - the
// actual money-out step.
func PayRiderPayout(c *gin.Context) {
id := c.Param("id")

// Defect #12: same locked-transaction fix as ApproveRiderPayout above,
// applied to the approved->paid transition and settlement ledger entry.
var payout models.RiderPayout
var notFound bool
txErr := database.DB.Transaction(func(tx *gorm.DB) error {
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&payout, id).Error; err != nil {
notFound = true
return err
}
if payout.Status != "approved" {
return errors.New("Only approved payouts can be paid")
}
now := time.Now()
payout.Status = "paid"
payout.PaidAt = &now
if err := tx.Save(&payout).Error; err != nil {
return err
}
return services.PostRiderPayoutSettlementLedgerEntry(tx, payout.ID)
})

if txErr != nil {
if notFound {
c.JSON(http.StatusNotFound, gin.H{"error": "Rider payout not found"})
return
}
if txErr.Error() == "Only approved payouts can be paid" {
c.JSON(http.StatusBadRequest, gin.H{"error": txErr.Error()})
return
}
c.JSON(http.StatusInternalServerError, gin.H{"error": "Payout payment failed: " + txErr.Error()})
return
}

adminID := c.MustGet("user_id").(uint)
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "pay_rider_payout", "rider_payout", id, "approved->paid")
c.JSON(http.StatusOK, payout)
}
