package rpc

import (
	"net"
	"net/rpc"
	"strings"
	"testing"

	"github.com/flashcatcloud/ibex/src/types"
	"github.com/ugorji/go/codec"
)

type testHuatuo struct{}

func (testHuatuo) Report(in string, out *string) error { *out = in; return nil }

func TestSharedListener(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		serverFails, huatuoFails bool
	}{
		{name: "both registered"},
		{name: "Huatuo fails after Server", huatuoFails: true},
		{name: "Server fails before Huatuo", serverFails: true},
		{name: "both fail", serverFails: true, huatuoFails: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := &Service{Name: "Server", Receiver: new(Server)}
			huatuo := &Service{Name: "Huatuo", Receiver: testHuatuo{}}
			if tc.serverFails {
				base.Receiver = struct{}{} // net/rpc rejects receivers without exported methods.
			}
			if tc.huatuoFails {
				huatuo.Receiver = struct{}{}
			}
			s, err := newRPCServer(base, huatuo)
			if (base.Err != nil) != tc.serverFails || (huatuo.Err != nil) != tc.huatuoFails {
				t.Fatalf("unexpected registration errors: Server=%v, Huatuo=%v", base.Err, huatuo.Err)
			}
			if tc.serverFails && tc.huatuoFails {
				if err == nil || s != nil {
					t.Fatal("no usable services, expected startup failure")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			client := sharedClient(t, s)
			var response types.ReportResponse
			err = client.Call("Server.Report", types.ReportRequest{Ident: "test-agent"}, &response)
			checkMethod(t, "Server.Report", err, tc.serverFails)
			var got string
			err = client.Call("Huatuo.Report", "dev-doris", &got)
			checkMethod(t, "Huatuo.Report", err, tc.huatuoFails)
			if !tc.huatuoFails && got != "dev-doris" {
				t.Fatalf("Huatuo.Report returned %q", got)
			}
		})
	}
}

func TestDuplicateRegistrationPreservesOriginalReceiver(t *testing.T) {
	first := &Service{Name: "Server", Receiver: new(Server)}
	duplicate := &Service{Name: "Server", Receiver: testHuatuo{}}
	other := &Service{Name: "Huatuo", Receiver: testHuatuo{}}
	s, err := newRPCServer(first, duplicate, other)
	if err != nil || first.Err != nil || duplicate.Err == nil || other.Err != nil {
		t.Fatalf("startup=%v, original=%v, duplicate=%v, other=%v", err, first.Err, duplicate.Err, other.Err)
	}
	client := sharedClient(t, s)
	for method, want := range map[string]string{"Server.Ping": "pong", "Huatuo.Report": "dev-doris"} {
		var got string
		if err := client.Call(method, "dev-doris", &got); err != nil || got != want {
			t.Fatalf("%s: %q, %v", method, got, err)
		}
	}
}
func sharedClient(t *testing.T, s *rpc.Server) *rpc.Client {
	t.Helper()
	a, b := net.Pipe()
	var h codec.MsgpackHandle
	go s.ServeCodec(codec.MsgpackSpecRpc.ServerCodec(a, &h))
	c := rpc.NewClientWithCodec(codec.MsgpackSpecRpc.ClientCodec(b, &h))
	t.Cleanup(func() { c.Close() })
	return c
}
func checkMethod(t *testing.T, method string, err error, unavailable bool) {
	t.Helper()
	if unavailable {
		if err == nil || !strings.Contains(err.Error(), "can't find service") {
			t.Fatalf("%s should be unregistered, got %v", method, err)
		}
	} else if err != nil {
		t.Fatalf("%s: %v", method, err)
	}
}
