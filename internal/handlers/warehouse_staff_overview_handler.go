package handlers

import (
"net/http"
"strconv"
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
// GET /api/v1/warehouse/staff?page=&limit= (warehouse staff only)
// Read-only staff roster for the caller's own warehouse - current task,
// orders handled, and last activity are computed live, not stored fields.
// Batches the per-staff lookups (task/count/last-activity) into a handful
// of grouped queries instead of firing them once per staff row, so this
// stays O(1) queries regardless of headcount rather than O(n).
func GetWarehouseStaffOverview(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)

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

var total int64
database.DB.Model(&models.WarehouseStaff{}).Where("warehouse_id = ?", warehouseID).Count(&total)

var staffList []models.WarehouseStaff
offset := (page - 1) * limit
if err := database.DB.Where("warehouse_id = ?", warehouseID).Order("name ASC").
Offset(offset).Limit(limit).Find(&staffList).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load staff"})
return
}

staffIDs := make([]uint, len(staffList))
for i, s := range staffList {
staffIDs[i] = s.ID
}

// In-progress pick/pack task per staff (at most one of each, by design).
type activeTask struct {
StaffID uint
OrderID uint
}
var activePicks, activePacks []activeTask
if len(staffIDs) > 0 {
database.DB.Model(&models.PickingTask{}).
Select("picker_id as staff_id, order_id").
Where("picker_id IN ? AND warehouse_id = ? AND status = ?", staffIDs, warehouseID, "in_progress").
Scan(&activePicks)
database.DB.Model(&models.PackingTask{}).
Select("packer_id as staff_id, order_id").
Where("packer_id IN ? AND warehouse_id = ? AND status = ?", staffIDs, warehouseID, "in_progress").
Scan(&activePacks)
}
pickTaskByStaff := make(map[uint]uint, len(activePicks))
for _, t := range activePicks {
pickTaskByStaff[t.StaffID] = t.OrderID
}
packTaskByStaff := make(map[uint]uint, len(activePacks))
for _, t := range activePacks {
packTaskByStaff[t.StaffID] = t.OrderID
}

// Completed-order counts, grouped by staff in one query each instead of
// one query per staff member.
type countRow struct {
StaffID uint
Count   int64
}
var pickCounts, packCounts []countRow
if len(staffIDs) > 0 {
database.DB.Model(&models.PickingTask{}).
Select("picker_id as staff_id, count(*) as count").
Where("picker_id IN ? AND warehouse_id = ? AND status = ?", staffIDs, warehouseID, "completed").
Group("picker_id").Scan(&pickCounts)
database.DB.Model(&models.PackingTask{}).
Select("packer_id as staff_id, count(*) as count").
Where("packer_id IN ? AND warehouse_id = ? AND status = ?", staffIDs, warehouseID, "completed").
Group("packer_id").Scan(&packCounts)
}
pickCountByStaff := make(map[uint]int64, len(pickCounts))
for _, r := range pickCounts {
pickCountByStaff[r.StaffID] = r.Count
}
packCountByStaff := make(map[uint]int64, len(packCounts))
for _, r := range packCounts {
packCountByStaff[r.StaffID] = r.Count
}

// Most recent audit-log entry per staff member, grouped in one query.
type lastActivityRow struct {
StaffID  uint
LastSeen time.Time
}
var lastActivities []lastActivityRow
if len(staffIDs) > 0 {
database.DB.Model(&models.WarehouseAuditLog{}).
Select("staff_id, max(created_at) as last_seen").
Where("staff_id IN ? AND warehouse_id = ?", staffIDs, warehouseID).
Group("staff_id").Scan(&lastActivities)
}
lastActivityByStaff := make(map[uint]time.Time, len(lastActivities))
for _, r := range lastActivities {
lastActivityByStaff[r.StaffID] = r.LastSeen
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

if orderID, ok := pickTaskByStaff[s.ID]; ok {
task := "Picking Order #" + itoa(orderID)
row.CurrentTask = &task
} else if orderID, ok := packTaskByStaff[s.ID]; ok {
task := "Packing Order #" + itoa(orderID)
row.CurrentTask = &task
}

row.OrdersHandled = pickCountByStaff[s.ID] + packCountByStaff[s.ID]

if lastSeen, ok := lastActivityByStaff[s.ID]; ok {
lastSeenCopy := lastSeen
row.LastActivity = &lastSeenCopy
}

rows = append(rows, row)
}

c.JSON(http.StatusOK, gin.H{
"staff":       rows,
"page":        page,
"limit":       limit,
"total":       total,
"total_pages": int((total + int64(limit) - 1) / int64(limit)),
})
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
