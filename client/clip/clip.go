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
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"golang.design/x/clipboard"
)

const MaxBufferSize = 1024 * 1024 * 16 // bytes

func checkFileUrl(data []byte) ([]byte, bool) {
	filePath := strings.TrimPrefix(string(data), "file://")

	// use regex to check if the file path is valid
	rgx := regexp.MustCompile(`^/(?:[a-zA-Z0-9._\-\ \(\)\[\]]+/)*[a-zA-Z0-9._\-\ \(\)\[\]]*$`)
	if !rgx.MatchString(filePath) {
		return data, false
	}

	// Check if file exists
	_, err := os.Stat(filePath)
	if err != nil {
		log.Println("[error]", err)
		return data, false
	}

	// Read the file into a buffer
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		notify.NotifyText(fmt.Sprintf("🚫 Failed to read file: %s", err))
		return data, false
	}

	// checking the file type
	extension := filepath.Ext(filePath)
	if extension == ".png" || extension == ".jpg" || extension == ".jpeg" || extension == ".gif" {
		return fileData, true
	}
	return data, false
}

func sendData(data []byte, t common.BufType) error {
	if len(data) >= MaxBufferSize {
		notify.NotifyText("🚫 Copied data should be within 300KB.\nPlease try again.")
		return fmt.Errorf("buffer limit exceeded: %d bytes", len(data))
	}

	url := common.Host + "/clip/" + string(t)

	var ct string
	if t == common.ImageType {
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

func CopyToClipboard(t common.BufType, data []byte, ntf bool) {
	if t == common.TextType {
		clipboard.Write(clipboard.FmtText, data)
		if ntf {
			notify.NotifyText("⬇️ " + string(data))
		}
	} else {
		clipboard.Write(clipboard.FmtImage, data)
		if ntf {
			notify.NotifyImage("⬇️ Image", data)
		}
	}
}

func watchText(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ch := clipboard.Watch(ctx, clipboard.FmtText)
	for data := range ch {
		data, isFile := checkFileUrl(data)
		if isFile {
			err := sendData(data, common.ImageType)
			if err != nil {
				log.Println("[error]", err)
				continue
			}
			notify.NotifyImage("⬆️ Image", data)
		} else {
			err := sendData(data, common.TextType)
			if err != nil {
				log.Println("[error]", err)
				continue
			}
			notify.NotifyText("⬆️ " + string(data))
		}

	}
}

func watchImage(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ch := clipboard.Watch(ctx, clipboard.FmtImage)
	for data := range ch {
		err := sendData(data, common.ImageType)
		if err != nil {
			log.Println("[error]", err)
			continue
		}
		notify.NotifyImage("⬆️ Image", data)
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
		bt := common.TextType
		if res.Header.Get("Content-Type") == "application/octet-stream" {
			bt = common.ImageType
		}

		data, _ := io.ReadAll(res.Body)
		t := res.Header.Get("X-Buffer-TTL")
		ttl, _ := strconv.ParseInt(t, 10, 64)

		common.LatestTTL = ttl
		common.LatestBuffer = data
		CopyToClipboard(bt, data, true)

	} else if res.StatusCode != http.StatusNotModified && res.StatusCode != http.StatusNoContent {
		buf, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected response: %s", string(buf))
	}

	return nil
}
