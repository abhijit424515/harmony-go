package common

import (
	"context"

	"firebase.google.com/go/v4/db"
)

var (
	Ctx    context.Context
	Client *db.Client
)
