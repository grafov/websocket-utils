# Websocket utils

LDFLAGS="-s -X main.Build `date -u +%Y%m%d%H%M%S`"

all: wsclient wsechoserver

# Build with dependencies (for go >= 1.3)
wsclient: wsclient.go
	GOPATH=~/go go build -i -ldflags $(LDFLAGS) -o wsclient wsclient.go

# Build with dependencies (for go >= 1.3)
wsechoserver: wsechoserver.go
	GOPATH=~/go go build -i -ldflags $(LDFLAGS) -o wsechoserver wsechoserver.go

# Remove temporary files
clean:
	rm wsclient wsechoserver
