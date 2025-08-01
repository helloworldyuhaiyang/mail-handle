package internal

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helloworldyuhaiyang/mail-handle/pkg"
	"github.com/helloworldyuhaiyang/mail-handle/pkg/data"
	"github.com/sirupsen/logrus"
)

type OAuthCallback interface {
	GetAuthURL() string
	Callback(code string) error
}

type ApiServer struct {
	addr          string
	db            *data.DB
	oauthCallback OAuthCallback
	ginServer     *pkg.GinService
}

func NewApiServer(addr string, db *data.DB, oauthCallback OAuthCallback) *ApiServer {
	return &ApiServer{addr: addr, db: db, oauthCallback: oauthCallback, ginServer: pkg.NewGinServer(addr)}
}

func (s *ApiServer) Start() error {
	s.setupRoutes()
	go func() {
		logrus.Info("starting server...")
		if err := s.ginServer.Start(); err != nil {
			logrus.Errorf("server error: %v", err)
		}
	}()
	return nil
}

func (s *ApiServer) Stop(waitTime time.Duration) error {
	return s.ginServer.Stop(5 * time.Second)
}

func (a *ApiServer) setupRoutes() {
	v1 := a.ginServer.GinGroup("/api/v1")
	oauth := v1.Group("/oauth")
	{
		oauth.GET("/callback", a.GetGoogleOauth)
	}
}

func (a *ApiServer) GetGoogleOauth(c *gin.Context) {
	code := c.Query("code")
	if err := a.oauthCallback.Callback(code); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "success"})
}
