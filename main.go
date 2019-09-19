package main

import (
	"chat/impl"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	websocket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	conn, err := impl.InitConnection(websocket)
	if err != nil {
		return
	}

	go conn.ProcLoop()
	go conn.ReadLoop()
	go conn.WriteLoop()
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.ListenAndServe("0.0.0.0:7777", nil)
}
