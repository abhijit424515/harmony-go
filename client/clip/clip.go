package clip

import (
	"bytes"
	"context"
	"fmt"
	"harmony/client/common"
	"harmony/client/notify"
	"io"
	"log"
	"net/http"
	"strconv"
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

	buf, _ := io.ReadAll(res.Body)
	ttl_ := string(buf)
	ttl, _ := strconv.ParseInt(ttl_, 10, 64)

	common.LatestTTL = ttl
	common.LatestBuffer = data

	return nil
}

func CopyToClipboard(t BufType, data []byte) {
	var x string
	if t == TextType {
		clipboard.Write(clipboard.FmtText, data)
		x = "[R] Text: " + fmt.Sprintf("%d", len(data)) + " bytes"
	} else {
		clipboard.Write(clipboard.FmtImage, data)
		x = "[R] Image: " + fmt.Sprintf("%d", len(data)) + " bytes"
	}
	notify.Notify(x)
}

func watchText(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ch := clipboard.Watch(ctx, clipboard.FmtText)
	for data := range ch {
		err := sendData(data, TextType)
		if err != nil {
			log.Println("[error]", err)
			continue
		}

		x := "[S] Text: " + fmt.Sprintf("%d", len(data)) + " bytes"
		notify.Notify(x)
	}
}

func watchImage(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ch := clipboard.Watch(ctx, clipboard.FmtImage)
	for data := range ch {
		err := sendData(data, ImageType)
		if err != nil {
			log.Println("[error]", err)
			continue
		}

		x := "[S] Image: " + fmt.Sprintf("%d", len(data)) + " bytes"
		notify.Notify(x)
	}
}

func Watch() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(common.Ctx)
	defer cancel()

	wg.Add(2)
	go watchText(ctx, &wg)
	go watchImage(ctx, &wg)
	wg.Wait()
}

func GetBuffer() error {
	url := common.Host + "/buffer"
	if common.LatestTTL != 0 {
		url += "?ttl=" + fmt.Sprintf("%d", common.LatestTTL)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	res, err := common.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		bt := TextType
		if res.Header.Get("Content-Type") == "application/octet-stream" {
			bt = ImageType
		}

		data, _ := io.ReadAll(res.Body)
		t := res.Header.Get("X-Buffer-TTL")
		ttl, _ := strconv.ParseInt(t, 10, 64)

		common.LatestTTL = ttl
		common.LatestBuffer = data
		CopyToClipboard(bt, data)

	} else if res.StatusCode != http.StatusNotModified && res.StatusCode != http.StatusNoContent {
		buf, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected response: %s", string(buf))
	}

	return nil
}
