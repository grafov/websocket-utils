# Websocket utils

export GOPATH=$(HOME)/go
LDFLAGS="-s -X main.Build `date -u +%Y%m%d%H%M%S` -X main.Version 0.1.2"

all: wsclient wsechoserver

# To build with dependencies for go >= 1.3 use `go build -i` and skip step below
deps:
	go get github.com/gorilla/websocket
	go get github.com/kdar/factorlog
	go get github.com/peterh/liner

wsclient: wsclient.go
	go build -ldflags $(LDFLAGS) -o wsclient wsclient.go

wsechoserver: wsechoserver.go
	go build -ldflags $(LDFLAGS) -o wsechoserver wsechoserver.go

build: wsclient wsechoserver

test:
	go test -v

# Remove temporary files
clean:
	rm wsclient wsechoserver
