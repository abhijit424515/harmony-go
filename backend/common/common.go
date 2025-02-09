package common

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/redis/go-redis/v9"
)

const (
	Lifetime = 5 * time.Minute
)

var (
	Ctx context.Context
	Dbc *dynamodb.Client
	Rdb *redis.Client
)
