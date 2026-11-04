package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RequestPayload 通用请求结构体
type RequestPayload struct {
	Title    string      `json:"title"`
	Content  string      `json:"content"`
	Text     string      `json:"text"`
	Msg      string      `json:"msg"`
	Message  string      `json:"message"`
	Channel  interface{} `json:"channel"`  // 支持 string 或 []string
	Channels interface{} `json:"channels"` // 支持 string 或 []string
	APIKey   string      `json:"api-key"`  // JSON 兼容 api-key
	APIKey2  string      `json:"api_key"`  // JSON 兼容 api_key
}

// ChannelResult 单个渠道推送结果
type ChannelResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ResponsePayload Webhook 响应结构体
type ResponsePayload struct {
	Code    int                      `json:"code"`
	Message string                   `json:"message"`
	Results map[string]ChannelResult `json:"results,omitempty"`
}

func main() {
	cfg := LoadConfig()

	// 打印已启用的渠道列表
	var enabledList []string
	for _, ch := range allChannels {
		if ch.IsEnabled(cfg) {
			enabledList = append(enabledList, ch.Name())
		}
	}
	log.Printf("[Startup] 服务启动成功，监听端口: %s\n", cfg.Port())
	if cfg.APIKey() != "" {
		log.Println("[Startup] 安全认证: 已启用 API-Key 校验")
	} else {
		log.Println("[Startup] 安全认证: 未配置 API_KEY，允许公开调用")
	}
	log.Printf("[Startup] 当前已配置并激活的渠道 (%d 个): %v\n", len(enabledList), enabledList)

	mux := http.NewServeMux()

	// 健康检查路由
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		renderJSON(w, http.StatusOK, map[string]interface{}{
			"status":           "ok",
			"enabled_channels": enabledList,
		})
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		renderJSON(w, http.StatusOK, map[string]interface{}{
			"service":          "webhook-notify",
			"status":           "running",
			"enabled_channels": enabledList,
		})
	})

	// Webhook / 推送路由（兼容 GET 与 POST）
	handler := func(w http.ResponseWriter, r *http.Request) {
		handleNotification(w, r, cfg)
	}

	mux.HandleFunc("GET /send", handler)
	mux.HandleFunc("POST /send", handler)
	mux.HandleFunc("GET /webhook", handler)
	mux.HandleFunc("POST /webhook", handler)

	server := &http.Server{
		Addr:         ":" + cfg.Port(),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 45 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server Listen error: %v\n", err)
	}
}

// handleNotification 核心请求处理逻辑
func handleNotification(w http.ResponseWriter, r *http.Request, cfg *Config) {
	var payload RequestPayload

	// 1. 解析请求数据
	if r.Method == http.MethodGet {
		q := r.URL.Query()
		payload.Title = q.Get("title")
		payload.Content = q.Get("content")
		if payload.Content == "" {
			payload.Content = q.Get("text")
		}
		if payload.Content == "" {
			payload.Content = q.Get("msg")
		}
		if payload.Content == "" {
			payload.Content = q.Get("message")
		}
		payload.Channel = q.Get("channel")
		if payload.Channel == "" {
			payload.Channel = q.Get("channels")
		}
		payload.APIKey = q.Get("api-key")
		if payload.APIKey == "" {
			payload.APIKey = q.Get("api_key")
		}
	} else if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil && len(bodyBytes) > 0 {
				_ = json.Unmarshal(bodyBytes, &payload)
			}
		} else {
			// 表单形式
			_ = r.ParseForm()
			payload.Title = r.FormValue("title")
			payload.Content = r.FormValue("content")
			if payload.Content == "" {
				payload.Content = r.FormValue("text")
			}
			if payload.Content == "" {
				payload.Content = r.FormValue("msg")
			}
			if payload.Content == "" {
				payload.Content = r.FormValue("message")
			}
			payload.Channel = r.FormValue("channel")
			if payload.Channel == "" {
				payload.Channel = r.FormValue("channels")
			}
			payload.APIKey = r.FormValue("api-key")
			if payload.APIKey == "" {
				payload.APIKey = r.FormValue("api_key")
			}
		}

		// 检查 URL query 是否有补全
		if payload.Title == "" {
			payload.Title = r.URL.Query().Get("title")
		}
		if payload.Content == "" {
			payload.Content = r.URL.Query().Get("content")
		}
		if payload.Channel == nil || payload.Channel == "" {
			payload.Channel = r.URL.Query().Get("channel")
		}
		if payload.APIKey == "" {
			payload.APIKey = r.URL.Query().Get("api-key")
		}
		if payload.APIKey == "" {
			payload.APIKey = r.URL.Query().Get("api_key")
		}
	}

	// 2. API-Key 安全校验
	expectedKey := cfg.APIKey()
	if expectedKey != "" {
		providedKey := payload.APIKey
		if providedKey == "" {
			providedKey = payload.APIKey2
		}
		if providedKey == "" {
			providedKey = r.Header.Get("X-API-Key")
		}
		if providedKey == "" {
			providedKey = r.Header.Get("api-key")
		}
		if providedKey == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				providedKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if providedKey != expectedKey {
			renderJSON(w, http.StatusUnauthorized, ResponsePayload{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized: invalid or missing api-key",
			})
			return
		}
	}

	// 3. 内容基础检验
	if payload.Title == "" && payload.Content == "" {
		renderJSON(w, http.StatusBadRequest, ResponsePayload{
			Code:    http.StatusBadRequest,
			Message: "title and content cannot both be empty",
		})
		return
	}
	if payload.Title == "" {
		payload.Title = "消息推送"
	}

	// 4. SKIP_PUSH_TITLE 检查
	skipTitles := cfg.Get("SKIP_PUSH_TITLE")
	if skipTitles != "" {
		for _, st := range strings.Split(skipTitles, "\n") {
			if strings.TrimSpace(st) == payload.Title {
				log.Printf("[Notify] 标题 %s 命中 SKIP_PUSH_TITLE，跳过推送\n", payload.Title)
				renderJSON(w, http.StatusOK, ResponsePayload{
					Code:    http.StatusOK,
					Message: fmt.Sprintf("Title '%s' skipped by SKIP_PUSH_TITLE", payload.Title),
				})
				return
			}
		}
	}

	// 5. 一言（HITOKOTO）追加
	finalContent := payload.Content
	if cfg.GetBool("HITOKOTO") {
		hCtx, hCancel := context.WithTimeout(r.Context(), 3*time.Second)
		quote := fetchHitokoto(hCtx)
		hCancel()
		if quote != "" {
			finalContent += "\n\n" + quote
		}
	}

	// 6. 解析需要推送的渠道列表
	targetChannels := extractTargetChannels(payload.Channel, payload.Channels)

	// 匹配需要推送的 Channel 对象
	var channelsToDispatch []Channel
	if len(targetChannels) > 0 {
		// 用户明确指定了渠道
		for _, target := range targetChannels {
			found := false
			for _, ch := range allChannels {
				if matchChannel(ch, target) {
					if ch.IsEnabled(cfg) {
						channelsToDispatch = append(channelsToDispatch, ch)
						found = true
					} else {
						log.Printf("[Notify] 渠道 %s 未配置对应环境变量，无法发送\n", ch.Name())
					}
					break
				}
			}
			if !found {
				log.Printf("[Notify] 未找到或未配置名为 %s 的推送渠道\n", target)
			}
		}
	} else {
		// 用户未指定渠道，默认推送所有已配置/已启用的渠道
		for _, ch := range allChannels {
			if ch.IsEnabled(cfg) {
				channelsToDispatch = append(channelsToDispatch, ch)
			}
		}
	}

	if len(channelsToDispatch) == 0 {
		renderJSON(w, http.StatusBadRequest, ResponsePayload{
			Code:    http.StatusBadRequest,
			Message: "No available or matching push channels configured",
		})
		return
	}

	// 7. 并发推送
	results := make(map[string]ChannelResult)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, ch := range channelsToDispatch {
		wg.Add(1)
		go func(c Channel) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			start := time.Now()
			err := c.Send(ctx, cfg, payload.Title, finalContent)
			duration := time.Since(start)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				log.Printf("[Notify][%s] 推送失败 (%v): %v\n", c.Name(), duration, err)
				results[c.Name()] = ChannelResult{
					Success: false,
					Message: err.Error(),
				}
			} else {
				log.Printf("[Notify][%s] 推送成功 (%v)\n", c.Name(), duration)
				results[c.Name()] = ChannelResult{
					Success: true,
					Message: "ok",
				}
			}
		}(ch)
	}

	wg.Wait()

	// 8. 汇总返回
	renderJSON(w, http.StatusOK, ResponsePayload{
		Code:    http.StatusOK,
		Message: "Notification dispatched",
		Results: results,
	})
}

// extractTargetChannels 从参数中提取目标渠道数组
func extractTargetChannels(c1, c2 interface{}) []string {
	var rawList []string
	add := func(v interface{}) {
		if v == nil {
			return
		}
		switch val := v.(type) {
		case string:
			for _, item := range strings.Split(val, ",") {
				if s := strings.TrimSpace(item); s != "" {
					rawList = append(rawList, strings.ToLower(s))
				}
			}
		case []interface{}:
			for _, item := range val {
				if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
					rawList = append(rawList, strings.ToLower(strings.TrimSpace(s)))
				}
			}
		case []string:
			for _, item := range val {
				if s := strings.TrimSpace(item); s != "" {
					rawList = append(rawList, strings.ToLower(s))
				}
			}
		}
	}

	add(c1)
	add(c2)
	return rawList
}

// matchChannel 匹配渠道名及别名
func matchChannel(ch Channel, name string) bool {
	name = strings.ToLower(name)
	if strings.ToLower(ch.Name()) == name {
		return true
	}
	for _, alias := range ch.Aliases() {
		if strings.ToLower(alias) == name {
			return true
		}
	}
	return false
}

// renderJSON 响应 JSON
func renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
