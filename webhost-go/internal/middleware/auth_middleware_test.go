package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"webhost-go/webhost-go/internal/middleware"
	"webhost-go/webhost-go/internal/services/user_service/authn/token"
)

func TestRequireUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tm := token.NewJWTManager("test-secret", time.Minute)
	mw := middleware.NewAuthMiddleware(tm)

	// 올바른 user 토큰 생성
	userToken, _ := tm.Generate("user@example.com", token.RoleUser)
	// 올바른 admin 토큰 생성
	adminToken, _ := tm.Generate("admin@example.com", token.RoleAdmin)
	// 유효하지 않은 토큰
	invalidToken := "this.is.not.valid.jwt"

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{"No token", "", http.StatusUnauthorized},
		{"Invalid token", invalidToken, http.StatusUnauthorized},
		{"Wrong role (admin)", adminToken, http.StatusForbidden},
		{"Correct role (user)", userToken, http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 요청 및 라우터 설정
			router := gin.New()
			router.GET("/test", mw.RequireUser(), func(c *gin.Context) {
				c.String(http.StatusOK, "success")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}
