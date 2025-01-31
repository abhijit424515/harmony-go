package main

import (
	"context"
	"harmony/backend/api"
	"harmony/backend/common"
	"harmony/backend/db"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func checkEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	for _, v := range []string{"PORT"} {
		if os.Getenv(v) == "" {
			log.Fatalf("[error] %s env var not set", v)
		}
	}
}

func main() {
	common.Ctx = context.Background()
	checkEnv()

	db.Setup()
	api.Setup()
}
