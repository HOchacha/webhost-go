package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webhost-go/webhost-go/internal/services/hosting_service"
	"webhost-go/webhost-go/internal/services/user_service"
)

type HostingHandler struct {
	HostingService hosting_service.Service // 인터페이스를 사용
	UserService    user_service.Service
}

type CreateHostingRequest struct {
	UserID   int64  `json:"user_id" binding:"required"`
	Username string `json:"username" binding:"required"`
}

func NewHostingHandler(h hosting_service.Service, u user_service.Service) *HostingHandler {
	return &HostingHandler{HostingService: h, UserService: u}
}

func (h *HostingHandler) CreateHosting(c *gin.Context) {
	targetEmail := c.Param("username")

	user, err := h.UserService.GetUserByEmail(targetEmail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "유저 정보를 불러올 수 없습니다: " + err.Error()})
		return
	}

	// VM 생성 및 DB 등록
	hosting, err := h.HostingService.CreateHosting(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "호스팅 생성 실패: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "VM 생성 완료",
		"hostname": hosting.VMName,
		"ip":       hosting.IPAddress,
		"ssh_port": hosting.SSHPort,
		"proxy":    hosting.ProxyPath,
	})
}
