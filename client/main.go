package main

import (
	"harmony/client/clip"
)

func main() {
	clip.SetupClipboard()
	c := make(chan int)
	go func() {
		clip.Watch()
		c <- 1
	}()
	<-c
}
