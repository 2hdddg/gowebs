package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type Options struct {
	OnReceive     func(b []byte)
	TimeoutInSecs int
	BufferSize    int
}

type Connection interface {
	Send(msg []byte)
	Close()
}

type ClientConnection struct {
	sendChan chan []byte
}

type ServerConnection struct {
	options  Options
	conn     *websocket.Conn
	sendChan chan []byte
}

func (c *ClientConnection) Send(msg []byte) {
	c.sendChan <- msg
}

func Accept(w http.ResponseWriter, r *http.Request, o *Options) (*ServerConnection, error) {

	upgrader := websocket.Upgrader{
		ReadBufferSize:  o.BufferSize,
		WriteBufferSize: o.BufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	conn.SetReadLimit(int64(o.BufferSize))

	server := ServerConnection{
		options:  *o,
		conn:     conn,
		sendChan: make(chan []byte),
	}

	go serve(server.conn, server.sendChan)

	return &server, nil
}

func (s *ServerConnection) Close() {
	close(s.sendChan)
}

func serve(conn *websocket.Conn, sendChan chan []byte) {
	pingPeriod := 2 * time.Second
	pingTimeout := 3 * time.Second
	closedChan := make(chan bool)

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	// Reader
	go func() {
		conn.SetReadDeadline(time.Now().Add(pingTimeout))
		conn.SetPongHandler(func(string) error {
			//fmt.Println("Received ping")
			conn.SetReadDeadline(time.Now().Add(pingTimeout))
			return nil
		})
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Reader", err)
				break
			} else {
				fmt.Println("Received msg:", msg)
			}
		}
		closedChan <- true
	}()

	writeWait := 1 * time.Second

	stopReader := func() {
		fmt.Println("Stopping reader")
		ticker.Stop()
		//conn.SetReadDeadline(time.Now())
	}

	connected := true
	for connected {
		select {
		case msg, more := <-sendChan:
			if !more {
				fmt.Println("Closed on this end")
				stopReader()
			}
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				fmt.Println(err)
				stopReader()
			}
			w.Write(msg)
			err = w.Close()
			if err != nil {
				fmt.Println(err)
				stopReader()
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				fmt.Println(err)
				stopReader()
			}
		case <-closedChan:
			connected = false
		}
	}

	fmt.Println("Connection closed")
}

func Connect(conn net.Conn, u *url.URL, o *Options) (*ClientConnection, error) {
	hdr := http.Header{"origin": []string{"localhost"}}
	wsconn, _, err := websocket.NewClient(conn, u, hdr, 1024, 1024)
	if err != nil {
		return nil, err
	}
	myconn := ClientConnection{
		sendChan: make(chan []byte),
	}

	go serve(wsconn, myconn.sendChan)

	return &myconn, nil
}

