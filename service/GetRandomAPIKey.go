package service

import (
	"log"
	"net/http"
	"time"
	"math/rand"
	"encoding/json"

)

type apiKey struct {
	APIURL string `json:"api_url"`
	APIKey string `json:"api_key"`
}

var apiKeys []apiKey

// apiKeysCache 用于缓存 API URL 和 API Key
var apiKeysCache struct {
	keys     []apiKey
	lastLoad time.Time
}

// getAPIKeys 返回 API URL 和 API Key，如果它们未被缓存，则从 API 中获取它们
func getAPIKeys() []apiKey {
	
	// 检查缓存中是否有结果
	if !apiKeysCache.lastLoad.IsZero() && time.Since(apiKeysCache.lastLoad) < 10*time.Minute {
		log.Println("Using cached API keys")
		return apiKeysCache.keys
	}

	// 创建 HTTP POST 请求
	req, err := http.NewRequest(http.MethodGet, Config.BaseUrl + "/api/keys", nil)
	if err != nil {
		log.Fatal("Error creating API keys request: ", err)
	}

	// 设置 HTTP 头
	req.Header.Set("Authorization", "Bearer " + Config.BaseKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error getting API keys: ", err)
	}
	defer resp.Body.Close()
	// 输出响应状态码
	log.Println("Using New API keys")

	// 解码响应
	if err := json.NewDecoder(resp.Body).Decode(&apiKeys); err != nil {
		log.Fatal("Error decoding API keys: ", err)
	}

	// 将结果存储在缓存中
	apiKeysCache.keys = apiKeys
	apiKeysCache.lastLoad = time.Now()

	return apiKeys
}

// getRandomAPIKey 返回一个随机的 API URL 和 API Key
func GetRandomAPIKey() (string, string) {
	keys := getAPIKeys()

	// 随机选择一个 API URL 和 API Key
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(keys))
	return keys[index].APIURL, keys[index].APIKey
}