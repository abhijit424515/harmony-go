package main

import (
	"harmony/clip"
	"harmony/db"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func checkEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	l := []string{"DB_URL", "FIREBASE_CONFIG"}

	for _, v := range l {
		if os.Getenv(v) == "" {
			log.Fatalf("%s env var not set", v)
		}
	}
}

func main() {
	checkEnv()
	db.Setup()

	c := make(chan int)
	go func() {
		clip.Watch()
		c <- 1
	}()
	<-c
}
