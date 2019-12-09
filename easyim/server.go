package easyim

import (
	"log"
	"net/http"
)

type WebSocketChat interface {
	GetSendClientMsg() (str string, err error)
	AccessClientMsg(msg string) (err error)
}

func NewHub(w WebSocketChat) (ws *Hub) {
	ws = &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clientMap:  make(map[int]*Client),
	}
	ws.WebSocketChat = w
	return
}

func ServeHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, `D:\project\go-mongo\examples\chat\home.html`)
}

// serveWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil) //升级协议
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.clientKey = client.CreateClientId()
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}


