package models

import "time"

type Order struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	UserID      uint        `gorm:"not null" json:"user_id"`
	User        User        `gorm:"foreignKey:UserID" json:"-"`
	TotalAmount float64     `gorm:"not null" json:"total_amount"`
	Status      string      `gorm:"default:pending" json:"status"` // pending/confirmed/shipped/delivered/cancelled
	Items       []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   uint      `gorm:"not null" json:"order_id"`
	ProductID uint      `gorm:"not null" json:"product_id"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	Price     float64   `gorm:"not null" json:"price"` // price at time of order
	CreatedAt time.Time `json:"created_at"`
}
