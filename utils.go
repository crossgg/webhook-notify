package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var defaultHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
}

// encodeRFC2047 按照 RFC 2047 进行 UTF-8 Base64 编码
func encodeRFC2047(text string) string {
	b64 := base64.StdEncoding.EncodeToString([]byte(text))
	return fmt.Sprintf("=?utf-8?B?%s?=", b64)
}

// hmacSha256 计算 HMAC-SHA256
func hmacSha256(key, message []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(message)
	return h.Sum(nil)
}

// fetchHitokoto 获取一条随机一言
func fetchHitokoto(ctx context.Context) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://v1.hitokoto.cn/", nil)
	if err != nil {
		return ""
	}
	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var data struct {
		Hitokoto string `json:"hitokoto"`
		From     string `json:"from"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return ""
	}
	if data.Hitokoto == "" {
		return ""
	}
	return fmt.Sprintf("%s    ----%s", data.Hitokoto, data.From)
}

// postJSON 发送 JSON 请求
func postJSON(ctx context.Context, client *http.Client, reqURL string, headers map[string]string, payload interface{}) ([]byte, int, error) {
	if client == nil {
		client = defaultHTTPClient
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return respBody, resp.StatusCode, nil
}

// postForm 发送表单请求 (x-www-form-urlencoded)
func postForm(ctx context.Context, client *http.Client, reqURL string, headers map[string]string, data url.Values) ([]byte, int, error) {
	if client == nil {
		client = defaultHTTPClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return respBody, resp.StatusCode, nil
}

// createProxyHTTPClient 为需要代理的服务（如 Telegram）创建 Client
func createProxyHTTPClient(proxyHost, proxyPort, proxyAuth string) *http.Client {
	if proxyHost == "" || proxyPort == "" {
		return defaultHTTPClient
	}
	var proxyURLStr string
	if proxyAuth != "" && !strings.Contains(proxyHost, "@") {
		proxyURLStr = fmt.Sprintf("http://%s@%s:%s", proxyAuth, proxyHost, proxyPort)
	} else {
		proxyURLStr = fmt.Sprintf("http://%s:%s", proxyHost, proxyPort)
	}

	parsedURL, err := url.Parse(proxyURLStr)
	if err != nil {
		return defaultHTTPClient
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(parsedURL),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	return &http.Client{
		Transport: transport,
		Timeout:   20 * time.Second,
	}
}
