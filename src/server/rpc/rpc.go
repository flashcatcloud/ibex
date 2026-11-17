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

// Service is one named receiver on the shared listener. Start sets Err
// synchronously so the caller can start only the jobs that use this receiver.
type Service struct {
	Name     string
	Receiver any
	Err      error
}

// Start binds the shared port when at least one receiver registered. The bool
// reports whether the built-in Server methods (ordinary Ibex and FlashAI) are ready.
func Start(listen string, additional ...*Service) (bool, error) {
	base := &Service{Name: "Server", Receiver: new(Server)}
	server, err := newRPCServer(base, additional...)
	if err != nil {
		return false, err
	}
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return false, fmt.Errorf("listen for ibex RPC on %s: %w", listen, err)
	}
	fmt.Println("rpc.listening:", listener.Addr())
	go serve(listener, server)
	return base.Err == nil, nil
}

func newRPCServer(base *Service, additional ...*Service) (*rpc.Server, error) {
	server := rpc.NewServer()
	registered := 0
	var lastErr error
	for _, service := range append([]*Service{base}, additional...) {
		if service == nil {
			lastErr = fmt.Errorf("nil ibex RPC service")
			logger.Errorf("%v", lastErr)
			continue
		}
		receiver := reflect.ValueOf(service.Receiver)
		if !receiver.IsValid() || (receiver.Kind() == reflect.Ptr && receiver.IsNil()) {
			service.Err = fmt.Errorf("nil receiver")
		} else {
			service.Err = server.RegisterName(service.Name, service.Receiver)
		}
		if service.Err != nil {
			lastErr = fmt.Errorf("register ibex RPC service %q: %w", service.Name, service.Err)
			logger.Errorf("%v", lastErr)
			continue
		}
		registered++
	}
	if registered == 0 {
		return nil, fmt.Errorf("no ibex RPC service registered: %w", lastErr)
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
