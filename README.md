# webhook-notify: 轻量级聚合推送 Webhook 网关

基于国内主流青龙/自动化脚本生态的 `notify.py` 完整重构。采用 **Go 静态编译** 与 **Alpine 基础镜像**，拥有极致轻量化和极低资源开销的优势。

- 📦 **超小体积**：整体 Docker 镜像仅约 **18 MB**（含完整 HTTPS 根证书与时区配置）。
- 🖥️ **全平台支持**：官方镜像支持 `linux/amd64` 与 `linux/arm64`（云服务器、树莓派、Apple Silicon、群晖 NAS 等均可原生运行）。
- ⚡ **超低资源**：日常内存占用仅约 **5 MB ~ 10 MB**，冷启动以毫秒计。
- 🌐 **全渠道兼容**：完整实现原 `notify.py` 拥有的 **25+ 种推送渠道**。
- ⚙️ **灵活配置**：支持 `config.json` 文件与环境变量双重注入，**环境变量优先级最高**。
- 🎯 **指定渠道发送**：支持请求中传入 `channel` 参数定向推送，不传则默认群发至所有已配置渠道。
- 🔐 **API-Key 安全防护**：支持全局配置 `api-key`，防止公网端口被恶意滥用。
- 🔄 **通用 Webhook 接口**：原生支持 `GET` 与 `POST`（JSON/Form）双调用协议。

---

## 一、 快速启动

### 方式 1：Docker Run（直接拉取官方镜像，纯环境变量启动）

```bash
docker run -d \
  --name webhook-notify \
  --restart unless-stopped \
  -p 5755:5755 \
  -e API_KEY="your_secret_key" \
  -e BARK_PUSH="https://api.day.app/your_token/" \
  -e QYWX_KEY="your_wecom_bot_key" \
  ghcr.io/crossgg/webhook-notify:latest
```

### 方式 2：使用 config.json 挂载启动

1. 复制配置模板：
   ```bash
   cp config.json.example config.json
   ```
2. 在 `config.json` 中填入你的密钥（留空的渠道不会被激活）。
3. 使用 Docker 挂载运行：
   ```bash
   docker run -d \
     --name webhook-notify \
     --restart unless-stopped \
     -p 5755:5755 \
     -v $(pwd)/config.json:/app/config.json:ro \
     ghcr.io/crossgg/webhook-notify:latest
   ```

### 方式 3：Docker Compose

```bash
docker compose up -d
```

---

## 二、 Webhook 调用说明

服务提供统一接口路径：`/send`（同时 alias `/webhook`）。

### 1. 安全鉴权（API-Key）
若配置了 `API_KEY`（在 `config.json` 或环境变量中设置），请求需携带 API Key，支持以下任意一种方式：
- **Header**：`X-API-Key: <key>` 或 `api-key: <key>` 或 `Authorization: Bearer <key>`
- **URL Query**：`?api-key=<key>` 或 `?api_key=<key>`
- **POST JSON Body**：`"api-key": "<key>"` 或 `"api_key": "<key>"`

> 若未配置 `API_KEY`，则不开启权限校验，允许直接公开调用。

---

### 2. GET 调用方式
适用于浏览器一键触发、iOS 快捷指令、简单脚本或 curl。

#### 示例 1：群发至所有已配置的渠道
```http
GET http://127.0.0.1:5755/send?title=服务通知&content=系统备份已完成&api-key=your_secret_key
```

#### 示例 2：定向推送到指定渠道（如 Bark 和 企业微信机器人）
```http
GET http://127.0.0.1:5755/send?title=CPU告警&content=负载超过90%&channel=bark,wecom_bot&api-key=your_secret_key
```

---

### 3. POST 调用方式（通用 JSON）

#### 示例 1：通用 POST 请求
```bash
curl -X POST http://127.0.0.1:5755/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_secret_key" \
  -d '{
    "title": "部署成功",
    "content": "版本 v1.0.2 发布成功并上线",
    "channel": "bark"
  }'
```

#### 示例 2：多渠道定向推送
```json
{
  "api-key": "your_secret_key",
  "title": "数据库告警",
  "content": "检测到主从同步延迟 50 秒",
  "channel": ["wecom_bot", "tg"]
}
```

> **字段兼容性**：
> - 标题字段：`title`
> - 内容字段兼容：`content`、`text`、`msg`、`message`
> - 渠道字段兼容：`channel`、`channels`（可为逗号分隔字符串如 `"bark,wecom"`，也可为数组 `["bark", "wecom"]`）

---

### 4. 响应格式

```json
{
  "code": 200,
  "message": "Notification dispatched",
  "results": {
    "bark": {
      "success": true,
      "message": "ok"
    },
    "wecom_bot": {
      "success": true,
      "message": "ok"
    }
  }
}
```

---

## 三、 支持的所有渠道与配置参数

| 渠道标识 (`channel`) | 别名 (`aliases`) | 激活所需关键配置 | 说明 |
| :--- | :--- | :--- | :--- |
| **bark** | `bark_push` | `BARK_PUSH` | Bark iOS 推送 |
| **wecom_bot** | `wecom`, `qywx_bot`, `qywx` | `QYWX_KEY` | 企业微信群机器人 |
| **wecom_app** | `qywx_app`, `wx_app` | `QYWX_AM` | 企业微信自建应用 (`corpid,secret,user,agentid`) |
| **dingtalk** | `dingding`, `dingding_bot`, `dd` | `DD_BOT_TOKEN`, `DD_BOT_SECRET` | 钉钉自定义机器人（带签名校验） |
| **feishu** | `feishu_bot`, `fs` | `FSKEY` (可选 `FSSECRET`) | 飞书自定义机器人（支持加签） |
| **telegram** | `tg`, `telegram_bot` | `TG_BOT_TOKEN`, `TG_USER_ID` | Telegram 机器人（支持代理配置） |
| **pushplus** | `pushplus_bot` | `PUSH_PLUS_TOKEN` | PushPlus 推送加微信公众号 |
| **pushdeer** | `deer` | `DEER_KEY` | PushDeer 消息推送 |
| **serverj** | `serverchan`, `ftqq` | `PUSH_KEY` | Server酱（支持 Turbo 版与旧版） |
| **ntfy** | - | `NTFY_TOPIC` | ntfy 推送服务（支持自建服务） |
| **wxpusher** | `wxpusher_bot` | `WXPUSHER_APP_TOKEN`, (`TOPIC_IDS`/`UIDS`) | WxPusher 微信消息服务（应用Token） |
| **wxpusher_spt**| - | `WXPUSHER_SPT_LIST` | WxPusher 极简推送（SPT 码） |
| **smtp** | `mail`, `email` | `SMTP_SERVER`, `SMTP_EMAIL`, `SMTP_PASSWORD` | 邮件推送（支持 SSL 465 与普通端口） |
| **go_cqhttp** | `gobot`, `qqbot` | `GOBOT_URL`, `GOBOT_QQ` | QQ 机器人推送 |
| **gotify** | - | `GOTIFY_URL`, `GOTIFY_TOKEN` | Gotify 自建推送服务 |
| **igot** | - | `IGOT_PUSH_KEY` | iGot 聚合推送 |
| **chat** | `synology`, `synology_chat` | `CHAT_URL`, `CHAT_TOKEN` | 群晖 Synology Chat |
| **weplus** | `weplus_bot` | `WE_PLUS_BOT_TOKEN` | 微加机器人 |
| **qmsg** | `qmsg_bot` | `QMSG_KEY`, `QMSG_TYPE` | Qmsg 酱 QQ 消息 |
| **aibotk** | - | `AIBOTK_KEY`, `AIBOTK_TYPE`, `AIBOTK_NAME` | 智能微秘书微信助手 |
| **pushme** | - | `PUSHME_KEY` | PushMe 消息推送 |
| **chronocat** | - | `CHRONOCAT_URL`, `CHRONOCAT_QQ`, `CHRONOCAT_TOKEN` | Chronocat Satori QQ 协议 |
| **wpush** | - | `WPUSH_APIKEY` | WPUSH 综合微信推送 |
| **openilink** | - | `OPENILINK_APP_TOKEN` | OpeniLink 智能消息推送 |
| **custom** | `webhook`, `custom_notify` | `WEBHOOK_URL`, `WEBHOOK_METHOD` | 自定义 Webhook（支持 `$title` 与 `$content` 模板替换） |
| **console** | - | `CONSOLE=true` | 控制台日志输出 |

---

## 四、 全局辅助配置

| 配置名 | 默认值 | 作用说明 |
| :--- | :--- | :--- |
| `API_KEY` | *(空)* | Webhook 密钥。为空则不鉴权，设置后必须校验 |
| `PORT` | `5755` | 服务监听端口 |
| `HITOKOTO` | `false` | 是否在推送消息末尾自动追加一条精美一言句子 |
| `SKIP_PUSH_TITLE` | *(空)* | 需要跳过的推送标题列表（换行分隔），命中则静默忽略 |
| `CONFIG_PATH` | *(自动查找)* | 指定加载的自定义 `config.json` 文件路径 |

---

## 五、 详细字段获取方法

关于每个渠道的详细作用、配置项含义以及如何在各个平台后台申请获取 Token / Key / Webhook 地址，请参阅：
👉 **[各渠道配置字段含义与获取教程 (CHANNELS_GUIDE.md)](./CHANNELS_GUIDE.md)**

