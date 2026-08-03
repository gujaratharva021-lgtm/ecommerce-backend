package utils

import (
"log"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// LogAudit records an admin action. Failures are logged but never block
// the request — an audit log write should never break the actual operation.
func LogAudit(adminID uint, adminPhone, action, entityType, entityID, details string) {
entry := models.AuditLog{
AdminID:    adminID,
AdminPhone: adminPhone,
Action:     action,
EntityType: entityType,
EntityID:   entityID,
Details:    details,
}
if err := database.DB.Create(&entry).Error; err != nil {
log.Printf("Failed to write audit log (%s %s#%s): %v", action, entityType, entityID, err)
}
}
