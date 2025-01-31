package main

import (
	"context"
	"fmt"
	"harmony/client/auth"
	"harmony/client/clip"
	"harmony/client/common"

	"golang.design/x/clipboard"
)

func setup() {
	common.Host = "http://localhost:6553"
	common.Ctx = context.TODO()

	err := clipboard.Init()
	if err != nil {
		panic(err)
	}
}

func main() {
	setup()

	err := auth.SetUserId()
	if err != nil {
		fmt.Println("[error]", err)
		return
	}

	c := make(chan int)
	go func() {
		clip.Watch()
		c <- 1
	}()
	<-c
}
