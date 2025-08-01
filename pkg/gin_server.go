package pkg

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"time"
)

// GinService 启动一个httpserver 对外提供服务。会依赖各个组件的业务系统
type GinService struct {
	local      string
	ginEngine  *gin.Engine
	httpServer *http.Server
}

func NewGinServer(local string) *GinService {
	ginEngine := gin.Default()

	return &GinService{
		local:     local,
		ginEngine: ginEngine,
		httpServer: &http.Server{
			Handler: ginEngine,
		},
	}
}

func (h *GinService) GinGroup(relativePath string) *gin.RouterGroup {
	return h.ginEngine.Group(relativePath)
}

func (h *GinService) GinEngine() *gin.Engine {
	return h.ginEngine
}

// Start 会阻塞
func (h *GinService) Start() error {
	// 设置服务器监听请求端口
	l, err := net.Listen("tcp4", h.local)
	if err != nil {
		return err
	}

	err = h.httpServer.Serve(l)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	} else {
		return err
	}
}

func (h *GinService) Stop(waitTime time.Duration) error {
	withTimeout, cancelFunc := context.WithTimeout(context.Background(), waitTime)
	defer cancelFunc()
	err := h.httpServer.Shutdown(withTimeout)
	if err != nil {
		return err
	} else {
		return nil
	}
}
