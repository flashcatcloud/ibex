package rpc

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"reflect"
	"time"

	"github.com/toolkits/pkg/logger"
	"github.com/ugorji/go/codec"
)

type Server int

// RegisterService adds an application-owned service to the existing Ibex listener.
type RegisterService func(*rpc.Server) error

// Start returns after all services are registered and the TCP address is bound.
func Start(listen string, additional ...RegisterService) error {
	server, err := newRPCServer(additional)
	if err != nil {
		return fmt.Errorf("register ibex RPC service: %w", err)
	}

	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return fmt.Errorf("listen for ibex RPC on %s: %w", listen, err)
	}

	fmt.Println("rpc.listening:", listener.Addr())
	go serve(listener, server)
	return nil
}

func newRPCServer(additional []RegisterService) (*rpc.Server, error) {
	server := rpc.NewServer()
	if err := server.Register(new(Server)); err != nil {
		return nil, err
	}
	for _, register := range additional {
		if err := register(server); err != nil {
			return nil, err
		}
	}
	return server, nil
}

func serve(l net.Listener, server *rpc.Server) {
	var mh codec.MsgpackHandle
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))

	duration := time.Duration(100) * time.Millisecond

	for {
		conn, err := l.Accept()
		if err != nil {
			logger.Warningf("listener accept error: %v", err)
			time.Sleep(duration)
			continue
		}

		var bufconn = struct {
			io.Closer
			*bufio.Reader
			*bufio.Writer
		}{conn, bufio.NewReader(conn), bufio.NewWriter(conn)}

		go server.ServeCodec(codec.MsgpackSpecRpc.ServerCodec(bufconn, &mh))
	}
}
