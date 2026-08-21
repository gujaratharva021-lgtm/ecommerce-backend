package models

import "time"

// RiderCODDeposit records a delivery partner physically depositing cash
// they collected on COD orders into the company's bank (SRS 12.9). Once
// verified by an admin, this moves the recorded balance from Cash (1001,
// where COD sales are already booked) into Bank (1002).
type RiderCODDeposit struct {
ID                uint       `gorm:"primaryKey" json:"id"`
DeliveryPartnerID uint       `gorm:"not null;index" json:"delivery_partner_id"`
Amount            float64    `gorm:"not null" json:"amount"`
DepositDate       time.Time  `gorm:"not null" json:"deposit_date"`
Status            string     `gorm:"not null;default:pending;index" json:"status"` // pending/verified
Note              string     `json:"note,omitempty"`
VerifiedByID      *uint      `json:"verified_by_id,omitempty"`
VerifiedAt        *time.Time `json:"verified_at,omitempty"`
CreatedByID       uint       `gorm:"not null" json:"created_by_id"`
CreatedAt         time.Time  `json:"created_at"`
}

type RiderCODDepositRequest struct {
DeliveryPartnerID uint    `json:"delivery_partner_id" binding:"required"`
Amount            float64 `json:"amount" binding:"required,gt=0"`
DepositDate       string  `json:"deposit_date" binding:"required"`
Note              string  `json:"note"`
}

// RiderPayout is one settlement period's earnings for a delivery partner
// (SRS 12.11): accrued when approved (Debit Rider Delivery Expense, Credit
// Rider Payable), settled when paid (Debit Rider Payable, Credit Bank) -
// mirrors the Expense approval workflow's two-step ledger posting.
type RiderPayout struct {
ID                uint       `gorm:"primaryKey" json:"id"`
DeliveryPartnerID uint       `gorm:"not null;index" json:"delivery_partner_id"`
PeriodFrom        time.Time  `gorm:"not null" json:"period_from"`
PeriodTo          time.Time  `gorm:"not null" json:"period_to"`
DeliveredCount    int        `gorm:"not null" json:"delivered_count"`
Amount            float64    `gorm:"not null" json:"amount"`
Status            string     `gorm:"not null;default:pending;index" json:"status"` // pending/approved/paid
ApprovedByID      *uint      `json:"approved_by_id,omitempty"`
ApprovedAt        *time.Time `json:"approved_at,omitempty"`
PaidAt            *time.Time `json:"paid_at,omitempty"`
CreatedByID       uint       `gorm:"not null" json:"created_by_id"`
CreatedAt         time.Time  `json:"created_at"`
}

type RiderPayoutRequest struct {
DeliveryPartnerID uint   `json:"delivery_partner_id" binding:"required"`
PeriodFrom        string `json:"period_from" binding:"required"`
PeriodTo          string `json:"period_to" binding:"required"`
}
