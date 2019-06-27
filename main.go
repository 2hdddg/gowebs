package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
	//"golang.org/x/net/websocket"
)

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := Accept(w, r, &Options{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	time.Sleep(300 * time.Second)
	fmt.Println("Closing...")
	conn.Close()

	/*
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defer conn.Close()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Read err")
				fmt.Println(err)
				//os.Exit(1)
				break
			}
			fmt.Println(message)
		}
		fmt.Println("Failed to read")
	*/
}

func server() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r)
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func client() {
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//defer conn.Close()

	url := url.URL{
		Scheme: "ws",
		Host:   "localhost:8080",
		Path:   "/ws",
	}
	/*
		hdr := http.Header{"origin": []string{"localhost"}}
		wsconn, _, err := websocket.NewClient(conn, &url, hdr, 1024, 1024)
		//wsconn.SetWriteDeadline(
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	*/
	wsconn, err := Connect(conn, &url, &Options{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {
		wsconn.Send([]byte("Hello"))
		time.Sleep(5 * time.Second)
	}
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Need param")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		fmt.Println("Server")
		server()
	case "client":
		fmt.Println("Client")
		client()
	default:
		fmt.Println("Unknown")
		os.Exit(1)
	}
}
