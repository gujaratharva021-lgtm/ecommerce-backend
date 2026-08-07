package models

import "time"

// Address is a saved delivery address belonging to a user.
// A user can have multiple addresses; IsDefault marks the one used
// automatically at checkout if the request doesn't specify address_id.
type Address struct {
ID        uint      `gorm:"primaryKey" json:"id"`
UserID    uint      `gorm:"not null;index" json:"user_id"`
User      User      `gorm:"foreignKey:UserID" json:"-"`
Label     string    `json:"label"` // e.g. "Home", "Office"
FullName  string    `gorm:"not null" json:"full_name"`
Phone     string    `gorm:"not null" json:"phone"`
Line1     string    `gorm:"not null" json:"line1"`
Line2     string    `json:"line2"`
City      string    `gorm:"not null" json:"city"`
State     string    `gorm:"not null" json:"state"`
Pincode   string    `gorm:"not null" json:"pincode"`
// Lat/Lng are optional GPS coordinates captured by the app (e.g. via
// a map picker or device location) when the address is saved. Used to
// auto-assign the nearest available delivery partner to an order.
Lat       *float64  `json:"lat,omitempty"`
Lng       *float64  `json:"lng,omitempty"`
IsDefault bool      `gorm:"default:false" json:"is_default"`
CreatedAt time.Time `json:"created_at"`
UpdatedAt time.Time `json:"updated_at"`
}

// AddressRequest is the body for POST /addresses and PUT /addresses/:id
type AddressRequest struct {
Label     string   `json:"label"`
FullName  string   `json:"full_name" binding:"required"`
Phone     string   `json:"phone" binding:"required,len=10,numeric"`
Line1     string   `json:"line1" binding:"required"`
Line2     string   `json:"line2"`
City      string   `json:"city" binding:"required"`
State     string   `json:"state" binding:"required"`
Pincode   string   `json:"pincode" binding:"required,len=6,numeric"`
Lat       *float64 `json:"lat"`
Lng       *float64 `json:"lng"`
IsDefault bool     `json:"is_default"`
}
