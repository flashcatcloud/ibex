package agentd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

type Agentd struct {
	ConfigFile string
	Version    string
}

type AgentdOption func(*Agentd)

func SetConfigFile(f string) AgentdOption {
	return func(s *Agentd) {
		s.ConfigFile = f
	}
}

func SetVersion(v string) AgentdOption {
	return func(s *Agentd) {
		s.Version = v
	}
}

// Run run agentd
func Run(opts ...AgentdOption) {
	code := 1
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	agentd := Agentd{
		ConfigFile: filepath.Join("etc", "agentd.conf"),
		Version:    "not specified",
	}

	for _, opt := range opts {
		opt(&agentd)
	}

	cleanFunc, err := agentd.initialize()
	if err != nil {
		fmt.Println("agentd init fail:", err)
		os.Exit(code)
	}

EXIT:
	for {
		sig := <-sc
		fmt.Println("received signal:", sig.String())
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			code = 0
			break EXIT
		case syscall.SIGHUP:
			// reload configuration?
		default:
			break EXIT
		}
	}

	cleanFunc()
	fmt.Println("agentd exited")
	os.Exit(code)
}

func (s Agentd) initialize() (func(), error) {
	fmt.Println("agentd init")
	return func() {}, nil
}
