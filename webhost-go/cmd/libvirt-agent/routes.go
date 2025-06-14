package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webhost-go/webhost-go/cmd/libvirt-agent/agent"
)

func registerRoutes(r *gin.Engine) {
	r.POST("/api/template/download", agent.DownloadTemplateHandler) // 추가됨
	r.POST("/api/vm/start", startVM)
	r.POST("/api/vm/stop", stopVM)
	r.DELETE("/api/vm/delete", deleteVM)
}

func startVM(c *gin.Context) {
	var req agent.VMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := agent.StartVM(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "VM started successfully"})
}

func stopVM(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing VM name"})
		return
	}
	if err := agent.StopVM(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "VM stopped successfully"})
}

func deleteVM(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing VM name"})
		return
	}
	if err := agent.DeleteVM(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "VM deleted successfully"})
}
