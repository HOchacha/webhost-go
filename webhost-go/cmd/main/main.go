package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
	"webhost-go/webhost-go/internal/dependency_injector"
)

func main() {
	// 1. Gin 라우터 초기화
	r := gin.Default()

	// "testuser:testpass@tcp(127.0.0.1:3306)/testdb?parseTime=true"
	ai := &dependency_injector.AppInitializer{
		DB: dependency_injector.DBConfig{
			User:     "testuser",
			Password: "testpass",
			Host:     "127.0.0.1",
			Port:     3306,
			Name:     "testdb",
		},
		JWTSecret: "outcider112@dankook.ac.kr",
		TokenTTL:  30 * time.Minute,
	}

	// 2. DI 컨테이너 생성
	registry, err := ai.InitApp()
	if err != nil {
		panic(err)
	}

	// 3. 라우팅 등록
	dependency_injector.RegisterRoutes(r, registry)

	// 4. 서버 실행
	if err := r.Run(":5050"); err != nil {
		log.Fatalf("Fail to run a server: %v", err)
	}
}
