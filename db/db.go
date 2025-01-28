package db

import (
	"context"
	"harmony/clip"
	"harmony/common"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

type TestData struct {
	Score int `json:"score"`
}

func Setup() {
	common.Ctx = context.Background()
	clip.SetupClipboard()

	configPath := os.Getenv("FIREBASE_CONFIG")
	if configPath == "" {
		log.Fatal("FIREBASE_CONFIG env var not set")
	}

	dbc := &firebase.Config{DatabaseURL: os.Getenv("DB_URL")}

	opt := option.WithCredentialsFile(configPath)
	app, err := firebase.NewApp(common.Ctx, dbc, opt)
	if err != nil {
		log.Fatalf("[error] initializing app: %v", err)
	}

	c_, err := app.Database(common.Ctx)
	if err != nil {
		log.Fatalln("error in creating firebase DB client: ", err)
	}
	common.Client = c_
}

// func Insert() {
// 	ref := common.Client.NewRef("test/1")

// 	data := TestData{
// 		Score: 40,
// 	}

// 	if err := ref.Set(common.Ctx, data); err != nil {
// 		log.Fatal(err)
// 	}
// }
