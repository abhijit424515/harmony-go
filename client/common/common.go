package common

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

var (
	Ctx          context.Context
	Client       *http.Client
	Host         string
	LatestTTL    int64
	LatestBuffer []byte
)

type BufType string

const (
	TextType  BufType = "text"
	ImageType BufType = "image"
)

func ClearScreen() {
	fmt.Fprint(os.Stdout, "\033[H\033[2J")
}
