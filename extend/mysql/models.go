package mysql

import (
	"fmt"
	"strconv"
	"time"
	"test/extend/conf"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // mysql
)

// DB 当前数据库连接
var DB *gorm.DB

// Setup MySQL 数据库配置
func Setup() {
	var err error
	//读取并拼接配置文件中的连接字符串
	var connectString = fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.DBConf.User,
		conf.DBConf.Password,
		conf.DBConf.Host+":"+strconv.Itoa(conf.DBConf.Port),
		conf.DBConf.DBName,
	)
	//gorm.open方法里面ping了并且close了
	//值得注意的是orm提供的方法是自己做了close这个行为的，但是如果你使用到了未close的方法请注意自己主动释放资源
	//具体下次再总结
	DB, err = gorm.Open(conf.DBConf.DBType, connectString)
	if err != nil {
		fmt.Printf("mysql connect error %v", err)
		time.Sleep(10 * time.Second) // 若连接失败，则延时10秒重新连接,防止网络抖动
		DB, err = gorm.Open(conf.DBConf.DBType, connectString)
		if err != nil {
			panic(err.Error())
		}
	}

	if DB.Error != nil {
		fmt.Printf("database error %v", DB.Error)
	}

	//设置orm操作数据库的表前缀
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return conf.DBConf.TablePrefix + defaultTableName
	}

	DB.LogMode(conf.DBConf.Debug)
	DB.SingularTable(true)
	//数据库连接池最大连接数以及最大空闲数
	DB.DB().SetMaxIdleConns(100)
	DB.DB().SetMaxOpenConns(200)

	// 全局禁用表名复数
	DB.SingularTable(true) // 如果设置为true,`User`的默认表名为`user`,使用`TableName`设置的表名不受影响

}
