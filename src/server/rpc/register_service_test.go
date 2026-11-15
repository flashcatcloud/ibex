package rpc

import (
	"net"
	"net/rpc"
	"testing"

	"github.com/ugorji/go/codec"
)

type testHuatuo struct{}

func (testHuatuo) Report(in string, out *string) error { *out = in; return nil }
func TestSharedListener(t *testing.T) {
	server, err := newRPCServer([]RegisterService{func(s *rpc.Server) error { return s.RegisterName("Huatuo", testHuatuo{}) }})
	if err != nil {
		t.Fatal(err)
	}
	a, b := net.Pipe()
	var h codec.MsgpackHandle
	go server.ServeCodec(codec.MsgpackSpecRpc.ServerCodec(a, &h))
	c := rpc.NewClientWithCodec(codec.MsgpackSpecRpc.ClientCodec(b, &h))
	defer c.Close()
	for method, want := range map[string]string{"Server.Ping": "pong", "Huatuo.Report": "dev-doris"} {
		var got string
		if err := c.Call(method, "dev-doris", &got); err != nil || got != want {
			t.Fatalf("%s: got %q, err %v", method, got, err)
		}
	}
}
