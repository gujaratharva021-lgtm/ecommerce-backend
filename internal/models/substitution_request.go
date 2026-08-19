package models

import "time"

// Substitution request statuses
const (
	SubstitutionStatusPending  = "pending"
	SubstitutionStatusApproved = "approved"
	SubstitutionStatusRejected = "rejected"
)

// SubstitutionRequest is created by a picker/packer when the original
// product for an order item is unavailable or short, and they want to
// swap in a replacement product. A warehouse_manager (or store manager)
// then approves or rejects it. Mirrors the WarehouseException pattern.
type SubstitutionRequest struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	OrderID            uint       `gorm:"not null;index" json:"order_id"`
	Order              Order      `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	PickingTaskItemID  *uint      `gorm:"index" json:"picking_task_item_id,omitempty"`
	OriginalProductID  uint       `gorm:"not null" json:"original_product_id"`
	OriginalProduct    Product    `gorm:"foreignKey:OriginalProductID" json:"original_product,omitempty"`
	SubstituteProductID uint      `gorm:"not null" json:"substitute_product_id"`
	SubstituteProduct  Product    `gorm:"foreignKey:SubstituteProductID" json:"substitute_product,omitempty"`
	Quantity           int        `gorm:"not null;default:1" json:"quantity"`
	Reason             string     `json:"reason"`
	WarehouseID        uint       `gorm:"not null;index" json:"warehouse_id"`
	RequestedByID      uint       `gorm:"not null" json:"requested_by_id"` // WarehouseStaff.ID
	Status             string     `gorm:"default:pending;index" json:"status"`
	DecidedByID        *uint      `json:"decided_by_id,omitempty"`
	DecisionNote       string     `json:"decision_note,omitempty"`
	DecidedAt          *time.Time `json:"decided_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// CreateSubstitutionRequest is the body for POST /warehouse/substitutions
type CreateSubstitutionRequest struct {
	OrderID             uint   `json:"order_id" binding:"required"`
	PickingTaskItemID   *uint  `json:"picking_task_item_id"`
	OriginalProductID   uint   `json:"original_product_id" binding:"required"`
	SubstituteProductID uint   `json:"substitute_product_id" binding:"required"`
	Quantity            int    `json:"quantity" binding:"required,min=1"`
	Reason              string `json:"reason"`
}

// DecideSubstitutionRequest is the body for PUT /warehouse/substitutions/:id/approve
// and PUT /warehouse/substitutions/:id/reject
type DecideSubstitutionRequest struct {
	Note string `json:"note"`
}
