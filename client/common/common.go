package common

import (
	"context"
	"net/http"
)

var (
	Ctx    context.Context
	Client *http.Client
	Host   string
)
