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
