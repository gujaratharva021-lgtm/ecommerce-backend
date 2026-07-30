package models

import "time"

// User is identified by phone number only — no email/password.
// Login happens via OTP sent to the phone (see OTP model + auth handlers).
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Phone     string    `gorm:"uniqueIndex;not null" json:"phone"`
	Role      string    `gorm:"default:customer" json:"role"` // customer / admin
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Request/response DTOs

// SendOTPRequest is the body for POST /auth/send-otp
// Phone is expected as a 10-digit Indian mobile number (no country code, no spaces).
type SendOTPRequest struct {
	Phone string `json:"phone" binding:"required,len=10,numeric"`
}

// VerifyOTPRequest is the body for POST /auth/verify-otp
type VerifyOTPRequest struct {
	Phone string `json:"phone" binding:"required,len=10,numeric"`
	OTP   string `json:"otp" binding:"required,len=6,numeric"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// UpdateProfileRequest is the body for PUT /auth/me. Phone is intentionally
// excluded — it's the login identity, so changing it goes through a
// separate OTP-verified flow, not a plain profile edit.
type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required"`
}
