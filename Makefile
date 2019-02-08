# Websocket utils

export GOPATH=$(HOME)/go
LDFLAGS="-s -X main.Build=`date -u +%Y%m%d%H%M%S` -X main.Version=0.4"
INSTALL=install

all: wsclient wsechoserver

wsclient: wsclient.go
	go build -ldflags $(LDFLAGS) -o wsclient wsclient.go

wsechoserver: wsechoserver.go
	go build -ldflags $(LDFLAGS) -o wsechoserver wsechoserver.go

build: wsclient wsechoserver

test:
	go test -v

install: all
	sudo $(INSTALL) wsclient /usr/local/bin
	sudo $(INSTALL) wsechoserver /usr/local/bin

# Remove temporary files
clean:
	rm wsclient wsechoserver

.PHONY: all wsclient wsechoserver install test clean
