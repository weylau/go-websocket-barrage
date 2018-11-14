package connection

import (
	"errors"
	"github.com/gorilla/websocket"
	"sync"
)

type Connection struct {
	wsConn    *websocket.Conn
	inMsg     chan []byte
	outMsg    chan []byte
	closeChan chan byte

	mutex   sync.Mutex
	isClose bool
}

func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConn:    wsConn,
		inMsg:     make(chan []byte, 1000),
		outMsg:    make(chan []byte, 1000),
		closeChan: make(chan byte, 1),
	}

	//启动读协程
	go conn.readLoop()

	//启动写协程
	go conn.writeLoop()
	return
}

//API
func (conn *Connection) ReadMessage() (data []byte, err error) {
	select {
	case data = <-conn.inMsg:
	case <-conn.closeChan:
		err = errors.New("connection is closed")
	}

	return
}

func (conn *Connection) WriteMessage(data []byte) (err error) {
	select {
	case conn.outMsg <- data:
	case <-conn.closeChan:
		err = errors.New("connection is closed")
	}

	return
}

func (conn *Connection) Close() {
	//线程安全可重复调用
	conn.wsConn.Close()

	conn.mutex.Lock()
	if !conn.isClose {
		close(conn.closeChan)
		conn.isClose = true
	}
	conn.mutex.Unlock()
}

func (conn *Connection) readLoop() {
	var (
		data []byte
		err  error
	)
	for {
		if _, data, err = conn.wsConn.ReadMessage(); err != nil {
			conn.Close()
			return
		}
		select {
		case conn.inMsg <- data:
		case <-conn.closeChan:
			//关闭逻辑
			conn.Close()
			return
		}

	}
}

func (conn *Connection) writeLoop() {
	var (
		data []byte
		err  error
	)

	for {
		select {
		case data = <-conn.outMsg:
		case <-conn.closeChan:
			//关闭逻辑
			conn.Close()
			return
		}

		if err = conn.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			conn.Close()
			return
		}
	}
}
