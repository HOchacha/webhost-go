package main

import (
	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
	"os"
	"webhost-go/webhost-go/cmd/nginx-agent/nginx"
)

var (
	log    = logging.MustGetLogger("api")
	format = logging.MustStringFormatter(
		`%{color}[%{level:.4s}] %{time:2006/01/02 - 15:04:05}%{color:reset} ▶ %{message}`,
	)
)

func initLogger() {
	// For demo purposes, create two backend for os.Stderr.
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)

	// For messages written to backend2 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	backend2Formatter := logging.NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend1
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")

	// Set the backends to be used.
	logging.SetBackend(backend1Leveled, backend2Formatter)
}

func main() {
	// Initialize logger.
	initLogger()
	log.Info("Successfully initialized logger")

	// If non-root user was running the program, exit. We need root permission.
	if os.Geteuid() != 0 {
		log.Error("Please run this program as root")
		return
	}

	// initialize server app
	router := gin.Default()

	manager := nginx.NewNginxManager(
		"/usr/local/nginx/conf/sites-available",
		"/usr/local/nginx/conf/sites-available/locations", // HTTP 프록시 설정 파일 경로
		"/usr/local/nginx/conf/stream.d",                  // stream 설정 디렉토리
	)

	server := &Server{Manager: manager}
	server.RegisterRoutes(router)

	if err := router.Run(":5003"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
