package models

import "time"

// Wishlist is a saved product for a user to buy later.
// Uniqueness of (UserID, ProductID) is enforced at the DB level so a
// product can't be added to the same user's wishlist twice.
type Wishlist struct {
ID        uint      `gorm:"primaryKey" json:"id"`
UserID    uint      `gorm:"not null;index:idx_wishlist_user_product,unique" json:"user_id"`
User      User      `gorm:"foreignKey:UserID" json:"-"`
ProductID uint      `gorm:"not null;index:idx_wishlist_user_product,unique" json:"product_id"`
Product   Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
CreatedAt time.Time `json:"created_at"`
}

// WishlistRequest is the body for POST /wishlist
type WishlistRequest struct {
ProductID uint `json:"product_id" binding:"required"`
}
