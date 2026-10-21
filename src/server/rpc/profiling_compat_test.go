package rpc

import (
	"github.com/flashcatcloud/ibex/src/types"
	"github.com/ugorji/go/codec"
	"testing"
)

func TestProfilingReceiptWireCompatibility(t *testing.T) {
	type oldResponse struct {
		Message     string
		AssignTasks []types.AssignTask
	}
	var data []byte
	h := &codec.MsgpackHandle{}
	current := types.ReportResponse{AssignTasks: []types.AssignTask{{Id: 42, Clock: 10, Action: "start"}}, ProfilingReceipts: []types.ProfilingReceipt{{Id: 41, Clock: 9, ProfilingTaskID: "parent"}}}
	if err := codec.NewEncoderBytes(&data, h).Encode(current); err != nil {
		t.Fatal(err)
	}
	var old oldResponse
	if err := codec.NewDecoderBytes(data, h).Decode(&old); err != nil {
		t.Fatal(err)
	}
	if len(old.AssignTasks) != 1 || old.AssignTasks[0].Id != 42 {
		t.Fatal("old Agent assignment changed")
	}
	data = nil
	if err := codec.NewEncoderBytes(&data, h).Encode(old); err != nil {
		t.Fatal(err)
	}
	var next types.ReportResponse
	if err := codec.NewDecoderBytes(data, h).Decode(&next); err != nil {
		t.Fatal(err)
	}
	if len(next.ProfilingReceipts) != 0 || len(next.AssignTasks) != 1 {
		t.Fatal("old server inferred a receipt")
	}
}
