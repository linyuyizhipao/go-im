package redis

import (
	"github.com/go-redis/redis"
	"test/extend/conf"
)

var Client *redis.Client

//获取redis连接。
func InitRedis(){

	redisConfig := conf.RedisConf
	Client = redis.NewClient(&redis.Options{
		Addr:     redisConfig.Host,
		Password: redisConfig.Password,
		DB:       redisConfig.DBNum,
		PoolSize:redisConfig.MaxActive,
		MinIdleConns : redisConfig.MaxIdle,
	})
	_, err := Client.Ping().Result()
	if err != nil {
		panic(err.Error())
	}
}
