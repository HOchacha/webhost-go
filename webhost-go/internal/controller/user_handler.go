package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webhost-go/webhost-go/internal/services/user_service"
)

type UserHandler struct {
	userSvc user_service.Service
}

func NewUserHandler(s user_service.Service) *UserHandler {
	return &UserHandler{userSvc: s}
}

// POST /register
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"` // 원래는 required,email로 이메일 형식 검사를 수행하였으나
		// 일시적으로 중단함
		Password string `json:"password" binding:"required"`
		Name     string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	err := h.userSvc.Signup(req.Email, req.Password, req.Name)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "회원가입 성공"})
}

// POST /login
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	token, err := h.userSvc.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.userSvc.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사용자 목록을 불러오는 데 실패했습니다"})
		return
	}

	var result []gin.H
	for _, u := range users {
		result = append(result, gin.H{
			"id":    u.ID,
			"email": u.Email,
			"name":  u.Name,
			"role":  u.Role,
		})
	}

	c.JSON(http.StatusOK, result)
}

func (h *UserHandler) UpdateUserHandler(c *gin.Context) {
	targetEmail := c.Param("username")

	var req struct {
		Name        string `json:"name"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 형식입니다"})
		return
	}

	err := h.userSvc.UpdateUser(targetEmail, req.Name, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사용자 정보 수정 실패"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "사용자 정보가 성공적으로 수정되었습니다"})
}

func (h *UserHandler) DelteUser(c *gin.Context) {
	targetEmail := c.Param("username")

	err := h.userSvc.DeleteUserByEmail(targetEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사용자 삭제 실패"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "계정이 성공적으로 삭제되었습니다"})
}

func (h *UserHandler) GetUserInfo(ctx *gin.Context) {
	targetEmail := ctx.Param("username")

	user, err := h.userSvc.GetUserByEmail(targetEmail)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "사용자 정보를 불러올 수 없습니다"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
	})
}
