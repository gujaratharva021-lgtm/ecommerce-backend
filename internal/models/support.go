package models

import "time"

// SupportTicket represents a customer support request, optionally linked to
// an order. Status flows: open -> in_progress -> resolved -> closed.
type SupportTicket struct {
ID        uint      `gorm:"primaryKey" json:"id"`
UserID    uint      `gorm:"not null" json:"user_id"`
OrderID   *uint     `json:"order_id,omitempty"`
Subject   string    `gorm:"not null" json:"subject"`
Status    string    `gorm:"default:open" json:"status"` // open | in_progress | resolved | closed
Priority  string    `gorm:"default:normal" json:"priority"` // low | normal | high
CreatedAt time.Time `json:"created_at"`
UpdatedAt time.Time `json:"updated_at"`
}

// SupportMessage is a single message within a ticket's thread. SenderType
// distinguishes customer replies from admin/agent replies.
type SupportMessage struct {
ID         uint      `gorm:"primaryKey" json:"id"`
TicketID   uint      `gorm:"not null" json:"ticket_id"`
SenderID   uint      `gorm:"not null" json:"sender_id"`
SenderType string    `gorm:"not null" json:"sender_type"` // "customer" | "admin"
Message    string    `gorm:"not null" json:"message"`
CreatedAt  time.Time `json:"created_at"`
}

// CreateTicketRequest is the customer request body for POST /support/tickets.
type CreateTicketRequest struct {
OrderID *uint  `json:"order_id"`
Subject string `json:"subject" binding:"required"`
Message string `json:"message" binding:"required"`
}

// ReplyRequest is the request body for POST /support/tickets/:id/messages
// (used by both customer and admin endpoints).
type ReplyRequest struct {
Message string `json:"message" binding:"required"`
}
