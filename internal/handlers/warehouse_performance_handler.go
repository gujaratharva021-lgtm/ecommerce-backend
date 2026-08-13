package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// StaffPerformanceRow is one staff member's aggregated fulfillment stats.
type StaffPerformanceRow struct {
StaffID           uint    `json:"staff_id"`
StaffName         string  `json:"staff_name"`
OrdersPicked      int64   `json:"orders_picked"`
OrdersPacked      int64   `json:"orders_packed"`
AvgPickingMinutes float64 `json:"avg_picking_minutes"`
AvgPackingMinutes float64 `json:"avg_packing_minutes"`
TotalItemsPicked  int64   `json:"total_items_picked"`
CleanPicks        int64   `json:"clean_picks"`
AccuracyRate      float64 `json:"accuracy_rate"` // % of items picked clean (not unavailable/short)
ExceptionsCaused  int64   `json:"exceptions_caused"`
}

// buildStaffPerformance computes performance rows for the given staff IDs
// within one warehouse. Shared by GetWarehouseStaffPerformance (all staff)
// and GetMyPerformance (single staff) so the two stay consistent.
func buildStaffPerformance(warehouseID uint, staffIDs []uint) []StaffPerformanceRow {
var staff []models.WarehouseStaff
q := database.DB.Where("warehouse_id = ?", warehouseID)
if len(staffIDs) > 0 {
q = q.Where("id IN ?", staffIDs)
}
q.Find(&staff)

rows := make([]StaffPerformanceRow, 0, len(staff))
for _, s := range staff {
row := StaffPerformanceRow{StaffID: s.ID, StaffName: s.Name}

database.DB.Model(&models.PickingTask{}).
Where("picker_id = ? AND warehouse_id = ? AND status = ?", s.ID, warehouseID, "completed").
Count(&row.OrdersPicked)

database.DB.Model(&models.PackingTask{}).
Where("packer_id = ? AND warehouse_id = ? AND status = ?", s.ID, warehouseID, "completed").
Count(&row.OrdersPacked)

database.DB.Model(&models.PickingTask{}).
Where("picker_id = ? AND warehouse_id = ? AND status = ? AND started_at IS NOT NULL AND completed_at IS NOT NULL", s.ID, warehouseID, "completed").
Select("COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - started_at))), 0)").
Scan(&row.AvgPickingMinutes)
row.AvgPickingMinutes = row.AvgPickingMinutes / 60

database.DB.Model(&models.PackingTask{}).
Where("packer_id = ? AND warehouse_id = ? AND status = ? AND started_at IS NOT NULL AND completed_at IS NOT NULL", s.ID, warehouseID, "completed").
Select("COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - started_at))), 0)").
Scan(&row.AvgPackingMinutes)
row.AvgPackingMinutes = row.AvgPackingMinutes / 60

// Accuracy: of all picking_task_items decided by this staff's tasks,
// what fraction ended up "picked" (clean) vs unavailable/short.
database.DB.Model(&models.PickingTaskItem{}).
Joins("JOIN picking_tasks ON picking_tasks.id = picking_task_items.picking_task_id").
Where("picking_tasks.picker_id = ? AND picking_tasks.warehouse_id = ? AND picking_task_items.status != ?", s.ID, warehouseID, models.PickItemPending).
Count(&row.TotalItemsPicked)

database.DB.Model(&models.PickingTaskItem{}).
Joins("JOIN picking_tasks ON picking_tasks.id = picking_task_items.picking_task_id").
Where("picking_tasks.picker_id = ? AND picking_tasks.warehouse_id = ? AND picking_task_items.status = ?", s.ID, warehouseID, models.PickItemPicked).
Count(&row.CleanPicks)

if row.TotalItemsPicked > 0 {
row.AccuracyRate = float64(row.CleanPicks) / float64(row.TotalItemsPicked) * 100
}

database.DB.Model(&models.WarehouseException{}).
Where("staff_id = ? AND warehouse_id = ?", s.ID, warehouseID).
Count(&row.ExceptionsCaused)

rows = append(rows, row)
}
return rows
}

// GetWarehouseStaffPerformance godoc
// GET /api/v1/warehouse/staff/performance (warehouse staff only)
// Manager-style view: every active staff member's fulfillment stats for
// the caller's own warehouse.
func GetWarehouseStaffPerformance(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
rows := buildStaffPerformance(warehouseID, nil)
c.JSON(http.StatusOK, gin.H{"staff_performance": rows})
}

// GetMyPerformance godoc
// GET /api/v1/warehouse/staff/performance/me (warehouse staff only)
func GetMyPerformance(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
staffID := c.MustGet("staff_id").(uint)
rows := buildStaffPerformance(warehouseID, []uint{staffID})
if len(rows) == 0 {
c.JSON(http.StatusNotFound, gin.H{"error": "No performance data found"})
return
}
c.JSON(http.StatusOK, rows[0])
}
