package main

import (
	"context"
	"fmt"
	"harmony/client/auth"
	"harmony/client/clip"
	"harmony/client/common"
	"log"
	"net/http"
	"net/http/cookiejar"
	"time"

	"golang.design/x/clipboard"
)

func setup() error {
	common.Host = "http://localhost:6554"
	common.Ctx = context.TODO()

	jar, _ := cookiejar.New(nil)
	common.Client = &http.Client{Jar: jar}

	logged_in, err := auth.CreateOrRestoreCookies()
	if err != nil {
		return err
	}

	err = clipboard.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize clipboard: %w", err)
	}

	if !logged_in {
		err = auth.SignIn()
		if err != nil {
			return fmt.Errorf("failed to sign in: %w", err)
		}
	} else {
		common.ClearScreen()
		fmt.Println("You are already signed in!")
	}

	return nil
}

func main() {
	err := setup()
	if err != nil {
		log.Fatal("[error]", err)
		return
	}

	go func() {
		for {
			err := clip.GetBuffer()
			if err != nil {
				log.Println("[error]", err)
				return
			}

			err = auth.SaveCookies()
			if err != nil {
				log.Println("[error]", err)
				return
			}

			time.Sleep(5 * time.Second)
		}
	}()

	c := make(chan int)
	go func() {
		clip.Watch()
		c <- 1
	}()
	<-c
}
