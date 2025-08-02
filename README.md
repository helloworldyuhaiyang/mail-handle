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
├── deploy/                 # 部署相关文件
│   ├── setup.sh           # 自动化部署脚本
│   ├── docker-compose.yaml # Docker编排文件
│   ├── Dockerfile         # Docker镜像构建文件
│   └── env.example        # 环境变量示例
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

## 快速体验 (推荐)

使用自动化部署脚本快速体验项目，无需手动配置环境：

### 前置要求

确保系统已安装：
- **Go 1.23+** - 用于构建应用程序
- **Docker** - 用于容器化部署
- **Docker Compose** - 用于多容器编排

### 一键部署

1. **克隆项目**
```bash
git clone <repository-url>
cd mail-handle
```

2. **运行部署脚本**
```bash
cd deploy
chmod +x setup.sh
./setup.sh
```

3. **获取Gmail API凭证**
脚本会引导您完成Gmail API凭证的获取：
- 访问 [Google Cloud Console](https://console.cloud.google.com/)
- 创建项目并启用Gmail API
- 创建OAuth2.0凭证，重定向URI: `http://localhost:8082/api/v1/oauth/callback`
- 下载凭证文件并重命名为 `gmail-credentials.json`
- 将文件放到 `deploy/dist/config/` 目录

4. **完成Gmail认证**
部署完成后，访问认证URL完成Gmail授权：
```
http://localhost:8082/api/v1/oauth/auth
```

### 部署后的服务

- **Gmail认证**: http://localhost:8082/api/v1/oauth/auth

### 常用命令

```bash
# 查看应用日志
docker-compose logs -f mail-handle

# 停止服务
docker-compose down

# 重启服务
docker-compose restart

# 查看服务状态
docker-compose ps
```

### 配置自定义

编辑 `deploy/.env` 文件来自定义配置：
- 数据库密码
- 邮件检查频率
- 转发关键词
- 日志级别等

## 快速开始开发

### 1. 环境准备

确保已安装：
- Go 
- MySQL

### 2. 获取Gmail API凭证

1. 访问 [Google Cloud Console](https://console.cloud.google.com/)
2. 创建新项目或选择现有项目
3. 启用Gmail API
4. 创建OAuth2.0凭证, 重定向 URI: http://localhost:8082/api/v1/oauth/callback
5. 下载凭证文件，重命名为 `gmail-credentials.json` 并放到 `config/` 目录(或者更改配置文件指定gmail.credentials_file的值)

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

