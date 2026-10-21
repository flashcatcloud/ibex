package models

import (
	"github.com/flashcatcloud/ibex/src/types"
	"testing"
)

func TestProfilingReceiptRequiresMatchingPersistedResult(t *testing.T) {
	r := types.ReportTask{Status: "success", Stdout: "full output", Stderr: ""}
	for _, stored := range []*TaskHost{nil, {Status: "timeout"}, {Status: "success", Stdout: "truncated"}, {Status: "success", Stdout: "full output", Stderr: "different"}} {
		if profilingReceiptMatchesResult(stored, r) {
			t.Fatal("accepted missing or mismatched result")
		}
	}
	if !profilingReceiptMatchesResult(&TaskHost{Status: "success", Stdout: "full output"}, r) {
		t.Fatal("persisted result rejected")
	}
}
func TestProfilingReceiptRequiresOptIn(t *testing.T) {
	p := "cpt-123e4567-e89b-12d3-a456-426614174000"
	header := "# CP_RETENTION_V1 task=" + p + "\n"
	for _, meta := range []*TaskMeta{nil, {Script: header}, {SystemCaller: "flashai", Script: header}, {SystemCaller: "continuous-profiling", Script: "echo x\n" + header}} {
		if profilingReceiptMatchesMeta(meta, p) {
			t.Fatal("ordinary task accepted")
		}
	}
	if !profilingReceiptMatchesMeta(&TaskMeta{SystemCaller: "continuous-profiling", Script: "#!/bin/sh\n" + header}, p) {
		t.Fatal("valid metadata rejected")
	}
}
