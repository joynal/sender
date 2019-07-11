package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/joho/godotenv"

	webpush "github.com/SherClockHolmes/webpush-go"
	"sender/core"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	// prepare configs
	dbUrl := os.Getenv("MONGODB_URL")
	dbName := os.Getenv("DB_NAME")
	// Db connection stuff
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx = context.WithValue(ctx, core.DbURL, dbUrl)
	db, err := core.ConfigDB(ctx, dbName)
	if err != nil {
		log.Fatalf("database configuration failed: %v", err)
	}

	fmt.Println("Connected to MongoDB!")

	client, err := pubsub.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		log.Fatal(err)
	}

	subscriberCol := db.Collection("notificationsubscribers")
	var subscriberData core.SubscriberPayload
	sub := client.Subscription("notification")
	cctx, _ := context.WithCancel(ctx)

	err = sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		msg.Ack()
		err = json.Unmarshal(msg.Data, &subscriberData)
		if err != nil {
			fmt.Println("json err:", err)
		}

		// Decode subscription
		s := &webpush.Subscription{}
		err = json.Unmarshal([]byte(subscriberData.PushEndpoint), s)

		if err != nil {
			fmt.Println("endpoint err:", err)
		}

		// Send Notification
		res, err := webpush.SendNotification([]byte(subscriberData.Data), s, &subscriberData.Options)

		fmt.Println("push response: ", res)

		// TODO: find the correct code for unsubscribe
		if err != nil {
			fmt.Println("webpush error:", err)
			_, err = subscriberCol.UpdateOne(
				ctx,
				bson.M{"_id": subscriberData.SubscriberID},
				bson.M{"$set": bson.M{"status": "unSubscribed"}})

			if err != nil {
				fmt.Println("db update err:", err)
			}
		}
	})

	if err != nil {
		fmt.Println("topic receive error: ", err)
	}
}

