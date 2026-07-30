package models

import "time"

// Notification stores a record of every SMS/message we attempted to send.
type Notification struct {
ID        uint      `json:"id" gorm:"primaryKey"`
OrderID   *uint     `json:"order_id,omitempty"`
Phone     string    `json:"phone"`
Message   string    `json:"message"`
Type      string    `json:"type"`
Status    string    `json:"status"`
CreatedAt time.Time `json:"created_at"`
}
