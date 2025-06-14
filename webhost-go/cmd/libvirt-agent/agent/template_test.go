package agent_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"webhost-go/webhost-go/cmd/libvirt-agent/agent" // <- 실제 모듈 경로로 바꿔주세요
)

func TestDownloadTemplateHandler(t *testing.T) {
	// 테스트용 디렉터리 생성
	testDir := t.TempDir()
	agent.TemplateDir = testDir // agent 패키지에서 TemplateDir을 변수로 노출해야 합니다.

	// 테스트용 HTTP 서버 생성 (모의 템플릿 URL 응답)
	testFileContent := []byte("test image data")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(testFileContent)
	}))
	defer ts.Close()

	// JSON 요청 생성
	body := map[string]string{"url": ts.URL + "/template.img"}
	jsonBody, _ := json.Marshal(body)

	// Gin 엔진 설정
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/download", agent.DownloadTemplateHandler)

	// 요청 실행
	req := httptest.NewRequest(http.MethodPost, "/download", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 결과 검증
	assert.Equal(t, http.StatusOK, w.Code)

	// 결과 파일 확인
	expectedPath := filepath.Join(testDir, "template.img")
	data, err := os.ReadFile(expectedPath)
	assert.NoError(t, err)
	assert.Equal(t, testFileContent, data)
}
