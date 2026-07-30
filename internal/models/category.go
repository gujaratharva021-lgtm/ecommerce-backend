package models

import "time"

type Category struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CategoryRequest is the body for POST/PUT /admin/categories (admin only).
type CategoryRequest struct {
	Name     string `json:"name" binding:"required"`
	ImageURL string `json:"image_url"`
}
