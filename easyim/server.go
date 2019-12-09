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

// websocket 处理器，处理client发送过来的msg
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil) //升级协议
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.clientKey = client.CreateClientId() //生成client_id标识,后面业务code处理完毕后会根据此client_id去去client到达回复消息的目的
	client.hub.register <- client  //将client放入全局map,此map的关联关系是保证消息点到点的base保证

	go client.writePump()  //实时监控有木有待发送到client的消息，一旦存在就会负责将此消息发送到client中，不牵扯任务业务，只保证消息到达client
	go client.readPump()  //实时监控client的消息，一旦存在就会负责将此消息发送到AccessClientMsg方法中，不牵扯任务业务，只保证消息到达AccessClientMsg
}


