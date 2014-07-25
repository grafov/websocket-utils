// Simple websocket server echoed requests back to client.
package main

/*
 Copyleft 2014 Alexander I.Grafov aka Axel <grafov@gmail.com>

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU General Public License as published by
 the Free Software Foundation, either version 3 of the License, or
 (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU General Public License for more details.

 You should have received a copy of the GNU General Public License
 along with this program.  If not, see <http://www.gnu.org/licenses/>.

 असतो मा सद्गमय
 तमसो मा ज्योतिर्गमय
 मृत्योर्मामृतं गमय।
*/

import (
	"flag"
	"github.com/gorilla/websocket"
	"github.com/kdar/factorlog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	VERSION    = "1.0"
	WS_TIMEOUT = 12 * time.Second
)

var (
	Build      string // value set by compiler
	log        *factorlog.FactorLog
	listen     = flag.String("listen", "127.1:48084", "listen HTTP requests at addr:port")
	ping       = flag.Bool("ping", true, "ping clients and response to ping requests")
	wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}
)

// We need initialize logging before all other modules.
// It don't matter in regular program run but critical for unit tests.
func init() {
	var (
		logFmt string
		debug  = flag.Bool("debug", false, "debug output")
		verb   = flag.Bool("verb", false, "verbose output")
	)

	flag.Parse()

	if *debug {
		// brief format for debug
		logFmt = `%{Color "red" "ERROR"}%{Color "red" "FATAL"}%{Color "yellow" "WARN"}%{Color "green" "INFO"}%{Color "cyan" "DEBUG"}%{Color "blue" "TRACE"}[%{Date} %{Time}] [%{SEVERITY}:%{File}:%{Line}] %{Message}%{Color "reset"}`
	} else {
		// short format for production
		logFmt = `%{Color "red" "ERROR"}%{Color "red" "FATAL"}%{Color "yellow" "WARN"}%{Color "green" "INFO"}%{Color "cyan" "DEBUG"}%{Color "blue" "TRACE"}[%{Date} %{Time}] [%{SEVERITY}] %{Message}%{Color "reset"}`
	}

	log = factorlog.New(os.Stdout, factorlog.NewStdFormatter(logFmt))

	switch {
	case *debug:
		log.SetMinMaxSeverity(factorlog.TRACE, factorlog.PANIC)
	case *verb:
		log.SetMinMaxSeverity(factorlog.INFO, factorlog.PANIC)
	default:
		log.SetMinMaxSeverity(factorlog.WARN, factorlog.PANIC)
	}
}

func main() {
	log.Infof("Websocket echo server ver. %s (build %s)", VERSION, Build)
	log.Infof("© 2014, TightVideo Ltd.")

	wsEcho()

	terminate := make(chan os.Signal)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Infoln("interrupted by operator")
}

//
func wsEcho() {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {

		var (
			wscon *websocket.Conn
			err   error
		)

		if r.URL.Path != "/ws" {
			http.Error(w, "resource not found", http.StatusNotFound)
			log.Errorf("%d not found\n", http.StatusNotFound)
			return
		}
		if r.Method != "GET" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			log.Errorf("%d not found\n", http.StatusMethodNotAllowed)
			return
		}

		wscon, err = wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "upgrade to websocket protocol failed", 400)
			log.Fatalf("upgrade to websocket protocol failed with %s", err)
			return
		}

		go func(conn *websocket.Conn) {

			defer conn.Close()
			for {
				messageType, p, err := conn.ReadMessage()
				if err != nil {
					log.Errorf("read with %s", err)
					return
				}
				log.Tracef("server receive message (type %d): %s", messageType, p)
				if err = conn.WriteMessage(messageType, p); err != nil {
					log.Info("close")
					return
				}
			}
		}(wscon)
		go pongDispatcher(wscon)
		go pingDispatcher(wscon)
	})

	go func(mux *http.ServeMux) {
		if err := http.ListenAndServe(*listen, mux); err != nil {
			log.Fatalf("WS echo server failed at %s with %s", *listen, err)
		}
	}(mux)
	log.Infof("WS echo server listens at %s", *listen)

}

// Set timeout depends on the ping
func pongDispatcher(s *websocket.Conn) {
	// s.SetReadDeadline(time.Now().Add(WS_TIMEOUT))
	// s.SetPongHandler(func(arg string) error {
	// 	s.SetReadDeadline(time.Now().Add(WS_TIMEOUT))
	// 	return nil
	// })
}

func pingDispatcher(s *websocket.Conn) {
	for {
		if s == nil {
			log.Info("can't ping as connection closed")
			return
		}
		if err := s.WriteControl(websocket.PingMessage, []byte(time.Now().String()), time.Now().Add(WS_TIMEOUT)); err == nil {
			log.Tracef("ping sent")
		} else {
			log.Error("ping failed")
			return
		}
		time.Sleep(WS_TIMEOUT)
	}
}
