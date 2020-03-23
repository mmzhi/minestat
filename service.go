package main

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/eci"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

const minecraftPort = "25565"

// 运行容器的信息
var containerGroup *eci.DescribeContainerGroupsContainerGroup0
var minestat *Minestat
var lastZeroPlayerTime *time.Time


func InitMinecraftEci(){
	// 获取正在运行的容器
	containerGroup, _ = DescribeContainerGroup()

	// 启动容器定食检测
	go func() {
		checkMinestat()
	}()

	// 启动容器定食清理
	go func() {
		t := time.NewTicker(time.Minute * 15)
		defer t.Stop()

		for {
			<-t.C
			DescribeContainerGroup()
		}
	}()
}

// 创建 minecraft
func startMinecraft()  {
	if c, _ := DescribeContainerGroup(); c != nil {
		logger.Debug("游戏已启动容器")
	}

	logger.Debug("正在启动容器")
	if _, err := CreateContainerGroup(); err != nil {
		return
	} else {
		containerGroup, _ = DescribeContainerGroup()
	}

	t := time.NewTicker(time.Second * 10)
	defer t.Stop()

	i := 0

	logger.Debug("启动容器完成，开始检测启动状态")
	x1 := false
	for  {
		<- t.C

		if i > 90 {
			logger.Debug("超时未启动，退出启动程序")
			break
		}
		i++

		if containerGroup == nil || containerGroup.Status != "Running" || containerGroup.IntranetIp == "" {
			containerGroup, _ = DescribeContainerGroup()
			if containerGroup == nil || containerGroup.Status != "Running" || containerGroup.IntranetIp == "" {
				continue
			}
		}

		if x1 == false {
			logger.Debug("容器启动完成，正在启动游戏")
			x1 = true
		}

		minestat = GetMinestat(containerGroup.IntranetIp, minecraftPort)
		if !minestat.Online {
			continue
		}

		logger.Debug("启动游戏完成")
		break
	}
}


// 检测 minecraft 的运行状态
func checkMinestat() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()

	for {
		<- t.C

		// 查找容器是否运行
		if containerGroup == nil || containerGroup.Status != "Running" || containerGroup.IntranetIp == "" {
			containerGroup, _ = DescribeContainerGroup()
			if containerGroup == nil {
				logger.Debug("容器未运行")
				lastZeroPlayerTime = nil
				minestat = nil
				continue
			}else if containerGroup.Status != "Running" {
				logger.Debug("初始化未完成")
				lastZeroPlayerTime = nil
				minestat = nil
				continue
			} else if containerGroup.IntranetIp == "" {
				logger.Debug("未能正确获取容器IP")
				lastZeroPlayerTime = nil
				minestat = nil
				continue
			}
		}

		logger.Debug("容器IP为:" + containerGroup.IntranetIp)

		// 查找 minecraft 是否运行
		minestat = GetMinestat(containerGroup.IntranetIp, minecraftPort)
		if !minestat.Online {
			lastZeroPlayerTime = nil
			logger.Debug("minecraft没有运行")
			continue
		}

		// 查找当前人数
		if player, _ := strconv.Atoi(minestat.CurrentPlayers); player > 0 {
			lastZeroPlayerTime = nil
			logger.Debug("当前有人在玩minecraft")
			continue
		}

		// 查找是否第一次检测到时间
		if lastZeroPlayerTime == nil {
			t := time.Now()
			lastZeroPlayerTime = &t
			logger.Debug("没有人玩minecraft了，开始计时")
			continue
		}

		// 查找是否到时间了
		duration := time.Now().Sub(*lastZeroPlayerTime)
		if duration > time.Minute * 15 {
			DeleteContainerGroup(containerGroup.ContainerGroupId)
			containerGroup = nil
			minestat = nil
			lastZeroPlayerTime = nil
			logger.Debug("超过15分钟没人玩了，关闭服务")
			continue
		}

		logger.Debug("暂时没有人玩minecraft")
	}
}




// 开启服务器
func InitProxy() {
	lis, err := net.Listen("tcp", "0.0.0.0:25565")
	if err != nil {
		logger.Error(err)
		panic(err)
	}

	go func() {
		defer lis.Close()
		for {
			conn, err := lis.Accept()
			if err != nil {
				logger.Error("建立连接错误:", err)
				continue
			}
			go handle(conn)
		}
	}()
}


func handle(sconn net.Conn) {
	defer sconn.Close()
	ip, ok := getIP()
	if !ok {
		return
	}
	dconn, err := net.Dial("tcp", ip)
	if err != nil {
		logger.Error("连接失败:", ip, err)
		return
	}
	ExitChan := make(chan bool, 1)
	go func(sconn net.Conn, dconn net.Conn, Exit chan bool) {
		_, err := io.Copy(dconn, sconn)
		if err != nil && strings.Index(err.Error(), "use of closed network connection") == 0 {
			logger.Error("发送数据失败:", err)
		}
		ExitChan <- true
	}(sconn, dconn, ExitChan)
	go func(sconn net.Conn, dconn net.Conn, Exit chan bool) {
		_, err := io.Copy(sconn, dconn)
		if err != nil && strings.Index(err.Error(), "use of closed network connection") == 0 {
			logger.Error("接收数据失败:", err)
		}
		ExitChan <- true
	}(sconn, dconn, ExitChan)
	<-ExitChan
	dconn.Close()
}

func getIP() (string, bool) {
	if containerGroup == nil || containerGroup.Status != "Running" || containerGroup.IntranetIp == "" {
		return "", false
	}
	return containerGroup.IntranetIp + ":" + minecraftPort, true
}