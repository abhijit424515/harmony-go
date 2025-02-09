package main

import (
	"context"
	"harmony/backend/api"
	"harmony/backend/cache"
	"harmony/backend/common"
	"harmony/backend/db"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func checkEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("[error] loading .env file: %v", err)
	}

	for _, v := range []string{"PORT", "JWT_SK", "REDIS_HOST", "REDIS_PWD"} {
		if os.Getenv(v) == "" {
			log.Fatalf("[error] %s env var not set", v)
		}
	}
}

func main() {
	common.Ctx = context.Background()
	checkEnv()

	cache.Setup()
	db.Setup()
	api.Setup()
}
