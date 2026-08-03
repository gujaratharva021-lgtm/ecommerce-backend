package firebase

import (
"context"
"log"

firebase "firebase.google.com/go/v4"
"firebase.google.com/go/v4/messaging"
"google.golang.org/api/option"
)

var Client *messaging.Client

// InitFirebase loads the service account credentials and initializes the
// Firebase Cloud Messaging client used to send push notifications.
func InitFirebase(credentialsPath string) {
opt := option.WithCredentialsFile(credentialsPath)
app, err := firebase.NewApp(context.Background(), nil, opt)
if err != nil {
log.Printf("Firebase init failed (push notifications disabled): %v", err)
return
}

client, err := app.Messaging(context.Background())
if err != nil {
log.Printf("Firebase messaging client failed (push notifications disabled): %v", err)
return
}

Client = client
log.Println("Firebase initialized successfully")
}
