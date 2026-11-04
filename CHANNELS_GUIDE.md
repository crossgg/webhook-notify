# 推送服务配置字段说明与获取教程

本文档详细说明了所有支持的推送渠道、各配置字段的具体含义，以及在各个平台后台申请和获取对应密钥/Token的详细操作指引。

---

## 目录
- [一、 全局通用配置](#一-全局通用配置)
- [二、 办公与群机器人推送](#二-办公与群机器人推送)
  - [1. 企业微信群机器人 (wecom_bot)](#1-企业微信群机器人-wecom_bot)
  - [2. 企业微信自建应用 (wecom_app)](#2-企业微信自建应用-wecom_app)
  - [3. 钉钉自定义机器人 (dingtalk)](#3-钉钉自定义机器人-dingtalk)
  - [4. 飞书自定义机器人 (feishu)](#4-飞书自定义机器人-feishu)
- [三、 手机端与微信生态轻量通知](#三-手机端与微信生态轻量通知)
  - [5. Bark (iOS 推送)](#5-bark-ios-推送)
  - [6. PushPlus 推送加 (pushplus)](#6-pushplus-推送加-pushplus)
  - [7. WxPusher 微信推送 (wxpusher)](#7-wxpusher-微信推送-wxpusher)
  - [8. PushDeer (pushdeer)](#8-pushdeer-pushdeer)
  - [9. Server酱 Turbo 版 (serverj)](#9-server酱-turbo-版-serverj)
  - [10. WPUSH (wpush)](#10-wpush-wpush)
  - [11. PushMe (pushme)](#11-pushme-pushme)
  - [12. iGot 聚合推送 (igot)](#12-igot-聚合推送-igot)
  - [13. 微加机器人 (weplus)](#13-微加机器人-weplus)
- [四、 国际与自建服务推送](#四-国际与自建服务推送)
  - [14. Telegram 机器人 (telegram)](#14-telegram-机器人-telegram)
  - [15. ntfy (ntfy)](#15-ntfy-ntfy)
  - [16. Gotify 自建推送 (gotify)](#16-gotify-自建推送-gotify)
  - [17. Synology Chat 群晖通知 (chat)](#17-synology-chat-群晖通知-chat)
  - [18. OpeniLink (openilink)](#18-openilink-openilink)
- [五、 邮件与 QQ 机器人](#五-邮件与-qq-机器人)
  - [19. SMTP 邮件推送 (smtp)](#19-smtp-邮件推送-smtp)
  - [20. go-cqhttp (go_cqhttp)](#20-go-cqhttp-go_cqhttp)
  - [21. Chronocat (chronocat)](#21-chronocat-chronocat)
  - [22. Qmsg 酱 (qmsg)](#22-qmsg-酱-qmsg)
  - [23. 智能微秘书 (aibotk)](#23-智能微秘书-aibotk)
- [六、 自定义 Webhook 扩展](#六-自定义-webhook-扩展)
  - [24. 自定义 Webhook (custom)](#24-自定义-webhook-custom)

---

## 一、 全局通用配置

| 字段名 | 默认值 | 含义说明 | 获取/配置方法 |
| :--- | :--- | :--- | :--- |
| `API_KEY` | *(空)* | Webhook 调用鉴权密钥 | 自定义任意复杂字符串（如 `Secr3t_996_Key`）。留空则不开启校验，公网部署强烈建议设置 |
| `PORT` | `5755` | 服务监听端口 | 默认 `5755`，可通过修改此值或环境变量调整 |
| `HITOKOTO` | `false` | 一言开关（随机诗词金句） | 填 `true` 时，会在推送内容末尾自动附加一条精选一言句子 |
| `SKIP_PUSH_TITLE` | *(空)* | 忽略推送的标题列表 | 多个标题用**换行符**分隔。当接收到的消息标题匹配时，将直接静默忽略不发 |
| `CONSOLE` | `false` | 控制台输出打印 | 填 `true` 时，会将所有推送内容同步输出在 Docker/服务标准日志中 |

---

## 二、 办公与群机器人推送

### 1. 企业微信群机器人 (wecom_bot)
- **触发渠道标识**：`wecom_bot`（别名：`wecom`, `qywx_bot`, `qywx`）
- **配置字段**：
  - `QYWX_KEY`：群机器人的 Webhook Key (**必填**)
  - `QYWX_ORIGIN`：企微 API 代理地址（可选，默认 `https://qyapi.weixin.qq.com`）
- **获取方法**：
  1. 打开企业微信电脑端，进入任意内部群聊。
  2. 点击右上角 `...` -> **添加群机器人** -> **新建**。
  3. 输入机器人名字（如“系统告警”），点击添加。
  4. 复制生成的 Webhook 地址，形如：`https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=693a91f6-7xxx-4bc4-97a0-0ec2sifa5aaa`。
  5. 取 URL 中 `key=` 后面的字符串填入 `QYWX_KEY` 即可。

---

### 2. 企业微信自建应用 (wecom_app)
- **触发渠道标识**：`wecom_app`（别名：`qywx_app`, `wx_app`）
- **配置字段**：
  - `QYWX_AM`：应用配置元组，格式为 `corpid,corpsecret,touser,agentid[,media_id]` (**必填**)
  - `QYWX_ORIGIN`：API 代理地址（可选，默认 `https://qyapi.weixin.qq.com`）
- **获取方法**：
  1. 登录 [企业微信管理后台](https://work.weixin.qq.com/)。
  2. **获取 corpid**：点击【我的企业】 -> 最下方可看到【企业ID】（即 `corpid`）。
  3. **创建自建应用**：点击【应用管理】 -> 【应用】 -> 【自建】 -> 【创建应用】。
     - 上传图标，输入应用名称（如“消息中心”），可见范围选择全体或指定人员。
     - 创建完成后，进入该应用详情页：
       - 查看并记录【AgentId】（形如 `1000002`）。
       - 查看【Secret】，点击发送到企业微信客户端查看（即 `corpsecret`）。
  4. **指定接收人**：
     - 如果发送给所有人，`touser` 填写 `@all`。
     - 如果发送给指定人，可在【通讯录】中点击对应员工查看其【账号】（即 `touser`）。
  5. 组合成字符串填入 `QYWX_AM`，例如：`ww1234567890abcdef,secret1234567890,@all,1000002`。

---

### 3. 钉钉自定义机器人 (dingtalk)
- **触发渠道标识**：`dingtalk`（别名：`dingding`, `dingding_bot`, `dd`）
- **配置字段**：
  - `DD_BOT_TOKEN`：机器人的 access_token (**必填**)
  - `DD_BOT_SECRET`：安全设置的“加签”密钥 (**必填**)
- **获取方法**：
  1. 打开钉钉电脑客户端，进入需要接收通知的钉钉群。
  2. 点击右上角群设置 -> 【智能群助手】 -> 【添加机器人】 -> 选择【自定义】。
  3. 机器人名称自定义，在【安全设置】中勾选 **【加签】**。
  4. 界面会显示形如 `SECxxxxxx` 的密钥，复制并填入 `DD_BOT_SECRET`。
  5. 点击完成，会生成一段 Webhook URL，形如：
     `https://oapi.dingtalk.com/robot/send?access_token=c814b7xxxxxxxxxx`
  6. 复制 `access_token=` 后面的一长串字符串，填入 `DD_BOT_TOKEN`。

---

### 4. 飞书自定义机器人 (feishu)
- **触发渠道标识**：`feishu`（别名：`feishu_bot`, `fs`）
- **配置字段**：
  - `FSKEY`：机器人的 Webhook Token (**必填**)
  - `FSSECRET`：安全设置的签名校验密钥（可选，但强烈建议开启）
- **获取方法**：
  1. 打开飞书电脑客户端，进入目标群聊。
  2. 点击右上角 `...` -> 【设置】 -> 【群机器人】 -> 【添加机器人】 -> 【自定义机器人】。
  3. 设置机器人名称与描述。
  4. 在安全设置中建议勾选 **【签名校验】**，复制生成的密钥填入 `FSSECRET`。
  5. 点击保存后获得 Webhook 地址，形如：
     `https://open.feishu.cn/open-apis/bot/v2/hook/7b2e987f-xxxx-4444-xxxx-xxxxxxxxxxxx`
  6. 复制末尾的 UUID 字符串填入 `FSKEY`。

---

## 三、 手机端与微信生态轻量通知

### 5. Bark (iOS 推送)
- **触发渠道标识**：`bark`（别名：`bark_push`）
- **配置字段**：
  - `BARK_PUSH`：设备推送地址或 Key (**必填**)
  - `BARK_SOUND`：推送铃声（可选）
  - `BARK_GROUP`：消息分组名称（可选）
  - `BARK_ICON`：自定义通知图标 URL（可选）
  - `BARK_LEVEL`：时效性，如 `active`, `timeSensitive`, `passive`（可选）
  - `BARK_URL`：点击推送后跳转的 URL（可选）
  - `BARK_ARCHIVE`：是否本地存档 `1`/`0`（可选）
- **获取方法**：
  1. 在 iPhone 的 App Store 搜索下载 **Bark** 应用。
  2. 打开 App，首页中央即可看到您的专属推送 URL，形如：`https://api.day.app/DxHcxXXXXXXXXXXXXXX/`。
  3. 可以直接将整串完整 URL（或仅后面的设备 Key）填入 `BARK_PUSH`。

---

### 6. PushPlus 推送加 (pushplus)
- **触发渠道标识**：`pushplus`（别名：`pushplus_bot`）
- **配置字段**：
  - `PUSH_PLUS_TOKEN`：推送加用户令牌 (**必填**)
  - `PUSH_PLUS_USER`：群组编码（可选，一对多群发时使用）
  - `PUSH_PLUS_TEMPLATE`：消息模板，支持 `html`, `txt`, `json`, `markdown`（默认 `html`）
  - `PUSH_PLUS_CHANNEL`：发送渠道，支持 `wechat`, `webhook`, `mail`, `sms`（默认 `wechat`）
  - `PUSH_PLUS_WEBHOOK`：webhook 编码（可选）
  - `PUSH_PLUS_CALLBACKURL`：发送结果回调地址（可选）
  - `PUSH_PLUS_TO`：好友令牌/企微用户ID（可选）
- **获取方法**：
  1. 浏览器访问 [PushPlus 官网](https://www.pushplus.plus/)。
  2. 点击右上角微信扫码登录并关注公众号。
  3. 登录后进入【一对一推送】，复制显示的【token】填入 `PUSH_PLUS_TOKEN`。

---

### 7. WxPusher 微信推送 (wxpusher / wxpusher_spt)
- **方式 A：标准应用推送（支持多人、多主题）**
  - **触发渠道标识**：`wxpusher`（别名：`wxpusher_bot`）
  - **配置字段**：
    - `WXPUSHER_APP_TOKEN`：应用 Token (**必填**)
    - `WXPUSHER_UIDS`：用户 UID，多个用英文分号 `;` 分隔（与主题ID二选一）
    - `WXPUSHER_TOPIC_IDS`：主题 ID，多个用英文分号 `;` 分隔
  - **获取方法**：
    1. 访问 [WxPusher 管理后台](https://wxpusher.zjiecode.com/admin/) 注册登录。
    2. 点击【应用管理】 -> 【创建应用】，创建后获取 `appToken`。
    3. 点击【用户管理】 -> 关注应用二维码，在用户列表即可查看并复制自己的 `UID`（形如 `UID_xxxxxx`）。
- **方式 B：极简推送（SPT 扫码即用）**
  - **触发渠道标识**：`wxpusher_spt`
  - **配置字段**：
    - `WXPUSHER_SPT_LIST`：极简推送码，多个用英文逗号 `,` 分隔（最多10个）
  - **获取方法**：
    1. 访问 [WxPusher SPT 文档](https://wxpusher.zjiecode.com/docs/#/?id=spt)。
    2. 用微信直接扫描文档中的“极简推送”二维码，即刻返回一串 `spt` 码，复制填入即可。

---

### 8. PushDeer (pushdeer)
- **触发渠道标识**：`pushdeer`（别名：`deer`）
- **配置字段**：
  - `DEER_KEY`：PushDeer 推送密钥 (**必填**)
  - `DEER_URL`：自建服务 API 地址（可选，默认使用官方云服务 `https://api2.pushdeer.com/message/push`）
- **获取方法**：
  1. 在 iOS App Store 下载 PushDeer 或使用微信快应用（Android）。
  2. 打开 App，在【Key】页面点击右上角【+】号，输入备注（如“服务器”）生成新的 Key。
  3. 复制生成的 Key 填入 `DEER_KEY`。

---

### 9. Server酱 Turbo 版 (serverj)
- **触发渠道标识**：`serverj`（别名：`serverchan`, `ftqq`）
- **配置字段**：
  - `PUSH_KEY`：Server酱 SendKey (**必填**)
- **获取方法**：
  1. 浏览器访问 [Server酱·Turbo版官网](https://sct.ftqq.com/)。
  2. 点击右上角微信扫码登录。
  3. 点击进入【SendKey】页面，复制你的 SendKey（形如 `SCTxxxxxxxx`）填入 `PUSH_KEY`。

---

### 10. WPUSH (wpush)
- **触发渠道标识**：`wpush`
- **配置字段**：
  - `WPUSH_APIKEY`：WPUSH API Key (**必填**)
  - `WPUSH_CHANNEL`：推送渠道，默认 `wechat`，可选 `app/sms/mail/webhook/dingtalk/feishu/wechat_work`
  - `WPUSH_TOPIC_CODE`：广播 Topic 编码（可选）
- **获取方法**：
  1. 访问 [WPUSH 官网](https://wpush.cn/) 并注册登录。
  2. 前往【设置】 -> 【API Key】生成并复制 Key。

---

### 11. PushMe (pushme)
- **触发渠道标识**：`pushme`
- **配置字段**：
  - `PUSHME_KEY`：PushMe 密钥 (**必填**)
  - `PUSHME_URL`：自建服务地址（可选，默认 `https://push.i-i.me/`）
- **获取方法**：
  1. 访问 [PushMe 官网](https://push.i-i.me/)。
  2. 微信扫码关注公众号并绑定，复制返回的 Key。

---

### 12. iGot 聚合推送 (igot)
- **触发渠道标识**：`igot`
- **配置字段**：
  - `IGOT_PUSH_KEY`：iGot 推送 Key (**必填**)
- **获取方法**：
  1. 访问 [iGot 官网](https://push.hellyw.com/)。
  2. 微信扫码关注并根据提示获取专属 Key。

---

### 13. 微加机器人 (weplus)
- **触发渠道标识**：`weplus`（别名：`weplus_bot`）
- **配置字段**：
  - `WE_PLUS_BOT_TOKEN`：用户令牌 (**必填**)
  - `WE_PLUS_BOT_RECEIVER`：消息接收者标识（可选）
  - `WE_PLUS_BOT_VERSION`：调用版本（默认 `pro`）
- **获取方法**：
  1. 访问 [微加机器人官网](https://www.weplusbot.com/) 注册登录并获取 Token。

---

## 四、 国际与自建服务推送

### 14. Telegram 机器人 (telegram)
- **触发渠道标识**：`telegram`（别名：`tg`, `telegram_bot`）
- **配置字段**：
  - `TG_BOT_TOKEN`：Telegram Bot Token (**必填**)
  - `TG_USER_ID`：接收消息的用户/群组 ID (**必填**)
  - `TG_API_HOST`：自建反向代理 Host（可选，若国内机器无法直连官方 API 时可配置）
  - `TG_PROXY_HOST`：HTTP 代理 IP/域名（可选）
  - `TG_PROXY_PORT`：HTTP 代理端口（可选）
  - `TG_PROXY_AUTH`：HTTP 代理认证账号密码，格式 `user:password`（可选）
- **获取方法**：
  1. 在 Telegram 中搜索 `@BotFather`，发送 `/start`。
  2. 发送 `/newbot`，按提示输入机器人的名字和用户名（必须以 `bot` 结尾）。
  3. 创建完成后，BotFather 会给出 HTTP API Token（形如 `1234567890:ABCdefGhIjkLmNoPqRsTuVwXyZ`），填入 `TG_BOT_TOKEN`。
  4. 获取自己的 `TG_USER_ID`：
     - 在 Telegram 搜索 `@userinfobot` 或 `@getmyid_bot`。
     - 点击 Start，机器人会回复你的 `Id`（纯数字，如 `987654321`），将其填入 `TG_USER_ID`。
     - **重要**：记得先主动向你刚创建的 Bot 发送任意一条文字（如 `/start`），允许它向你发消息。

---

### 15. ntfy (ntfy)
- **触发渠道标识**：`ntfy`
- **配置字段**：
  - `NTFY_TOPIC`：消息主题 Topic 名称 (**必填**)
  - `NTFY_URL`：服务器地址（默认 `https://ntfy.sh`，可改为自建实例）
  - `NTFY_PRIORITY`：优先级，默认为 `3`（可选 1~5，5为紧急）
  - `NTFY_TOKEN`：Token 鉴权（可选）
  - `NTFY_USERNAME` / `NTFY_PASSWORD`：Basic Auth 账号密码（可选）
  - `NTFY_ACTIONS`：按钮动作 JSON/Header（可选）
- **获取方法**：
  1. 手机下载安装 ntfy App（Android 可在 Google Play 或 F-Droid 下载，iOS 在 App Store 下载）。
  2. 任意构想一个专属于你的独特 Topic（例如 `my_homelab_notify_889`）。
  3. 在手机 App 内点击【+】订阅这个 Topic。
  4. 将该 Topic 名称填入 `NTFY_TOPIC` 即可，零注册秒用。

---

### 16. Gotify 自建推送 (gotify)
- **触发渠道标识**：`gotify`
- **配置字段**：
  - `GOTIFY_URL`：自建 Gotify 服务地址，如 `https://gotify.yourdomain.com` (**必填**)
  - `GOTIFY_TOKEN`：应用 Application Token (**必填**)
  - `GOTIFY_PRIORITY`：推送优先级，默认为 `0`
- **获取方法**：
  1. 登录您自建的 Gotify 网页控制台。
  2. 点击顶部【APPS】 -> 【CREATE APPLICATION】。
  3. 输入应用名称，点击创建，复制生成的 `Token` 填入 `GOTIFY_TOKEN`。

---

### 17. Synology Chat 群晖通知 (chat)
- **触发渠道标识**：`chat`（别名：`synology`, `synology_chat`）
- **配置字段**：
  - `CHAT_URL`：群晖 Chat Webhook 前缀 URL (**必填**)
  - `CHAT_TOKEN`：Webhook Token (**必填**)
- **获取方法**：
  1. 打开群晖 DSM，启动 Synology Chat。
  2. 点击左上角用户头像 -> 【整合】 -> 【传入的 Webhook】。
  3. 点击【创建】，选择要投递的频道，系统会生成 Webhook URL。
  4. 将 URL 与 Token 对应拆分填入 `CHAT_URL` 和 `CHAT_TOKEN`。

---

### 18. OpeniLink (openilink)
- **触发渠道标识**：`openilink`
- **配置字段**：
  - `OPENILINK_APP_TOKEN`：在 OpeniLink Hub 后台安装 App 后获取的 app_token (**必填**)
  - `OPENILINK_HUB_URL`：Hub 地址（默认 `https://hub.openilink.com`）
  - `OPENILINK_CONTEXT_TOKEN`：会话上下文标识（可选）
- **获取方法**：
  1. 参阅 [OpeniLink 官方文档](https://openilink.com/docs/hub/apps) 在管理后台获取。

---

## 五、 邮件与 QQ 机器人

### 19. SMTP 邮件推送 (smtp)
- **触发渠道标识**：`smtp`（别名：`mail`, `email`）
- **配置字段**：
  - `SMTP_SERVER`：SMTP 服务器与端口，形如 `smtp.qq.com:465` 或 `smtp.163.com:465` (**必填**)
  - `SMTP_SSL`：是否启用 SSL，填写 `true` 或 `false`（465 端口请填 `true`）
  - `SMTP_EMAIL`：发件人邮箱地址 (**必填**)
  - `SMTP_PASSWORD`：邮箱密码或专用的 **SMTP 授权码** (**必填**)
  - `SMTP_NAME`：发件人显示昵称（可选，如“系统告警中心”）
  - `SMTP_EMAIL_TO`：收件邮箱，默认发给自己；若有多个接收人用分号 `;` 分隔
- **获取方法**（以主流 QQ 邮箱与 163 邮箱为例）：
  - **QQ 邮箱**：
    1. 网页登录 QQ 邮箱 -> 点击顶部【设置】 -> 【账户】。
    2. 向下滚动找到【POP3/IMAP/SMTP/Exchange/CardDAV/CalDAV服务】。
    3. 开启 `POP3/SMTP服务`，按提示发送短信验证，获取一串 16 位的**授权码**（填入 `SMTP_PASSWORD`，注意不是 QQ 登录密码）。
    4. `SMTP_SERVER` 填写 `smtp.qq.com:465`，`SMTP_SSL` 填 `true`。
  - **163 邮箱**：
    1. 网页登录 163 邮箱 -> 【设置】 -> 【POP3/SMTP/IMAP】。
    2. 开启 `POP3/SMTP服务`，新增授权密码。
    3. `SMTP_SERVER` 填写 `smtp.163.com:465`，`SMTP_SSL` 填 `true`。

---

### 20. go-cqhttp (go_cqhttp)
- **触发渠道标识**：`go_cqhttp`（别名：`gobot`, `qqbot`）
- **配置字段**：
  - `GOBOT_URL`：go-cqhttp 的 HTTP 地址，如 `http://127.0.0.1:5700/send_private_msg` 或 `/send_group_msg`
  - `GOBOT_QQ`：目标 QQ 用户或群，私聊填 `user_id=123456`，群聊填 `group_id=123456`
  - `GOBOT_TOKEN`：go-cqhttp 配置的 access_token（可选）

---

### 21. Chronocat (chronocat)
- **触发渠道标识**：`chronocat`
- **配置字段**：
  - `CHRONOCAT_URL`：Chronocat 服务 API 基础地址
  - `CHRONOCAT_QQ`：目标，如 `user_id=123456` 或 `group_id=654321`
  - `CHRONOCAT_TOKEN`：Chronocat Satori Token
- **获取方法**：
  1. 查阅您自建的 Chronocat 客户端配置获取。

---

### 22. Qmsg 酱 (qmsg)
- **触发渠道标识**：`qmsg`（别名：`qmsg_bot`）
- **配置字段**：
  - `QMSG_KEY`：Qmsg 酱 API Key (**必填**)
  - `QMSG_TYPE`：发送方式，如 `send` (私聊) 或 `group` (群聊)
- **获取方法**：
  1. 访问 [Qmsg 酱官网](https://qmsg.zendee.cn/) 登录，添加指定的推送 QQ 即可在后台查看 Key。

---

### 23. 智能微秘书 (aibotk)
- **触发渠道标识**：`aibotk`
- **配置字段**：
  - `AIBOTK_KEY`：智能微秘书 API Key
  - `AIBOTK_TYPE`：发送目标类型，`room` (群) 或 `contact` (个人好友)
  - `AIBOTK_NAME`：群名称或好友昵称
- **获取方法**：
  1. 访问 [智能微秘书官网](http://wechat.aibotk.com/) 个人中心查看 API Key。

---

## 六、 自定义 Webhook 扩展

### 24. 自定义 Webhook (custom)
- **触发渠道标识**：`custom`（别名：`webhook`, `custom_notify`）
- **适用场景**：将告警转发给任意未内置的第三方系统（如自建网站、Slack、Discord、自动化平台等）。
- **配置字段**：
  - `WEBHOOK_URL`：第三方接口 URL，URL 中可包含变量 `$title` 与 `$content` (**必填**)
  - `WEBHOOK_METHOD`：HTTP 方法，如 `POST` 或 `GET` (**必填**)
  - `WEBHOOK_CONTENT_TYPE`：请求的 Content-Type（如 `application/json`）
  - `WEBHOOK_HEADERS`：自定义请求头，多行用换行符分隔（如 `Authorization: Bearer xxx\nUser-Agent: Notify`）
  - `WEBHOOK_BODY`：请求体模板，支持包含 `$title` 和 `$content` 占位符
- **使用示例**：
  ```json
  {
    "WEBHOOK_URL": "https://api.example.com/alert",
    "WEBHOOK_METHOD": "POST",
    "WEBHOOK_CONTENT_TYPE": "application/json",
    "WEBHOOK_BODY": "{\"topic\":\"$title\",\"detail\":\"$content\"}"
  }
  ```
  在发送消息时，服务会自动把 `$title` 和 `$content` 替换为实际接收到的标题与消息内容。
