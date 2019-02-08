Websocket utilities [![Is maintained?](http://stillmaintained.com/grafov/websocket-utils.png)](http://stillmaintained.com/grafov/websocket-utils)
===================

Utilities for debug websocket (RFC4655) connections.
These utilities will not dig deep and will not show you internals of the protocols (use Wireshark instead). But with them you just have simple way to establish connection with your websocket server and send test messages or connect to echo server for testing your client.
Currently are:

* wsclient — allows you to establish websocket connection with websocket server and send text messages from command line.
* wsechoserver — just reply back messages it received from websocket client.

Build [![Build Status](https://cloud.drone.io/api/badges/grafov/websocket-utils/status.svg)](https://cloud.drone.io/grafov/websocket-utils)
=======

The utilities use external libraries:

	github.com/gorilla/websocket
	github.com/kdar/factorlog
	github.com/peterh/liner

They are vendored with `dep` so you don't need to load anything else.

Makefile supplied for using [`make`](http://www.gnu.org/software/make/)
instead of direct call of go compiler.

	make build && make install

Echo server
===========

Run it with `-verb` and/or `-debug` args. By default it listen on
localhost:48084 but you may point it to another host:port with
`-listen` parameter.

Client
======

After connect to websocket server client waits for input. Any input
will send to remote server. Except two special cases:

1. If you start text with `!` followed by command(s) then external
   command(s) will executed and its output will send to a server:

	`! ls *.txt; date`

2. If you start text with `<` followed by filename then this file
   loads and its content send to server.

	`< sample.json`

Any other string sequences will send as is.

Try it for example as:

	wsclient -bind=echo.websocket.org
