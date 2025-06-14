package agent

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const templateDir = "/var/lib/libvirt/images/templates"

type TemplateRequest struct {
	URL string `json:"url" binding:"required"`
}

func downloadTemplateHandler(c *gin.Context) {
	var req TemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "유효하지 않은 요청 형식"})
		return
	}

	filename := filepath.Base(req.URL)
	dstPath := filepath.Join(templateDir, filename)

	// 다운로드
	out, err := os.Create(dstPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "파일 생성 실패", "detail": err.Error()})
		return
	}
	defer out.Close()

	resp, err := http.Get(req.URL)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": "템플릿 다운로드 실패", "detail": err.Error()})
		return
	}
	defer resp.Body.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "파일 저장 실패", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "템플릿 다운로드 및 저장 성공",
		"filename": filename,
	})
}
