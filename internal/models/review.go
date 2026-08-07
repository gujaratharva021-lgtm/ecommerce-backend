package models

import "time"

// Review is a user's rating + comment for a product.
// Uniqueness of (UserID, ProductID) is enforced so a user can only
// review a given product once (they can update it via PUT instead).
type Review struct {
ID        uint      `gorm:"primaryKey" json:"id"`
UserID    uint      `gorm:"not null;index:idx_review_user_product,unique" json:"user_id"`
User      User      `gorm:"foreignKey:UserID" json:"-"`
ProductID uint      `gorm:"not null;index:idx_review_user_product,unique" json:"product_id"`
Rating    int       `gorm:"not null" json:"rating"` // 1-5
Comment   string    `json:"comment"`
CreatedAt time.Time `json:"created_at"`
UpdatedAt time.Time `json:"updated_at"`
}

// ReviewRequest is the body for POST/PUT /products/:id/reviews
type ReviewRequest struct {
Rating  int    `json:"rating" binding:"required,min=1,max=5"`
Comment string `json:"comment"`
}
