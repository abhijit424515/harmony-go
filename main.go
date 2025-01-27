package main

import (
	"context"
	"harmony/clip"

	firebase "firebase.google.com/go/v4"
)

func setup() *firebase.App {
	clip.SetupClipboard()

	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		println("[error] initializing app: %v\n", err)
	}

	return app
}

func main() {
	setup()

	c := make(chan int)
	go func() {
		clip.Watch()
		c <- 1
	}()
	<-c
}
