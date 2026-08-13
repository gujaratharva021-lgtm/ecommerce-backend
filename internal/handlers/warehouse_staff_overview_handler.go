package handlers

import (
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// StaffOverviewRow is one row of the warehouse-scoped staff list, combining
// the WarehouseStaff record with a live "what are they doing right now"
// snapshot derived from PickingTask/PackingTask, plus lifetime counts and
// last-seen timestamp from the audit log. Read-only - staff CRUD stays in
// the admin panel.
type StaffOverviewRow struct {
ID            uint       `json:"id"`
Name          string     `json:"name"`
Phone         string     `json:"phone"`
Role          string     `json:"role"`
IsActive      bool       `json:"is_active"`
CurrentTask   *string    `json:"current_task"`
OrdersHandled int64      `json:"orders_handled"`
LastActivity  *time.Time `json:"last_activity"`
}

// GetWarehouseStaffOverview godoc
// GET /api/v1/warehouse/staff (warehouse staff only)
// Read-only staff roster for the caller's own warehouse - current task,
// orders handled, and last activity are computed live, not stored fields.
func GetWarehouseStaffOverview(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)

var staffList []models.WarehouseStaff
if err := database.DB.Where("warehouse_id = ?", warehouseID).Order("name ASC").Find(&staffList).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load staff"})
return
}

rows := make([]StaffOverviewRow, 0, len(staffList))
for _, s := range staffList {
row := StaffOverviewRow{
ID:       s.ID,
Name:     s.Name,
Phone:    s.Phone,
Role:     s.Role,
IsActive: s.IsActive,
}

// Current task: an in-progress picking task first, else an in-progress packing task.
var pickTask models.PickingTask
if err := database.DB.Where("picker_id = ? AND warehouse_id = ? AND status = ?", s.ID, warehouseID, "in_progress").
First(&pickTask).Error; err == nil {
task := "Picking Order #" + itoa(pickTask.OrderID)
row.CurrentTask = &task
} else {
var packTask models.PackingTask
if err := database.DB.Where("packer_id = ? AND warehouse_id = ? AND status = ?", s.ID, warehouseID, "in_progress").
First(&packTask).Error; err == nil {
task := "Packing Order #" + itoa(packTask.OrderID)
row.CurrentTask = &task
}
}

// Orders handled: distinct orders where this staff completed picking or packing.
var pickCount int64
database.DB.Model(&models.PickingTask{}).
Where("picker_id = ? AND warehouse_id = ? AND status = ?", s.ID, warehouseID, "completed").
Count(&pickCount)
var packCount int64
database.DB.Model(&models.PackingTask{}).
Where("packer_id = ? AND warehouse_id = ? AND status = ?", s.ID, warehouseID, "completed").
Count(&packCount)
row.OrdersHandled = pickCount + packCount

// Last activity: most recent audit-log entry for this staff member.
var lastLog models.WarehouseAuditLog
if err := database.DB.Where("staff_id = ? AND warehouse_id = ?", s.ID, warehouseID).
Order("created_at DESC").First(&lastLog).Error; err == nil {
row.LastActivity = &lastLog.CreatedAt
}

rows = append(rows, row)
}

c.JSON(http.StatusOK, gin.H{"staff": rows})
}

func itoa(id uint) string {
if id == 0 {
return "0"
}
digits := []byte{}
for id > 0 {
digits = append([]byte{byte('0' + id%10)}, digits...)
id /= 10
}
return string(digits)
}
