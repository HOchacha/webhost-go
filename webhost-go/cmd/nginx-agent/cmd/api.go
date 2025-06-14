package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webhost-go/webhost-go/cmd/nginx-agent/nginx"
)

type Server struct {
	Manager *nginx.NginxManager
}

func (s *Server) RegisterRoutes(router *gin.Engine) {
	router.POST("/api/nginx/register", s.registerAgent)
}

func (s *Server) registerAgent(c *gin.Context) {
	var agent nginx.AgentInfo

	// bind json into struct
	if err := c.ShouldBindJSON(&agent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// update nginx configuration
	if err := s.Manager.AddHTTPConfig(agent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "nginx config update failed: " + err.Error()})
		return
	}

	if err := s.Manager.Reload(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "nginx reload failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "nginx configuration updated and reloaded"})
}
