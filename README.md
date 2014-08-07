Websocket utilities [![Is maintained?](http://stillmaintained.com/grafov/websocket-utils.png)](http://stillmaintained.com/grafov/websocket-utils)
===================

Utilities for debug websocket (RFC4655) connections.
These utilities will not dig deep and will not show you internals of the protocols (use Wireshark instead). But with them you just have simple way to establish connection with your websocket server and send test messages or connect to echo server for testing your client.
Currently are:

* wsclient — allows you to establish websocket connection with server and send text messages from command line.
* wsechoserver — just reply back messages it received.

Install
=======

Binary downloads [![Build Status](https://drone.io/github.com/grafov/websocket-utils/status.png)](https://drone.io/github.com/grafov/websocket-utils/latest)
----------------

Prebuilt binaries provided for Linux x86_64:

* https://drone.io/github.com/grafov/websocket-utils/files/wsechoserver
* https://drone.io/github.com/grafov/websocket-utils/files/wsclient

Download and place to $PATH.

Build from source
-----------------

You need go environment then install external packages:

    go get github.com/gorilla/websocket
    go get github.com/kdar/factorlog
    go get github.com/peterh/liner

Then install and updat package with:

    go get github.com/grafov/websocket-utils

For go 1.3 just use `build -i` for autoinstall external packages during build.

Makefile supplied to use `make` instead direct usage of go compiler.

    make deps
    make build

Echo server
===========

Just run it with `-verb` and/or `-debug` args. By default it listen on localhost:48084 but you may point it to another host:port with `-listen` parameter.

Client
======

After connect to a websocket server client waits for input. Any input will send to remote server. Except two special cases:

1. If you starts text with `!` followed by command(s) then external command(s) will executed and its output will send to server:

    `! ls *.txt; date`

2. If you starts text with `<` followed by filename then this file loads and its content send to server.

    `< sample.json`

Any other string sequences will send as is.
