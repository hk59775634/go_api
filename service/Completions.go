package service

import (
	"log"
	"io"
	"io/ioutil"
	"time"
	"fmt"
	"github.com/google/uuid"
	"bytes"
	"regexp"
	"strings"
	"net/http"
	"encoding/json"
	"github.com/fatih/color"
	"github.com/pkoukk/tiktoken-go"
)


// CompletionRequest represents a request structure for chat completion API.
type CompletionRequest struct {
	Model            string                  `json:"model"`
	Prompt           string                `json:"prompt,omitempty"`
	Suffix           string                  `json:"suffix,omitempty"`
	MaxTokens        int                     `json:"max_tokens,omitempty"`
	Temperature      float32                 `json:"temperature,omitempty"`
	TopP             float32                 `json:"top_p,omitempty"`
	N                int                     `json:"n,omitempty"`
	Stream           bool                    `json:"stream,omitempty"`
	Logprobs         int                     `json:"logprobs,omitempty"`
	Echo             bool                    `json:"echo,omitempty"`
	Stop             []string                `json:"stop,omitempty"`
	PresencePenalty  float32                 `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32                 `json:"frequency_penalty,omitempty"`
	Best_of          int                     `json:"best_of,omitempty"`
	LogitBias        map[string]int          `json:"logit_bias,omitempty"`
	User             string                  `json:"user,omitempty"`
}


func Completions(w http.ResponseWriter, r *http.Request) {
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


	// 解析请求体以获取 stream 值
	var CompletionReq CompletionRequest
	err = json.Unmarshal(requestBodyBuf.Bytes(), &CompletionReq)
	if err != nil {
		log.Println("Error unmarshalling request body: ", err.Error())
		http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
		return
	}

	isStream := CompletionReq.Stream
	prompt := CompletionReq.Prompt
	// promptStr := strings.Join(prompt, "")
	model := CompletionReq.Model
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		err = fmt.Errorf("EncodingForModel: %v", err)
		fmt.Println(err)
		return
	}
	// prompt_token := NumTokensFromMessages(prompt,model)
	prompt_token := tkm.Encode(prompt, nil, nil)
	prompt_tokens := len(prompt_token)
	log.Println("prompt: ", green(prompt))
	log.Println("prompt_token: ", green(prompt_tokens))
    // 获取随机的 API URL 和 API Key
    target, apiKey := GetRandomAPIKey()
    // 拼接目标URL
    targetURL := target + "/v1/completions"
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

	// 根据 stream 值选择不同的处理方式
	if isStream {
		var fullResponse string
		buf := make([]byte, 1024)
		var fullContent strings.Builder
		for {
			if n, err := responseBody.Read(buf); err == io.EOF || n == 0 {
				break
			} else if err != nil {
				log.Println("Error while reading respbody: ", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else {
				fullResponse += string(buf[:n])
				fullContent.Write(buf[:n])

				if _, err = w.Write(buf[:n]); err != nil {
					log.Println("Error while writing resp: ", err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.(http.Flusher).Flush()
			}
		}
		// log.Println("Full response: ", green(fullResponse))

		// 提取 content 字段的值
		contentRegexp := regexp.MustCompile(`"text":"(.*?)"`)
		contentMatches := contentRegexp.FindAllStringSubmatch(fullContent.String(), -1)

		// 合并 content 字段的值
		var combinedContent strings.Builder
		for _, match := range contentMatches {
			if len(match) > 1 {
				combinedContent.WriteString(match[1])
			}
		}
		log.Println("Combined content: ", green(combinedContent.String()))

		tkm, err := tiktoken.EncodingForModel(model)
		if err != nil {
			err = fmt.Errorf("EncodingForModel: %v", err)
			fmt.Println(err)
			return
		}
		combined_token := tkm.Encode(combinedContent.String(), nil, nil)
		combined_tokens := len(combined_token)

		log.Println("Tokens used (stream mode): ", green(combined_tokens))
		total_tokens := prompt_tokens + combined_tokens
		log.Println("Total tokens used (stream mode): ", green(total_tokens))
		log.Println("model: ", green(model))
		StoreTokenUsage(userKey, sessionID, model, prompt_tokens, combined_tokens, total_tokens)
	} else {
		// 从响应体中解析 total_tokens
		bodyBytes, err := ioutil.ReadAll(responseBody)
		if err != nil {
			log.Println("Error reading response body: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var responseObj map[string]interface{}
		err = json.Unmarshal(bodyBytes, &responseObj)
		if err != nil {
			log.Println("Error unmarshalling response body: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		usage, ok := responseObj["usage"].(map[string]interface{})
		if !ok {
			log.Println("Error retrieving usage from response")
			http.Error(w, "Error retrieving usage from response", http.StatusInternalServerError)
			return
		}

		completiontokens, ok := usage["completion_tokens"].(float64)
		if !ok {
			log.Println("Error retrieving completion_tokens from response")
			http.Error(w, "Error retrieving completion_tokens from response", http.StatusInternalServerError)
			return
		}

		choicesArr, ok := responseObj["choices"].([]interface{})
		if !ok {
			log.Println("Error retrieving choices from response")
			http.Error(w, "Error retrieving choices from response", http.StatusInternalServerError)
			return
		}

		if len(choicesArr) < 1 {
			log.Println("No choices found in response")
			http.Error(w, "No choices found in response", http.StatusInternalServerError)
			return
		}

		choice, ok := choicesArr[0].(map[string]interface{})
		if !ok {
			log.Println("Error retrieving choice from response")
			http.Error(w, "Error retrieving choice from response", http.StatusInternalServerError)
			return
		}

		// message, ok := choice["message"].(map[string]interface{})
		// if !ok {
		// 	log.Println("Error retrieving message from response")
		// 	http.Error(w, "Error retrieving message from response", http.StatusInternalServerError)
		// 	return
		// }
		text, ok := choice["text"].(string)
		if !ok {
			log.Println("Error retrieving content from response")
			http.Error(w, "Error retrieving content from response", http.StatusInternalServerError)
			return
		}

		log.Println("Combined content: ", green(text))

		// 输出 tokens 使用情况
		log.Println("completion_tokens: ", green(int(completiontokens)))
		total_tokens := prompt_tokens + int(completiontokens)
		log.Println("Total tokens used: ", green(total_tokens))
		log.Println("model: ", green(model))
		StoreTokenUsage(userKey, sessionID, model, prompt_tokens, int(completiontokens), total_tokens)
		// 将响应写回客户端
		_, err = w.Write(bodyBytes)
		if err != nil {
			log.Println("Error writing response body: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

