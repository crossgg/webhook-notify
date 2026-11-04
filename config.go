package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Config 管理全局配置项
type Config struct {
	mu     sync.RWMutex
	values map[string]string
}

var globalConfig *Config

// NewConfig 创建新的配置对象
func NewConfig() *Config {
	return &Config{
		values: make(map[string]string),
	}
}

// LoadConfig 按优先级加载配置：
// 1. 默认值
// 2. config.json 文件（若存在）
// 3. 环境变量（优先级最高，覆盖文件中的同名配置）
func LoadConfig() *Config {
	cfg := NewConfig()

	// 1. 默认值
	cfg.Set("PORT", "5755")
	cfg.Set("HITOKOTO", "false")
	cfg.Set("CONSOLE", "false")
	cfg.Set("PUSH_PLUS_TEMPLATE", "html")
	cfg.Set("PUSH_PLUS_CHANNEL", "wechat")
	cfg.Set("WE_PLUS_BOT_VERSION", "pro")
	cfg.Set("WPUSH_CHANNEL", "wechat")
	cfg.Set("NTFY_URL", "https://ntfy.sh")
	cfg.Set("NTFY_PRIORITY", "3")
	cfg.Set("QYWX_ORIGIN", "https://qyapi.weixin.qq.com")
	cfg.Set("SMTP_SSL", "false")

	// 2. 读取配置文件 config.json
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		candidates := []string{
			"config.json",
			"/app/config.json",
			filepath.Join(".", "config.json"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				configPath = c
				break
			}
		}
	}

	if configPath != "" {
		if data, err := os.ReadFile(configPath); err == nil {
			var rawMap map[string]interface{}
			if err := json.Unmarshal(data, &rawMap); err == nil {
				log.Printf("[Config] 成功加载配置文件: %s\n", configPath)
				for k, v := range rawMap {
					normKey := strings.ToUpper(strings.ReplaceAll(k, "-", "_"))
					strVal := fmt.Sprintf("%v", v)
					cfg.Set(normKey, strVal)
				}
			} else {
				log.Printf("[Config] 解析 JSON 文件失败: %v\n", err)
			}
		}
	}

	// 3. 环境变量覆盖（优先级最高）
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 && parts[1] != "" {
			normKey := strings.ToUpper(strings.ReplaceAll(parts[0], "-", "_"))
			cfg.Set(normKey, parts[1])
		}
	}

	globalConfig = cfg
	return cfg
}

func (c *Config) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	normKey := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
	c.values[normKey] = strings.TrimSpace(value)
}

func (c *Config) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	normKey := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
	return c.values[normKey]
}

func (c *Config) GetBool(key string) bool {
	val := strings.ToLower(c.Get(key))
	return val == "true" || val == "1" || val == "yes"
}

func (c *Config) GetInt(key string, def int) int {
	val := c.Get(key)
	if val == "" {
		return def
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return n
}

func (c *Config) Has(key string) bool {
	return c.Get(key) != ""
}

// APIKey 获取配置的 api-key
func (c *Config) APIKey() string {
	if k := c.Get("API_KEY"); k != "" {
		return k
	}
	return c.Get("APIKEY")
}

// Port 获取服务监听端口
func (c *Config) Port() string {
	if p := c.Get("PORT"); p != "" {
		return p
	}
	return "5755"
}
