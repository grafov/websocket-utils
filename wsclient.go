// Simple websocket client with CLI shell. It can send text messages to ws server.
package main

/*
 Copyleft 2014 Alexander I.Grafov aka Axel <grafov@gmail.com>

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU General Public License as published by
 the Free Soft ware Foundation, either version 3 of the License, or
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
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kdar/factorlog"
	"github.com/peterh/liner"
)

var (
	Version        string // value set by compiler
	Build          string // value set by compiler
	err            error
	log            *factorlog.FactorLog
	bindURL        *url.URL
	history        = "~/.wsclient_history"
	ping           = flag.Bool("ping", true, "ping clients and response to ping requests")
	headers        = flag.String("headers", "", "file with custom HTTP headers for client handshake")
	wsClient       wscon
	stamp          time.Time
	wsTestUpgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}
	dialer = websocket.Dialer{
		HandshakeTimeout: 3 * time.Second,
	}
)

/* We need initialize logging before all other modules.
This order don't matter in regular program run but critical for unit tests.

This init procedure needs only for logging. If you don't want log messages
then you may remove this function and comment out all log.entries.
*/
func init() {
	var (
		logFmt string
		debug  = flag.Bool("debug", false, "debug dispatch")
		verb   = flag.Bool("verb", false, "verbose dispatch")
		bind   = flag.String("bind", "127.1:48084", "bind HTTP requests to addr:port/path")
	)

	flag.Parse()
	bindURL, err = url.Parse(*bind)

	if *debug {
		// brief format for debug
		logFmt = `%{Color "red" "ERROR"}%{Color "red" "FATAL"}%{Color "yellow" "WARN"}%{Color "green" "INFO"}%{Color "cyan" "DEBUG"}%{Color "blue" "TRACE"}[%{Date} %{Time}] [%{SEVERITY}:%{File}:%{Line}] %{Message}%{Color "reset"}`
	} else {
		// short format for production
		logFmt = `%{Color "red" "ERROR"}%{Color "red" "FATAL"}%{Color "yellow" "WARN"}%{Color "green" "INFO"}%{Color "cyan" "DEBUG"}%{Color "blue" "TRACE"}[%{Date} %{Time}] [%{SEVERITY}] %{Message}%{Color "reset"}`
	}

	log = factorlog.New(os.Stderr, factorlog.NewStdFormatter(logFmt))

	switch {
	case *debug:
		log.SetMinMaxSeverity(factorlog.TRACE, factorlog.PANIC)
	case *verb:
		log.SetMinMaxSeverity(factorlog.INFO, factorlog.PANIC)
	default:
		log.SetMinMaxSeverity(factorlog.WARN, factorlog.PANIC)
	}

	if err != nil {
		log.Criticalf("bad url %s", *bind)
		os.Exit(1)
	}
	if bindURL.Path == "" {
		bindURL.Path = "/"
	}
}

func main() {
	log.Infof("Websocket client ver. %s (build %s)", Version, Build)

	wsclosed := make(chan bool)
	wsInit()
	go wsClient.textIn()
	go wsClient.textOut()
	go wsClient.dispatch(wsclosed)
	//wsClient.pongDispatcher(8 * time.Second)

	terminate := make(chan os.Signal)
	signal.Notify(terminate, os.Interrupt)

	select {
	case <-terminate:
		log.Infoln("interrupted by operator")
	case <-wsclosed:
		log.Infoln("websocket connection closed")
	}
}

// Establishes websocket connection.
func wsInit() {
	h := make(http.Header)
	if *headers != "" {
		if customHeaders, err := ioutil.ReadFile(*headers); err != nil {
			log.Errorf("can't read HTTP headers from %s", *headers)
		} else {
			for _, line := range strings.Split(string(customHeaders), "\n") {
				keyval := strings.SplitN(line, ":", 2)
				if len(keyval) == 2 {
					h.Add(keyval[0], keyval[1])
				}
			}
			if h.Get("Host") == "" {
				// if host skiped in the headers then it added as RFC2616 requires
				h.Set("Host", bindURL.Host)
			}
		}
	}
	var proto string
	if bindURL.Scheme != "" {
		proto = bindURL.Scheme
	} else {
		proto = "ws"
	}
	wsClient.Conn, _, err = dialer.Dial(fmt.Sprintf("%s://%s%s", proto, bindURL.Host, bindURL.Path), h)
	if err != nil {
		log.Fatalf("can't connect to WS server with %s", err)
	}
	wsClient.Inbox = make(chan wsmsg, 32)
	wsClient.TextIn = make(chan *bytes.Buffer, 8)
	wsClient.TextOut = make(chan *bytes.Buffer, 8)
}

// Реквизиты соединения и каналы обмена данными.
type wscon struct {
	Conn    *websocket.Conn
	Inbox   chan wsmsg         // полученные сообщения
	TextIn  chan *bytes.Buffer // текст для отправки
	TextOut chan *bytes.Buffer // результат для отображения
}

// Для обмена приёмника с диспетчером.
type wsmsg struct {
	Op      int
	Payload *bytes.Buffer
}

func (c wscon) dispatch(closed chan bool) {
	ticker := time.Tick(6 * time.Second)

	go c.receiveLoop()

	for {
		select {
		case text := <-c.TextIn:
			err = c.send(websocket.TextMessage, text)
		case msg := <-c.Inbox:
			switch msg.Op {
			case websocket.TextMessage, websocket.BinaryMessage:
				c.TextOut <- msg.Payload
			case websocket.CloseMessage:
				err = c.send(websocket.CloseMessage, nil)
			}
		case <-ticker:
			if err = c.Conn.WriteControl(websocket.PingMessage, []byte(fmt.Sprintf("client at %s", time.Now().String())), time.Now().Add(10*time.Second)); err == nil {
				log.Tracef("ping sent by client")
			} else {
				log.Trace("can't send ping: %s", err)
			}
		}
		if err != nil {
			log.Error(err)
			c.Conn.Close()
			closed <- true
			return
		}
		err = nil
	}
}

func (c wscon) receiveLoop() {
	for {
		op, r, err := c.Conn.NextReader()
		if err != nil {
			log.Errorf("read failed with %s", err)
			return
		}
		buffer := new(bytes.Buffer)
		_, err = buffer.ReadFrom(r)
		if err != nil {
			log.Errorf("payload read for op %d failed with %s", op, err)
			return
		}
		log.Tracef("got: %s", buffer.Bytes())
		c.Inbox <- wsmsg{Op: op, Payload: buffer}
	}
}

func (c wscon) send(op int, data *bytes.Buffer) error {
	if writeTo, err := c.Conn.NextWriter(op); err == nil {
		defer writeTo.Close()
		if data != nil {
			log.Trace("send data")
			writeTo.Write(data.Bytes())
		}
		return nil
	} else {
		log.Error(err)
		return err
	}
}

// Собирает данные для отправки из stdin
func (c wscon) textIn() {

	line := liner.NewLiner()
	defer func() {
		line.Close()
		if f, err := os.Create(history); err != nil {
			log.Error("Error writing history file: ", err)
		} else {
			line.WriteHistory(f)
			f.Close()
		}
	}()

	if f, err := os.Open(history); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	for {
		if text, err := line.Prompt(fmt.Sprintf("%s%s >> ", bindURL.Host, bindURL.Path)); err != nil {
			log.Print("Error reading line: ", err)
		} else {
			cmdline := strings.TrimSpace(text)
			if len(cmdline) == 0 {
				continue
			}
			switch cmdline[0] {
			case '!': // get output from external cmd
				cmd := exec.Command("bash", "-c", cmdline[1:])
				if extout, err := cmd.Output(); err != nil {
					log.Errorf("External command %s failed", cmd)
					continue
				} else {
					text = string(extout)
				}
			case '<': // load from file
				if data, err := ioutil.ReadFile(cmdline[1:]); err != nil {
					log.Errorf("can't load file %s", cmdline[1:])
					continue
				} else {
					text = string(data)
				}
			}
			c.TextIn <- bytes.NewBufferString(text)
			stamp = time.Now()
			line.AppendHistory(text)
		}
	}
}

// Собирает данные для отправки из stdin
func (c wscon) textOut() {
	for {
		out := <-c.TextOut
		fmt.Printf("\n<< [%s] %s", time.Since(stamp), out.String())
		fmt.Printf("\n%s%s >> ", bindURL.Host, bindURL.Path)
	}
}

// Set timeout depends on the ping
func (c wscon) pongDispatcher(timeout time.Duration) {
	c.Conn.SetReadDeadline(time.Now().Add(timeout))
	c.Conn.SetPongHandler(func(arg string) error {
		c.Conn.SetReadDeadline(time.Now().Add(timeout))
		return nil
	})
}
