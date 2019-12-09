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

