package service

import (
	"log"
	"net/http"
	"time"
	"github.com/fatih/color"
)
// tokenCache 用于缓存 token 的有效性
var tokenCache = make(map[string]cacheEntry)
//  fakeTokenCache 用于缓存 无效 token 。减少对平台的请求
var fakeTokenCache = make(map[string]cacheEntry)

// cacheEntry 包含一个标志，用于指示 token 是否有效，以及最后一次缓存的时间
type cacheEntry struct {
	valid  bool
	expiry time.Time
}

// isValidToken 检查 token 是否合法
func IsValidToken(token string) bool {
	// 设置一个绿色的颜色函数，用于打印日志
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// 检查缓存中是否有结果
	entry, ok := tokenCache[token]
	if ok && time.Now().Before(entry.expiry) {
		log.Println("缓存命中用户token: \n", green(token))
		return entry.valid
	}
	// 检查缓存中是否有结果
	entry, ok = fakeTokenCache[token]
	if ok && time.Now().Before(entry.expiry) {
		log.Println("缓存命中无效token: \n", red(token))
		log.Println(red("==========================验证失败。停止响应======================================="))
		return false
	}

	// 如果token不是以ns_开头，直接返回false
	if len(token) < 3 || token[:3] != "ns_" {
		log.Println("格式错误的非法token: \n", red(token))
		log.Println(red("==========================验证失败。停止响应======================================="))
		return false
	}

	// 使用 Config.BaseUrl + "/api/check" 构造一个请求
	req, err := http.NewRequest(http.MethodPost, Config.BaseUrl + "/api/check", nil)
	if err != nil {
		log.Println("Error creating token validation request: ", err.Error())
		return false
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error sending token validation request: ", err.Error())
		return false
	}
	defer resp.Body.Close()

	valid := resp.StatusCode == http.StatusOK
	if valid {
		// 将结果存储在缓存中
		entry = cacheEntry{
			valid:  valid,
			expiry: time.Now().Add(10 * time.Minute),
		}
		tokenCache[token] = entry
		log.Println("非缓存命中用户token: \n", green(token))
	}else{
		// 将结果存储在缓存中
		entry = cacheEntry{
			valid:  valid,
			expiry: time.Now().Add(10 * time.Minute),
		}
		fakeTokenCache[token] = entry
		log.Println("经查询为非法token: \n", red(token))
		log.Println(red("==========================验证失败。停止响应======================================="))
	}
	return valid
}