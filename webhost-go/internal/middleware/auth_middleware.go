package middleware

import (
	"net/http"
	"strings"
	"webhost-go/webhost-go/internal/services/user_service/authn/token"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	tokenManager token.TokenManager
}

func NewAuthMiddleware(tm token.TokenManager) *AuthMiddleware {
	return &AuthMiddleware{tokenManager: tm}
}

func (a *AuthMiddleware) RequireReader() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "토큰이 없습니다."})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// 토큰 검증
		result := a.tokenManager.Validate(tokenStr)
		if !result.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "유효하지 않은 토큰"})
			return
		}

		// Role 검사
		if token.Role(result.Claims.Role) != token.RoleReader {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Reader 권한이 필요합니다"})
			return
		}

		// 인증 정보 context에 저장
		c.Set("auth", result)
		c.Next()
	}
}

func (a *AuthMiddleware) RequireSelfOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		authVal, exists := c.Get("auth")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "인증 정보가 없습니다"})
			return
		}

		result, ok := authVal.(*token.TokenValidationResult)
		if !ok || !result.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "유효하지 않은 인증 정보"})
			return
		}

		requested := c.Param("username")
		authEmail := result.Claims.Email
		authRole := result.Claims.Role

		if authEmail != requested && authRole != token.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "접근 권한이 없습니다"})
			return
		}

		c.Next()
	}
}

func (a *AuthMiddleware) RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorization 헤더 추출
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "토큰이 없습니다"})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// 토큰 검증
		result := a.tokenManager.Validate(tokenStr)
		if !result.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "유효하지 않은 토큰"})
			return
		}

		// Role 검사
		if token.Role(result.Claims.Role) != token.RoleUser {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user 권한이 필요합니다"})
			return
		}

		// 인증 정보 context에 저장
		c.Set("auth", result)
		c.Next()
	}
}

func (a *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorization 헤더 추출
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "토큰이 없습니다"})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// 토큰 검증
		result := a.tokenManager.Validate(tokenStr)
		if !result.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "유효하지 않은 토큰"})
			return
		}

		// Role 검사
		if token.Role(result.Claims.Role) != token.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user 권한이 필요합니다"})
			return
		}

		// 인증 정보 context에 저장
		c.Set("auth", result)
		c.Next()
	}
}
