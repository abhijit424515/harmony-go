package common

import (
	"context"
	"database/sql"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	Lifetime = 5 * time.Minute
)

var (
	Ctx context.Context
	Rdb *redis.Client
	Db  *sql.DB
)
