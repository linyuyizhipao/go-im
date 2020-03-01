package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"test/easyim"
	"test/extend/conf"
	"test/extend/logs"
	"test/extend/redis"
	"time"
)
var hub *easyim.Hub

//消费者
func sendMsgToClient() {
	for {
		if str, err := redis.Client.Do("lpop", "chat_test").String();err == nil{

			if str == "" {
				time.Sleep(time.Microsecond * 3)
				continue
			}

			log.Info().Msgf("我是要发送到客户端的消息%s:",str)
			easyim.MessageJsonByService <- str
		}else{
			log.Error().Msgf("lpop发生了错误%s:",err.Error())
			time.Sleep(time.Microsecond * 300)
		}
	}
}

// client发送过来的消息，生产者
func getMsgToClient() {
	for msg := range easyim.MessageJsonByClient {
		log.Info().Msgf("给客户端发送过来的消息:%s",msg)

		go func(){
			b, err := redis.Client.Do("lpush", "chat_test", msg).Bool()
			if err != nil {
				log.Error().Msg(fmt.Sprintf("lpush%s发生了错误，错误信息为：%s", msg, err.Error()))
			}
			if b != true {
				log.Error().Msg(fmt.Sprintf("lpush发生失败了，错误信息为：%s", msg))
			}
		}()

	}
}


func main(){
	conf.Setup()  //基本配置初始化
	logs.Initlog()  // 日志初始化
	redis.InitRedis()  // 缓存初始化
	addr := fmt.Sprintf(":%d",conf.ServerConf.Port)
	hub = easyim.NewHub()
	hub.Run()

	http.HandleFunc("/", easyim.ServeHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		easyim.ServeWs(hub, w, r)
	})

	go sendMsgToClient()
	go getMsgToClient()

	log.Info().Msg("启动成功")
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Error().Msgf("ListenAndServe:%s",err.Error())
	}
}

