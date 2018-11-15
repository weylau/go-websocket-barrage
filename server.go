package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"go-websocket-barrage/connection"
	"net/http"
	"time"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			//允许跨域
			return true
		},
	}
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	//完成握手应答
	var (
		wsConn *websocket.Conn
		err    error
		data   []byte
		conn   *connection.Connection
	)
	if wsConn, err = upgrader.Upgrade(w, r, nil); err != nil {
		return
	}
	defer wsConn.Close()

	if conn, err = connection.InitConnection(wsConn); err != nil {
		return
	}

	go func() {
		var (
			err error
		)
		for {
			if err = conn.WriteMessage([]byte("heartbeat")); err != nil {
				return
			}
			time.Sleep(time.Second)
		}
	}()

	for {
		if data, err = conn.ReadMessage(); err != nil {
			return
		}

		if err = conn.WriteMessage(data); err != nil {
			return
		}
	}
}
func main() {
	http.HandleFunc("/ws", wsHandler)
	http.ListenAndServe("localhost:7777", nil)
	count := 0
	for {
		fmt.Println("count:", count)
		count++
		time.Sleep(time.Second)
	}
}
