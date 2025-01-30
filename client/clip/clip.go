package clip

import (
	"bytes"
	"context"
	"fmt"
	"harmony/client/notify"
	"net/http"
	"sync"

	"golang.design/x/clipboard"
)

func SetupClipboard() {
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}
}

const HOST = "http://localhost:6553"
const USERID = "fa3c07ea-7d10-46d6-8efc-6abc056f2139"

type BufType string

const (
	TextType  BufType = "text"
	ImageType BufType = "image"
)

func sendData(userid string, data []byte, t BufType) error {
	url := HOST + "/clip/" + string(t) + "?user_id=" + userid

	var ct string
	if t == ImageType {
		ct = "application/octet-stream"
	} else {
		ct = "text/plain"
	}

	res, err := http.Post(url, ct, bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func WatchText(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ch := clipboard.Watch(ctx, clipboard.FmtText)
	for data := range ch {
		err := sendData(USERID, data, TextType)
		if err != nil {
			fmt.Println("[error]", err)
			continue
		}

		x := "Text: " + fmt.Sprintf("%d", len(data)) + " bytes"
		notify.Notify(x)
	}
}

func WatchImage(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ch := clipboard.Watch(ctx, clipboard.FmtImage)
	for data := range ch {
		err := sendData(USERID, data, ImageType)
		if err != nil {
			fmt.Println("[error]", err)
			continue
		}

		x := "Image: " + fmt.Sprintf("%d", len(data)) + " bytes"
		notify.Notify(x)
	}
}

func Watch() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(2)
	go WatchText(ctx, &wg)
	go WatchImage(ctx, &wg)
	wg.Wait()
}
