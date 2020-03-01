package logs

import (
	"test/extend/conf"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup 日志初始化设置
func Initlog() {
	level := strings.ToLower(conf.LoggerConf.Level) //获取配置文件中配置的日志异常级别，并转全小写
	//根据不同的日志级别设置zerolog组件 打印的日志格式
	switch  level{
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if conf.LoggerConf.FilePath != "" {
		files,_ :=  os.OpenFile(conf.LoggerConf.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		//是否在终端输出日志，及输出格式
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:     files,
		})

	}else{
		files := os.Stderr
		//是否在终端输出日志，及输出格式
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:     files,
		})

	}
}
