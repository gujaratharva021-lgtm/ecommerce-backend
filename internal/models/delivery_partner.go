package models

import (
	"time"

	"gorm.io/gorm"
)

// DeliveryPartner is a delivery person the admin can create/manage.
// The partner logs in via phone + OTP (same flow as customers) and can
// then push their live location while out for delivery.
type DeliveryPartner struct {
ID                 uint       `gorm:"primaryKey" json:"id"`
Name               string     `gorm:"not null" json:"name"`
Phone              string     `gorm:"not null;uniqueIndex" json:"phone"`
VehicleNumber      string     `json:"vehicle_number"`
IsActive           bool       `gorm:"default:true" json:"is_active"`
CurrentLat         *float64   `json:"current_lat,omitempty"`
CurrentLng         *float64   `json:"current_lng,omitempty"`
LastLocationUpdate *time.Time `json:"last_location_update,omitempty"`
CreatedAt          time.Time  `json:"created_at"`
UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
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

// UpdateLocationRequest is the body for PUT /delivery/location (delivery partner only)
type UpdateLocationRequest struct {
Lat float64 `json:"lat" binding:"required"`
Lng float64 `json:"lng" binding:"required"`
}

