package models

import "time"

// OrderHandover records the warehouse-to-delivery-partner handover for one
// order. Created when warehouse staff verifies and hands over a
// ready_for_dispatch order to the assigned delivery partner.
type OrderHandover struct {
ID                uint      `gorm:"primaryKey" json:"id"`
OrderID           uint      `gorm:"not null;uniqueIndex" json:"order_id"`
Order             Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
WarehouseID       uint      `gorm:"not null;index" json:"warehouse_id"`
WarehouseStaffID  uint      `gorm:"not null" json:"warehouse_staff_id"`
DeliveryPartnerID uint      `gorm:"not null;index" json:"delivery_partner_id"`
PackageCount      int       `gorm:"not null;default:1" json:"package_count"`
HandedOverAt      time.Time `json:"handed_over_at"`
CreatedAt         time.Time `json:"created_at"`
}
