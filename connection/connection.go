package connection

import (
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
	var wg *sync.WaitGroup
	wg.Add(1)
	//启动读协程
	go conn.readLoop(wg)
	wg.Add(1)
	//启动写协程
	go conn.writeLoop(wg)
	go func() {
		wg.Wait()
		conn.Close()
	}()
	return
}

//API
func (conn *Connection) ReadMessage() (data []byte, err error) {
	data = <-conn.inMsg
	return
}

func (conn *Connection) WriteMessage(data []byte) (err error) {
	conn.outMsg <- data
	return
}

func (conn *Connection) Close() {
	//线程安全可重复调用
	conn.wsConn.Close()
	close(conn.closeChan)
}

func (conn *Connection) readLoop(wg *sync.WaitGroup) {
	defer wg.Done()
	var (
		data []byte
		err  error
	)
	for {
		if _, data, err = conn.wsConn.ReadMessage(); err != nil {
			return
		}
		select {
		case conn.inMsg <- data:
		case <-conn.closeChan:
			//关闭逻辑
			return
		}

	}
}

func (conn *Connection) writeLoop(wg *sync.WaitGroup) {
	defer wg.Done()
	var (
		data []byte
		err  error
	)

	for {
		select {
		case data = <-conn.outMsg:
		case <-conn.closeChan:
			return
		}
		if err = conn.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}
