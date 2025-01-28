package clip

import (
	"context"
	"fmt"
	"harmony/common"
	"harmony/notify"
	"sync"

	"golang.design/x/clipboard"
)

func SetupClipboard() {
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}
}

func WatchText(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ch := clipboard.Watch(ctx, clipboard.FmtText)
	for data := range ch {
		x := "Text: " + fmt.Sprintf("%d", len(data)) + " bytes"
		notify.Notify(x)
	}
}

func WatchImage(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ch := clipboard.Watch(ctx, clipboard.FmtImage)
	for data := range ch {
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
