package impl

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	msgType int
	data    []byte
}

type Connection struct {
	wsConn  *websocket.Conn
	inChan  chan *Message
	outChan chan *Message

	closeChan chan byte
	isClosed  bool
	mutex     sync.Mutex
}

func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConn:    wsConn,
		inChan:    make(chan *Message, 1000),
		outChan:   make(chan *Message, 1000),
		closeChan: make(chan byte, 1),
		isClosed:  false,
	}

	return
}

func (conn *Connection) ProcLoop() {
	go func() {
		for {
			time.Sleep(2 * time.Second)
			if err := conn.wsWrite(websocket.TextMessage, []byte("heartbeat.")); err != nil {
				fmt.Println("heatbeat fail.")
				conn.Close()
				break
			}
		}
	}()

	for {
		msg, err := conn.wsRead()
		if err != nil {
			fmt.Println("read fail.")
			break
		}
		fmt.Println(string(msg.data))
		err = conn.wsWrite(msg.msgType, msg.data)
		if err != nil {
			fmt.Println("write fail.")
			break
		}
	}
}

func (conn *Connection) wsRead() (*Message, error) {
	select {
	case msg := <-conn.inChan:
		return msg, nil
	case <-conn.closeChan:
	}
	return nil, errors.New("websocket closed")
}

func (conn *Connection) wsWrite(msgType int, data []byte) error {
	select {
	case conn.outChan <- &Message{msgType, data}:
	case <-conn.closeChan:
		return errors.New("closed")
	}
	return nil
}

func (conn *Connection) Close() {
	conn.wsConn.Close()
	conn.mutex.Lock()

	if !conn.isClosed {
		close(conn.closeChan)
		conn.isClosed = true
	}
	conn.mutex.Unlock()
}

func (conn *Connection) ReadLoop() {
	for {
		msgType, data, err := conn.wsConn.ReadMessage()

		if err != nil {
			goto ERR
		}

		req := &Message{
			msgType,
			data,
		}

		select {
		case conn.inChan <- req:
		case <-conn.closeChan:
			goto ERR
		}

	}
ERR:
	conn.Close()
}

func (conn *Connection) WriteLoop() {
	for {
		select {
		case msg := <-conn.outChan:
			if err := conn.wsConn.WriteMessage(msg.msgType, msg.data); err != nil {
				goto ERR
			}
		case <-conn.closeChan:

		}
	}
ERR:
	conn.Close()
}
