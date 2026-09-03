package storage

import (
	"context"
	"time"

	"github.com/ccfos/nightingale/v6/storage"
	"github.com/redis/go-redis/v9"
)

type Redis redis.Cmdable

var Cache Redis

const DEFAULT = time.Hour

func InitRedis(cfg storage.RedisConfig) (err error) {
	Cache, err = storage.NewRedis(cfg)
	if err != nil {
		return err
	}

	return IdInit()
}

func CacheMGet(ctx context.Context, keys []string) [][]byte {
	return storage.MGet(ctx, Cache, keys)
}

const IDINITIAL = 1 << 32

// IdInit 初始化发号器。必须用 SetNX：这个 id 是 edge 与 center 网络不通时给任务兜底用的，
// 要求跨重启不重复。无条件 Set 会让每次重启都从 IDINITIAL 重新发号，发出与历史任务相同的 id。
func IdInit() error {
	return Cache.SetNX(context.Background(), "id", IDINITIAL, 0).Err()
}

func IdGet() (int64, error) {
	return Cache.Incr(context.Background(), "id").Result()
}
