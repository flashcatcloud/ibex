package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/toolkits/pkg/logger"
	"github.com/ulricqin/ibex/src/pkg/tlsx"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Address  string
	Username string
	Password string
	DB       int
	UseTLS   bool
	tlsx.ClientConfig
	RedisType        string
	MasterName       string
	SentinelUsername string
	SentinelPassword string
}

type Redis redis.Cmdable

var Cache Redis
var DEFAULT = time.Hour

func NewRedis(cfg RedisConfig) (Redis, error) {
	var redisClient Redis
	switch cfg.RedisType {
	case "standalone", "":
		redisOptions := &redis.Options{
			Addr:     cfg.Address,
			Username: cfg.Username,
			Password: cfg.Password,
			DB:       cfg.DB,
		}

		if cfg.UseTLS {
			tlsConfig, err := cfg.TLSConfig()
			if err != nil {
				fmt.Println("failed to init redis tls config:", err)
				os.Exit(1)
			}
			redisOptions.TLSConfig = tlsConfig
		}

		redisClient = redis.NewClient(redisOptions)

	case "cluster":
		redisOptions := &redis.ClusterOptions{
			Addrs:    strings.Split(cfg.Address, ","),
			Username: cfg.Username,
			Password: cfg.Password,
		}

		if cfg.UseTLS {
			tlsConfig, err := cfg.TLSConfig()
			if err != nil {
				fmt.Println("failed to init redis tls config:", err)
				os.Exit(1)
			}
			redisOptions.TLSConfig = tlsConfig
		}

		redisClient = redis.NewClusterClient(redisOptions)

	case "sentinel":
		redisOptions := &redis.FailoverOptions{
			MasterName:       cfg.MasterName,
			SentinelAddrs:    strings.Split(cfg.Address, ","),
			Username:         cfg.Username,
			Password:         cfg.Password,
			DB:               cfg.DB,
			SentinelUsername: cfg.SentinelUsername,
			SentinelPassword: cfg.SentinelPassword,
		}

		if cfg.UseTLS {
			tlsConfig, err := cfg.TLSConfig()
			if err != nil {
				fmt.Println("failed to init redis tls config:", err)
				os.Exit(1)
			}
			redisOptions.TLSConfig = tlsConfig
		}

		redisClient = redis.NewFailoverClient(redisOptions)

	default:
		fmt.Println("failed to init redis , redis type is illegal:", cfg.RedisType)
		os.Exit(1)
	}

	err := redisClient.Ping(context.Background()).Err()
	if err != nil {
		fmt.Println("failed to ping redis:", err)
		os.Exit(1)
	}
	return redisClient, nil
}

func CacheMGet(ctx context.Context, keys []string) [][]byte {
	var vals [][]byte
	pipe := Cache.Pipeline()
	for _, key := range keys {
		pipe.Get(ctx, key)
	}
	cmds, _ := pipe.Exec(ctx)

	for i, key := range keys {
		cmd := cmds[i]
		if errors.Is(cmd.Err(), redis.Nil) {
			continue
		}

		if cmd.Err() != nil {
			logger.Errorf("failed to get key: %s, err: %s", key, cmd.Err())
			continue
		}
		val := []byte(cmd.(*redis.StringCmd).Val())
		vals = append(vals, val)
	}

	return vals
}
