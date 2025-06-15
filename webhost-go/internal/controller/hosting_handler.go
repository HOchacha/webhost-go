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

func (h *HostingHandler) GetVMStatus(c *gin.Context) {
	email := c.Param("username")

	active, err := h.HostingService.GetVMStatus(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "상태 조회 실패: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"name": active.VMName, "running": active.Active})
}

func (h *HostingHandler) GetVMDetail(c *gin.Context) {
	email := c.Param("username")

	hosting, info, err := h.HostingService.GetVMDetail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "상세 정보 조회 실패: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hosting": hosting,
		"info":    info,
	})
}

func (h *HostingHandler) StartVM(c *gin.Context) {
	email := c.Param("username")
	if err := h.HostingService.StartVM(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "VM 시작 실패: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "VM 시작 완료"})
}

func (h *HostingHandler) StopVM(c *gin.Context) {
	email := c.Param("username")
	if err := h.HostingService.StopVM(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "VM 중지 실패: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "VM 중지 완료"})
}

func (h *HostingHandler) DeleteVM(c *gin.Context) {
	email := c.Param("username")
	if err := h.HostingService.DeleteVM(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "VM 삭제 실패: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "VM 삭제 완료"})
}
