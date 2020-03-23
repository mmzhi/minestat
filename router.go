package main

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

var router *gin.Engine

func initRouter() {

	if Config.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 关闭 gin 默认日志打印功能，使用自定义的
	gin.DisableConsoleColor()
	gin.DefaultWriter = ioutil.Discard

	router = gin.Default()
	router.Use(getGinLogFun())

	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	router.GET("/state", func(c *gin.Context) {
		if containerGroup == nil {
			c.JSON(200, gin.H{ "text" : "启动游戏" })
		} else if containerGroup.Status != "Running" || containerGroup.IntranetIp == ""  {
			c.JSON(200, gin.H{ "text" : "正在启动..." })
		} else if minestat == nil || !minestat.Online  {
			c.JSON(200, gin.H{ "text" : "正在启动..." })
		} else {
			c.JSON(200, gin.H{ "text" : "已经启动" })
		}
	})

	router.POST("/play", func(c *gin.Context) {
		go func() {
			startMinecraft()
		}()
		c.JSON(200, gin.H{ })
	})
}
