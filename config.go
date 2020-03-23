package main

import (
	"github.com/spf13/viper"
	"log"
)

type config struct {
	Debug 	bool
	Port 	int
	Container *struct {
		Nfs 			string
		Image			string
	}
	Aliyun *struct {
		AccessKey 		string
		SecretKey 		string
		RegionId 		string
		ZoneId 			string
		SecurityGroupId string
		VSwitchId 		string
	}
}

var Config config

func initConfig() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath("/etc/minestat/")   // path to look for the config file in
	viper.AddConfigPath("$HOME/.minestat")  // call multiple times to add many search paths
	viper.AddConfigPath(".")               // optionally look for config in the working directory
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		log.Fatalf("读取配置文件失败, %v", err)
	}

	err = viper.Unmarshal(&Config)
	if err != nil {
		log.Fatalf("转换配置文件失败, %v", err)
	}

	if Config.Port == 0 {
		Config.Port = 8080
	}
}