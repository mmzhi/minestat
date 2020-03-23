package main

import "strconv"

func main() {

	// 初始化配置
	initConfig()

	// 初始化日志功能
	initLog()

	// 初始化ECI
	InitEci()

	// 开启转发服务器
	InitProxy()

	// 初始化游戏功能
	InitMinecraftEci()

	// 初始化路由
	initRouter()
	router.Run(":" + strconv.Itoa(Config.Port))
}
