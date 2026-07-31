package services

import (
"context"
"log"

"firebase.google.com/go/v4/messaging"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
fb "github.com/gujaratharva021-lgtm/ecommerce-backend/internal/firebase"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// SendPushToAll sends a push notification with the given title/body to
// every registered device token. Invalid/unregistered tokens are removed
// from the database automatically.
func SendPushToAll(title string, body string) {
if fb.Client == nil {
log.Println("Firebase not initialized, skipping push notification")
return
}

var tokens []models.DeviceToken
if err := database.DB.Find(&tokens).Error; err != nil {
log.Printf("Failed to load device tokens: %v", err)
return
}
if len(tokens) == 0 {
log.Println("No device tokens registered, skipping push notification")
return
}

ctx := context.Background()

for _, t := range tokens {
msg := &messaging.Message{
Notification: &messaging.Notification{
Title: title,
Body:  body,
},
Token: t.Token,
}

_, err := fb.Client.Send(ctx, msg)
if err != nil {
log.Printf("Push failed for token %s: %v", t.Token, err)
if messaging.IsRegistrationTokenNotRegistered(err) || messaging.IsInvalidArgument(err) {
database.DB.Delete(&t)
}
continue
}
}

log.Printf("Push notification sent to %d device(s): %s", len(tokens), title)
}

// SendPushToUser sends a push notification to all device tokens linked to
// a specific user (e.g. for order status updates). If the user has no
// linked tokens (never registered one while logged in), this is a no-op.
func SendPushToUser(userID uint, title string, body string) {
if fb.Client == nil {
log.Println("Firebase not initialized, skipping push notification")
return
}

var tokens []models.DeviceToken
if err := database.DB.Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
log.Printf("Failed to load device tokens for user %d: %v", userID, err)
return
}
if len(tokens) == 0 {
return
}

ctx := context.Background()

for _, t := range tokens {
msg := &messaging.Message{
Notification: &messaging.Notification{
Title: title,
Body:  body,
},
Token: t.Token,
}

_, err := fb.Client.Send(ctx, msg)
if err != nil {
log.Printf("Push failed for token %s: %v", t.Token, err)
if messaging.IsRegistrationTokenNotRegistered(err) || messaging.IsInvalidArgument(err) {
database.DB.Delete(&t)
}
continue
}
}

log.Printf("Push notification sent to user %d (%d device(s)): %s", userID, len(tokens), title)
}
