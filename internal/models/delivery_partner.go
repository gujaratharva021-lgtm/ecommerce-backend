package models

import "time"

// DeliveryPartner is a delivery person the admin can assign to orders.
// Kept simple (no login) — the admin manages assignment and status
// updates; the partner is contacted via phone/WhatsApp with order details.
type DeliveryPartner struct {
ID            uint      `gorm:"primaryKey" json:"id"`
Name          string    `gorm:"not null" json:"name"`
Phone         string    `gorm:"not null;uniqueIndex" json:"phone"`
VehicleNumber string    `json:"vehicle_number"`
IsActive      bool      `gorm:"default:true" json:"is_active"`
CreatedAt     time.Time `json:"created_at"`
UpdatedAt     time.Time `json:"updated_at"`
}

// DeliveryPartnerRequest is the body for POST/PUT /admin/delivery-partners
type DeliveryPartnerRequest struct {
Name          string `json:"name" binding:"required"`
Phone         string `json:"phone" binding:"required,len=10,numeric"`
VehicleNumber string `json:"vehicle_number"`
IsActive      *bool  `json:"is_active"`
}

// AssignDeliveryPartnerRequest is the body for PUT /admin/orders/:id/assign-delivery
type AssignDeliveryPartnerRequest struct {
DeliveryPartnerID uint `json:"delivery_partner_id" binding:"required"`
}
