package main

import (
	"fmt"
	"github.com/toolkits/pkg/net/tcpx"
	"github.com/toolkits/pkg/runner"
	"github.com/ulricqin/ibex/src/agentd"
	"github.com/ulricqin/ibex/src/server"
)

// VERSION go build -ldflags "-X main.VERSION=x.x.x"
var VERSION = "not specified"

func NewServerCmd() {
	printEnv()

	tcpx.WaitHosts()

	var opts []server.ServerOption
	opts = append(opts, server.SetVersion(VERSION))
	// parse config file

	server.Run(true, opts...)
}

func NewEdgeServerCmd() {
	printEnv()

	tcpx.WaitHosts()

	var opts []server.ServerOption
	opts = append(opts, server.SetVersion(VERSION))

	server.Run(false, opts...)
}

func NewAgentdCmd() {
	printEnv()
	var opts []agentd.AgentdOption
	opts = append(opts, agentd.SetVersion(VERSION))

	agentd.Run(opts...)
}

func printEnv() {
	runner.Init()
	fmt.Println("runner.cwd:", runner.Cwd)
	fmt.Println("runner.hostname:", runner.Hostname)
	fmt.Println("runner.fd_limits:", runner.FdLimits())
	fmt.Println("runner.vm_limits:", runner.VMLimits())
}
