package mon

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"os"
	"time"
)



var Client *mongo.Database


func Dbconnect() {
	want, err := readpref.New(readpref.SecondaryMode) //表示只使用辅助节点
	if err != nil {
		checkErr(err)
	}
	wc := writeconcern.New(writeconcern.WMajority())
	readconcern.Majority()
	//链接mongo服务
	opt := options.Client().ApplyURI("mongodb://localhost:27017")
	opt.SetLocalThreshold(3 * time.Second)     //只使用与mongo操作耗时小于3秒的
	opt.SetMaxConnIdleTime(5 * time.Second)    //指定连接可以保持空闲的最大毫秒数
	opt.SetMaxPoolSize(10)                    //使用最大的连接数
	opt.SetReadPreference(want)                //表示只使用辅助节点
	opt.SetReadConcern(readconcern.Majority()) //指定查询应返回实例的最新数据确认为，已写入副本集中的大多数成员
	opt.SetWriteConcern(wc)                    //请求确认写操作传播到大多数mongod实例
	ctx := GetContext()

	if database, err := mongo.Connect(ctx, opt); err != nil {
		checkErr(err)
	}else{
		//UseSession(client)
		//判断服务是否可用
		if err = database.Ping(ctx, readpref.Primary()); err != nil {
			checkErr(err)
		}

		Client = database.Database("testing_base")
	}
}


func GetContext() (ctx context.Context) {
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	return
}

func checkErr(err error) {
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("没有查到数据")
			os.Exit(0)
		} else {
			fmt.Println(err)
			os.Exit(0)
		}

	}
}