package handlers

import (
"net/http"
"strconv"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
)

// ---- Vendors ----

// ListVendors godoc
// GET /api/v1/admin/finance/vendors?is_active=&page=&limit=
func ListVendors(c *gin.Context) {
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

db := database.DB.Model(&models.Vendor{})
if isActive := c.Query("is_active"); isActive != "" {
db = db.Where("is_active = ?", isActive == "true")
}

var total int64
db.Count(&total)

var vendors []models.Vendor
offset := (page - 1) * limit
if err := db.Order("name ASC").Offset(offset).Limit(limit).Find(&vendors).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vendors"})
return
}

c.JSON(http.StatusOK, gin.H{
"vendors":     vendors,
"page":        page,
"limit":       limit,
"total":       total,
"total_pages": int((total + int64(limit) - 1) / int64(limit)),
})
}

// CreateVendor godoc
// POST /api/v1/admin/finance/vendors
func CreateVendor(c *gin.Context) {
var req models.VendorRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

isActive := true
if req.IsActive != nil {
isActive = *req.IsActive
}

vendor := models.Vendor{
Name:        req.Name,
ContactName: req.ContactName,
Phone:       req.Phone,
Email:       req.Email,
GSTIN:       req.GSTIN,
Address:     req.Address,
IsActive:    isActive,
}
if err := database.DB.Create(&vendor).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vendor"})
return
}

adminID := c.MustGet("user_id").(uint)
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "create_vendor", "vendor", strconv.Itoa(int(vendor.ID)), "created")

c.JSON(http.StatusCreated, vendor)
}

// UpdateVendor godoc
// PUT /api/v1/admin/finance/vendors/:id
func UpdateVendor(c *gin.Context) {
id := c.Param("id")
var vendor models.Vendor
if err := database.DB.First(&vendor, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Vendor not found"})
return
}

var req models.VendorRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

vendor.Name = req.Name
vendor.ContactName = req.ContactName
vendor.Phone = req.Phone
vendor.Email = req.Email
vendor.GSTIN = req.GSTIN
vendor.Address = req.Address
if req.IsActive != nil {
vendor.IsActive = *req.IsActive
}

if err := database.DB.Save(&vendor).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vendor"})
return
}

adminID := c.MustGet("user_id").(uint)
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "update_vendor", "vendor", id, "updated")

c.JSON(http.StatusOK, vendor)
}

// DeleteVendor godoc
// DELETE /api/v1/admin/finance/vendors/:id
func DeleteVendor(c *gin.Context) {
id := c.Param("id")
if err := database.DB.Delete(&models.Vendor{}, id).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vendor"})
return
}

adminID := c.MustGet("user_id").(uint)
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "delete_vendor", "vendor", id, "deleted")

c.JSON(http.StatusOK, gin.H{"success": true})
}

// ---- Vendor Bills (Accounts Payable) ----

type vendorBillResponse struct {
models.VendorBill
Status string `json:"status"`
}

func toVendorBillResponse(b models.VendorBill) vendorBillResponse {
return vendorBillResponse{VendorBill: b, Status: models.VendorBillStatus(b)}
}

// ListVendorBills godoc
// GET /api/v1/admin/finance/vendor-bills?vendor_id=&status=&page=&limit=
func ListVendorBills(c *gin.Context) {
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

db := database.DB.Model(&models.VendorBill{}).Preload("Vendor")
if vendorID := c.Query("vendor_id"); vendorID != "" {
db = db.Where("vendor_id = ?", vendorID)
}

var all []models.VendorBill
if err := db.Order("bill_date DESC").Find(&all).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vendor bills"})
return
}

// Status filter applied in-memory since it's a derived field, not a
// stored column - the dataset here is small enough (finance/admin
// tooling, not customer-facing) that this doesn't need to be pushed to SQL.
statusFilter := c.Query("status")
var filtered []models.VendorBill
for _, b := range all {
if statusFilter == "" || models.VendorBillStatus(b) == statusFilter {
filtered = append(filtered, b)
}
}

total := len(filtered)
start := (page - 1) * limit
end := start + limit
if start > total {
start = total
}
if end > total {
end = total
}
page_slice := filtered[start:end]

responses := make([]vendorBillResponse, len(page_slice))
for i, b := range page_slice {
responses[i] = toVendorBillResponse(b)
}

var totalOutstanding float64
for _, b := range all {
totalOutstanding += b.Amount - b.AmountPaid
}

c.JSON(http.StatusOK, gin.H{
"bills":             responses,
"page":              page,
"limit":             limit,
"total":             total,
"total_pages":       (total + limit - 1) / limit,
"total_outstanding": totalOutstanding,
})
}

// CreateVendorBill godoc
// POST /api/v1/admin/finance/vendor-bills
func CreateVendorBill(c *gin.Context) {
var req models.VendorBillRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var vendor models.Vendor
if err := database.DB.First(&vendor, req.VendorID).Error; err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Vendor not found"})
return
}

billDate, err := time.Parse("2006-01-02", req.BillDate)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bill_date, use YYYY-MM-DD"})
return
}

var dueDate *time.Time
if req.DueDate != "" {
d, err := time.Parse("2006-01-02", req.DueDate)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due_date, use YYYY-MM-DD"})
return
}
dueDate = &d
}

adminID := c.MustGet("user_id").(uint)
bill := models.VendorBill{
VendorID:    req.VendorID,
BillNumber:  req.BillNumber,
Amount:      req.Amount,
GSTAmount:   req.GSTAmount,
BillDate:    billDate,
DueDate:     dueDate,
Note:        req.Note,
CreatedByID: adminID,
}
if err := database.DB.Create(&bill).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "DEBUG: " + err.Error()})
return
}

adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "create_vendor_bill", "vendor_bill", strconv.Itoa(int(bill.ID)), "created")

database.DB.Preload("Vendor").First(&bill, bill.ID)
c.JSON(http.StatusCreated, toVendorBillResponse(bill))
}

// PayVendorBill godoc
// POST /api/v1/admin/finance/vendor-bills/:id/pay
// Records a (possibly partial) payment against a bill. Rejects overpayment
// rather than silently clamping, since a payment amount typed wrong should
// surface as an error, not get quietly capped and hide the mistake.
func PayVendorBill(c *gin.Context) {
id := c.Param("id")
var bill models.VendorBill
if err := database.DB.First(&bill, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Vendor bill not found"})
return
}

var req models.VendorBillPaymentRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

remaining := bill.Amount - bill.AmountPaid
if req.Amount > remaining {
c.JSON(http.StatusBadRequest, gin.H{"error": "Payment exceeds remaining balance of the bill"})
return
}

bill.AmountPaid += req.Amount
if err := database.DB.Save(&bill).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record payment"})
return
}

adminID := c.MustGet("user_id").(uint)
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "pay_vendor_bill", "vendor_bill", id, "paid")

database.DB.Preload("Vendor").First(&bill, bill.ID)
c.JSON(http.StatusOK, toVendorBillResponse(bill))
}

// DeleteVendorBill godoc
// DELETE /api/v1/admin/finance/vendor-bills/:id
func DeleteVendorBill(c *gin.Context) {
id := c.Param("id")
if err := database.DB.Delete(&models.VendorBill{}, id).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vendor bill"})
return
}

adminID := c.MustGet("user_id").(uint)
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "delete_vendor_bill", "vendor_bill", id, "deleted")

c.JSON(http.StatusOK, gin.H{"success": true})
}
