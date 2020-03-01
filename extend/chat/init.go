package chat

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"test/extend/redis"
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)
var addr = flag.String("addr", ":8080", "http service address")

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	clientMap map[string]*Client
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	clientKey string
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		clientMap:  make(map[string]*Client),
	}
}

func (h *Hub) run() {
	go func() {
		for {
			messageJson, err := sendClientSend()
			if err != nil {
				time.Sleep(time.Millisecond * 2)
				continue
			}
			clientId := gjson.Get(messageJson, CLIENT_MSG_MAP_KEY_NAME).String()
			if clientId == "" {
				log.Error().Msgf("这个消息不合法，不存在约定好了并且生产好了的的client_id的key")
				continue
			}

			message := []byte(messageJson)
			if client, ok := h.clientMap[clientId]; ok {
				client.send <- message
			}
		}
	}()

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
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Msgf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		msgMap := map[string]interface{}{}
		if err := json.Unmarshal(message, &msgMap); err != nil {
			log.Error().Msgf("msg:%s在解析到map中的时候发生了错误%s", string(message), err.Error())
			continue
		}
		msgMap[CLIENT_MSG_MAP_KEY_NAME] = c.clientKey
		msg, _ := json.Marshal(msgMap)
		messageStr := string(msg)
		clientMsgProduct(messageStr)
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
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

//获取client发送过来的消息
func clientMsgProduct(msg string) {
	b, err := redis.Redis.Do("lpush", "chat", msg).Bool()
	if err != nil {
		log.Error().Msgf("lpush%s发生了错误，错误信息为：%s", msg, err.Error())
	}

	if b != true {
		log.Error().Msgf("lpush%s发生失败了，错误信息为：%s", msg)
	}
}

//获取将要发送到client的消息
func sendClientSend() (msg string, err error) {
	msg, err = redis.Redis.Do("rpop", "chat").String()
	return
}
