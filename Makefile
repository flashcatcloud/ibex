.PHONY: start build

ifeq ($(shell uname),Darwin)
 PLATFORM="darwin"
else
 ifeq ($(OS),Windows_NT)
  PLATFORM="windows"
 else
  PLATFORM="linux"
 endif
endif

APP_BIN=ibex
APP_VER=0.5.0

all: build

build:
	CGO_ENABLED=0 go build -ldflags "-w -s -X main.VERSION=$(APP_VER)" -o $(APP_BIN) ./src/cmd

start_server:
	./$(APP_BIN) server -c ./etc/server.conf

start_agentd:
	./$(APP_BIN) agentd -c ./etc/agentd.conf

pack: build
	tar zcvf $(APP_BIN)-$(APP_VER).tar.gz etc sql $(APP_BIN)
