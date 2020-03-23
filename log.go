package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

var logger *log.Entry


func initLog() {
	l := log.New()

	if Config.Debug {
		l.SetLevel(log.DebugLevel)
	} else {
		l.SetLevel(log.InfoLevel)
	}


	// 输出格式是普通文字格式
	l.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	l.SetOutput(os.Stdout)
	logger = log.NewEntry(l)
}

func getGinLogFun() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()
		// 结束时间
		endTime := time.Now()
		// 执行时间
		latencyTime := endTime.Sub(startTime)
		// 请求方式
		reqMethod := c.Request.Method
		// 请求路由
		reqUri := c.Request.RequestURI
		// 状态码
		statusCode := c.Writer.Status()
		// 请求IP
		clientIP := c.ClientIP()
		// 日志格式
		logger.Debugf("| %3d | %13v | %15s | %s | %s |",
			statusCode,
			latencyTime,
			clientIP,
			reqMethod,
			reqUri,
		)
	}
}