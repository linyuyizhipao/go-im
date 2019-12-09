package main

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/tidwall/gjson"
	"net/http"
	"test/easyim"
)

type webSckt struct {
	addr string
	easyim.WebSocketChat
}

func (*webSckt) GetSendClientMsg() (str string, err error) {
	str, err = redisClient.Do("rpop", "chat_test").String()
	clientId := gjson.Get(str, easyim.CLIENT_MSG_MAP_KEY_NAME).String()
	if clientId == "" {
		err = errors.New("没有获取到数据")
	}
	return
}

func (*webSckt)  AccessClientMsg(msg string) (err error) {
	b, err := redisClient.Do("lpush", "chat_test", msg).Bool()
	if err != nil {
		err = errors.New(fmt.Sprintf("lpush%s发生了错误，错误信息为：%s", msg, err.Error()))
	}
	if b != true {
		err = errors.New(fmt.Sprintf("lpush%s发生失败了，错误信息为：%s", msg))
	}
	return
}


func main(){
	ws := new(webSckt)
	ws.addr = ":8080"
	hub := easyim.NewHub(ws)
	hub.Run()

	http.HandleFunc("/", easyim.ServeHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		easyim.ServeWs(hub, w, r)
	})
	err := http.ListenAndServe(ws.addr, nil)
	if err != nil {
		fmt.Println("ListenAndServe", err)
	}
}




func init() {
	InitRedis()
}

var redisClient *redis.Client

func InitRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:6379",
		Password:     "",
		DB:           0,
		PoolSize:     30,
		MinIdleConns: 30,
	})
	_, err := redisClient.Ping().Result()
	if err != nil {
		panic(err.Error())
	}
}