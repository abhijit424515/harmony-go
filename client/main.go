package main

import (
	"context"
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
	common.Host = "http://localhost:6553"
	common.Ctx = context.TODO()

	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	common.Client = &http.Client{
		Jar: jar,
	}
	logged_in, err := auth.CreateOrRestoreCookies()
	if err != nil {
		return err
	}

	err = clipboard.Init()
	if err != nil {
		panic(err)
	}

	if !logged_in {
		err = auth.SignIn()
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	err := setup()
	if err != nil {
		log.Println("[error]", err)
		return
	}

	go func() {
		for {
			err := auth.SaveCookies()
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
