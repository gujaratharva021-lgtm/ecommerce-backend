package services

import (
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// LogWarehouseAction writes one WarehouseAuditLog row. Fire-and-forget by
// design (best-effort) - a logging failure should never block the actual
// warehouse action that already succeeded.
func LogWarehouseAction(warehouseID, staffID uint, staffName, action, entityType, entityID, before, after string) {
log := models.WarehouseAuditLog{
WarehouseID: warehouseID,
StaffID:     staffID,
StaffName:   staffName,
Action:      action,
EntityType:  entityType,
EntityID:    entityID,
BeforeValue: before,
AfterValue:  after,
}
database.DB.Create(&log)
}
