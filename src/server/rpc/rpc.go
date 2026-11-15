package rpc

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"os"
	"reflect"
	"time"

	"github.com/toolkits/pkg/logger"
	"github.com/ugorji/go/codec"
)

type Server int

// RegisterService adds an application-owned service to the existing Ibex listener.
type RegisterService func(*rpc.Server) error

func Start(listen string, additional ...RegisterService) {
	go serve(listen, additional)
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

func serve(listen string, additional []RegisterService) {
	server, err := newRPCServer(additional)
	if err != nil {
		logger.Errorf("register ibex RPC service: %v", err)
		return
	}

	l, err := net.Listen("tcp", listen)
	if err != nil {
		fmt.Printf("fail to listen on: %s, error: %v\n", listen, err)
		os.Exit(1)
	}

	fmt.Println("rpc.listening:", listen)

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
