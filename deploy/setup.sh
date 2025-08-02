#!/bin/bash

set -e

echo "🚀 Mail Handle 自动化部署脚本"
echo "=============================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_step() {
    echo -e "${PURPLE}📋 $1${NC}"
}

print_success() {
    echo -e "${CYAN}🎉 $1${NC}"
}

# Function to show progress bar
show_progress() {
    local current=$1
    local total=$2
    local width=50
    local percentage=$((current * 100 / total))
    local completed=$((width * current / total))
    local remaining=$((width - completed))
    
    printf "\r["
    printf "%${completed}s" | tr ' ' '#'
    printf "%${remaining}s" | tr ' ' '-'
    printf "] %d%%" $percentage
}

# Function to wait for user input with timeout
wait_for_input() {
    local prompt="$1"
    local timeout="${2:-30}"
    local default="${3:-n}"
    
    echo -n "$prompt (${timeout}s): "
    read -t $timeout -n 1 -r || true
    echo
    
    if [[ -z "$REPLY" ]]; then
        REPLY="$default"
    fi
}

# Check if we're in the right directory
if [ ! -f "../go.mod" ]; then
    print_error "请在 mail-handle 项目根目录下运行此脚本"
    exit 1
fi

# Show deployment overview
echo ""
print_info "部署流程概览："
echo "1. 📦 检查系统依赖"
echo "2. 🔨 构建应用程序"
echo "3. 🔑 配置 Gmail 凭证"
echo "4. 🐳 构建 Docker 镜像"
echo "5. 🚀 启动应用程序"
echo "6. 📊 查看运行状态"
echo ""

# Check prerequisites
print_step "步骤 1: 检查系统依赖..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go 未安装。请先安装 Go 1.19+"
    echo ""
    echo "安装指南："
    echo "  macOS: brew install go"
    echo "  Ubuntu: sudo apt-get install golang-go"
    echo "  CentOS: sudo yum install golang"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
if [[ $(echo "$GO_VERSION" | cut -d. -f1) -lt 1 ]] || [[ $(echo "$GO_VERSION" | cut -d. -f2) -lt 19 ]]; then
    print_error "Go 版本过低。需要 Go 1.19+，当前版本: $GO_VERSION"
    exit 1
fi

print_status "Go 版本: $GO_VERSION"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    print_error "Docker 未安装。请先安装 Docker"
    echo ""
    echo "安装指南："
    echo "  macOS: https://docs.docker.com/desktop/install/mac-install/"
    echo "  Ubuntu: https://docs.docker.com/engine/install/ubuntu/"
    echo "  CentOS: https://docs.docker.com/engine/install/centos/"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose 未安装。请先安装 Docker Compose"
    echo ""
    echo "安装指南："
    echo "  https://docs.docker.com/compose/install/"
    exit 1
fi

print_status "Docker 和 Docker Compose 已安装"

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    print_error "Docker 守护进程未运行。请启动 Docker"
    echo ""
    echo "启动方法："
    echo "  macOS: 打开 Docker Desktop 应用"
    echo "  Linux: sudo systemctl start docker"
    exit 1
fi

print_status "Docker 守护进程正在运行"

# Step 2: Build the application
print_step "步骤 2: 构建应用程序..."

cd ..

# Clean previous builds
print_info "清理之前的构建文件..."
rm -rf dist/

# Build for Linux (for Docker)
print_info "构建 Linux 版本..."
show_progress 1 3
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/mail-handle-linux cmd/main.go
show_progress 2 3

if [ ! -f "dist/mail-handle-linux" ]; then
    print_error "构建失败！请检查代码是否有错误"
    echo ""
    echo "常见问题："
    echo "1. 检查 go.mod 文件是否存在"
    echo "2. 运行 'go mod tidy' 更新依赖"
    echo "3. 检查 cmd/main.go 文件是否存在"
    exit 1
fi

show_progress 3 3
echo ""
print_status "应用程序构建成功"


cd deploy
# Copy configuration files
print_info "复制配置文件..."
mkdir -p dist/config && cp ../config/default.yaml dist/config/


# Step 3: Gmail credentials setup
print_step "步骤 3: 配置 Gmail 凭证..."

# Check if gmail-credentials.json exists
if [ ! -f "dist/config/gmail-credentials.json" ]; then
    print_warning "Gmail 凭证文件未找到！"
    echo ""
    echo "📋 详细获取 Gmail API 凭证步骤："
    echo ""
    echo "1️⃣ 访问 Google Cloud Console"
    echo "   https://console.cloud.google.com/"
    echo ""
    echo "2️⃣ 创建新项目或选择现有项目"
    echo "   - 点击页面顶部的项目选择器"
    echo "   - 点击 '新建项目' 或选择现有项目"
    echo ""
    echo "3️⃣ 启用 Gmail API"
    echo "   - 在左侧菜单中选择 'API 和服务' > '库'"
    echo "   - 搜索 'Gmail API' 并点击"
    echo "   - 点击 '启用' 按钮"
    echo ""
    echo "4️⃣ 创建 OAuth2.0 凭证"
    echo "   - 在左侧菜单中选择 'API 和服务' > '凭证'"
    echo "   - 点击 '创建凭证' > 'OAuth 客户端 ID'"
    echo "   - 应用类型选择 'Web 应用程序'"
    echo "   - 名称：输入任意名称（如 'Mail Handle'）"
    echo "   - 重定向 URI：http://localhost:8082/api/v1/oauth/callback"
    echo "   - 点击 '创建'"
    echo ""
    echo "5️⃣ 下载凭证文件"
    echo "   - 创建完成后会显示客户端 ID 和客户端密钥"
    echo "   - 点击 '下载 JSON' 按钮"
    echo ""
    echo "6️⃣ 放置凭证文件"
    echo "   - 将下载的文件重命名为 'gmail-credentials.json'"
    echo "   - 放到当前目录的 'dist/config/' 文件夹中"
    echo ""
    
    echo "📄 凭证文件格式示例: client_secret_933662119290-6piqu80k5nnhtj9bikeqfu28oq9u1qh4.apps.googleusercontent.com.json"
    
    echo "💡 提示："
    echo "- 确保重定向 URI 完全匹配：http://localhost:8082/api/v1/oauth/callback"
    echo "- 如果遇到权限问题，请确保已启用 Gmail API"
    echo "- 首次使用需要等待几分钟才能生效"
    echo ""
    
    wait_for_input "准备好 gmail-credentials.json 文件后按 Enter 继续" 60 "y"
    
    if [ ! -f "dist/config/gmail-credentials.json" ]; then
        print_error "Gmail 凭证文件仍未找到。请添加文件后重新运行脚本。"
        echo ""
        echo "文件应该位于：$(pwd)/dist/config/gmail-credentials.json"
        exit 1
    fi
fi

print_status "Gmail 凭证文件已找到"

# Step 4: Create environment file
print_step "步骤 4: 创建环境配置文件..."

if [ ! -f ".env" ]; then
    print_info "从 env.example 创建 .env 文件..."
    cp env.example .env
    print_status ".env 文件已创建。您可以编辑它来自定义配置。"
else
    print_status ".env 文件已存在"
fi

# Step 5: Build Docker image
print_step "步骤 5: 构建 Docker 镜像..."

echo ""
echo "🐳 Docker 镜像选项："
echo "1. 使用远程镜像（推荐，快速）"
echo "2. 构建本地镜像（需要时间，但确保最新版本）"
echo ""

wait_for_input "选择镜像构建方式 (1/2，默认2)" 30 "2"

if [[ "$REPLY" == "2" ]]; then
    print_info "构建本地 Docker 镜像..."
    echo "这可能需要几分钟时间，请耐心等待..."
    
    # Build the image with correct context
    docker build -t mail-handle:latest -f Dockerfile ..
    
    if [ $? -eq 0 ]; then
        print_status "Docker 镜像构建成功"
        
        # Update docker-compose to use local image
        sed -i.bak 's|registry.cn-hangzhou.aliyuncs.com/helloworldyu/mail-handle:beta|mail-handle:latest|g' docker-compose.yaml
        print_status "已更新 docker-compose.yaml 使用本地镜像"
    else
        print_error "Docker 镜像构建失败"
        exit 1
    fi
else
    print_info "使用远程 Docker 镜像"
fi

# Step 6: Start the application
print_step "步骤 6: 启动应用程序..."

# Stop any existing containers
print_info "停止现有容器..."
docker-compose down 2>/dev/null || true

# Start the application
print_info "启动应用程序..."
docker-compose up -d

# Wait a moment for containers to start
echo "等待容器启动..."
for i in {1..10}; do
    show_progress $i 10
    sleep 1
done
echo ""

# Check if containers are running
if docker-compose ps | grep -q "Up"; then
    print_success "应用程序启动成功！"
else
    print_error "应用程序启动失败。请检查日志："
    docker-compose logs
    exit 1
fi

# Step 7: Show status and next steps
echo ""
print_success "部署完成！"
echo "=============="
echo ""
print_status "应用程序状态："
docker-compose ps
echo ""

# Show useful commands
echo "📋 常用命令："
echo ""
print_info "查看实时日志："
echo "  docker-compose logs -f mail-handle"
echo ""
print_info "停止应用程序："
echo "  docker-compose down"
echo ""
print_info "重启应用程序："
echo "  docker-compose restart"
echo ""
print_info "查看容器状态："
echo "  docker-compose ps"
echo ""

# Check if Gmail authentication is needed
print_info "检查 Gmail 认证状态..."
sleep 3

if docker-compose logs mail-handle 2>/dev/null | grep -q "Gmail client needs authentication"; then
    print_warning "需要 Gmail 认证！, 第一次请一定要打开"
    echo ""
    echo "🔐 Gmail 认证步骤："
    echo ""
    echo "1. 访问认证 URL："
    echo "   http://localhost:8082/api/v1/oauth/auth"
    echo ""
    echo "2. 在浏览器中完成 Gmail 授权"
    echo "   - 选择要授权的 Gmail 账户"
    echo "   - 点击 '允许' 授权应用访问"
    echo ""
    echo "3. 认证完成后重启应用程序："
    echo "   docker-compose restart mail-handle"
    echo ""
    echo "💡 提示："
    echo "- 如果认证失败，请检查 gmail-credentials.json 文件是否正确"
    echo "- 确保重定向 URI 配置正确"
    echo "- 首次认证可能需要等待几分钟"
    echo ""
    
    wait_for_input "完成 Gmail 认证后按 Enter 继续查看日志" 60 "y"
else
    print_status "Gmail 认证已完成"
fi

echo ""
print_success "应用程序已准备就绪！"
echo "🌐 API 服务地址: http://localhost:8082"
echo "🔐 Gmail 认证: http://localhost:8082/api/v1/oauth/auth"
echo ""

# Ask if user wants to view logs
wait_for_input "是否查看实时日志？(y/n)" 10 "y"

if [[ "$REPLY" =~ ^[Yy]$ ]]; then
    echo ""
    print_info "开始查看实时日志..."
    echo "按 Ctrl+C 退出日志查看"
    echo ""
    docker-compose logs -f mail-handle
else
    echo ""
    print_info "如需查看日志，请运行："
    echo "  docker-compose logs -f mail-handle"
fi 