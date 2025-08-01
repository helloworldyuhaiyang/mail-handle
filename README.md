# Mail-Handle

一个基于Gmail API的智能邮件转发系统，能够自动识别重要邮件并根据关键词转发给指定联系人。

## 项目功能

- 🔐 **Gmail OAuth2认证** - 安全的Gmail API访问
- 📧 **智能邮件监控** - 定时检查未读邮件
- 🎯 **关键词识别** - 根据邮件主题关键词自动转发
- 📤 **自动转发** - 将重要邮件转发给指定联系人
- 🗄️ **数据库管理** - 管理转发目标联系人信息
- ⏰ **定时任务** - 可配置的邮件检查频率

## 项目架构

```
mail-handle/
├── cmd/                    # 应用程序入口
│   └── main.go            # 主程序，包含run和auth命令
├── config/                 # 配置文件
│   └── default.yaml       # 默认配置（服务器、数据库、Gmail等）
├── internal/              # 内部业务逻辑
│   ├── api/              # HTTP API服务
│   ├── db/               # 数据库操作
│   ├── mail/             # 邮件服务接口
│   ├── models/           # 数据模型
│   ├── repo/             # 数据仓库层
│   └── schedule/         # 定时任务调度器
└── pkg/                  # 公共包
    ├── app/              # 应用程序框架
    ├── data/             # 数据库连接
    ├── gmail/            # Gmail API客户端
    └── gin_server.go     # HTTP服务器
```

## 包结构说明

### cmd/
- **main.go**: 程序入口，提供两个主要命令：
  - `run`: 启动邮件处理服务
  - `auth`: Gmail OAuth2认证

## 快速开始

### 1. 环境准备

确保已安装：
- Go 
- MySQL

### 2. 获取Gmail API凭证

1. 访问 [Google Cloud Console](https://console.cloud.google.com/)
2. 创建新项目或选择现有项目
3. 启用Gmail API
4. 创建OAuth2.0凭证
5. 下载凭证文件，重命名为 `gmail-credentials.json` 并放到 `config/` 目录

### 3. 配置数据库

1. 创建MySQL数据库：
```sql
CREATE DATABASE mail_handle CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

2. 创建转发目标表：
```sql
CREATE TABLE forward_targets (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    email VARCHAR(128) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

3. 插入测试数据：
```sql
INSERT INTO forward_targets (name, email) VALUES 
('yang', 'helloworldyang9@gmail.com')
('张三', 'zhangsan@example.com'),
('李四', 'lisi@example.com');
```

### 4. 配置应用

修改 `config/default.yaml` 中的数据库连接信息：
```yaml
database:
  dsn: "用户名:密码@tcp(127.0.0.1:3306)/mail_handle?charset=utf8mb4&parseTime=True&loc=Local"
```

### 5. 首次认证

选择以下任一方式进行Gmail认证：

**方式一：命令行认证(不推荐)**
```bash
mail-handle auth
```
复制控制台输出的URL到浏览器，获取授权码后粘贴回控制台。

**方式二：浏览器认证**
1. 启动服务：`mail-handle run`
2. 浏览器访问：`http://127.0.0.1:8082/api/v1/oauth/auth`
3. 完成Gmail授权

### 6. 启动服务

```bash
mail-handle run
```

服务启动后会自动：
- 每分钟检查一次未读邮件
- 识别包含关键词（如"urgent"、"important"、"priority"）的邮件
- 根据邮件主题中的联系人名称自动转发

## 邮件转发规则

系统会解析邮件主题，格式为：`关键词-联系人名称: 邮件内容`

例如：
- `urgent-张三: 请立即处理`
- `important-李四: 项目进度汇报`

系统会：
1. 提取关键词和联系人名称
2. 在数据库中查找对应的邮箱地址
3. 将邮件转发给该联系人

## 配置说明

### 定时任务配置
```yaml
scheduler:
  fetch_interval: "0 */1 * * * *"  # 每分钟执行一次
  forward_keywords: ["urgent", "important", "priority"]  # 转发关键词
```

### 日志配置
```yaml
log:
  level: "info"    # 日志级别：debug, info, warn, error
  format: "text"   # 日志格式：text 或 json
```

