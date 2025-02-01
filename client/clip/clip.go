package clip

import (
	"bytes"
	"context"
	"fmt"
	"harmony/client/common"
	"harmony/client/notify"
	"log"
	"net/http"
	"sync"

	"golang.design/x/clipboard"
)

type BufType string

const (
	TextType  BufType = "text"
	ImageType BufType = "image"
)

func sendData(data []byte, t BufType) error {
	url := common.Host + "/clip/" + string(t)

	var ct string
	if t == ImageType {
		ct = "application/octet-stream"
	} else {
		ct = "text/plain"
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", ct)

	res, err := common.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func CopyToClipboard(data []byte, t BufType) {
	if t == TextType {
		clipboard.Write(clipboard.FmtText, data)
	} else {
		clipboard.Write(clipboard.FmtImage, data)
	}
}

func WatchText(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ch := clipboard.Watch(ctx, clipboard.FmtText)
	for data := range ch {
		err := sendData(data, TextType)
		if err != nil {
			log.Println("[error]", err)
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
		err := sendData(data, ImageType)
		if err != nil {
			log.Println("[error]", err)
			continue
		}

		x := "Image: " + fmt.Sprintf("%d", len(data)) + " bytes"
		notify.Notify(x)
	}
}

func Watch() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(common.Ctx)
	defer cancel()

	wg.Add(2)
	go WatchText(ctx, &wg)
	go WatchImage(ctx, &wg)
	wg.Wait()
}
