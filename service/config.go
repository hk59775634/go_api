package service

import (
	"encoding/json"
	"os"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

var Config configStruct

type configStruct struct {
	BaseUrl string `json:"BaseUrl"`
	BaseKey string `json:"BaseKey"`
}

func ReadConfig() error {
	// 构建根目录下的 config.json 文件路径
	configPath := filepath.Join("config.json")

	// 读取 config.json 文件内容
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	// 解析 JSON 数据
	err = json.Unmarshal(data, &Config)
	if err != nil {
		return fmt.Errorf("error decoding config file: %v", err)
	}

	// 读取配置文件中的值。如果没有设置，那么使用默认值
	if Config.BaseUrl == "" {
		Config.BaseUrl = "https://www.chat-api.net"
	}
	if Config.BaseKey == "" {
		Config.BaseKey = ""
	}

	// 如果环境变量设置了 BaseUrl 和 BaseKey ，那么使用环境变量的值
	if os.Getenv("BaseUrl") != "" {
		Config.BaseUrl = os.Getenv("BaseUrl")
	}
	if os.Getenv("BaseKey") != "" {
		Config.BaseKey = os.Getenv("BaseKey")
	}
	return nil
}
