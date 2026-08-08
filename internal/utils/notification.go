package utils

import (
"fmt"
"log"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// SendNotification "sends" an SMS notification. For now this just logs to
// console and saves a record in the database. When a real SMS gateway is
// integrated, only the body of this function needs to change.
func SendNotification(phone string, message string, notifType string, orderID *uint) {
status := "sent"

// Simulate sending — replace this block with real SMS gateway call later.
log.Printf("[SMS -> %s] (%s) %s\n", phone, notifType, message)

record := models.Notification{
OrderID: orderID,
Phone:   phone,
Message: message,
Type:    notifType,
Status:  status,
}

if err := database.DB.Create(&record).Error; err != nil {
fmt.Println("Failed to save notification record:", err)
}
}
