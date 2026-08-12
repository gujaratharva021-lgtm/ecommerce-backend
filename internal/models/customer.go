package models

import "time"

// CustomerListQuery is the query for GET /admin/customers
type CustomerListQuery struct {
Search  string `form:"search"` // matches name or phone
Status  string `form:"status"` // "blocked", "active", or empty for all
Sort    string `form:"sort"`   // newest, oldest, most_orders, highest_spent
Page    int    `form:"page,default=1"`
Limit   int    `form:"limit,default=20"`
}

// CustomerSummary is a row in the admin customer list - includes aggregated
// order stats that aren't on the User model itself.
type CustomerSummary struct {
ID          uint       `json:"id"`
Name        string     `json:"name"`
Phone       string     `json:"phone"`
IsBlocked   bool       `json:"is_blocked"`
CreatedAt   time.Time  `json:"created_at"`
TotalOrders int64      `json:"total_orders"`
TotalSpent  float64    `json:"total_spent"`
LastOrderAt *time.Time `json:"last_order_at"`
}

// CustomerListResponse wraps paginated customer results.
type CustomerListResponse struct {
Customers  []CustomerSummary `json:"customers"`
Page       int               `json:"page"`
Limit      int               `json:"limit"`
Total      int64             `json:"total"`
TotalPages int               `json:"total_pages"`
}

// CustomerDetail is the response for GET /admin/customers/:id
type CustomerDetail struct {
ID          uint               `json:"id"`
Name        string             `json:"name"`
Phone       string             `json:"phone"`
IsBlocked   bool               `json:"is_blocked"`
CreatedAt   time.Time          `json:"created_at"`
TotalOrders int64              `json:"total_orders"`
TotalSpent  float64            `json:"total_spent"`
Orders      []Order            `json:"orders"`
Addresses   []Address          `json:"addresses"`
Wallet      *Wallet            `json:"wallet,omitempty"`
Transactions []WalletTransaction `json:"wallet_transactions,omitempty"`
}
