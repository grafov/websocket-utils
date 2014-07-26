Websocket utils
===============

Simple CLI utils for testing websocket (RFC4655) connections.
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

* go get github.com/gorilla/websocket
* go get github.com/kdar/factorlog
* go get github.com/peterh/liner

Then install and updat package with:

  go get github.com/grafov/websocket-utils

For go 1.3 just use `build -i` for autoinstall external packages during build.

Makefile supplied to use `make` instead direct usage of go compiler.

    make deps
    make build
