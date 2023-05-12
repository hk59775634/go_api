package main

import (
	"log"
	"net/http"
	"go_api/service"
)

func main() {
	// 读取配置文件或环境变量。使用方法：service.Config.BaseUrl。如果是service包内部使用，可以直接使用 Config.BaseUrl
	err := service.ReadConfig()
	if err != nil {
		log.Fatal("Failed to read config:", err)
	}

	// 设置路由
	http.HandleFunc("/v1/chat/completions", service.Chatcompletions)
	http.HandleFunc("/v1/completions", service.Completions)
	// v1/images/generations
	http.HandleFunc("/v1/images/generations", service.ImagesGenerations)
	// 服务响应测试接口
	http.HandleFunc("/ping", service.Ping)
	http.HandleFunc("/test", service.Test)
	// 启动服务
	http.ListenAndServe(":5000", nil)
}
