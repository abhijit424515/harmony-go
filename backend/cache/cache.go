package cache

import (
	"harmony/backend/common"
	"log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func Set(uid string, timestamp int64) {
	common.Rdb.Set(common.Ctx, uid, timestamp, common.Lifetime)
}

func Get(uid string) int64 {
	result, err := common.Rdb.Get(common.Ctx, uid).Result()
	if err != nil {
		return 0
	}

	t, _ := strconv.ParseInt(result, 10, 64)
	return t
}

func Setup() {
	common.Rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Username: "default",
		Password: os.Getenv("REDIS_PWD"),
		DB:       0,
	})

	_, err := common.Rdb.Ping(common.Ctx).Result()
	if err != nil {
		log.Fatal("[error] redis connection failed")
	}
}
