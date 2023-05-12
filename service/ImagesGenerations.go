package service

import (
	"log"
	"io"
	"io/ioutil"
	"time"
	"bytes"
	"strings"
	"net/http"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/fatih/color"
)

type ImagesRequest struct {
	Prompt           string                  `json:"prompt"`
	N                int                     `json:"n,omitempty"`
	Size             string                  `json:"size,omitempty"`
	ResponseFormat   string                  `json:"response_format,omitempty"`
	User             string                  `json:"user,omitempty"`
}

// ImagesGenerations 生成图片
func ImagesGenerations(w http.ResponseWriter, r *http.Request) {
	log.Println("==========================新的请求处理过程=========================================")
	// 设置一个绿色的颜色函数，用于打印日志
	green := color.New(color.FgGreen).SprintFunc()

	// 检查 token 是否合法
	userKey := strings.Split(r.Header.Get("Authorization"), " ")[1]
    if !IsValidToken(userKey) {
		// 返回一个json格式的错误信息. 包含401状态码和错误信息
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
        return
    }

	// 为每一个请求创建一个新的会话id.暂时无用
	sessionID := uuid.New().String()
	log.Println("Session ID: ", green(sessionID))	

	// 复制请求体到缓冲区
	var requestBodyBuf bytes.Buffer
	_, err := io.Copy(&requestBodyBuf, r.Body)
	if err != nil {
		log.Println("Error copying request body to buffer: ", err.Error())
		http.Error(w, "Error copying request body to buffer", http.StatusInternalServerError)
		return
	}
	requestBody := bytes.NewReader(requestBodyBuf.Bytes())
	defer r.Body.Close()

	// 记录请求体相关信息.暂时无用
	log.Println("User request IP:PORT : ", green(r.RemoteAddr))
	log.Println("User request body(json): ", green(requestBodyBuf.String()))

	// 解析请求体以获取 参数
	var ImagesReq ImagesRequest
	err = json.Unmarshal(requestBodyBuf.Bytes(), &ImagesReq)
	if err != nil {
		log.Println("Error unmarshalling request body: ", err.Error())
		http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
		return
	}
	prompt := ImagesReq.Prompt
	log.Println("Prompt: ", green(prompt))
	n := ImagesReq.N
	log.Println("N: ", green(n))
	size := ImagesReq.Size
	log.Println("Size: ", green(size))
	response_format := ImagesReq.ResponseFormat
	log.Println("ResponseFormat: ", green(response_format))
	// 根据size值不同定价不同。1024×1024 0.02美元，512x512 0.018美元，256×256 0.016美元
	// 成本 := size对应的美元乘以n。
	if size == "1024x1024" {
		Images_cost := 2000000 * n
		// 打印计算出的成本到控制台
		log.Println("Cost: ", green(Images_cost))
	} else if size == "512x512" {
		Images_cost := 1800000 * n
		// 打印计算出的成本到控制台
		log.Println("Cost: ", green(Images_cost))
	} else if size == "256x256" {
		Images_cost := 1600000 * n
		// 打印计算出的成本到控制台
		log.Println("Cost: ", green(Images_cost))
	} else {
		Images_cost := 2000000 * n
		// 打印计算出的成本到控制台
		log.Println("Cost: ", green(Images_cost))
	}


	// 获取随机的 API URL 和 API Key
	target, apiKey := GetRandomAPIKey()
	// 拼接目标URL
	targetURL := target + "/v1/images/generations"
	log.Println("Using API url: ", green(targetURL))
	log.Println("Using API key: ", green(apiKey))
  

    // 创建OpenAI API HTTP请求
    proxyReq, err := http.NewRequest(r.Method, targetURL, requestBody)
    if err != nil {
        log.Println("Error creating proxy request: ", err.Error())
        http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
        return
    }
    // 将原始请求头复制到新请求中
    for headerKey, headerValues := range r.Header {
        for _, headerValue := range headerValues {
            proxyReq.Header.Add(headerKey, headerValue)
        }
    }
	// 修改请求头中的Authorization字段。换成key池中的随机key
    proxyReq.Header.Set("Authorization", "Bearer "+apiKey)

    // 默认超时时间设置为60s
    client := &http.Client{
        Timeout: 60 * time.Second,
    }

    // 向 OpenAI 发起 API 请求
    resp, err := client.Do(proxyReq)
    if err != nil {
        log.Println("Error sending proxy request: ", err.Error())
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    // 复制响应体到缓冲区
    var responseBodyBuf bytes.Buffer
    responseBody := io.TeeReader(resp.Body, &responseBodyBuf)

    // 将响应头复制到代理响应头中
    for key, values := range resp.Header {
        for _, value := range values {
            w.Header().Add(key, value)
        }
    }

    // 将响应状态码设置为原始
	w.WriteHeader(resp.StatusCode)

	// 响应体
	bodyBytes, err := ioutil.ReadAll(responseBody)
	if err != nil {
		log.Println("Error reading response body: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 将响应写回客户端
	_, err = w.Write(bodyBytes)
	if err != nil {
		log.Println("Error writing response body: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 记账
	StoreTokenUsage(userKey, sessionID, size, n, n, n)
}
