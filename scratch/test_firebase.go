package main

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
)

func main() {
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: "your_firebase_project_id"})
	if err != nil {
		fmt.Printf("NewApp error: %v\n", err)
		return
	}
	
	client, err := app.Auth(ctx)
	if err != nil {
		fmt.Printf("Auth error: %v\n", err)
		return
	}
	
	fmt.Printf("Auth client initialized: %v\n", client != nil)
}
