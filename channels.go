package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Channel 推送渠道通用接口
type Channel interface {
	Name() string
	Aliases() []string
	IsEnabled(cfg *Config) bool
	Send(ctx context.Context, cfg *Config, title, content string) error
}

// 注册所有渠道
var allChannels = []Channel{
	&BarkChannel{},
	&DingTalkChannel{},
	&FeishuChannel{},
	&WeComBotChannel{},
	&WeComAppChannel{},
	&TelegramChannel{},
	&PushPlusChannel{},
	&PushDeerChannel{},
	&ServerJChannel{},
	&NtfyChannel{},
	&WxPusherBotChannel{},
	&WxPusherSPTChannel{},
	&SMTPChannel{},
	&GoCQHTTPChannel{},
	&GotifyChannel{},
	&IGotChannel{},
	&SynologyChatChannel{},
	&WePlusChannel{},
	&QmsgChannel{},
	&AibotkChannel{},
	&PushMeChannel{},
	&ChronocatChannel{},
	&WpushChannel{},
	&OpeniLinkChannel{},
	&CustomWebhookChannel{},
	&ConsoleChannel{},
}

// ----------------------------------------------------
// 1. Bark
// ----------------------------------------------------
type BarkChannel struct{}

func (b *BarkChannel) Name() string       { return "bark" }
func (b *BarkChannel) Aliases() []string { return []string{"bark_push"} }
func (b *BarkChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("BARK_PUSH")
}
func (b *BarkChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	rawURL := cfg.Get("BARK_PUSH")
	var reqURL string
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		reqURL = rawURL
	} else {
		reqURL = "https://api.day.app/" + rawURL
	}

	payload := map[string]interface{}{
		"title": title,
		"body":  content,
	}
	if v := cfg.Get("BARK_ARCHIVE"); v != "" {
		payload["isArchive"] = v
	}
	if v := cfg.Get("BARK_GROUP"); v != "" {
		payload["group"] = v
	}
	if v := cfg.Get("BARK_SOUND"); v != "" {
		payload["sound"] = v
	}
	if v := cfg.Get("BARK_ICON"); v != "" {
		payload["icon"] = v
	}
	if v := cfg.Get("BARK_LEVEL"); v != "" {
		payload["level"] = v
	}
	if v := cfg.Get("BARK_URL"); v != "" {
		payload["url"] = v
	}

	body, status, err := postJSON(ctx, defaultHTTPClient, reqURL, nil, payload)
	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("HTTP %d: %s", status, string(body))
	}
	var res struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(body, &res); err == nil && res.Code != 200 {
		return fmt.Errorf("bark code: %d, body: %s", res.Code, string(body))
	}
	return nil
}

// ----------------------------------------------------
// 2. 钉钉机器人 (DingTalk)
// ----------------------------------------------------
type DingTalkChannel struct{}

func (d *DingTalkChannel) Name() string       { return "dingtalk" }
func (d *DingTalkChannel) Aliases() []string { return []string{"dingding", "dingding_bot", "dd"} }
func (d *DingTalkChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("DD_BOT_TOKEN") && cfg.Has("DD_BOT_SECRET")
}
func (d *DingTalkChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	secret := cfg.Get("DD_BOT_SECRET")
	stringToSign := timestamp + "\n" + secret
	hmacCode := hmacSha256([]byte(secret), []byte(stringToSign))
	sign := url.QueryEscape(base64.StdEncoding.EncodeToString(hmacCode))

	reqURL := fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s&timestamp=%s&sign=%s",
		cfg.Get("DD_BOT_TOKEN"), timestamp, sign)

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("%s\n\n%s", title, content),
		},
	}

	body, _, err := postJSON(ctx, defaultHTTPClient, reqURL, nil, payload)
	if err != nil {
		return err
	}
	var res struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.ErrCode != 0 {
		return fmt.Errorf("dingtalk errcode: %d, msg: %s", res.ErrCode, res.ErrMsg)
	}
	return nil
}

// ----------------------------------------------------
// 3. 飞书机器人 (Feishu)
// ----------------------------------------------------
type FeishuChannel struct{}

func (f *FeishuChannel) Name() string       { return "feishu" }
func (f *FeishuChannel) Aliases() []string { return []string{"feishu_bot", "fs"} }
func (f *FeishuChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("FSKEY")
}
func (f *FeishuChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	reqURL := fmt.Sprintf("https://open.feishu.cn/open-apis/bot/v2/hook/%s", cfg.Get("FSKEY"))
	payload := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": fmt.Sprintf("%s\n\n%s", title, content),
		},
	}

	if secret := cfg.Get("FSSECRET"); secret != "" {
		timestamp := fmt.Sprintf("%d", time.Now().Unix())
		stringToSign := timestamp + "\n" + secret
		// 飞书签名：key为 stringToSign，消息体为空
		hmacCode := hmacSha256([]byte(stringToSign), []byte(""))
		sign := base64.StdEncoding.EncodeToString(hmacCode)
		payload["timestamp"] = timestamp
		payload["sign"] = sign
	}

	body, _, err := postJSON(ctx, defaultHTTPClient, reqURL, nil, payload)
	if err != nil {
		return err
	}
	var res struct {
		Code       int    `json:"code"`
		StatusCode int    `json:"StatusCode"`
		Msg        string `json:"msg"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.Code != 0 && res.StatusCode != 0 {
		return fmt.Errorf("feishu code: %d, msg: %s", res.Code, res.Msg)
	}
	return nil
}

// ----------------------------------------------------
// 4. 企业微信机器人 (WeCom Bot)
// ----------------------------------------------------
type WeComBotChannel struct{}

func (w *WeComBotChannel) Name() string       { return "wecom_bot" }
func (w *WeComBotChannel) Aliases() []string { return []string{"wecom", "qywx_bot", "qywx"} }
func (w *WeComBotChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("QYWX_KEY")
}
func (w *WeComBotChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	origin := cfg.Get("QYWX_ORIGIN")
	if origin == "" {
		origin = "https://qyapi.weixin.qq.com"
	}
	reqURL := fmt.Sprintf("%s/cgi-bin/webhook/send?key=%s", origin, cfg.Get("QYWX_KEY"))

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("%s\n\n%s", title, content),
		},
	}

	body, _, err := postJSON(ctx, defaultHTTPClient, reqURL, nil, payload)
	if err != nil {
		return err
	}
	var res struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.ErrCode != 0 {
		return fmt.Errorf("wecom_bot errcode: %d, msg: %s", res.ErrCode, res.ErrMsg)
	}
	return nil
}

// ----------------------------------------------------
// 5. 企业微信应用 (WeCom App)
// ----------------------------------------------------
type WeComAppChannel struct{}

func (w *WeComAppChannel) Name() string       { return "wecom_app" }
func (w *WeComAppChannel) Aliases() []string { return []string{"qywx_app", "wx_app"} }
func (w *WeComAppChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("QYWX_AM")
}
func (w *WeComAppChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	origin := cfg.Get("QYWX_ORIGIN")
	if origin == "" {
		origin = "https://qyapi.weixin.qq.com"
	}

	parts := strings.Split(cfg.Get("QYWX_AM"), ",")
	if len(parts) < 4 {
		return errors.New("QYWX_AM format error: expected corpid,corpsecret,touser,agentid[,media_id]")
	}
	corpID := parts[0]
	corpSecret := parts[1]
	toUser := parts[2]
	agentID := parts[3]
	mediaID := ""
	if len(parts) >= 5 {
		mediaID = parts[4]
	}

	// 1. 获取 token
	tokenURL := fmt.Sprintf("%s/cgi-bin/gettoken?corpid=%s&corpsecret=%s", origin, corpID, corpSecret)
	tokenResp, _, err := postJSON(ctx, defaultHTTPClient, tokenURL, nil, nil)
	if err != nil {
		return fmt.Errorf("gettoken failed: %w", err)
	}
	var tokenRes struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(tokenResp, &tokenRes); err != nil {
		return err
	}
	if tokenRes.AccessToken == "" {
		return fmt.Errorf("gettoken error: %d, %s", tokenRes.ErrCode, tokenRes.ErrMsg)
	}

	// 2. 发送消息
	sendURL := fmt.Sprintf("%s/cgi-bin/message/send?access_token=%s", origin, tokenRes.AccessToken)
	var payload map[string]interface{}
	if mediaID == "" {
		payload = map[string]interface{}{
			"touser":  toUser,
			"msgtype": "text",
			"agentid": agentID,
			"text": map[string]string{
				"content": fmt.Sprintf("%s\n\n%s", title, content),
			},
			"safe": "0",
		}
	} else {
		payload = map[string]interface{}{
			"touser":  toUser,
			"msgtype": "mpnews",
			"agentid": agentID,
			"mpnews": map[string]interface{}{
				"articles": []map[string]string{
					{
						"title":          title,
						"thumb_media_id": mediaID,
						"author":         "Notification",
						"content":        strings.ReplaceAll(content, "\n", "<br/>"),
						"digest":         content,
					},
				},
			},
		}
	}

	sendResp, _, err := postJSON(ctx, defaultHTTPClient, sendURL, nil, payload)
	if err != nil {
		return err
	}
	var res struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(sendResp, &res); err != nil {
		return err
	}
	if res.ErrCode != 0 {
		return fmt.Errorf("wecom_app errcode: %d, msg: %s", res.ErrCode, res.ErrMsg)
	}
	return nil
}

// ----------------------------------------------------
// 6. Telegram Bot
// ----------------------------------------------------
type TelegramChannel struct{}

func (t *TelegramChannel) Name() string       { return "telegram" }
func (t *TelegramChannel) Aliases() []string { return []string{"tg", "telegram_bot"} }
func (t *TelegramChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("TG_BOT_TOKEN") && cfg.Has("TG_USER_ID")
}
func (t *TelegramChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	client := createProxyHTTPClient(
		cfg.Get("TG_PROXY_HOST"),
		cfg.Get("TG_PROXY_PORT"),
		cfg.Get("TG_PROXY_AUTH"),
	)

	apiHost := cfg.Get("TG_API_HOST")
	if apiHost == "" {
		apiHost = "https://api.telegram.org"
	}
	reqURL := fmt.Sprintf("%s/bot%s/sendMessage", strings.TrimRight(apiHost, "/"), cfg.Get("TG_BOT_TOKEN"))

	form := url.Values{}
	form.Set("chat_id", cfg.Get("TG_USER_ID"))
	form.Set("text", fmt.Sprintf("%s\n\n%s", title, content))
	form.Set("disable_web_page_preview", "true")

	body, _, err := postForm(ctx, client, reqURL, nil, form)
	if err != nil {
		return err
	}
	var res struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if !res.OK {
		return fmt.Errorf("telegram error: %s", res.Description)
	}
	return nil
}

// ----------------------------------------------------
// 7. PushPlus
// ----------------------------------------------------
type PushPlusChannel struct{}

func (p *PushPlusChannel) Name() string       { return "pushplus" }
func (p *PushPlusChannel) Aliases() []string { return []string{"pushplus_bot"} }
func (p *PushPlusChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("PUSH_PLUS_TOKEN")
}
func (p *PushPlusChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	payload := map[string]interface{}{
		"token":       cfg.Get("PUSH_PLUS_TOKEN"),
		"title":       title,
		"content":     content,
		"topic":       cfg.Get("PUSH_PLUS_USER"),
		"template":    cfg.Get("PUSH_PLUS_TEMPLATE"),
		"channel":     cfg.Get("PUSH_PLUS_CHANNEL"),
		"webhook":     cfg.Get("PUSH_PLUS_WEBHOOK"),
		"callbackUrl": cfg.Get("PUSH_PLUS_CALLBACKURL"),
		"to":          cfg.Get("PUSH_PLUS_TO"),
	}

	body, _, err := postJSON(ctx, defaultHTTPClient, "https://www.pushplus.plus/send", nil, payload)
	if err == nil {
		var res struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
		if err := json.Unmarshal(body, &res); err == nil && res.Code == 200 {
			return nil
		}
	}

	// 备选地址
	body, _, err = postJSON(ctx, defaultHTTPClient, "http://pushplus.hxtrip.com/send", nil, payload)
	if err != nil {
		return err
	}
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.Code != 200 {
		return fmt.Errorf("pushplus code: %d, msg: %s", res.Code, res.Msg)
	}
	return nil
}

// ----------------------------------------------------
// 8. PushDeer
// ----------------------------------------------------
type PushDeerChannel struct{}

func (p *PushDeerChannel) Name() string       { return "pushdeer" }
func (p *PushDeerChannel) Aliases() []string { return []string{"deer"} }
func (p *PushDeerChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("DEER_KEY")
}
func (p *PushDeerChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	reqURL := cfg.Get("DEER_URL")
	if reqURL == "" {
		reqURL = "https://api2.pushdeer.com/message/push"
	}

	form := url.Values{}
	form.Set("text", title)
	form.Set("desp", content)
	form.Set("type", "markdown")
	form.Set("pushkey", cfg.Get("DEER_KEY"))

	body, _, err := postForm(ctx, defaultHTTPClient, reqURL, nil, form)
	if err != nil {
		return err
	}
	var res struct {
		Content struct {
			Result []interface{} `json:"result"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if len(res.Content.Result) == 0 {
		return fmt.Errorf("pushdeer failed: %s", string(body))
	}
	return nil
}

// ----------------------------------------------------
// 9. Server酱 (ServerJ)
// ----------------------------------------------------
type ServerJChannel struct{}

func (s *ServerJChannel) Name() string       { return "serverj" }
func (s *ServerJChannel) Aliases() []string { return []string{"serverchan", "ftqq"} }
func (s *ServerJChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("PUSH_KEY")
}
func (s *ServerJChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	pushKey := cfg.Get("PUSH_KEY")
	var reqURL string

	re := regexp.MustCompile(`sctp(\d+)t`)
	matches := re.FindStringSubmatch(pushKey)
	if len(matches) > 1 {
		reqURL = fmt.Sprintf("https://%s.push.ft07.com/send/%s.send", matches[1], pushKey)
	} else {
		reqURL = fmt.Sprintf("https://sctapi.ftqq.com/%s.send", pushKey)
	}

	form := url.Values{}
	form.Set("text", title)
	form.Set("desp", strings.ReplaceAll(content, "\n", "\n\n"))

	body, _, err := postForm(ctx, defaultHTTPClient, reqURL, nil, form)
	if err != nil {
		return err
	}
	var res struct {
		Code    int    `json:"code"`
		Errno   int    `json:"errno"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.Code != 0 && res.Errno != 0 {
		return fmt.Errorf("serverj error: %s", res.Message)
	}
	return nil
}

// ----------------------------------------------------
// 10. ntfy
// ----------------------------------------------------
type NtfyChannel struct{}

func (n *NtfyChannel) Name() string       { return "ntfy" }
func (n *NtfyChannel) Aliases() []string { return []string{} }
func (n *NtfyChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("NTFY_TOPIC")
}
func (n *NtfyChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	baseURL := strings.TrimRight(cfg.Get("NTFY_URL"), "/")
	if baseURL == "" {
		baseURL = "https://ntfy.sh"
	}
	reqURL := fmt.Sprintf("%s/%s", baseURL, cfg.Get("NTFY_TOPIC"))

	priority := cfg.Get("NTFY_PRIORITY")
	if priority == "" {
		priority = "3"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(content))
	if err != nil {
		return err
	}
	req.Header.Set("Title", encodeRFC2047(title))
	req.Header.Set("Priority", priority)
	req.Header.Set("Icon", "https://qn.whyour.cn/logo.png")

	if token := cfg.Get("NTFY_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else if u, p := cfg.Get("NTFY_USERNAME"), cfg.Get("NTFY_PASSWORD"); u != "" && p != "" {
		authStr := base64.StdEncoding.EncodeToString([]byte(u + ":" + p))
		req.Header.Set("Authorization", "Basic "+authStr)
	}
	if actions := cfg.Get("NTFY_ACTIONS"); actions != "" {
		req.Header.Set("Actions", encodeRFC2047(actions))
	}

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ntfy HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ----------------------------------------------------
// 11. WxPusher (AppToken 方式)
// ----------------------------------------------------
type WxPusherBotChannel struct{}

func (w *WxPusherBotChannel) Name() string       { return "wxpusher" }
func (w *WxPusherBotChannel) Aliases() []string { return []string{"wxpusher_bot"} }
func (w *WxPusherBotChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("WXPUSHER_APP_TOKEN") && (cfg.Has("WXPUSHER_TOPIC_IDS") || cfg.Has("WXPUSHER_UIDS"))
}
func (w *WxPusherBotChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	var topicIDs []int
	for _, idStr := range strings.Split(cfg.Get("WXPUSHER_TOPIC_IDS"), ";") {
		if s := strings.TrimSpace(idStr); s != "" {
			if id, err := strconv.Atoi(s); err == nil {
				topicIDs = append(topicIDs, id)
			}
		}
	}

	var uids []string
	for _, uid := range strings.Split(cfg.Get("WXPUSHER_UIDS"), ";") {
		if s := strings.TrimSpace(uid); s != "" {
			uids = append(uids, s)
		}
	}

	payload := map[string]interface{}{
		"appToken":      cfg.Get("WXPUSHER_APP_TOKEN"),
		"content":       fmt.Sprintf("<h1>%s</h1><br/><div style='white-space: pre-wrap;'>%s</div>", title, content),
		"summary":       title,
		"contentType":   2,
		"topicIds":      topicIDs,
		"uids":          uids,
		"verifyPayType": 0,
	}

	body, _, err := postJSON(ctx, defaultHTTPClient, "https://wxpusher.zjiecode.com/api/send/message", nil, payload)
	if err != nil {
		return err
	}
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.Code != 1000 {
		return fmt.Errorf("wxpusher code: %d, msg: %s", res.Code, res.Msg)
	}
	return nil
}

// ----------------------------------------------------
// 12. WxPusher (SPT 极简方式)
// ----------------------------------------------------
type WxPusherSPTChannel struct{}

func (w *WxPusherSPTChannel) Name() string       { return "wxpusher_spt" }
func (w *WxPusherSPTChannel) Aliases() []string { return []string{} }
func (w *WxPusherSPTChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("WXPUSHER_SPT_LIST")
}
func (w *WxPusherSPTChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	var spts []string
	for _, item := range strings.Split(cfg.Get("WXPUSHER_SPT_LIST"), ",") {
		if s := strings.TrimSpace(item); s != "" {
			spts = append(spts, s)
		}
	}
	if len(spts) == 0 {
		return errors.New("WXPUSHER_SPT_LIST is empty")
	}

	payload := map[string]interface{}{
		"content":     fmt.Sprintf("<h1>%s</h1><br/><div style='white-space: pre-wrap;'>%s</div>", title, content),
		"summary":     title,
		"contentType": 2,
	}
	if len(spts) == 1 {
		payload["spt"] = spts[0]
	} else {
		payload["sptList"] = spts
	}

	body, _, err := postJSON(ctx, defaultHTTPClient, "https://wxpusher.zjiecode.com/api/send/message/simple-push", nil, payload)
	if err != nil {
		return err
	}
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.Code != 1000 {
		return fmt.Errorf("wxpusher spt code: %d, msg: %s", res.Code, res.Msg)
	}
	return nil
}

// ----------------------------------------------------
// 13. SMTP 邮件
// ----------------------------------------------------
type SMTPChannel struct{}

func (s *SMTPChannel) Name() string       { return "smtp" }
func (s *SMTPChannel) Aliases() []string { return []string{"mail", "email"} }
func (s *SMTPChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("SMTP_SERVER") && cfg.Has("SMTP_EMAIL") && cfg.Has("SMTP_PASSWORD")
}
func (s *SMTPChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	smtpServer := cfg.Get("SMTP_SERVER")
	host, port, err := net.SplitHostPort(smtpServer)
	if err != nil {
		host = smtpServer
		if cfg.GetBool("SMTP_SSL") {
			port = "465"
		} else {
			port = "25"
		}
		smtpServer = net.JoinHostPort(host, port)
	}

	fromEmail := cfg.Get("SMTP_EMAIL")
	password := cfg.Get("SMTP_PASSWORD")
	fromName := cfg.Get("SMTP_NAME")
	if fromName == "" {
		fromName = fromEmail
	}

	toEmailRaw := cfg.Get("SMTP_EMAIL_TO")
	if toEmailRaw == "" {
		toEmailRaw = fromEmail
	}
	var toEmails []string
	for _, e := range strings.FieldsFunc(toEmailRaw, func(r rune) bool { return r == ';' || r == '；' || r == ',' }) {
		if s := strings.TrimSpace(e); s != "" {
			toEmails = append(toEmails, s)
		}
	}
	if len(toEmails) == 0 {
		toEmails = []string{fromEmail}
	}

	headerFrom := fmt.Sprintf("%s <%s>", encodeRFC2047(fromName), fromEmail)
	headerTo := strings.Join(toEmails, ", ")
	headerSubject := encodeRFC2047(title)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		headerFrom, headerTo, headerSubject, content)

	auth := smtp.PlainAuth("", fromEmail, password, host)

	if cfg.GetBool("SMTP_SSL") {
		tlsConfig := &tls.Config{
			ServerName: host,
		}
		conn, err := tls.Dial("tcp", smtpServer, tlsConfig)
		if err != nil {
			return fmt.Errorf("tls.Dial failed: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return err
		}
		defer client.Close()

		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("auth failed: %w", err)
		}
		if err := client.Mail(fromEmail); err != nil {
			return err
		}
		for _, addr := range toEmails {
			if err := client.Rcpt(addr); err != nil {
				return err
			}
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(msg)); err != nil {
			return err
		}
		return w.Close()
	}

	return smtp.SendMail(smtpServer, auth, fromEmail, toEmails, []byte(msg))
}

// ----------------------------------------------------
// 14. go-cqhttp
// ----------------------------------------------------
type GoCQHTTPChannel struct{}

func (g *GoCQHTTPChannel) Name() string       { return "go_cqhttp" }
func (g *GoCQHTTPChannel) Aliases() []string { return []string{"gobot", "qqbot"} }
func (g *GoCQHTTPChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("GOBOT_URL") && cfg.Has("GOBOT_QQ")
}
func (g *GoCQHTTPChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	reqURL := fmt.Sprintf("%s?access_token=%s&%s&message=%s",
		cfg.Get("GOBOT_URL"),
		url.QueryEscape(cfg.Get("GOBOT_TOKEN")),
		cfg.Get("GOBOT_QQ"),
		url.QueryEscape(fmt.Sprintf("标题:%s\n内容:%s", title, content)),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var res struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}
	if res.Status != "ok" {
		return fmt.Errorf("go_cqhttp status: %s", res.Status)
	}
	return nil
}

// ----------------------------------------------------
// 15. Gotify
// ----------------------------------------------------
type GotifyChannel struct{}

func (g *GotifyChannel) Name() string       { return "gotify" }
func (g *GotifyChannel) Aliases() []string { return []string{} }
func (g *GotifyChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("GOTIFY_URL") && cfg.Has("GOTIFY_TOKEN")
}
func (g *GotifyChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	reqURL := fmt.Sprintf("%s/message?token=%s", strings.TrimRight(cfg.Get("GOTIFY_URL"), "/"), cfg.Get("GOTIFY_TOKEN"))
	payload := map[string]interface{}{
		"title":    title,
		"message":  content,
		"priority": cfg.GetInt("GOTIFY_PRIORITY", 0),
	}
	body, _, err := postJSON(ctx, defaultHTTPClient, reqURL, nil, payload)
	if err != nil {
		return err
	}
	var res struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.ID == 0 {
		return fmt.Errorf("gotify failed: %s", string(body))
	}
	return nil
}

// ----------------------------------------------------
// 16. iGot
// ----------------------------------------------------
type IGotChannel struct{}

func (i *IGotChannel) Name() string       { return "igot" }
func (i *IGotChannel) Aliases() []string { return []string{} }
func (i *IGotChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("IGOT_PUSH_KEY")
}
func (i *IGotChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	reqURL := fmt.Sprintf("https://push.hellyw.com/%s", cfg.Get("IGOT_PUSH_KEY"))
	form := url.Values{}
	form.Set("title", title)
	form.Set("content", content)

	body, _, err := postForm(ctx, defaultHTTPClient, reqURL, nil, form)
	if err != nil {
		return err
	}
	var res struct {
		Ret    int    `json:"ret"`
		ErrMsg string `json:"errMsg"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.Ret != 0 {
		return fmt.Errorf("igot error: %s", res.ErrMsg)
	}
	return nil
}

// ----------------------------------------------------
// 17. Synology Chat
// ----------------------------------------------------
type SynologyChatChannel struct{}

func (s *SynologyChatChannel) Name() string       { return "chat" }
func (s *SynologyChatChannel) Aliases() []string { return []string{"synology", "synology_chat"} }
func (s *SynologyChatChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("CHAT_URL") && cfg.Has("CHAT_TOKEN")
}
func (s *SynologyChatChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	reqURL := cfg.Get("CHAT_URL") + cfg.Get("CHAT_TOKEN")
	payloadMap := map[string]string{
		"text": fmt.Sprintf("%s\n%s", title, content),
	}
	pBytes, _ := json.Marshal(payloadMap)
	form := url.Values{}
	form.Set("payload", string(pBytes))

	_, status, err := postForm(ctx, defaultHTTPClient, reqURL, nil, form)
	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("chat HTTP %d", status)
	}
	return nil
}

// ----------------------------------------------------
// 18. 微加机器人 (WePlus)
// ----------------------------------------------------
type WePlusChannel struct{}

func (w *WePlusChannel) Name() string       { return "weplus" }
func (w *WePlusChannel) Aliases() []string { return []string{"weplus_bot"} }
func (w *WePlusChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("WE_PLUS_BOT_TOKEN")
}
func (w *WePlusChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	template := "txt"
	if len(content) > 800 {
		template = "html"
	}
	payload := map[string]interface{}{
		"token":    cfg.Get("WE_PLUS_BOT_TOKEN"),
		"title":    title,
		"content":  content,
		"template": template,
		"receiver": cfg.Get("WE_PLUS_BOT_RECEIVER"),
		"version":  cfg.Get("WE_PLUS_BOT_VERSION"),
	}
	body, _, err := postJSON(ctx, defaultHTTPClient, "https://www.weplusbot.com/send", nil, payload)
	if err != nil {
		return err
	}
	var res struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.Code != 200 {
		return fmt.Errorf("weplus code: %d, body: %s", res.Code, string(body))
	}
	return nil
}

// ----------------------------------------------------
// 19. Qmsg酱
// ----------------------------------------------------
type QmsgChannel struct{}

func (q *QmsgChannel) Name() string       { return "qmsg" }
func (q *QmsgChannel) Aliases() []string { return []string{"qmsg_bot"} }
func (q *QmsgChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("QMSG_KEY") && cfg.Has("QMSG_TYPE")
}
func (q *QmsgChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	reqURL := fmt.Sprintf("https://qmsg.zendee.cn/%s/%s", cfg.Get("QMSG_TYPE"), cfg.Get("QMSG_KEY"))
	form := url.Values{}
	form.Set("msg", fmt.Sprintf("%s\n\n%s", title, strings.ReplaceAll(content, "----", "-")))

	body, _, err := postForm(ctx, defaultHTTPClient, reqURL, nil, form)
	if err != nil {
		return err
	}
	var res struct {
		Code   int    `json:"code"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.Code != 0 {
		return fmt.Errorf("qmsg error: %s", res.Reason)
	}
	return nil
}

// ----------------------------------------------------
// 20. 智能微秘书 (Aibotk)
// ----------------------------------------------------
type AibotkChannel struct{}

func (a *AibotkChannel) Name() string       { return "aibotk" }
func (a *AibotkChannel) Aliases() []string { return []string{} }
func (a *AibotkChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("AIBOTK_KEY") && cfg.Has("AIBOTK_TYPE") && cfg.Has("AIBOTK_NAME")
}
func (a *AibotkChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	var reqURL string
	payload := map[string]interface{}{
		"apiKey": cfg.Get("AIBOTK_KEY"),
		"message": map[string]interface{}{
			"type":    1,
			"content": fmt.Sprintf("【通知】\n\n%s\n%s", title, content),
		},
	}
	if cfg.Get("AIBOTK_TYPE") == "room" {
		reqURL = "https://api-bot.aibotk.com/openapi/v1/chat/room"
		payload["roomName"] = cfg.Get("AIBOTK_NAME")
	} else {
		reqURL = "https://api-bot.aibotk.com/openapi/v1/chat/contact"
		payload["name"] = cfg.Get("AIBOTK_NAME")
	}

	body, _, err := postJSON(ctx, defaultHTTPClient, reqURL, nil, payload)
	if err != nil {
		return err
	}
	var res struct {
		Code  int    `json:"code"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.Code != 0 {
		return fmt.Errorf("aibotk error: %s", res.Error)
	}
	return nil
}

// ----------------------------------------------------
// 21. PushMe
// ----------------------------------------------------
type PushMeChannel struct{}

func (p *PushMeChannel) Name() string       { return "pushme" }
func (p *PushMeChannel) Aliases() []string { return []string{} }
func (p *PushMeChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("PUSHME_KEY")
}
func (p *PushMeChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	reqURL := cfg.Get("PUSHME_URL")
	if reqURL == "" {
		reqURL = "https://push.i-i.me/"
	}
	form := url.Values{}
	form.Set("push_key", cfg.Get("PUSHME_KEY"))
	form.Set("title", title)
	form.Set("content", content)
	if v := cfg.Get("date"); v != "" {
		form.Set("date", v)
	}
	if v := cfg.Get("type"); v != "" {
		form.Set("type", v)
	}

	body, status, err := postForm(ctx, defaultHTTPClient, reqURL, nil, form)
	if err != nil {
		return err
	}
	if status != 200 || string(body) != "success" {
		return fmt.Errorf("pushme status %d: %s", status, string(body))
	}
	return nil
}

// ----------------------------------------------------
// 22. Chronocat
// ----------------------------------------------------
type ChronocatChannel struct{}

func (c *ChronocatChannel) Name() string       { return "chronocat" }
func (c *ChronocatChannel) Aliases() []string { return []string{} }
func (c *ChronocatChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("CHRONOCAT_URL") && cfg.Has("CHRONOCAT_QQ") && cfg.Has("CHRONOCAT_TOKEN")
}
func (c *ChronocatChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	rawQQ := cfg.Get("CHRONOCAT_QQ")
	userRegex := regexp.MustCompile(`user_id=(\d+)`)
	groupRegex := regexp.MustCompile(`group_id=(\d+)`)

	userMatches := userRegex.FindAllStringSubmatch(rawQQ, -1)
	groupMatches := groupRegex.FindAllStringSubmatch(rawQQ, -1)

	reqURL := fmt.Sprintf("%s/api/message/send", strings.TrimRight(cfg.Get("CHRONOCAT_URL"), "/"))
	headers := map[string]string{
		"Authorization": "Bearer " + cfg.Get("CHRONOCAT_TOKEN"),
	}

	type target struct {
		chatType int
		id       string
	}
	var targets []target
	for _, m := range userMatches {
		if len(m) > 1 {
			targets = append(targets, target{chatType: 1, id: m[1]})
		}
	}
	for _, m := range groupMatches {
		if len(m) > 1 {
			targets = append(targets, target{chatType: 2, id: m[1]})
		}
	}

	if len(targets) == 0 {
		return errors.New("no user_id or group_id matched in CHRONOCAT_QQ")
	}

	var errList []string
	for _, t := range targets {
		payload := map[string]interface{}{
			"peer": map[string]interface{}{
				"chatType": t.chatType,
				"peerUin":  t.id,
			},
			"elements": []map[string]interface{}{
				{
					"elementType": 1,
					"textElement": map[string]string{
						"content": fmt.Sprintf("%s\n\n%s", title, content),
					},
				},
			},
		}
		_, status, err := postJSON(ctx, defaultHTTPClient, reqURL, headers, payload)
		if err != nil || status != 200 {
			errList = append(errList, fmt.Sprintf("target %s failed (HTTP %d, err %v)", t.id, status, err))
		}
	}

	if len(errList) > 0 {
		return errors.New(strings.Join(errList, "; "))
	}
	return nil
}

// ----------------------------------------------------
// 23. WPUSH
// ----------------------------------------------------
type WpushChannel struct{}

func (w *WpushChannel) Name() string       { return "wpush" }
func (w *WpushChannel) Aliases() []string { return []string{} }
func (w *WpushChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("WPUSH_APIKEY")
}
func (w *WpushChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	payload := map[string]interface{}{
		"apikey":  cfg.Get("WPUSH_APIKEY"),
		"title":   title,
		"content": content,
		"channel": cfg.Get("WPUSH_CHANNEL"),
	}
	if topic := cfg.Get("WPUSH_TOPIC_CODE"); topic != "" {
		payload["topic_code"] = topic
	}
	body, _, err := postJSON(ctx, defaultHTTPClient, "https://api.wpush.cn/api/v1/send", nil, payload)
	if err != nil {
		return err
	}
	var res struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.Code != 0 {
		return fmt.Errorf("wpush error: %s", res.Message)
	}
	return nil
}

// ----------------------------------------------------
// 24. OpeniLink
// ----------------------------------------------------
type OpeniLinkChannel struct{}

func (o *OpeniLinkChannel) Name() string       { return "openilink" }
func (o *OpeniLinkChannel) Aliases() []string { return []string{} }
func (o *OpeniLinkChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("OPENILINK_APP_TOKEN")
}
func (o *OpeniLinkChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	baseURL := strings.TrimRight(cfg.Get("OPENILINK_HUB_URL"), "/")
	if baseURL == "" {
		baseURL = "https://hub.openilink.com"
	}
	reqURL := fmt.Sprintf("%s/bot/v1/message/send", baseURL)
	headers := map[string]string{
		"Authorization": "Bearer " + cfg.Get("OPENILINK_APP_TOKEN"),
	}
	payload := map[string]interface{}{
		"type":    "text",
		"content": fmt.Sprintf("%s\n\n%s", title, content),
	}
	if token := cfg.Get("OPENILINK_CONTEXT_TOKEN"); token != "" {
		payload["context_token"] = token
	}

	body, _, err := postJSON(ctx, defaultHTTPClient, reqURL, headers, payload)
	if err != nil {
		return err
	}
	var res struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if !res.OK {
		return fmt.Errorf("openilink error: %s", res.Error)
	}
	return nil
}

// ----------------------------------------------------
// 25. 自定义 Webhook (Custom)
// ----------------------------------------------------
type CustomWebhookChannel struct{}

func (c *CustomWebhookChannel) Name() string       { return "custom" }
func (c *CustomWebhookChannel) Aliases() []string { return []string{"webhook", "custom_notify"} }
func (c *CustomWebhookChannel) IsEnabled(cfg *Config) bool {
	return cfg.Has("WEBHOOK_URL") && cfg.Has("WEBHOOK_METHOD")
}
func (c *CustomWebhookChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	rawURL := cfg.Get("WEBHOOK_URL")
	method := strings.ToUpper(cfg.Get("WEBHOOK_METHOD"))
	rawBody := cfg.Get("WEBHOOK_BODY")
	contentType := cfg.Get("WEBHOOK_CONTENT_TYPE")
	rawHeaders := cfg.Get("WEBHOOK_HEADERS")

	formattedURL := strings.ReplaceAll(rawURL, "$title", url.QueryEscape(title))
	formattedURL = strings.ReplaceAll(formattedURL, "$content", url.QueryEscape(content))

	var reqBody io.Reader
	if rawBody != "" {
		escapedTitle := strings.ReplaceAll(title, "\n", "\\n")
		escapedContent := strings.ReplaceAll(content, "\n", "\\n")
		replacedBody := strings.ReplaceAll(rawBody, "$title", escapedTitle)
		replacedBody = strings.ReplaceAll(replacedBody, "$content", escapedContent)
		reqBody = strings.NewReader(replacedBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, formattedURL, reqBody)
	if err != nil {
		return err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if rawHeaders != "" {
		for _, line := range strings.Split(rawHeaders, "\n") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// ----------------------------------------------------
// 26. 控制台输出 (Console)
// ----------------------------------------------------
type ConsoleChannel struct{}

func (c *ConsoleChannel) Name() string       { return "console" }
func (c *ConsoleChannel) Aliases() []string { return []string{} }
func (c *ConsoleChannel) IsEnabled(cfg *Config) bool {
	return cfg.GetBool("CONSOLE")
}
func (c *ConsoleChannel) Send(ctx context.Context, cfg *Config, title, content string) error {
	log.Printf("\n[CONSOLE NOTIFY]\n标题: %s\n内容:\n%s\n", title, content)
	return nil
}
