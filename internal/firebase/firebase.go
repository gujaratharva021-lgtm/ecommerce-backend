package firebase

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var Client *messaging.Client

// InitFirebase loads the service account credentials and initializes the
// Firebase Cloud Messaging client used to send push notifications.
//
// It first checks the FIREBASE_CREDENTIALS_JSON env var (the full service
// account JSON as a string — used on Render, where we can't upload a file).
// If that's not set, it falls back to reading the JSON from credentialsPath
// (used for local development, where secrets/firebase-service-account.json
// exists on disk).
func InitFirebase(credentialsPath string) {
	var opt option.ClientOption

	if jsonCreds := os.Getenv("FIREBASE_CREDENTIALS_JSON"); jsonCreds != "" {
		opt = option.WithCredentialsJSON([]byte(jsonCreds))
	} else {
		opt = option.WithCredentialsFile(credentialsPath)
	}

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