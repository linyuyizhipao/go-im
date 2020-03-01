package easyim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"math/rand"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	CLIENT_MSG_MAP_KEY_NAME = "client_id"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
	MessageJsonByService =  make(chan string,10)
	MessageJsonByClient =  make(chan string,10)
)

type Hub struct {
	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	clientMap map[int]*Client
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	clientKey int

}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//只要有client来了websocket的话就会注册到 h.clientMap 这里面来
//同理只要断开，就会从  h.clientMap 中剔除掉
func (h *Hub) Run() {
	go func(){
		for {
			select {
			case client := <-h.register:
				h.clientMap[client.clientKey] = client
			case client := <-h.unregister:
				if _, ok := h.clientMap[client.clientKey]; ok {
					delete(h.clientMap, client.clientKey)
				}
			}
		}
	}()

	go func() {

		for messageStr := range MessageJsonByService {
			clientId64 := gjson.Get(messageStr, CLIENT_MSG_MAP_KEY_NAME).Int()
			if clientId64 <= 0  {
				log.Error().Msg("没有获取到数据")
				return
			}
			clientId := int(clientId64)
			if client :=h.clientMap[clientId];client != nil {
				client.send <- []byte(messageStr)
			}
		}

	}()
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	if err:=c.conn.SetReadDeadline(time.Now().Add(pongWait));err!=nil{
		fmt.Println(err)
	}
	c.conn.SetPongHandler(func(string) error {
		if err:=c.conn.SetReadDeadline(time.Now().Add(pongWait));err!=nil{
			fmt.Println(err)
		}
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Msgf("error:%s", err.Error())
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		msgMap := map[string]interface{}{}
		if err := json.Unmarshal(message, &msgMap); err != nil {
			log.Info().Msg("没有消息")
			continue
		}
		msgMap[CLIENT_MSG_MAP_KEY_NAME] = c.clientKey
		msg, _ := json.Marshal(msgMap)
		messageStr := string(msg)

		MessageJsonByClient <- messageStr
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.conn.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				fmt.Println(err)
			}

			if !ok {
				// The hub closed the channel.
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					fmt.Println(err)
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				fmt.Println(err)
			}

			n := len(c.send)
			for i := 0; i < n; i++ {
				if _, err := w.Write(newline); err != nil {
					fmt.Println(err)
				}
				if _, err := w.Write(<-c.send); err != nil {
					fmt.Println(err)
				}
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				fmt.Println(err)
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) CreateClientId() (randInt int) {
	randInt = rand.Int()
	return
}