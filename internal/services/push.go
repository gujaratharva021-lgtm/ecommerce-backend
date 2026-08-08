package services

import (
"context"
"log"

"firebase.google.com/go/v4/messaging"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
fb "github.com/gujaratharva021-lgtm/ecommerce-backend/internal/firebase"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// fcmMulticastBatchSize is FCM's hard limit on tokens per multicast send.
const fcmMulticastBatchSize = 500

// SendPushToAll sends a push notification with the given title/body to
// every registered device token. Invalid/unregistered tokens are removed
// from the database automatically.
//
// L-11: this used to loop and call fb.Client.Send once per device (one
// network round-trip per token). It now batches tokens into groups of up
// to 500 (FCM's multicast limit) and uses SendEachForMulticast, cutting the
// number of calls to Firebase from O(devices) to O(devices/500).
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
sent := 0
for start := 0; start < len(tokens); start += fcmMulticastBatchSize {
end := start + fcmMulticastBatchSize
if end > len(tokens) {
end = len(tokens)
}
batch := tokens[start:end]

tokenStrings := make([]string, len(batch))
for i, t := range batch {
tokenStrings[i] = t.Token
}

msg := &messaging.MulticastMessage{
Notification: &messaging.Notification{
Title: title,
Body:  body,
},
Tokens: tokenStrings,
}

resp, err := fb.Client.SendEachForMulticast(ctx, msg)
if err != nil {
log.Printf("Multicast push failed for batch of %d: %v", len(batch), err)
continue
}
sent += resp.SuccessCount

for i, r := range resp.Responses {
if r.Success {
continue
}
log.Printf("Push failed for token %s: %v", batch[i].Token, r.Error)
if messaging.IsRegistrationTokenNotRegistered(r.Error) || messaging.IsInvalidArgument(r.Error) {
database.DB.Delete(&batch[i])
}
}
}
log.Printf("Push notification sent to %d/%d device(s): %s", sent, len(tokens), title)
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

// SendPushToPartner sends a push notification to all device tokens linked
// to a specific delivery partner (e.g. when a new order is assigned to
// them). If the partner has no linked tokens, this is a no-op.
func SendPushToPartner(partnerID uint, title string, body string) {
if fb.Client == nil {
log.Println("Firebase not initialized, skipping push notification")
return
}
var tokens []models.DeviceToken
if err := database.DB.Where("delivery_partner_id = ?", partnerID).Find(&tokens).Error; err != nil {
log.Printf("Failed to load device tokens for delivery partner %d: %v", partnerID, err)
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
log.Printf("Push notification sent to delivery partner %d (%d device(s)): %s", partnerID, len(tokens), title)
}
