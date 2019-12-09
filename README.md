# golang 开箱即用的im

##1. 模仿gin的调用方式开启一个websocket服务

调用示例：
```go
package main

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
```
##核心点：
* 1.webSckt 结构体继承了 easyim.WebSocketChat接口，所以你必须实现
    * 1.1 GetSendClientMsg 方法：此函数用于处理client发送过来的msg，需要你自己将msg发送到你自己业务中的队列中去,充当生产者的角色
    * 1.2 AccessClientMsg 方法：此函数用于从队列中拉取需要发送给client的数据,充当消费者的角色

* 2 示例中使用的时redis充当消息队列，但是其实上面的2方法只需要你保证消息处理与消息获取，及AccessClientMsg时每当有client有消息来到的时候就会执行它一次
   GetSendClientMsg 是只要你想给client发送一条消息，你只需要保证它能够获取到你想要发送的消息便可，框架自会保证消息点到点的稳定到达,值得你注意的是你需
   要保证GetSendClientMsg获取到的json格式应该为{"client_id":"123"},及AccessClientMsg中存在的那个client_id的值。