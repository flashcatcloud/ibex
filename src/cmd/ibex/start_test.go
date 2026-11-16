package ibex

import (
	"bytes"
	"context"
	"errors"
	"net"
	stdrpc "net/rpc"
	"runtime/pprof"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/ccfos/nightingale/v6/alert/aconf"
	"github.com/ccfos/nightingale/v6/conf"
	"github.com/flashcatcloud/ibex/src/server/config"
	"github.com/flashcatcloud/ibex/src/server/rpc"
	"github.com/flashcatcloud/ibex/src/storage"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// ServerStart initializes the Redis ID counter before bringing up RPC.
type startupRedis struct{ redis.Cmdable }

func (startupRedis) Set(context.Context, string, interface{}, time.Duration) *redis.StatusCmd {
	return redis.NewStatusResult("OK", nil)
}

func TestServerStartRPCFailure(t *testing.T) {
	oldConfig, oldCache, oldDB, oldPort := config.C, storage.Cache, storage.DB, HttpPort
	config.C = new(config.Config)
	t.Cleanup(func() { config.C, storage.Cache, storage.DB, HttpPort = oldConfig, oldCache, oldDB, oldPort })

	t.Run("duplicate registration", func(t *testing.T) {
		duplicate := func(s *stdrpc.Server) error { return s.Register(new(rpc.Server)) }
		err := ServerStart(true, nil, startupRedis{}, nil, aconf.HeartbeatConfig{}, nil, gin.New(), nil,
			conf.Ibex{RPCListen: "127.0.0.1:0"}, 0, duplicate)
		if err == nil || !strings.Contains(err.Error(), "service already defined: Server") {
			t.Fatalf("expected duplicate service error, got %v", err)
		}
		assertNoCenterTimers(t)
	})

	t.Run("missing port", func(t *testing.T) {
		err := ServerStart(true, nil, startupRedis{}, nil, aconf.HeartbeatConfig{}, nil, gin.New(), nil,
			conf.Ibex{RPCListen: "127.0.0.1"}, 0)
		var addrErr *net.AddrError
		if !errors.As(err, &addrErr) {
			t.Fatalf("expected an invalid listen address error, got %v", err)
		}
		assertNoCenterTimers(t)
	})

	t.Run("occupied port", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		defer listener.Close()
		err = ServerStart(true, nil, startupRedis{}, nil, aconf.HeartbeatConfig{}, nil, gin.New(), nil,
			conf.Ibex{RPCListen: listener.Addr().String()}, 0)
		if !errors.Is(err, syscall.EADDRINUSE) {
			t.Fatalf("expected address already in use, got %v", err)
		}
		assertNoCenterTimers(t)
	})
}

// These loops live for the lifetime of the process. If either was launched,
// its stack exists even when it is sleeping between database operations.
func assertNoCenterTimers(t *testing.T) {
	t.Helper()
	var stacks bytes.Buffer
	if err := pprof.Lookup("goroutine").WriteTo(&stacks, 2); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Heartbeat", "Schedule"} {
		if strings.Contains(stacks.String(), "github.com/flashcatcloud/ibex/src/server/timer."+name+"(") {
			t.Errorf("timer.%s started after RPC initialization failed", name)
		}
	}
}
